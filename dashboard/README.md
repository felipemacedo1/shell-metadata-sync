# Dashboard - GitHub Metrics

Dashboard Next.js que exibe métricas do GitHub estaticamente no GitHub Pages.

## Quick Start

```bash
npm install
npm run dev
```

Acesse: http://localhost:3000

## Build

```bash
npm run build  # Copia JSONs e gera build estático em out/
```

## Variáveis

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

