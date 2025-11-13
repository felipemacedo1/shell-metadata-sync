// TypeScript types for GitHub Analytics data

export interface Repository {
  name: string;
  owner: string;
  description?: string;
  language?: string;
  url: string;
  updated_at: string;
}

export interface Profile {
  login: string;
  name: string;
  bio?: string;
  avatar_url: string;
  followers: number;
  following: number;
  public_repos: number;
  total_stars_received?: number;
  total_forks_received?: number;
  organizations?: string[];
  generated_at: string;
}

export interface DailyActivity {
  date: string;
  commits: number;
  prs?: number;
  issues?: number;
}

export interface ActivityData {
  metadata: {
    user: string;
    period: string;
    start_date: string;
    end_date: string;
    generated_at: string;
  };
  daily_metrics: Record<string, DailyActivity>;
}

export interface LanguageStats {
  bytes: number;
  repos: number;
  percentage: number;
}

export interface LanguagesData {
  metadata: {
    user: string;
    generated_at: string;
  };
  languages: Record<string, LanguageStats>;
  top_languages: string[];
}

export interface RepoContribution {
  repo: string;
  commits: number;
  additions: number;
  deletions: number;
  prs: number;
  issues: number;
}

export interface ContributionsData {
  metadata: {
    user: string;
    generated_at: string;
  };
  by_repo: RepoContribution[];
  summary: {
    total_commits: number;
    total_prs: number;
    total_issues: number;
    active_repos: number;
  };
}

export interface Metadata {
  last_sync: {
    repos?: string;
    activity?: string;
    stats?: string;
  };
  data_coverage: {
    repos_count: number;
    activity_days?: number;
    commits_tracked?: number;
  };
  version: string;
}

// Chart data interfaces
export interface ChartDataPoint {
  date: string;
  value: number;
  label?: string;
}

export interface LanguageChartData {
  name: string;
  value: number;
  percentage: number;
  color?: string;
}

export interface HeatmapValue {
  date: string;
  count: number;
}
