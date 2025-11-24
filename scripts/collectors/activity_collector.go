package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"dev-metadata-sync/scripts/storage"
	"dev-metadata-sync/scripts/utils"
)

type SearchCommitsResponse struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Author struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Message string `json:"message"`
		} `json:"commit"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	} `json:"items"`
}

type SearchIssuesResponse struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		State     string    `json:"state"`
		CreatedAt time.Time `json:"created_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		PullRequest *struct{} `json:"pull_request,omitempty"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	} `json:"items"`
}

type ActivityDaily struct {
	Date    string `json:"date"`
	Commits int    `json:"commits"`
	PRs     int    `json:"prs"`
	Issues  int    `json:"issues"`
}

type ActivityOutput struct {
	Metadata struct {
		User        string `json:"user"`
		Period      string `json:"period"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		GeneratedAt string `json:"generated_at"`
	} `json:"metadata"`
	DailyMetrics map[string]struct {
		Commits int `json:"commits"`
		PRs     int `json:"prs"`
		Issues  int `json:"issues"`
	} `json:"daily_metrics"`
}

func searchCommits(ctx context.Context, client *http.Client, username, token, startDate, endDate string) (map[string]int, error) {
	return searchCommitsWithOrg(ctx, client, username, "", token, startDate, endDate)
}

func searchCommitsWithOrg(ctx context.Context, client *http.Client, username, org, token, startDate, endDate string) (map[string]int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	// Se o período for maior que 90 dias, quebrar em chunks mensais
	daysDiff := int(end.Sub(start).Hours() / 24)
	if daysDiff > 90 {
		log.Printf("   📅 Period > 90 days, splitting into monthly chunks...")
		return searchCommitsInChunks(ctx, client, username, org, token, start, end)
	}

	return searchCommitsSinglePeriod(ctx, client, username, org, token, startDate, endDate)
}

func searchCommitsSinglePeriod(ctx context.Context, client *http.Client, username, org, token, startDate, endDate string) (map[string]int, error) {
	commits := make(map[string]int)
	page := 1
	perPage := 100
	
	rateLimiter := utils.NewRateLimitHandler(3, 1*time.Second)

	var query string
	if org != "" {
		query = fmt.Sprintf("author:%s org:%s committer-date:%s..%s", username, org, startDate, endDate)
	} else {
		query = fmt.Sprintf("author:%s committer-date:%s..%s", username, startDate, endDate)
	}
	
	for {
		urlStr := fmt.Sprintf("https://api.github.com/search/commits?q=%s&per_page=%d&page=%d",
			url.QueryEscape(query), perPage, page)

		resp, err := rateLimiter.RetryWithBackoff(ctx, func() (*http.Response, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
			if err != nil {
				return nil, err
			}

			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			req.Header.Set("Accept", "application/vnd.github.cloak-preview+json")

			return client.Do(req)
		})
		
		if err != nil {
			return nil, fmt.Errorf("failed after retries: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}
		
		utils.LogRateLimitInfo(resp)

		var result SearchCommitsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			date := item.Commit.Author.Date.Format("2006-01-02")
			commits[date]++
		}

		log.Printf("  Page %d: +%d commits (total so far: %d)", page, len(result.Items), len(commits))

		if len(result.Items) < perPage {
			break
		}

		page++
		time.Sleep(200 * time.Millisecond) // Rate limiting
	}

	return commits, nil
}

func searchCommitsInChunks(ctx context.Context, client *http.Client, username, org, token string, start, end time.Time) (map[string]int, error) {
	allCommits := make(map[string]int)
    
	// Processar mês por mês
	current := start
	monthCount := 0
	type chunk struct {
		start time.Time
		end   time.Time
		idx   int
	}
	var chunks []chunk

	for current.Before(end) {
		// Fim do chunk: 30 dias ou fim do período
		chunkEnd := current.AddDate(0, 0, 30)
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		monthCount++
		chunks = append(chunks, chunk{start: current, end: chunkEnd, idx: monthCount})
		current = chunkEnd.AddDate(0, 0, 1)
	}

	if monthCount == 0 {
		return allCommits, nil
	}

	// Paralelizar execução de chunks (limitado por MaxConcurrency)
	cfg := utils.LoadConfig()
	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 2
	}
	// Cap razoável para chunks paralelos para evitar bursts nos limites de busca
	if concurrency > 4 {
		concurrency = 4
	}

	log.Printf("   📦 Executing %d chunks with concurrency %d", monthCount, concurrency)

	// Channel de resultados por chunk
	type chunkResult struct {
		idx    int
		start  time.Time
		end    time.Time
		commits map[string]int
		err    error
	}

	results := make(chan chunkResult, len(chunks))

	// Função para processar um chunk
	processChunk := func(c chunk) error {
		startStr := c.start.Format("2006-01-02")
		endStr := c.end.Format("2006-01-02")
		log.Printf("   → Chunk %d: %s to %s (started)", c.idx, startStr, endStr)
		commits, err := searchCommitsSinglePeriod(ctx, client, username, org, token, startStr, endStr)
		results <- chunkResult{idx: c.idx, start: c.start, end: c.end, commits: commits, err: err}
		return nil
	}

	// Executar em paralelo com worker pool
	errs := utils.ProcessBatch(ctx, chunks, concurrency, func(c chunk) error {
		return processChunk(c)
	})

	// Fechar a channel de resultados depois do processamento
	// (a utils.ProcessBatch aguarda os workers e fecha automaticamente)

	close(results)

	// Agregar resultados
	totalDays := 0
	for res := range results {
		if res.err != nil {
			// Se houve erro em algum chunk, reporte e falhe
			return nil, fmt.Errorf("error in chunk %d: %w", res.idx, res.err)
		}
		// Merge commits
		for date, count := range res.commits {
			allCommits[date] += count
		}
		totalDays += len(res.commits)
		log.Printf("   ✓ Chunk %d: +%d days with commits (aggregated: %d days)", res.idx, len(res.commits), len(allCommits))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("errors executing chunks: %v", errs)
	}

	log.Printf("   ✅ Aggregated %d chunks, total: %d days with commits", monthCount, len(allCommits))
	return allCommits, nil
}

func searchPRs(ctx context.Context, client *http.Client, username, token, startDate, endDate string) (map[string]int, error) {
	return searchPRsWithOrg(ctx, client, username, "", token, startDate, endDate)
}

func searchPRsWithOrg(ctx context.Context, client *http.Client, username, org, token, startDate, endDate string) (map[string]int, error) {
	prs := make(map[string]int)
	page := 1
	perPage := 100

	var query string
	if org != "" {
		query = fmt.Sprintf("author:%s org:%s type:pr created:%s..%s", username, org, startDate, endDate)
	} else {
		query = fmt.Sprintf("author:%s type:pr created:%s..%s", username, startDate, endDate)
	}
	
	for {
		urlStr := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=%d&page=%d",
			url.QueryEscape(query), perPage, page)

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}

		var result SearchIssuesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			date := item.CreatedAt.Format("2006-01-02")
			prs[date]++
		}

		log.Printf("  Page %d: +%d PRs (total so far: %d)", page, len(result.Items), len(prs))

		if len(result.Items) < perPage {
			break
		}

		page++
		time.Sleep(200 * time.Millisecond)
	}

	return prs, nil
}

func searchIssues(ctx context.Context, client *http.Client, username, token, startDate, endDate string) (map[string]int, error) {
	return searchIssuesWithOrg(ctx, client, username, "", token, startDate, endDate)
}

func searchIssuesWithOrg(ctx context.Context, client *http.Client, username, org, token, startDate, endDate string) (map[string]int, error) {
	issues := make(map[string]int)
	page := 1
	perPage := 100

	var query string
	if org != "" {
		query = fmt.Sprintf("author:%s org:%s type:issue created:%s..%s", username, org, startDate, endDate)
	} else {
		query = fmt.Sprintf("author:%s type:issue created:%s..%s", username, startDate, endDate)
	}
	
	for {
		urlStr := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=%d&page=%d",
			url.QueryEscape(query), perPage, page)

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}

		var result SearchIssuesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			date := item.CreatedAt.Format("2006-01-02")
			issues[date]++
		}

		log.Printf("  Page %d: +%d issues (total so far: %d)", page, len(result.Items), len(issues))

		if len(result.Items) < perPage {
			break
		}

		page++
		time.Sleep(200 * time.Millisecond)
	}

	return issues, nil
}

func saveJSON(path string, v interface{}) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func getToken() string {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func main() {
	var (
		username string
		org      string
		token    string
		outFile  string
		mongoURI string
		days     int
	)

	config := utils.LoadConfig()
	logger := utils.NewLogger(config.EnableStructured)
	metrics := utils.NewMetrics()

	flag.StringVar(&username, "user", "felipemacedo1", "GitHub username")
	flag.StringVar(&org, "org", "", "GitHub organization (optional - if set, searches commits by user in org repos)")
	flag.StringVar(&token, "token", config.GitHubToken, "GitHub token (or set GH_TOKEN/GITHUB_TOKEN env)")
	flag.StringVar(&outFile, "out", "data/activity-daily.json", "output JSON file")
	flag.StringVar(&mongoURI, "mongo-uri", config.MongoURI, "MongoDB URI (or set MONGO_URI env)")
	flag.IntVar(&days, "days", 365, "number of days to fetch (default 365)")
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: config.HTTPTimeout}

	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	if org != "" {
		logger.Info("Collecting activity for user in organization", map[string]interface{}{
			"user": username,
			"org":  org,
			"days": days,
		})
	} else {
		logger.Info("Collecting activity for user", map[string]interface{}{
			"user": username,
			"days": days,
		})
	}
	log.Printf("   Period: %s to %s (%d days)", startStr, endStr, days)

	// Buscar commits
	log.Printf("🔍 Searching commits...")
	commits, err := searchCommitsWithOrg(ctx, client, username, org, token, startStr, endStr)
	if err != nil {
		log.Fatalf("❌ Error searching commits: %v", err)
	}
	log.Printf("✓ Found commits on %d days", len(commits))

	// Buscar PRs
	log.Printf("🔍 Searching pull requests...")
	prs, err := searchPRsWithOrg(ctx, client, username, org, token, startStr, endStr)
	if err != nil {
		log.Fatalf("❌ Error searching PRs: %v", err)
	}
	log.Printf("✓ Found PRs on %d days", len(prs))

	// Buscar issues
	log.Printf("🔍 Searching issues...")
	issues, err := searchIssuesWithOrg(ctx, client, username, org, token, startStr, endStr)
	if err != nil {
		log.Fatalf("❌ Error searching issues: %v", err)
	}
	log.Printf("✓ Found issues on %d days", len(issues))

	// Merge data por dia
	dailyMetrics := make(map[string]struct {
		Commits int `json:"commits"`
		PRs     int `json:"prs"`
		Issues  int `json:"issues"`
	})

	// Preencher todos os dias (inclusive zeros)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dailyMetrics[dateStr] = struct {
			Commits int `json:"commits"`
			PRs     int `json:"prs"`
			Issues  int `json:"issues"`
		}{
			Commits: commits[dateStr],
			PRs:     prs[dateStr],
			Issues:  issues[dateStr],
		}
	}

	// Gerar output JSON
	output := ActivityOutput{}
	output.Metadata.User = username
	output.Metadata.Period = "rolling_365_days"
	output.Metadata.StartDate = startStr
	output.Metadata.EndDate = endStr
	output.Metadata.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	output.DailyMetrics = dailyMetrics

	// Salvar JSON em múltiplos locais com validação
	// Determinar paths baseado no output file
	var paths []string
	if outFile == "data/activity-daily.json" || outFile == "data/activity-daily-secondary.json" {
		// Usar paths padrão (data/ e dashboard/public/data/)
		filename := filepath.Base(outFile)
		paths = storage.GetDefaultPaths(filename)
	} else {
		// Path customizado - salvar apenas no especificado
		paths = []string{outFile}
		// Criar diretório se necessário
		if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
			log.Fatalf("❌ Error creating output dir: %v", err)
		}
	}

	// Salvar em todos os paths com validação
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("❌ Error creating dir for %s: %v", path, err)
		}
		if err := utils.ValidateAndSaveJSON(path, output); err != nil {
			log.Fatalf("❌ Error saving to %s: %v", path, err)
		}
		log.Printf("✓ Saved and validated activity to: %s", path)
	}

	// Upsert no MongoDB (se configurado)
	if mongoURI != "" {
		mongoClient, err := storage.NewMongoClient(ctx, mongoURI)
		if err != nil {
			log.Printf("⚠️  MongoDB não disponível: %v", err)
		} else {
			defer mongoClient.Close()

			var activities []storage.DailyActivity
			for dateStr, metrics := range dailyMetrics {
				date, _ := time.Parse("2006-01-02", dateStr)
				activities = append(activities, storage.DailyActivity{
					User:         username,
					Date:         date,
					Commits:      metrics.Commits,
					PRsOpened:    metrics.PRs,
					IssuesOpened: metrics.Issues,
				})
			}

			if err := mongoClient.UpsertDailyActivitiesBatch(activities); err != nil {
				log.Printf("⚠️  Error upserting activities to MongoDB: %v", err)
			} else {
				log.Printf("✓ Daily activities batch upserted to MongoDB")
			}
		}
	} else {
		log.Printf("ℹ️  MongoDB URI not configured, skipping database upsert")
	}

	totalCommits := 0
	totalPRs := 0
	totalIssues := 0
	for _, m := range dailyMetrics {
		totalCommits += m.Commits
		totalPRs += m.PRs
		totalIssues += m.Issues
	}
	
	metrics.ItemsTotal = len(dailyMetrics)
	metrics.ItemsSuccess = len(dailyMetrics)
	metrics.Finish()

	logger.Success("Activity collection completed", map[string]interface{}{
		"total_commits": totalCommits,
		"total_prs":     totalPRs,
		"total_issues":  totalIssues,
		"days":          len(dailyMetrics),
		"metrics":       metrics.ToMap(),
	})
	
	log.Printf("📊 Summary:")
	log.Printf("   Total commits: %d", totalCommits)
	log.Printf("   Total PRs: %d", totalPRs)
	log.Printf("   Total issues: %d", totalIssues)
	log.Printf("   %s", metrics.String())
	log.Printf("🎉 Activity collection completed!")
}
