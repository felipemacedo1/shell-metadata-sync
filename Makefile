.PHONY: all build clean test install orchestrator help

# Variáveis
BUILD_DIR := bin
CMD_DIR := cmd
ORCHESTRATOR := orchestrator

# Build all
all: clean build

# Build orchestrator
orchestrator:
	@echo "🔨 Building orchestrator..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(ORCHESTRATOR) $(CMD_DIR)/$(ORCHESTRATOR)/main.go
	@echo "✅ Orchestrator built: $(BUILD_DIR)/$(ORCHESTRATOR)"

# Build (alias para orchestrator)
build: orchestrator

# Build collectors (antigos - para compatibilidade)
build-collectors:
	@echo "🔨 Building collectors..."
	@mkdir -p $(BUILD_DIR)
	@echo "  Building activity_collector..."
	@cd scripts/collectors && go build -o ../../$(BUILD_DIR)/activity_collector activity_collector.go
	@echo "  Building repos_collector..."
	@cd scripts/collectors && go build -o ../../$(BUILD_DIR)/repos_collector repos_collector.go
	@echo "  Building stats_collector..."
	@cd scripts/collectors && go build -o ../../$(BUILD_DIR)/stats_collector stats_collector.go
	@echo "  Building user_collector..."
	@cd scripts/collectors && go build -o ../../$(BUILD_DIR)/user_collector user_collector.go
	@echo "✅ Collectors built!"

# Build tudo (novo + antigo)
build-all: orchestrator build-collectors

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/$(ORCHESTRATOR)
	@echo "✅ Clean complete!"

# Clean tudo
clean-all:
	@echo "🧹 Cleaning all build artifacts..."
	@rm -rf $(BUILD_DIR)/*
	@echo "✅ Clean complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed!"

# Run orchestrator - pipeline full
run-full:
	@echo "🚀 Running full pipeline..."
	@go run $(CMD_DIR)/$(ORCHESTRATOR)/main.go -pipeline=full

# Run orchestrator - pipeline quick
run-quick:
	@echo "⚡ Running quick pipeline..."
	@go run $(CMD_DIR)/$(ORCHESTRATOR)/main.go -pipeline=quick

# Dry run
dry-run:
	@echo "🔍 Dry run mode..."
	@go run $(CMD_DIR)/$(ORCHESTRATOR)/main.go -dry-run -pipeline=full

# Help
help:
	@echo "📖 Available targets:"
	@echo ""
	@echo "  Build:"
	@echo "    make build              - Build orchestrator"
	@echo "    make build-collectors   - Build legacy collectors"
	@echo "    make build-all          - Build everything"
	@echo ""
	@echo "  Run:"
	@echo "    make run-full           - Run full pipeline"
	@echo "    make run-quick          - Run quick pipeline"
	@echo "    make dry-run            - Dry run (no execution)"
	@echo ""
	@echo "  Maintenance:"
	@echo "    make clean              - Clean orchestrator binary"
	@echo "    make clean-all          - Clean all binaries"
	@echo "    make test               - Run tests"
	@echo "    make install            - Install dependencies"
	@echo ""
	@echo "  Help:"
	@echo "    make help               - Show this help"
