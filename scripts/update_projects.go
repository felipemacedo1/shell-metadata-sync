package main

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log/slog"
    "math"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
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

// Output wraps the repositories with metadata
type Output struct {
    Metadata     Metadata `json:"metadata"`
    Repositories []Repo   `json:"repositories"`
}

// Metadata contains information about the generated data
type Metadata struct {
    GeneratedAt    time.Time `json:"generated_at"`
    TotalRepos     int       `json:"total_repos"`
    Users          []string  `json:"users"`
    RateLimitUsed  int       `json:"rate_limit_used,omitempty"`
    RateLimitLimit int       `json:"rate_limit_limit,omitempty"`
    RateLimitReset time.Time `json:"rate_limit_reset,omitempty"`
}

// cacheEntry stores ETag and data for a user
type cacheEntry struct {
    ETag  string    `json:"etag"`
    Data  []Repo    `json:"data"`
    Time  time.Time `json:"cached_at"`
}

// fetchRepos fetches public repositories for a given username using GitHub API v3.
// It accepts an optional token to increase rate limits and supports caching with ETag.
func fetchRepos(ctx context.Context, client *http.Client, user, token string, cache *cacheEntry, logger *slog.Logger) ([]Repo, *cacheEntry, error) {
    var repos []Repo
    newCache := &cacheEntry{Time: time.Now()}
    
    // GitHub paginates results. We'll loop until we get an empty page.
    page := 1
    maxRetries := 3
    
    for {
        var resp *http.Response
        var err error
        
        // Retry with exponential backoff
        for attempt := 0; attempt < maxRetries; attempt++ {
            req, reqErr := http.NewRequestWithContext(ctx, "GET", 
                fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&page=%d", user, page), nil)
            if reqErr != nil {
                return nil, nil, reqErr
            }
            
            if token != "" {
                req.Header.Set("Authorization", "token "+token)
            }
            req.Header.Set("Accept", "application/vnd.github.v3+json")
            
            // Add ETag for cache validation (only on first page)
            if page == 1 && cache != nil && cache.ETag != "" {
                req.Header.Set("If-None-Match", cache.ETag)
                logger.Debug("using cached etag", "user", user, "etag", cache.ETag)
            }

            resp, err = client.Do(req)
            if err == nil {
                break
            }
            
            if attempt < maxRetries-1 {
                backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
                logger.Warn("request failed, retrying", 
                    "user", user, 
                    "attempt", attempt+1, 
                    "backoff", backoff.String(),
                    "error", err.Error())
                time.Sleep(backoff)
            }
        }
        
        if err != nil {
            return nil, nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
        }
        defer resp.Body.Close()

        // Handle rate limiting
        if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
            resetTime := resp.Header.Get("X-RateLimit-Reset")
            remaining := resp.Header.Get("X-RateLimit-Remaining")
            
            if remaining == "0" {
                reset, _ := strconv.ParseInt(resetTime, 10, 64)
                resetAt := time.Unix(reset, 0)
                return nil, nil, fmt.Errorf("rate limit exceeded, resets at %s", resetAt.Format(time.RFC3339))
            }
        }

        // Handle 304 Not Modified (cache hit)
        if resp.StatusCode == http.StatusNotModified {
            logger.Info("cache hit", "user", user, "cached_repos", len(cache.Data))
            return cache.Data, cache, nil
        }

        if resp.StatusCode == http.StatusNotFound {
            logger.Warn("user not found", "user", user)
            return repos, newCache, nil
        }
        
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            return nil, nil, fmt.Errorf("github API error: status=%d body=%s", resp.StatusCode, string(body))
        }

        // Capture ETag on first page
        if page == 1 {
            newCache.ETag = resp.Header.Get("ETag")
            logger.Debug("captured etag", "user", user, "etag", newCache.ETag)
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
            return nil, nil, fmt.Errorf("failed to decode JSON: %w", err)
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

        logger.Debug("fetched page", "user", user, "page", page, "repos", len(raw))
        page++
    }

    newCache.Data = repos
    logger.Info("fetched repos", "user", user, "total", len(repos))
    return repos, newCache, nil
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

// loadCache loads the cache file for a user
func loadCache(cacheDir, user string) (*cacheEntry, error) {
    cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s.json", user))
    data, err := os.ReadFile(cachePath)
    if err != nil {
        return nil, err
    }
    
    var cache cacheEntry
    if err := json.Unmarshal(data, &cache); err != nil {
        return nil, err
    }
    
    return &cache, nil
}

// saveCache saves the cache file for a user
func saveCache(cacheDir, user string, cache *cacheEntry) error {
    if err := os.MkdirAll(cacheDir, 0o755); err != nil {
        return err
    }
    
    cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s.json", user))
    data, err := json.MarshalIndent(cache, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(cachePath, data, 0o644)
}

// validateOutput validates the generated JSON structure
func validateOutput(output *Output) error {
    if output.Metadata.TotalRepos != len(output.Repositories) {
        return fmt.Errorf("metadata count mismatch: expected %d, got %d", 
            output.Metadata.TotalRepos, len(output.Repositories))
    }
    
    seen := make(map[string]bool)
    for _, repo := range output.Repositories {
        if repo.Name == "" || repo.Owner == "" || repo.URL == "" {
            return fmt.Errorf("invalid repo: missing required fields (name=%s, owner=%s, url=%s)", 
                repo.Name, repo.Owner, repo.URL)
        }
        
        key := repo.Owner + "/" + repo.Name
        if seen[key] {
            return fmt.Errorf("duplicate repository: %s", key)
        }
        seen[key] = true
    }
    
    return nil
}

// generateChangelog compares old and new data and generates a changelog
func generateChangelog(oldFile, newOutput *Output, changelogPath string, logger *slog.Logger) error {
    if oldFile == nil {
        logger.Info("no previous data, skipping changelog")
        return nil
    }
    
    oldMap := make(map[string]Repo)
    for _, r := range oldFile.Repositories {
        oldMap[r.Owner+"/"+r.Name] = r
    }
    
    var added, updated, removed []string
    newMap := make(map[string]bool)
    
    for _, r := range newOutput.Repositories {
        key := r.Owner + "/" + r.Name
        newMap[key] = true
        
        if old, exists := oldMap[key]; !exists {
            added = append(added, key)
        } else if !old.UpdatedAt.Equal(r.UpdatedAt) {
            updated = append(updated, key)
        }
    }
    
    for _, r := range oldFile.Repositories {
        key := r.Owner + "/" + r.Name
        if !newMap[key] {
            removed = append(removed, key)
        }
    }
    
    if len(added) == 0 && len(updated) == 0 && len(removed) == 0 {
        logger.Info("no changes detected")
        return nil
    }
    
    // Generate changelog
    var changelog strings.Builder
    changelog.WriteString(fmt.Sprintf("# Changelog - %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
    
    if len(added) > 0 {
        changelog.WriteString(fmt.Sprintf("## Added (%d)\n", len(added)))
        for _, repo := range added {
            changelog.WriteString(fmt.Sprintf("- %s\n", repo))
        }
        changelog.WriteString("\n")
    }
    
    if len(updated) > 0 {
        changelog.WriteString(fmt.Sprintf("## Updated (%d)\n", len(updated)))
        for _, repo := range updated {
            changelog.WriteString(fmt.Sprintf("- %s\n", repo))
        }
        changelog.WriteString("\n")
    }
    
    if len(removed) > 0 {
        changelog.WriteString(fmt.Sprintf("## Removed (%d)\n", len(removed)))
        for _, repo := range removed {
            changelog.WriteString(fmt.Sprintf("- %s\n", repo))
        }
        changelog.WriteString("\n")
    }
    
    logger.Info("changelog generated", 
        "added", len(added), 
        "updated", len(updated), 
        "removed", len(removed))
    
    return os.WriteFile(changelogPath, []byte(changelog.String()), 0o644)
}

// computeHash computes SHA256 hash of the repositories data
func computeHash(repos []Repo) string {
    data, _ := json.Marshal(repos)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("%x", hash[:8])
}

func main() {
    // Command flags
    var outFile string
    var token string
    var usersFlag string
    var cacheDir string
    var changelogFile string
    var verbose bool
    var jsonLogs bool
    
    flag.StringVar(&outFile, "out", "data/projects.json", "output JSON file")
    flag.StringVar(&token, "token", os.Getenv("GH_TOKEN"), "GitHub token (or set GH_TOKEN env)")
    flag.StringVar(&usersFlag, "users", os.Getenv("GH_USERS"), "comma-separated list of GitHub users (or set GH_USERS env)")
    flag.StringVar(&cacheDir, "cache-dir", ".cache", "cache directory for ETags")
    flag.StringVar(&changelogFile, "changelog", "CHANGELOG.md", "changelog output file")
    flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
    flag.BoolVar(&jsonLogs, "json-logs", false, "output logs as JSON")
    flag.Parse()

    // Setup logger
    logLevel := slog.LevelInfo
    if verbose {
        logLevel = slog.LevelDebug
    }
    
    var handler slog.Handler
    if jsonLogs {
        handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
    } else {
        handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
    }
    logger := slog.New(handler)
    slog.SetDefault(logger)

    ctx := context.Background()
    client := &http.Client{Timeout: 30 * time.Second}

    // Parse users
    users := strings.Split(usersFlag, ",")
    for i := range users {
        users[i] = strings.TrimSpace(users[i])
    }
    
    logger.Info("starting collection", 
        "users", users, 
        "output", outFile,
        "cache_enabled", cacheDir != "")

    // Load previous output for changelog
    var oldOutput *Output
    if data, err := os.ReadFile(outFile); err == nil {
        if err := json.Unmarshal(data, &oldOutput); err != nil {
            logger.Warn("failed to load previous output for changelog", "error", err)
        }
    }

    var all []Repo
    var totalRateLimitUsed, totalRateLimitLimit int
    var rateLimitReset time.Time
    
    for _, u := range users {
        logger.Info("fetching repos", "user", u)
        
        // Load cache
        var cache *cacheEntry
        if cacheDir != "" {
            if c, err := loadCache(cacheDir, u); err == nil {
                cache = c
                logger.Debug("loaded cache", "user", u, "age", time.Since(c.Time).String())
            }
        }
        
        repos, newCache, err := fetchRepos(ctx, client, u, token, cache, logger)
        if err != nil {
            logger.Error("failed to fetch repos", "user", u, "error", err)
            os.Exit(2)
        }
        
        // Save cache
        if cacheDir != "" && newCache != nil {
            if err := saveCache(cacheDir, u, newCache); err != nil {
                logger.Warn("failed to save cache", "user", u, "error", err)
            } else {
                logger.Debug("saved cache", "user", u)
            }
        }
        
        all = append(all, repos...)
    }

    // Sort by owner/name to make deterministic output
    sort.Slice(all, func(i, j int) bool {
        if all[i].Owner == all[j].Owner {
            if all[i].Name == all[j].Name {
                return all[i].UpdatedAt.After(all[j].UpdatedAt)
            }
            return all[i].Name < all[j].Name
        }
        return all[i].Owner < all[j].Owner
    })

    // Create output with metadata
    output := Output{
        Metadata: Metadata{
            GeneratedAt:    time.Now().UTC(),
            TotalRepos:     len(all),
            Users:          users,
            RateLimitUsed:  totalRateLimitUsed,
            RateLimitLimit: totalRateLimitLimit,
            RateLimitReset: rateLimitReset,
        },
        Repositories: all,
    }

    // Validate output
    if err := validateOutput(&output); err != nil {
        logger.Error("validation failed", "error", err)
        os.Exit(2)
    }
    logger.Info("validation passed")

    // Ensure output directory exists
    if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
        logger.Error("failed to create output dir", "error", err)
        os.Exit(2)
    }

    // Save with indentation for readability
    if err := saveJSON(outFile, output); err != nil {
        logger.Error("failed to save JSON", "error", err)
        os.Exit(2)
    }

    logger.Info("saved output", 
        "file", outFile, 
        "repos", len(all),
        "hash", computeHash(all))

    // Generate changelog
    if changelogFile != "" {
        if err := generateChangelog(oldOutput, &output, changelogFile, logger); err != nil {
            logger.Warn("failed to generate changelog", "error", err)
        }
    }

    // Final summary
    logger.Info("collection complete", 
        "total_repos", len(all),
        "users", len(users),
        "duration", time.Since(output.Metadata.GeneratedAt).String())
}
