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
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "dev_metadata"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	fmt.Println("📊 Collections no MongoDB Atlas:")
	fmt.Println("================================")
	fmt.Println()

	colls, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	if len(colls) == 0 {
		fmt.Println("⚠️  Nenhuma collection encontrada")
		fmt.Println("💡 Os collectors não salvaram dados no MongoDB ainda")
		return
	}

	totalDocs := int64(0)
	for _, collName := range colls {
		count, _ := db.Collection(collName).CountDocuments(ctx, bson.M{})
		totalDocs += count
		
		if collName == "_test" {
			continue // Pular collection de teste
		}
		
		fmt.Printf("  📁 %s: %d documentos\n", collName, count)

		if count > 0 && count <= 3 {
			cursor, err := db.Collection(collName).Find(ctx, bson.M{}, options.Find().SetLimit(1))
			if err == nil {
				defer cursor.Close(ctx)
				var result bson.M
				if cursor.Next(ctx) {
					cursor.Decode(&result)
					fmt.Printf("     Exemplo (primeiros campos): ")
					i := 0
					for k, v := range result {
						if i >= 3 {
							break
						}
						if k != "_id" {
							fmt.Printf("%s=%v ", k, v)
						}
						i++
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("================================")
	fmt.Printf("Total: %d documentos em %d collections\n", totalDocs, len(colls))
}
