// Client-side API functions for fetching data
// Works with both Next.js API routes and static JSON files

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || '/api';
const USE_STATIC = process.env.NEXT_PUBLIC_USE_STATIC === 'true';

export interface Repository {
  name: string;
  owner: string;
  description?: string;
  language?: string;
  url: string;
  updated_at: string;
}

export interface ProfileData {
  login: string;
  name: string;
  bio: string;
  avatar_url: string;
  followers: number;
  following: number;
  public_repos: number;
  total_stars_received: number;
  total_forks_received: number;
  organizations: string[];
  generated_at: string;
}

export interface DailyMetric {
  commits: number;
  prs: number;
  issues: number;
}

export interface ActivityData {
  metadata: {
    user: string;
    period: string;
    start_date: string;
    end_date: string;
    generated_at: string;
  };
  daily_metrics: Record<string, DailyMetric>;
}

export interface LanguageData {
  metadata: {
    user: string;
    generated_at: string;
  };
  languages: Record<string, {
    bytes: number;
    repos: number;
    percentage: number;
  }>;
  top_languages: string[];
}

export interface ProjectsResponse {
  metadata: {
    generated_at: string;
    total_repos: number;
    users: string[];
  };
  repositories: Repository[];
}

// Fetch functions - trabalham com API routes ou arquivos estáticos
async function fetchData<T>(endpoint: string): Promise<T | null> {
  try {
    const url = USE_STATIC 
      ? `/data/${endpoint}.json`
      : `${API_BASE}/${endpoint}`;
    
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return await response.json();
  } catch (error) {
    console.error(`Error fetching ${endpoint}:`, error);
    return null;
  }
}

export async function fetchRepositories(): Promise<Repository[]> {
  const data = await fetchData<ProjectsResponse>('projects');
  return data?.repositories || [];
}

export async function fetchProfile(): Promise<ProfileData | null> {
  return fetchData<ProfileData>('profile');
}

export async function fetchActivity(): Promise<ActivityData | null> {
  return fetchData<ActivityData>('activity');
}

export async function fetchLanguages(): Promise<LanguageData | null> {
  return fetchData<LanguageData>('languages');
}

export async function fetchMetadata() {
  return fetchData<any>('metadata');
}

// Helper functions para transformar dados
export function transformActivityForChart(activityData: ActivityData | null) {
  if (!activityData) return [];
  
  return Object.entries(activityData.daily_metrics).map(([date, metrics]) => ({
    date,
    commits: metrics.commits,
    prs: metrics.prs,
    issues: metrics.issues
  }));
}

export function calculateStreak(activityData: ActivityData | null): number {
  if (!activityData) return 0;
  
  let currentStreak = 0;
  const sortedDates = Object.keys(activityData.daily_metrics).sort().reverse();
  
  for (const date of sortedDates) {
    const metrics = activityData.daily_metrics[date];
    if (metrics.commits > 0 || metrics.prs > 0 || metrics.issues > 0) {
      currentStreak++;
    } else {
      break;
    }
  }
  
  return currentStreak;
}

export function transformLanguagesForChart(languageData: LanguageData | null) {
  if (!languageData) return [];
  
  return Object.entries(languageData.languages).map(([name, data]) => ({
    name,
    value: data.bytes,
    percentage: data.percentage
  }));
}

export function calculateRepoStats(repos: Repository[]) {
  const totalRepos = repos.length;
  const reposWithLanguage = repos.filter(r => r.language).length;
  const languageCounts: Record<string, number> = {};
  
  repos.forEach(repo => {
    if (repo.language) {
      languageCounts[repo.language] = (languageCounts[repo.language] || 0) + 1;
    }
  });
  
  const topLanguage = Object.entries(languageCounts)
    .sort(([, a], [, b]) => b - a)[0]?.[0] || 'N/A';
  
  // Calculate recently updated (last 30 days)
  const thirtyDaysAgo = new Date();
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
  const recentlyUpdated = repos.filter(r => {
    const updatedAt = new Date(r.updated_at);
    return updatedAt >= thirtyDaysAgo;
  }).length;
  
  return {
    totalRepos,
    reposWithLanguage,
    topLanguage,
    languageCount: Object.keys(languageCounts).length,
    recentlyUpdated
  };
}
