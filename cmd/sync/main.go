package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	users := os.Getenv("GH_USERS")
	if users == "" {
		log.Fatal("GH_USERS environment variable not set")
	}

	userList := strings.Split(users, ",")
	
	fmt.Println("🚀 Sync to MongoDB Atlas")
	fmt.Println("Users:", users)
	fmt.Println()

	// User profiles
	fmt.Println("👤 Syncing user profiles...")
	for _, user := range userList {
		user = strings.TrimSpace(user)
		fmt.Printf("   → %s\n", user)
		if err := runCollector("user_collector", "-user="+user); err != nil {
			log.Printf("⚠️  Failed to sync user %s: %v", user, err)
		}
	}

	// Repositories
	fmt.Println("\n📚 Syncing repositories...")
	if err := runCollector("repos_collector", "-users="+users); err != nil {
		log.Printf("⚠️  Failed to sync repos: %v", err)
	}

	// Languages
	fmt.Println("\n💻 Syncing languages...")
	for _, user := range userList {
		user = strings.TrimSpace(user)
		fmt.Printf("   → %s\n", user)
		if err := runCollector("stats_collector", "-user="+user); err != nil {
			log.Printf("⚠️  Failed to sync stats for %s: %v", user, err)
		}
	}

	// Activity
	fmt.Println("\n📊 Syncing activity (90 days)...")
	for _, user := range userList {
		user = strings.TrimSpace(user)
		fmt.Printf("   → %s\n", user)
		if err := runCollector("activity_collector", "-user="+user, "-days=90"); err != nil {
			log.Printf("⚠️  Failed to sync activity for %s: %v", user, err)
		}
	}

	// Export JSONs
	fmt.Println("\n📦 Exporting to JSON...")
	if err := runBinary("export_from_mongo", "-out=data"); err != nil {
		log.Printf("⚠️  Failed to export: %v", err)
	}

	fmt.Println("\n✅ Sync complete!")
}

func runCollector(name string, args ...string) error {
	return runBinary(name, args...)
}

func runBinary(name string, args ...string) error {
	cmd := exec.Command("./bin/"+name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}
