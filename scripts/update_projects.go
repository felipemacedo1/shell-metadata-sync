package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "time"
)

// Repo contains the fields we expose in the projects.json file.
type Repo struct {
    Name        string    `json:"name"`
    Owner       string    `json:"owner"`
    Description string    `json:"description,omitempty"`
    Language    string    `json:"language,omitempty"`
    URL         string    `json:"url"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// fetchRepos fetches public repositories for a given username using GitHub API v3.
// It accepts an optional token to increase rate limits.
func fetchRepos(ctx context.Context, client *http.Client, user, token string) ([]Repo, error) {
    var repos []Repo
    // GitHub paginates results. We'll loop until we get an empty page.
    page := 1
    for {
        req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&page=%d", user, page), nil)
        if err != nil {
            return nil, err
        }
        if token != "" {
            req.Header.Set("Authorization", "token "+token)
        }
        req.Header.Set("Accept", "application/vnd.github.v3+json")

        resp, err := client.Do(req)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusNotFound {
            // No such user or repos; return empty slice.
            return repos, nil
        }
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            return nil, fmt.Errorf("github API error: status=%d body=%s", resp.StatusCode, string(body))
        }

        var raw []struct {
            Name        string    `json:"name"`
            Owner       struct{ Login string `json:"login"` } `json:"owner"`
            Description *string   `json:"description"`
            Language    *string   `json:"language"`
            HTMLURL     string    `json:"html_url"`
            UpdatedAt   time.Time `json:"updated_at"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
            return nil, err
        }

        if len(raw) == 0 {
            break
        }

        for _, r := range raw {
            repo := Repo{
                Name:      r.Name,
                Owner:     r.Owner.Login,
                URL:       r.HTMLURL,
                UpdatedAt: r.UpdatedAt,
            }
            if r.Description != nil {
                repo.Description = *r.Description
            }
            if r.Language != nil {
                repo.Language = *r.Language
            }
            repos = append(repos, repo)
        }

        page++
        // continue to next page
    }

    return repos, nil
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
    // Command flags
    var outFile string
    var token string
    flag.StringVar(&outFile, "out", "data/projects.json", "output JSON file")
    flag.StringVar(&token, "token", os.Getenv("GH_TOKEN"), "GitHub token (or set GH_TOKEN env)")
    flag.Parse()

    ctx := context.Background()
    client := &http.Client{Timeout: 30 * time.Second}

    users := []string{"felipemacedo1", "growthfolio"}
    var all []Repo
    for _, u := range users {
        fmt.Fprintf(os.Stderr, "Fetching repos for %s\n", u)
        repos, err := fetchRepos(ctx, client, u, token)
        if err != nil {
            fmt.Fprintf(os.Stderr, "error fetching %s: %v\n", u, err)
            os.Exit(2)
        }
        all = append(all, repos...)
    }

    // Sort by owner/name to make deterministic output, then by updated_at desc
    sort.Slice(all, func(i, j int) bool {
        if all[i].Owner == all[j].Owner {
            if all[i].Name == all[j].Name {
                return all[i].UpdatedAt.After(all[j].UpdatedAt)
            }
            return all[i].Name < all[j].Name
        }
        return all[i].Owner < all[j].Owner
    })

    // Ensure output directory exists
    if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
        fmt.Fprintf(os.Stderr, "error creating output dir: %v\n", err)
        os.Exit(2)
    }

    // Save with indentation for readability
    if err := saveJSON(outFile, all); err != nil {
        fmt.Fprintf(os.Stderr, "error saving JSON: %v\n", err)
        os.Exit(2)
    }

    fmt.Fprintf(os.Stderr, "Wrote %d repositories to %s\n", len(all), outFile)
}
