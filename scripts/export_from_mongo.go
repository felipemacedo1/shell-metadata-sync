package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	outDir := flag.String("out", "data", "Output directory for JSON files")
	flag.Parse()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "dev_metadata"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	fmt.Printf("📦 Exporting from MongoDB '%s' to '%s/'\n\n", dbName, *outDir)

	// Criar diretório de saída
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Export users → profile.json
	if err := exportProfile(ctx, db, *outDir); err != nil {
		log.Printf("⚠️  Error exporting profile: %v", err)
	}

	// Export repositories → projects.json
	if err := exportProjects(ctx, db, *outDir); err != nil {
		log.Printf("⚠️  Error exporting projects: %v", err)
	}

	// Export languages → languages.json
	if err := exportLanguages(ctx, db, *outDir); err != nil {
		log.Printf("⚠️  Error exporting languages: %v", err)
	}

	// Export activity → activity-daily.json
	if err := exportActivity(ctx, db, *outDir); err != nil {
		log.Printf("⚠️  Error exporting activity: %v", err)
	}

	// Create metadata.json
	metadata := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"source":       "mongodb",
		"database":     dbName,
	}
	if err := writeJSON(*outDir+"/metadata.json", metadata); err != nil {
		log.Printf("⚠️  Error writing metadata: %v", err)
	}

	fmt.Println("\n✅ Export completed!")
}

func exportProfile(ctx context.Context, db *mongo.Database, outDir string) error {
	fmt.Print("👤 Exporting profile... ")

	var users []map[string]interface{}
	cursor, err := db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	// Pegar primeiro usuário como profile principal
	profile := users[0]
	profile["generated_at"] = time.Now().Format(time.RFC3339)

	if err := writeJSON(outDir+"/profile.json", profile); err != nil {
		return err
	}

	fmt.Println("✅")
	return nil
}

func exportProjects(ctx context.Context, db *mongo.Database, outDir string) error {
	fmt.Print("📚 Exporting projects... ")

	var repos []map[string]interface{}
	cursor, err := db.Collection("repositories").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &repos); err != nil {
		return err
	}

	// Extrair users únicos
	usersMap := make(map[string]bool)
	for _, repo := range repos {
		if owner, ok := repo["owner"].(string); ok {
			usersMap[owner] = true
		}
	}

	users := []string{}
	for user := range usersMap {
		users = append(users, user)
	}

	output := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated_at": time.Now().Format(time.RFC3339),
			"total_repos":  len(repos),
			"users":        users,
		},
		"repositories": repos,
	}

	if err := writeJSON(outDir+"/projects.json", output); err != nil {
		return err
	}

	fmt.Printf("✅ (%d repos)\n", len(repos))
	return nil
}

func exportLanguages(ctx context.Context, db *mongo.Database, outDir string) error {
	fmt.Print("💻 Exporting languages... ")

	var languages []map[string]interface{}
	cursor, err := db.Collection("languages").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &languages); err != nil {
		return err
	}

	if len(languages) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	// Pegar primeiro documento
	output := languages[0]
	output["generated_at"] = time.Now().Format(time.RFC3339)

	if err := writeJSON(outDir+"/languages.json", output); err != nil {
		return err
	}

	fmt.Println("✅")
	return nil
}

func exportActivity(ctx context.Context, db *mongo.Database, outDir string) error {
	fmt.Print("📊 Exporting activity... ")

	var activities []map[string]interface{}
	cursor, err := db.Collection("activity").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &activities); err != nil {
		return err
	}

	if len(activities) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	// Converter para formato daily_metrics
	dailyMetrics := make(map[string]map[string]interface{})
	var user string
	for _, activity := range activities {
		date, _ := activity["date"].(string)
		user, _ = activity["user"].(string)

		dailyMetrics[date] = map[string]interface{}{
			"commits": activity["commits"],
			"prs":     activity["prs"],
			"issues":  activity["issues"],
		}
	}

	output := map[string]interface{}{
		"metadata": map[string]interface{}{
			"user":         user,
			"period":       fmt.Sprintf("%d days", len(activities)),
			"generated_at": time.Now().Format(time.RFC3339),
		},
		"daily_metrics": dailyMetrics,
	}

	if err := writeJSON(outDir+"/activity-daily.json", output); err != nil {
		return err
	}

	fmt.Printf("✅ (%d days)\n", len(activities))
	return nil
}

func writeJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
