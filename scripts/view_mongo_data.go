package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	uri := os.Getenv("MONGODB_URI")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("github_analytics")

	fmt.Println("📊 Dados no MongoDB Atlas - github_analytics")
	fmt.Println("============================================")
	fmt.Println()

	// Usuários
	fmt.Println("👤 Usuários:")
	cursor, _ := db.Collection("users").Find(ctx, bson.M{})
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user bson.M
		cursor.Decode(&user)
		name := user["name"]
		if name == nil || name == "" {
			name = user["login"]
		}
		fmt.Printf("  - %v (%s)\n", name, user["login"])
		fmt.Printf("    Followers: %v | Following: %v | Repos: %v\n", 
			user["followers"], user["following"], user["public_repos"])
		if user["bio"] != nil && user["bio"] != "" {
			fmt.Printf("    Bio: %s\n", user["bio"])
		}
	}

	// Repositórios
	fmt.Println("\n📚 Repositórios:")
	repoCount, _ := db.Collection("repositories").CountDocuments(ctx, bson.M{})
	fmt.Printf("  Total: %d repositórios\n", repoCount)

	// Top 10 por stars
	fmt.Println("\n  Top 10 mais estrelados:")
	opts := options.Find().SetSort(bson.D{{"stargazers_count", -1}}).SetLimit(10)
	cursor2, _ := db.Collection("repositories").Find(ctx, bson.M{}, opts)
	defer cursor2.Close(ctx)
	i := 1
	for cursor2.Next(ctx) {
		var repo bson.M
		cursor2.Decode(&repo)
		stars := 0
		if repo["stargazers_count"] != nil {
			stars = int(repo["stargazers_count"].(int32))
		}
		lang := "N/A"
		if repo["language"] != nil {
			lang = repo["language"].(string)
		}
		fmt.Printf("    %d. %s (%s): ⭐ %d\n", i, repo["name"], lang, stars)
		i++
	}

	// Linguagens
	fmt.Println("\n💻 Linguagens (contagem):")
	pipeline := []bson.M{
		{"$match": bson.M{"language": bson.M{"$ne": nil}}},
		{"$group": bson.M{
			"_id":   "$language",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	}
	cursor3, _ := db.Collection("repositories").Aggregate(ctx, pipeline)
	defer cursor3.Close(ctx)
	for cursor3.Next(ctx) {
		var result bson.M
		cursor3.Decode(&result)
		fmt.Printf("  - %s: %d repos\n", result["_id"], result["count"])
	}

	fmt.Println("\n============================================")
	userCount, _ := db.Collection("users").CountDocuments(ctx, bson.M{})
	fmt.Printf("✅ Total: %d usuários, %d repositórios\n", userCount, repoCount)
}
