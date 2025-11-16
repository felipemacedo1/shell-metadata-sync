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

type GitHubUser struct {
	Login             string    `json:"login"`
	Name              string    `json:"name"`
	Bio               string    `json:"bio"`
	AvatarURL         string    `json:"avatar_url"`
	Followers         int       `json:"followers"`
	Following         int       `json:"following"`
	PublicRepos       int       `json:"public_repos"`
	PublicGists       int       `json:"public_gists"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	TotalPrivateRepos int       `json:"total_private_repos"`
	OwnedPrivateRepos int       `json:"owned_private_repos"`
}

type GitHubOrg struct {
	Login       string `json:"login"`
	Description string `json:"description"`
}

type ProfileData struct {
	Login               string   `json:"login"`
	Name                string   `json:"name"`
	Bio                 string   `json:"bio"`
	AvatarURL           string   `json:"avatar_url"`
	Followers           int      `json:"followers"`
	Following           int      `json:"following"`
	PublicRepos         int      `json:"public_repos"`
	TotalStarsReceived  int      `json:"total_stars_received"`
	TotalForksReceived  int      `json:"total_forks_received"`
	Organizations       []string `json:"organizations"`
	GeneratedAt         string   `json:"generated_at"`
}

func fetchUser(ctx context.Context, client *http.Client, username, token string) (*GitHubUser, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
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
		return nil, fmt.Errorf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func fetchOrganizations(ctx context.Context, client *http.Client, username, token string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/orgs", username)
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
		return []string{}, nil // Retornar vazio se não conseguir buscar orgs
	}

	var orgs []GitHubOrg
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, err
	}

	var orgNames []string
	for _, org := range orgs {
		orgNames = append(orgNames, org.Login)
	}

	return orgNames, nil
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
		token    string
		outFile  string
		mongoURI string
	)

	flag.StringVar(&username, "user", "felipemacedo1", "GitHub username")
	flag.StringVar(&token, "token", os.Getenv("GH_TOKEN"), "GitHub token (or set GH_TOKEN env)")
	flag.StringVar(&outFile, "out", "data/profile.json", "output JSON file")
	flag.StringVar(&mongoURI, "mongo-uri", os.Getenv("MONGO_URI"), "MongoDB URI (or set MONGO_URI env)")
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}

	log.Printf("📡 Fetching user data for: %s", username)

	// Buscar dados do usuário
	user, err := fetchUser(ctx, client, username, token)
	if err != nil {
		log.Fatalf("❌ Error fetching user: %v", err)
	}

	log.Printf("✓ User: %s (%s)", user.Name, user.Login)
	log.Printf("  Followers: %d | Following: %d | Repos: %d", 
		user.Followers, user.Following, user.PublicRepos)

	// Buscar organizações
	orgs, err := fetchOrganizations(ctx, client, username, token)
	if err != nil {
		log.Printf("⚠️  Warning fetching organizations: %v", err)
		orgs = []string{}
	}

	if len(orgs) > 0 {
		log.Printf("✓ Organizations: %v", orgs)
	}

	// Gerar ProfileData para JSON estático
	profileData := ProfileData{
		Login:               user.Login,
		Name:                user.Name,
		Bio:                 user.Bio,
		AvatarURL:           user.AvatarURL,
		Followers:           user.Followers,
		Following:           user.Following,
		PublicRepos:         user.PublicRepos,
		TotalStarsReceived:  0, // Será calculado por stats_collector
		TotalForksReceived:  0, // Será calculado por stats_collector
		Organizations:       orgs,
		GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
	}

	// Salvar JSON em múltiplos locais
	var paths []string
	if outFile == "data/profile.json" || outFile == "data/profile-secondary.json" {
		filename := filepath.Base(outFile)
		paths = storage.GetDefaultPaths(filename)
	} else {
		paths = []string{outFile}
	}

	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("❌ Error creating dir for %s: %v", path, err)
		}
		if err := saveJSON(path, profileData); err != nil {
			log.Fatalf("❌ Error saving to %s: %v", path, err)
		}
		log.Printf("✓ Saved profile to: %s", path)
	}

	// Upsert no MongoDB (se configurado)
	if mongoURI != "" {
		mongoClient, err := storage.NewMongoClient(ctx, mongoURI)
		if err != nil {
			log.Printf("⚠️  MongoDB não disponível: %v", err)
		} else {
			defer mongoClient.Close()

			mongoUser := storage.User{
				Login:       user.Login,
				Name:        user.Name,
				Bio:         user.Bio,
				AvatarURL:   user.AvatarURL,
				Followers:   user.Followers,
				Following:   user.Following,
				PublicRepos: user.PublicRepos,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			}

			if err := mongoClient.UpsertUser(mongoUser); err != nil {
				log.Printf("⚠️  Error upserting user to MongoDB: %v", err)
			} else {
				log.Printf("✓ User upserted to MongoDB")
			}
		}
	} else {
		log.Printf("ℹ️  MongoDB URI not configured, skipping database upsert")
	}

	log.Printf("🎉 User collection completed!")
}
