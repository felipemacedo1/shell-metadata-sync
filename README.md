# Dev Metadata Sync

Automated GitHub data collection via GitHub Actions.

## Architecture

```
GitHub Actions (scheduled) → Go Collectors → MongoDB Atlas → JSON Export → GitHub Pages
```

## Features

- **Automated**: Runs every 6 hours via GitHub Actions
- **MongoDB Storage**: Persistent data in MongoDB Atlas
- **Static Export**: JSONs for GitHub Pages
- **Dashboard**: Next.js visualization

## Setup

### 1. Fork Repository

Configure secrets in Settings → Secrets and variables → Actions:

- `MONGODB_URI` - Connection string from MongoDB Atlas
- `MONGODB_DATABASE` - Database name (optional)
- `GH_USERS` - GitHub users/orgs to track

### 2. MongoDB Atlas

1. Create free cluster at cloud.mongodb.com
2. Configure Network Access (allow GitHub Actions)
3. Create database user
4. Copy connection string to MONGODB_URI secret

### 3. Run

Actions → Sync to MongoDB Atlas → Run workflow

## Local Development

```bash
cp .env.example .env
# Edit .env with your credentials
go build -o bin/test ./cmd/test && ./bin/test
go build -o bin/sync ./cmd/sync && ./bin/sync
```

## Structure

```
├── .github/workflows/   # GitHub Actions automation
├── cmd/                 # Go binaries (sync, test)
├── scripts/collectors/  # Data collectors
├── data/               # Exported JSONs
└── dashboard/          # Next.js app
```

## License

MIT
