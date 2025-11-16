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
	"sort"
	"time"

	"dev-metadata-sync/scripts/storage"
)

type GitHubRepo struct {
	Name        string    `json:"name"`
	Owner       struct{ Login string `json:"login"` } `json:"owner"`
	Description *string   `json:"description"`
	Language    *string   `json:"language"`
	Topics      []string  `json:"topics"`
	HTMLURL     string    `json:"html_url"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
	PushedAt    time.Time `json:"pushed_at"`
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	Watchers    int       `json:"watchers_count"`
	OpenIssues  int       `json:"open_issues_count"`
	Size        int       `json:"size"`
	DefaultBranch string  `json:"default_branch"`
}

type RepoOutput struct {
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Description string    `json:"description,omitempty"`
	Language    string    `json:"language,omitempty"`
	URL         string    `json:"url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MetadataOutput struct {
	LastSync struct {
		Repos string `json:"repos"`
	} `json:"last_sync"`
	DataCoverage struct {
		ReposCount int `json:"repos_count"`
	} `json:"data_coverage"`
	Version string `json:"version"`
}

func fetchRepos(ctx context.Context, client *http.Client, user, token string) ([]GitHubRepo, error) {
	var repos []GitHubRepo
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&page=%d", user, page)
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

		if resp.StatusCode == http.StatusNotFound {
			return repos, nil
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}

		var raw []GitHubRepo
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return nil, err
		}

		if len(raw) == 0 {
			break
		}

		repos = append(repos, raw...)
		page++
	}

	return repos, nil
}

func fetchLanguages(ctx context.Context, client *http.Client, owner, repo, token string) (map[string]int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", owner, repo)
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
		return nil, fmt.Errorf("status %d", resp.StatusCode)
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

func main() {
	var (
		users      string
		token      string
		outFile    string
		metaFile   string
		mongoURI   string
		withLangs  bool
	)

	flag.StringVar(&users, "users", "felipemacedo1,growthfolio", "comma-separated GitHub usernames")
	flag.StringVar(&token, "token", os.Getenv("GH_TOKEN"), "GitHub token (or set GH_TOKEN env)")
	flag.StringVar(&outFile, "out", "data/projects.json", "output JSON file")
	flag.StringVar(&metaFile, "meta", "data/metadata.json", "metadata JSON file")
	flag.StringVar(&mongoURI, "mongo-uri", os.Getenv("MONGO_URI"), "MongoDB URI (or set MONGO_URI env)")
	flag.BoolVar(&withLangs, "with-languages", false, "fetch language breakdown for each repo (slower)")
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}

	userList := []string{}
	for _, u := range []string{"felipemacedo1", "growthfolio"} {
		userList = append(userList, u)
	}

	var allRepos []RepoOutput
	var allMongoRepos []storage.Repository

	for _, user := range userList {
		log.Printf("📡 Fetching repos for: %s", user)

		repos, err := fetchRepos(ctx, client, user, token)
		if err != nil {
			log.Fatalf("❌ Error fetching %s: %v", user, err)
		}

		log.Printf("✓ Found %d repositories for %s", len(repos), user)

		for _, r := range repos {
			// Gerar output para JSON estático
			repoOut := RepoOutput{
				Name:      r.Name,
				Owner:     r.Owner.Login,
				URL:       r.HTMLURL,
				UpdatedAt: r.UpdatedAt,
			}
			if r.Description != nil {
				repoOut.Description = *r.Description
			}
			if r.Language != nil {
				repoOut.Language = *r.Language
			}

			allRepos = append(allRepos, repoOut)

			// Gerar objeto para MongoDB
			mongoRepo := storage.Repository{
				Name:          r.Name,
				Owner:         r.Owner.Login,
				Topics:        r.Topics,
				Stars:         r.Stars,
				Forks:         r.Forks,
				Watchers:      r.Watchers,
				OpenIssues:    r.OpenIssues,
				Size:          r.Size,
				DefaultBranch: r.DefaultBranch,
				CreatedAt:     r.CreatedAt,
				UpdatedAt:     r.UpdatedAt,
				PushedAt:      r.PushedAt,
			}
			if r.Description != nil {
				mongoRepo.Description = *r.Description
			}
			if r.Language != nil {
				mongoRepo.Language = *r.Language
			}

			// Fetch languages breakdown (opcional, lento)
			if withLangs {
				langs, err := fetchLanguages(ctx, client, r.Owner.Login, r.Name, token)
				if err != nil {
					log.Printf("⚠️  Warning fetching languages for %s/%s: %v", r.Owner.Login, r.Name, err)
				} else {
					mongoRepo.Languages = langs
				}
				time.Sleep(100 * time.Millisecond) // Rate limiting cortesia
			}

			allMongoRepos = append(allMongoRepos, mongoRepo)
		}
	}

	// Ordenar por owner/name
	sort.Slice(allRepos, func(i, j int) bool {
		if allRepos[i].Owner == allRepos[j].Owner {
			return allRepos[i].Name < allRepos[j].Name
		}
		return allRepos[i].Owner < allRepos[j].Owner
	})

	// Salvar projects.json
	// Salvar repos JSON em múltiplos locais
	var repoPaths []string
	if outFile == "data/projects.json" {
		repoPaths = storage.GetDefaultPaths("projects.json")
	} else {
		repoPaths = []string{outFile}
	}

	for _, path := range repoPaths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("❌ Error creating dir for %s: %v", path, err)
		}
		if err := saveJSON(path, allRepos); err != nil {
			log.Fatalf("❌ Error saving to %s: %v", path, err)
		}
		log.Printf("✓ Saved %d repositories to: %s", len(allRepos), path)
	}

	// Salvar metadata.json em múltiplos locais
	metadata := MetadataOutput{
		Version: "2.0.0",
	}
	metadata.LastSync.Repos = time.Now().UTC().Format(time.RFC3339)
	metadata.DataCoverage.ReposCount = len(allRepos)

	var metaPaths []string
	if metaFile == "data/metadata.json" {
		metaPaths = storage.GetDefaultPaths("metadata.json")
	} else {
		metaPaths = []string{metaFile}
	}

	for _, path := range metaPaths {
		if err := saveJSON(path, metadata); err != nil {
			log.Printf("⚠️  Warning saving metadata to %s: %v", path, err)
		} else {
			log.Printf("✓ Saved metadata to: %s", path)
		}
	}

	// Upsert no MongoDB (se configurado)
	if mongoURI != "" {
		mongoClient, err := storage.NewMongoClient(ctx, mongoURI)
		if err != nil {
			log.Printf("⚠️  MongoDB não disponível: %v", err)
		} else {
			defer mongoClient.Close()

			if err := mongoClient.UpsertRepositoriesBatch(allMongoRepos); err != nil {
				log.Printf("⚠️  Error upserting repos to MongoDB: %v", err)
			} else {
				log.Printf("✓ Repositories batch upserted to MongoDB")
			}
		}
	} else {
		log.Printf("ℹ️  MongoDB URI not configured, skipping database upsert")
	}

	log.Printf("🎉 Repository collection completed!")
}
