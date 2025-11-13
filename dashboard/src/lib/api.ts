import fs from 'fs';
import path from 'path';

const DATA_DIR = path.join(process.cwd(), '..', 'data');

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

function readJSONFile<T>(filename: string): T {
  const filePath = path.join(DATA_DIR, filename);
  const fileContent = fs.readFileSync(filePath, 'utf-8');
  return JSON.parse(fileContent) as T;
}

export async function fetchRepositories(): Promise<Repository[]> {
  try {
    return readJSONFile<Repository[]>('projects.json');
  } catch (error) {
    console.error('Error fetching repositories:', error);
    return [];
  }
}

export async function fetchProfile(): Promise<ProfileData | null> {
  try {
    return readJSONFile<ProfileData>('profile.json');
  } catch (error) {
    console.error('Error fetching profile:', error);
    return null;
  }
}

export async function fetchActivity(): Promise<ActivityData | null> {
  try {
    return readJSONFile<ActivityData>('activity-daily.json');
  } catch (error) {
    console.error('Error fetching activity:', error);
    return null;
  }
}

export async function fetchLanguages(): Promise<LanguageData | null> {
  try {
    return readJSONFile<LanguageData>('languages.json');
  } catch (error) {
    console.error('Error fetching languages:', error);
    return null;
  }
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
