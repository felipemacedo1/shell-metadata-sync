import { ProfileData, ActivityData, LanguageData, DailyMetric, Repository } from '@/lib/types';

export function aggregateProfiles(primary: ProfileData, secondary: ProfileData | null): ProfileData {
  if (!secondary) return primary;
  return {
    ...primary,
    followers: primary.followers + secondary.followers,
    following: primary.following + secondary.following,
    public_repos: primary.public_repos + secondary.public_repos,
    total_stars_received: (primary.total_stars_received || 0) + (secondary.total_stars_received || 0),
    total_forks_received: (primary.total_forks_received || 0) + (secondary.total_forks_received || 0),
    organizations: [...new Set([...(primary.organizations || []), ...(secondary.organizations || [])])],
  };
}

export function aggregateActivity(primary: ActivityData, secondary: ActivityData | null): ActivityData {
  if (!secondary) return primary;
  const mergedMetrics: Record<string, DailyMetric> = { ...primary.daily_metrics };
  Object.entries(secondary.daily_metrics).forEach(([date, metrics]) => {
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
  return { metadata: primary.metadata, daily_metrics: mergedMetrics, summary: primary.summary };
}

export function aggregateLanguages(primary: LanguageData, secondary: LanguageData | null): LanguageData {
  if (!secondary) return primary;
  const mergedLanguages: Record<string, { bytes: number; repos: number; percentage: number }> = {};
  let totalBytes = 0;
  [...new Set([...Object.keys(primary.languages), ...Object.keys(secondary.languages)])].forEach(lang => {
    const p = primary.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };
    const s = secondary.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };
    mergedLanguages[lang] = { bytes: p.bytes + s.bytes, repos: p.repos + s.repos, percentage: 0 };
    totalBytes += mergedLanguages[lang].bytes;
  });
  Object.keys(mergedLanguages).forEach(lang => {
    mergedLanguages[lang].percentage = (mergedLanguages[lang].bytes / totalBytes) * 100;
  });
  const topLanguages = Object.entries(mergedLanguages).sort(([, a], [, b]) => b.bytes - a.bytes).slice(0, 10).map(([name]) => name);
  return { metadata: primary.metadata, languages: mergedLanguages, top_languages: topLanguages };
}

export interface AggregatedData {
  profile: ProfileData;
  activity: ActivityData;
  languages: LanguageData;
  repositories: Repository[];
}

export function aggregateAllData(
  primaryProfile: ProfileData, secondaryProfile: ProfileData | null,
  primaryActivity: ActivityData, secondaryActivity: ActivityData | null,
  primaryLanguages: LanguageData, secondaryLanguages: LanguageData | null,
  repositories: Repository[]
): AggregatedData {
  return {
    profile: aggregateProfiles(primaryProfile, secondaryProfile),
    activity: aggregateActivity(primaryActivity, secondaryActivity),
    languages: aggregateLanguages(primaryLanguages, secondaryLanguages),
    repositories,
  };
}
