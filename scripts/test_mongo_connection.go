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
	// Carregar connection string do ambiente
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("❌ MONGODB_URI não definida. Configure no .env ou export MONGODB_URI=...")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "dev_metadata"
	}

	fmt.Println("🔌 Testando conexão com MongoDB Atlas...")
	fmt.Printf("📦 Database: %s\n\n", dbName)

	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectar
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ Erro ao conectar: %v", err)
	}
	defer client.Disconnect(ctx)

	// Testar ping
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ Erro no ping: %v", err)
	}

	fmt.Println("✅ Conexão estabelecida com sucesso!\n")

	// Listar databases
	db := client.Database(dbName)
	
	// Testar inserção simples
	collection := db.Collection("_test")
	testDoc := bson.M{
		"test":       true,
		"message":    "Teste de conexão",
		"timestamp":  time.Now(),
	}

	result, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		log.Fatalf("❌ Erro ao inserir documento de teste: %v", err)
	}

	fmt.Printf("✅ Documento de teste inserido: %v\n", result.InsertedID)

	// Limpar documento de teste
	_, err = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
	if err != nil {
		log.Printf("⚠️  Não foi possível deletar documento de teste: %v", err)
	} else {
		fmt.Println("✅ Documento de teste removido\n")
	}

	// Listar collections existentes
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Printf("⚠️  Erro ao listar collections: %v", err)
	} else {
		fmt.Println("📚 Collections existentes:")
		if len(collections) == 0 {
			fmt.Println("   (nenhuma collection ainda)")
		} else {
			for _, coll := range collections {
				// Contar documentos
				count, _ := db.Collection(coll).CountDocuments(ctx, bson.M{})
				fmt.Printf("   - %s (%d documentos)\n", coll, count)
			}
		}
	}

	fmt.Println("\n🎉 Teste concluído com sucesso!")
	fmt.Println("\n💡 Próximos passos:")
	fmt.Println("   1. Execute os collectors para popular o banco:")
	fmt.Println("      go run scripts/collectors/user_collector.go -user=felipemacedo1")
	fmt.Println("      go run scripts/collectors/repos_collector.go -users=felipemacedo1,growthfolio")
	fmt.Println("   2. Verifique os dados no MongoDB Atlas")
}
