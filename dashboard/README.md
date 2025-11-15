# GitHub Portfolio Dashboard

Modern portfolio dashboard showcasing GitHub stats evolution, activity tracking, and project contributions.

## 🚀 Features

- **Hero Section**: Profile overview with avatar, stats, and bio
- **Activity Heatmap**: GitHub-style contribution calendar (90 days)
- **Stats Grid**: Key metrics including commits, PRs, active days, and activity rate
- **Activity Timeline**: Weekly contribution trends with area charts
- **Language Distribution**: Visual breakdown of programming languages used
- **Repository Grid**: Filterable and searchable repository showcase
  - Search by name or description
  - Filter by programming language
  - Sort by updated date, name, stars, or language

## 🛠️ Tech Stack

- **Next.js 16** - React framework with App Router
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first styling
- **Recharts** - Interactive charts
- **React Calendar Heatmap** - Contribution heatmap
- **Lucide React** - Icon library

## 🏃 Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Access: http://localhost:3000

## 📦 Build & Deploy

```bash
# Build for production (includes data sync)
npm run build
```

This syncs data from `/data/` to `/public/data/` and generates static export in `./out/`

## 📊 Data Sources

Dashboard reads JSON files from `/public/data/`:

- `profile.json` - User profile information
- `projects.json` - Repository list
- `languages.json` - Language statistics
- `activity-daily.json` - Daily contributions
- `metadata.json` - Sync metadata

Data automatically synced from `/data/` via build script.

## 🎨 Components

```
components/
├── Hero.tsx                    # Profile hero section
├── StatsGrid.tsx              # Key metrics cards
├── RepositoryGrid.tsx         # Filterable repo list
├── ActivityTimeline.tsx       # Weekly activity chart
├── LanguageChart.tsx          # Language pie chart
└── charts/
    └── ContributionHeatmap.tsx # Activity heatmap
```

Crie `.env.local`:

```env
NEXT_PUBLIC_USE_STATIC=true  # true=static files, false=API routes
```

## Deploy

O workflow `.github/workflows/deploy-pages.yml` faz deploy automático no GitHub Pages quando há mudanças em `data/` ou `dashboard/`.

## Stack

- Next.js 16 + React 19
- Tailwind 4 + Tremor
- Recharts + Calendar Heatmap
- Static export para GitHub Pages

