package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/felipemacedo1/dev-metadata-sync/internal/orchestrator"
)

func main() {
	configPath := flag.String("config", "config/orchestrator.yml", "Path to configuration file")
	pipeline := flag.String("pipeline", "full", "Pipeline to execute (full, quick)")
	dryRun := flag.Bool("dry-run", false, "Dry run mode (don't execute tasks)")
	flag.Parse()

	// Load configuration
	fmt.Printf("📂 Loading config from: %s\n", *configPath)
	config, err := orchestrator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf("✅ Config loaded: %s (v%s)\n\n", config.Name, config.Version)

	if *dryRun {
		fmt.Println("🔍 DRY RUN MODE - No tasks will be executed\n")
		p, ok := config.GetPipeline(*pipeline)
		if !ok {
			log.Fatalf("❌ Pipeline '%s' not found", *pipeline)
		}
		fmt.Printf("Pipeline: %s\n", p.Name)
		fmt.Printf("Description: %s\n", p.Description)
		fmt.Printf("Enabled: %v\n", p.Enabled)
		fmt.Printf("Stages: %d\n", len(p.Stages))
		for i, stage := range p.Stages {
			fmt.Printf("  %d. %s (%d tasks, parallel=%v)\n", i+1, stage.Name, len(stage.Tasks), stage.Parallel)
		}
		return
	}

	// Create orchestrator
	orch := orchestrator.New(config)

	// Execute pipeline
	ctx := context.Background()
	if err := orch.ExecutePipeline(ctx, *pipeline); err != nil {
		orch.PrintReport()
		log.Fatalf("❌ Pipeline execution failed: %v", err)
	}

	// Print report
	orch.PrintReport()

	if !orch.Results.Success {
		os.Exit(1)
	}
}
