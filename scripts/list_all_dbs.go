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

	dbNames, _ := client.ListDatabaseNames(ctx, bson.M{})
	fmt.Println("🗄️  Databases disponíveis:")
	for _, name := range dbNames {
		if name == "admin" || name == "local" {
			continue
		}
		fmt.Printf("\n  📦 %s\n", name)
		db := client.Database(name)
		colls, _ := db.ListCollectionNames(ctx, bson.M{})
		for _, coll := range colls {
			count, _ := db.Collection(coll).CountDocuments(ctx, bson.M{})
			fmt.Printf("     📁 %s: %d docs\n", coll, count)
		}
	}
}
