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
	fmt.Println("🧪 MongoDB Connection Test")
	fmt.Println("==========================")
	fmt.Println()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("❌ MONGODB_URI not set")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "dev_metadata"
	}

	fmt.Println("📋 Configuration:")
	fmt.Printf("   Database: %s\n", dbName)
	fmt.Println()

	// Test connection
	fmt.Println("🔌 Testing connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ Ping failed: %v", err)
	}
	fmt.Println("✅ Connected successfully!")
	fmt.Println()

	// Test operations
	db := client.Database(dbName)
	testColl := db.Collection("_test")

	// Insert test document
	testDoc := bson.M{"test": "connection", "timestamp": time.Now()}
	result, err := testColl.InsertOne(ctx, testDoc)
	if err != nil {
		log.Printf("⚠️  Insert failed: %v", err)
	} else {
		fmt.Printf("✅ Test document inserted: %v\n", result.InsertedID)
	}

	// Delete test document
	if result != nil {
		if _, err := testColl.DeleteOne(ctx, bson.M{"_id": result.InsertedID}); err != nil {
			log.Printf("⚠️  Delete failed: %v", err)
		} else {
			fmt.Println("✅ Test document deleted")
		}
	}
	fmt.Println()

	// List collections
	fmt.Println("📚 Collections:")
	colls, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Printf("⚠️  List collections failed: %v", err)
	} else {
		if len(colls) == 0 {
			fmt.Println("   (none)")
		} else {
			for _, coll := range colls {
				if coll == "_test" {
					continue
				}
				count, _ := db.Collection(coll).CountDocuments(ctx, bson.M{})
				fmt.Printf("   - %s: %d docs\n", coll, count)
			}
		}
	}

	fmt.Println()
	fmt.Println("🎉 Test completed!")
}
