package storage

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

// MongoClient encapsula a conexão MongoDB
type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
}

// NewMongoClient cria uma nova instância do client MongoDB
func NewMongoClient(ctx context.Context, uri string) (*MongoClient, error) {
	if uri == "" {
		return nil, fmt.Errorf("MONGO_URI não configurado")
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	// Ping para verificar conexão
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("erro ao pingar MongoDB: %w", err)
	}

	dbName := "github_analytics"
	if envDB := os.Getenv("MONGO_DB"); envDB != "" {
		dbName = envDB
	}

	log.Printf("✓ Conectado ao MongoDB: %s", dbName)

	return &MongoClient{
		client: client,
		db:     client.Database(dbName),
		ctx:    ctx,
	}, nil
}

// Close fecha a conexão com o MongoDB
func (mc *MongoClient) Close() error {
	if mc.client != nil {
		return mc.client.Disconnect(mc.ctx)
	}
	return nil
}

// HealthCheck verifica se a conexão está ativa
func (mc *MongoClient) HealthCheck() error {
	return mc.client.Ping(mc.ctx, nil)
}

// --- User Collection ---

type User struct {
	ID           string    `bson:"_id"`
	Login        string    `bson:"login"`
	Name         string    `bson:"name,omitempty"`
	Bio          string    `bson:"bio,omitempty"`
	AvatarURL    string    `bson:"avatar_url,omitempty"`
	Followers    int       `bson:"followers"`
	Following    int       `bson:"following"`
	PublicRepos  int       `bson:"public_repos"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
	LastSynced   time.Time `bson:"last_synced"`
}

func (mc *MongoClient) UpsertUser(user User) error {
	collection := mc.db.Collection("users")
	
	user.LastSynced = time.Now()
	user.ID = user.Login

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(mc.ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("erro ao upsert user %s: %w", user.Login, err)
	}

	return nil
}

// --- Repository Collection ---

type Repository struct {
	ID            string            `bson:"_id"`
	Name          string            `bson:"name"`
	Owner         string            `bson:"owner"`
	Description   string            `bson:"description,omitempty"`
	Language      string            `bson:"language,omitempty"`
	Languages     map[string]int    `bson:"languages,omitempty"`
	Topics        []string          `bson:"topics,omitempty"`
	Stars         int               `bson:"stars"`
	Forks         int               `bson:"forks"`
	Watchers      int               `bson:"watchers"`
	OpenIssues    int               `bson:"open_issues"`
	Size          int               `bson:"size"`
	DefaultBranch string            `bson:"default_branch"`
	CreatedAt     time.Time         `bson:"created_at"`
	UpdatedAt     time.Time         `bson:"updated_at"`
	PushedAt      time.Time         `bson:"pushed_at"`
	LastSynced    time.Time         `bson:"last_synced"`
}

func (mc *MongoClient) UpsertRepository(repo Repository) error {
	collection := mc.db.Collection("repositories")
	
	repo.LastSynced = time.Now()
	repo.ID = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	filter := bson.M{"_id": repo.ID}
	update := bson.M{"$set": repo}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(mc.ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("erro ao upsert repository %s: %w", repo.ID, err)
	}

	return nil
}

// --- Daily Activity Collection ---

type DailyActivity struct {
	User              string    `bson:"user"`
	Date              time.Time `bson:"date"`
	Commits           int       `bson:"commits"`
	PRsOpened         int       `bson:"prs_opened"`
	PRsMerged         int       `bson:"prs_merged"`
	IssuesOpened      int       `bson:"issues_opened"`
	IssuesClosed      int       `bson:"issues_closed"`
	ReposContributed  []string  `bson:"repos_contributed,omitempty"`
	LastSynced        time.Time `bson:"last_synced"`
}

func (mc *MongoClient) UpsertDailyActivity(activity DailyActivity) error {
	collection := mc.db.Collection("daily_activity")
	
	activity.LastSynced = time.Now()

	// Normalizar data para midnight UTC
	activity.Date = time.Date(
		activity.Date.Year(),
		activity.Date.Month(),
		activity.Date.Day(),
		0, 0, 0, 0,
		time.UTC,
	)

	filter := bson.M{
		"user": activity.User,
		"date": activity.Date,
	}
	update := bson.M{"$set": activity}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(mc.ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("erro ao upsert daily_activity %s/%s: %w", 
			activity.User, activity.Date.Format("2006-01-02"), err)
	}

	return nil
}

// --- Contributions Collection ---

type Contribution struct {
	User       string    `bson:"user"`
	Repo       string    `bson:"repo"`
	Year       int       `bson:"year"`
	Month      int       `bson:"month"`
	Commits    int       `bson:"commits"`
	Additions  int       `bson:"additions"`
	Deletions  int       `bson:"deletions"`
	PRs        int       `bson:"prs"`
	Issues     int       `bson:"issues"`
	LastSynced time.Time `bson:"last_synced"`
}

func (mc *MongoClient) UpsertContribution(contrib Contribution) error {
	collection := mc.db.Collection("contributions")
	
	contrib.LastSynced = time.Now()

	filter := bson.M{
		"user":  contrib.User,
		"repo":  contrib.Repo,
		"year":  contrib.Year,
		"month": contrib.Month,
	}
	update := bson.M{"$set": contrib}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(mc.ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("erro ao upsert contribution %s/%s/%d-%02d: %w",
			contrib.User, contrib.Repo, contrib.Year, contrib.Month, err)
	}

	return nil
}

// --- Batch Operations ---

func (mc *MongoClient) UpsertRepositoriesBatch(repos []Repository) error {
	if len(repos) == 0 {
		return nil
	}

	collection := mc.db.Collection("repositories")
	
	var operations []mongo.WriteModel
	for _, repo := range repos {
		repo.LastSynced = time.Now()
		repo.ID = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

		filter := bson.M{"_id": repo.ID}
		update := bson.M{"$set": repo}
		
		operation := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		
		operations = append(operations, operation)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(mc.ctx, operations, opts)
	if err != nil {
		return fmt.Errorf("erro ao batch upsert repositories: %w", err)
	}

	log.Printf("✓ Batch upsert: %d repos (inserted: %d, modified: %d)",
		len(repos), result.InsertedCount, result.ModifiedCount)

	return nil
}

func (mc *MongoClient) UpsertDailyActivitiesBatch(activities []DailyActivity) error {
	if len(activities) == 0 {
		return nil
	}

	collection := mc.db.Collection("daily_activity")
	
	var operations []mongo.WriteModel
	for _, activity := range activities {
		activity.LastSynced = time.Now()
		
		// Normalizar data
		activity.Date = time.Date(
			activity.Date.Year(),
			activity.Date.Month(),
			activity.Date.Day(),
			0, 0, 0, 0,
			time.UTC,
		)

		filter := bson.M{
			"user": activity.User,
			"date": activity.Date,
		}
		update := bson.M{"$set": activity}
		
		operation := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		
		operations = append(operations, operation)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(mc.ctx, operations, opts)
	if err != nil {
		return fmt.Errorf("erro ao batch upsert daily_activity: %w", err)
	}

	log.Printf("✓ Batch upsert: %d activities (inserted: %d, modified: %d)",
		len(activities), result.InsertedCount, result.ModifiedCount)

	return nil
}

// --- Query Helpers ---

func (mc *MongoClient) GetUserActivity(username string, startDate, endDate time.Time) ([]DailyActivity, error) {
	collection := mc.db.Collection("daily_activity")
	
	filter := bson.M{
		"user": username,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	
	opts := options.Find().SetSort(bson.M{"date": 1})
	cursor, err := collection.Find(mc.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(mc.ctx)

	var activities []DailyActivity
	if err := cursor.All(mc.ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

func (mc *MongoClient) GetRepositories(owner string) ([]Repository, error) {
	collection := mc.db.Collection("repositories")
	
	filter := bson.M{"owner": owner}
	opts := options.Find().SetSort(bson.M{"updated_at": -1})
	
	cursor, err := collection.Find(mc.ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(mc.ctx)

	var repos []Repository
	if err := cursor.All(mc.ctx, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}
