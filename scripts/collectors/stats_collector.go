package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dev-metadata-sync/scripts/storage"
)

type LanguagesOutput struct {
	Metadata struct {
		User        string `json:"user"`
		GeneratedAt string `json:"generated_at"`
	} `json:"metadata"`
	Languages map[string]struct {
		Bytes      int     `json:"bytes"`
		Repos      int     `json:"repos"`
		Percentage float64 `json:"percentage"`
	} `json:"languages"`
	TopLanguages []string `json:"top_languages"`
}

func fetchRepos(ctx context.Context, client *http.Client, user, token string) ([]string, error) {
	return fetchReposWithOrg(ctx, client, user, "", token)
}

func fetchReposWithOrg(ctx context.Context, client *http.Client, user, org, token string) ([]string, error) {
	var repos []string
	page := 1

	var baseURL string
	if org != "" {
		baseURL = fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	} else {
		baseURL = fmt.Sprintf("https://api.github.com/users/%s/repos", user)
	}

	for {
		url := fmt.Sprintf("%s?per_page=100&page=%d", baseURL, page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
			return repos, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}

		var raw []struct {
			FullName string `json:"full_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return nil, err
		}

		if len(raw) == 0 {
			break
		}

		for _, r := range raw {
			repos = append(repos, r.FullName)
		}

		page++
	}

	return repos, nil
}

func fetchLanguages(ctx context.Context, client *http.Client, repoFullName, token string) (map[string]int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/languages", repoFullName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var languages map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&languages); err != nil {
		return nil, err
	}

	return languages, nil
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
	)

	flag.StringVar(&username, "user", "felipemacedo1", "GitHub username")
	flag.StringVar(&org, "org", "", "GitHub organization (optional)")
	flag.StringVar(&token, "token", getToken(), "GitHub token (or set GH_TOKEN/GITHUB_TOKEN env)")
	flag.StringVar(&outFile, "out", "data/languages.json", "output JSON file")
	flag.StringVar(&mongoURI, "mongo-uri", os.Getenv("MONGO_URI"), "MongoDB URI (or set MONGO_URI env)")
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}

	if org != "" {
		log.Printf("📡 Collecting language stats for organization: %s", org)
	} else {
		log.Printf("📡 Collecting language stats for: %s", username)
	}

	// Buscar lista de repos
	repos, err := fetchReposWithOrg(ctx, client, username, org, token)
	if err != nil {
		log.Fatalf("❌ Error fetching repos: %v", err)
	}

	log.Printf("✓ Found %d repositories", len(repos))

	// Agregar linguagens de todos os repos
	languageTotals := make(map[string]int)
	languageRepoCount := make(map[string]int)
	totalBytes := 0

	for i, repoFullName := range repos {
		log.Printf("  [%d/%d] Fetching languages for %s", i+1, len(repos), repoFullName)

		langs, err := fetchLanguages(ctx, client, repoFullName, token)
		if err != nil {
			log.Printf("    ⚠️  Warning: %v", err)
			continue
		}

		for lang, bytes := range langs {
			languageTotals[lang] += bytes
			languageRepoCount[lang]++
			totalBytes += bytes
		}

		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	log.Printf("✓ Aggregated %d languages across %d repos", len(languageTotals), len(repos))

	// Calcular percentages
	type langStat struct {
		Bytes      int
		Repos      int
		Percentage float64
	}

	languagesOutput := make(map[string]struct {
		Bytes      int     `json:"bytes"`
		Repos      int     `json:"repos"`
		Percentage float64 `json:"percentage"`
	})

	var topLanguages []struct {
		name  string
		bytes int
	}

	for lang, bytes := range languageTotals {
		percentage := 0.0
		if totalBytes > 0 {
			percentage = (float64(bytes) / float64(totalBytes)) * 100
		}

		languagesOutput[lang] = struct {
			Bytes      int     `json:"bytes"`
			Repos      int     `json:"repos"`
			Percentage float64 `json:"percentage"`
		}{
			Bytes:      bytes,
			Repos:      languageRepoCount[lang],
			Percentage: percentage,
		}

		topLanguages = append(topLanguages, struct {
			name  string
			bytes int
		}{lang, bytes})
	}

	// Ordenar top 5
	if len(topLanguages) > 1 {
		for i := 0; i < len(topLanguages)-1; i++ {
			for j := i + 1; j < len(topLanguages); j++ {
				if topLanguages[j].bytes > topLanguages[i].bytes {
					topLanguages[i], topLanguages[j] = topLanguages[j], topLanguages[i]
				}
			}
		}
	}

	topLangNames := []string{}
	maxTop := 5
	if len(topLanguages) < maxTop {
		maxTop = len(topLanguages)
	}
	for i := 0; i < maxTop; i++ {
		topLangNames = append(topLangNames, topLanguages[i].name)
	}

	// Gerar output
	output := LanguagesOutput{}
	output.Metadata.User = username
	output.Metadata.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	output.Languages = languagesOutput
	output.TopLanguages = topLangNames

	// Salvar JSON em múltiplos locais
	var paths []string
	if outFile == "data/languages.json" || outFile == "data/languages-secondary.json" {
		filename := filepath.Base(outFile)
		paths = storage.GetDefaultPaths(filename)
	} else {
		paths = []string{outFile}
	}

	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("❌ Error creating dir for %s: %v", path, err)
		}
		if err := saveJSON(path, output); err != nil {
			log.Fatalf("❌ Error saving to %s: %v", path, err)
		}
		log.Printf("✓ Saved languages to: %s", path)
	}

	// MongoDB update (enriquecer repositories collection)
	if mongoURI != "" {
		mongoClient, err := storage.NewMongoClient(ctx, mongoURI)
		if err != nil {
			log.Printf("⚠️  MongoDB não disponível: %v", err)
		} else {
			defer mongoClient.Close()

			// Aqui poderíamos enriquecer cada repo com suas linguagens
			// Mas isso já foi feito no repos_collector com flag --with-languages
			// Então só logamos
			log.Printf("✓ MongoDB available (language enrichment done via repos_collector --with-languages)")
		}
	} else {
		log.Printf("ℹ️  MongoDB URI not configured, skipping database operations")
	}

	log.Printf("📊 Summary:")
	log.Printf("   Total languages: %d", len(languageTotals))
	log.Printf("   Top 5: %v", topLangNames)
	log.Printf("   Total code bytes: %d", totalBytes)
	log.Printf("🎉 Stats collection completed!")
}
