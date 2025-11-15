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
  homepage?: string;
  stargazers_count?: number;
  forks_count?: number;
  open_issues_count?: number;
  created_at?: string;
  updated_at: string;
  topics?: string[];
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
  summary?: {
    total_commits: number;
    total_prs: number;
    total_issues: number;
    active_days: number;
  };
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

// Fetch functions - trabalham com dados estaticos em /data
async function fetchData<T>(endpoint: string): Promise<T | null> {
  try {
    // For server-side rendering, we need absolute URL or file system access
    // For client-side, relative URLs work
    const isServer = typeof window === 'undefined';
    const basePath = process.env.NODE_ENV === 'production' ? '/dev-metadata-sync' : '';
    
    let url: string;
    if (isServer) {
      // Server-side: use absolute URL or file system
      const fs = await import('fs/promises');
      const path = await import('path');
      const filePath = path.join(process.cwd(), 'public', 'data', `${endpoint}.json`);
      const data = await fs.readFile(filePath, 'utf-8');
      return JSON.parse(data);
    } else {
      // Client-side: use relative URL
      url = `${basePath}/data/${endpoint}.json`;
      const response = await fetch(url, { cache: 'no-store' });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      return await response.json();
    }
  } catch (error) {
    console.error(`Error fetching ${endpoint}:`, error);
    return null;
  }
}

export async function fetchRepositories(): Promise<Repository[]> {
  // projects.json is a direct array of repositories
  const data = await fetchData<Repository[]>('projects');
  return data || [];
}

export async function fetchProfile(): Promise<ProfileData | null> {
  return fetchData<ProfileData>('profile');
}

export async function fetchActivity(): Promise<ActivityData | null> {
  return fetchData<ActivityData>('activity-daily');
}

export async function fetchLanguages(): Promise<LanguageData | null> {
  return fetchData<LanguageData>('languages');
}

// Aggregate data from multiple users
export async function fetchAggregatedData() {
  const basePath = process.env.NODE_ENV === 'production' ? '/dev-metadata-sync' : '';
  
  // Try to fetch data from multiple sources
  const [primaryProfile, primaryActivity, primaryLanguages, primaryRepos] = await Promise.all([
    fetchProfile(),
    fetchActivity(),
    fetchLanguages(),
    fetchRepositories()
  ]);

  // Try to fetch secondary user data (growthfolio)
  const secondaryProfile = await fetchData<ProfileData>('profile-secondary').catch(() => null);
  const secondaryActivity = await fetchData<ActivityData>('activity-daily-secondary').catch(() => null);
  const secondaryLanguages = await fetchData<LanguageData>('languages-secondary').catch(() => null);

  // Aggregate profile (use primary as base, add secondary stats)
  let aggregatedProfile = primaryProfile;
  if (primaryProfile && secondaryProfile) {
    aggregatedProfile = {
      ...primaryProfile,
      // Keep primary user info (felipemacedo1)
      login: primaryProfile.login,
      name: primaryProfile.name || 'Felipe Macedo',
      bio: primaryProfile.bio,
      avatar_url: primaryProfile.avatar_url,
      // Aggregate stats
      followers: primaryProfile.followers + secondaryProfile.followers,
      following: primaryProfile.following + secondaryProfile.following,
      public_repos: primaryProfile.public_repos + secondaryProfile.public_repos,
      total_stars_received: (primaryProfile.total_stars_received || 0) + (secondaryProfile.total_stars_received || 0),
      total_forks_received: (primaryProfile.total_forks_received || 0) + (secondaryProfile.total_forks_received || 0),
      organizations: [...new Set([...(primaryProfile.organizations || []), ...(secondaryProfile.organizations || [])])],
      generated_at: primaryProfile.generated_at,
    };
  }

  // Aggregate activity (merge daily metrics)
  let aggregatedActivity = primaryActivity;
  if (primaryActivity && secondaryActivity) {
    const mergedMetrics: Record<string, DailyMetric> = { ...primaryActivity.daily_metrics };
    
    Object.entries(secondaryActivity.daily_metrics).forEach(([date, metrics]) => {
      if (mergedMetrics[date]) {
        mergedMetrics[date] = {
          commits: mergedMetrics[date].commits + metrics.commits,
          prs: mergedMetrics[date].prs + metrics.prs,
          issues: mergedMetrics[date].issues + metrics.issues,
        };
      } else {
        mergedMetrics[date] = metrics;
      }
    });

    aggregatedActivity = {
      metadata: primaryActivity.metadata,
      daily_metrics: mergedMetrics,
      summary: primaryActivity.summary,
    };
  }

  // Aggregate languages (merge bytes and repos)
  let aggregatedLanguages = primaryLanguages;
  if (primaryLanguages && secondaryLanguages) {
    const mergedLanguages: Record<string, { bytes: number; repos: number; percentage: number }> = {};
    let totalBytes = 0;

    // Combine languages from both sources
    [...Object.keys(primaryLanguages.languages), ...Object.keys(secondaryLanguages.languages)]
      .forEach(lang => {
        const primary = primaryLanguages.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };
        const secondary = secondaryLanguages.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };
        
        mergedLanguages[lang] = {
          bytes: primary.bytes + secondary.bytes,
          repos: primary.repos + secondary.repos,
          percentage: 0, // Will recalculate
        };
        totalBytes += mergedLanguages[lang].bytes;
      });

    // Recalculate percentages
    Object.keys(mergedLanguages).forEach(lang => {
      mergedLanguages[lang].percentage = (mergedLanguages[lang].bytes / totalBytes) * 100;
    });

    // Sort by bytes and get top languages
    const topLanguages = Object.entries(mergedLanguages)
      .sort(([, a], [, b]) => b.bytes - a.bytes)
      .slice(0, 10)
      .map(([name]) => name);

    aggregatedLanguages = {
      metadata: {
        user: 'felipemacedo1',
        generated_at: primaryLanguages.metadata.generated_at,
      },
      languages: mergedLanguages,
      top_languages: topLanguages,
    };
  }

  return {
    profile: aggregatedProfile,
    activity: aggregatedActivity,
    languages: aggregatedLanguages,
    repositories: primaryRepos, // Repos are already combined in projects.json
  };
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
