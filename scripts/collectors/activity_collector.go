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
	commits := make(map[string]int)
	page := 1
	perPage := 100

	var query string
	if org != "" {
		query = fmt.Sprintf("author:%s org:%s committer-date:%s..%s", username, org, startDate, endDate)
	} else {
		query = fmt.Sprintf("author:%s committer-date:%s..%s", username, startDate, endDate)
	}
	
	for {
		urlStr := fmt.Sprintf("https://api.github.com/search/commits?q=%s&per_page=%d&page=%d",
			url.QueryEscape(query), perPage, page)

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", "application/vnd.github.cloak-preview+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}

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

func main() {
	var (
		username string
		org      string
		token    string
		outFile  string
		mongoURI string
		days     int
	)

	flag.StringVar(&username, "user", "felipemacedo1", "GitHub username")
	flag.StringVar(&org, "org", "", "GitHub organization (optional - if set, searches commits by user in org repos)")
	flag.StringVar(&token, "token", os.Getenv("GH_TOKEN"), "GitHub token (or set GH_TOKEN env)")
	flag.StringVar(&outFile, "out", "data/activity-daily.json", "output JSON file")
	flag.StringVar(&mongoURI, "mongo-uri", os.Getenv("MONGO_URI"), "MongoDB URI (or set MONGO_URI env)")
	flag.IntVar(&days, "days", 365, "number of days to fetch (default 365)")
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: 60 * time.Second}

	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	if org != "" {
		log.Printf("📡 Collecting activity for: %s in organization %s", username, org)
	} else {
		log.Printf("📡 Collecting activity for: %s", username)
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

	// Salvar JSON
	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		log.Fatalf("❌ Error creating output dir: %v", err)
	}

	if err := saveJSON(outFile, output); err != nil {
		log.Fatalf("❌ Error saving JSON: %v", err)
	}

	log.Printf("✓ Saved activity to: %s", outFile)

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

	log.Printf("📊 Summary:")
	log.Printf("   Total commits: %d", totalCommits)
	log.Printf("   Total PRs: %d", totalPRs)
	log.Printf("   Total issues: %d", totalIssues)
	log.Printf("🎉 Activity collection completed!")
}
