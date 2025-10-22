# dev-metadata-sync

This repository generates a public JSON file containing combined public repositories from `felipemacedo1` and `growthfolio` and serves it from `data/projects.json` (suitable for GitHub Pages).

Usage

- The script is `scripts/update_projects.go`. It reads GH_TOKEN from the environment (optional) and writes `data/projects.json`.
- A GitHub Actions workflow runs every 6 hours or manually, compiles the Go program (Go 1.22), runs it, and commits changes if any.

Pages

Configure GitHub Pages to serve from the `data/` folder on the default branch (or `gh-pages` branch if you prefer). The JSON file will then be available publicly at `https://<owner>.github.io/<repo>/projects.json`.
# dev-metadata-sync