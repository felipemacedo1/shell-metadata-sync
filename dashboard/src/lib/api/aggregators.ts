import { ProfileData, ActivityData, LanguageData, DailyMetric, Repository } from '@/lib/types';

export function aggregateProfiles(
  primary: ProfileData | null,
  secondary: ProfileData | null
): ProfileData | null {
  if (!primary) return secondary;
  if (!secondary) return primary;

  console.log('🔗 Aggregating profiles:', { primary: primary.login, secondary: secondary.login });

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

export function aggregateActivity(
  primary: ActivityData | null,
  secondary: ActivityData | null
): ActivityData | null {
  if (!primary) return secondary;
  if (!secondary) return primary;

  console.log('🔗 Aggregating activity data');

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

  const totalCommits = Object.values(mergedMetrics).reduce((sum, m) => sum + m.commits, 0);
  const totalPRs = Object.values(mergedMetrics).reduce((sum, m) => sum + m.prs, 0);
  const totalIssues = Object.values(mergedMetrics).reduce((sum, m) => sum + m.issues, 0);
  
  console.log('📊 Merged activity:', { totalCommits, totalPRs, totalIssues, days: Object.keys(mergedMetrics).length });

  return {
    metadata: primary.metadata,
    daily_metrics: mergedMetrics,
    summary: {
      total_commits: totalCommits,
      total_prs: totalPRs,
      total_issues: totalIssues,
      active_days: Object.values(mergedMetrics).filter(m => m.commits > 0 || m.prs > 0 || m.issues > 0).length
    }
  };
}

export function aggregateLanguages(
  primary: LanguageData | null,
  secondary: LanguageData | null
): LanguageData | null {
  if (!primary) return secondary;
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

  const topLanguages = Object.entries(mergedLanguages)
    .sort(([, a], [, b]) => b.bytes - a.bytes)
    .slice(0, 10)
    .map(([name]) => name);

  return {
    metadata: primary.metadata,
    languages: mergedLanguages,
    top_languages: topLanguages,
  };
}

export interface AggregatedData {
  profile: ProfileData | null;
  activity: ActivityData | null;
  languages: LanguageData | null;
  repositories: Repository[];
}

export function aggregateAllData(
  primaryProfile: ProfileData | null,
  secondaryProfile: ProfileData | null,
  primaryActivity: ActivityData | null,
  secondaryActivity: ActivityData | null,
  primaryLanguages: LanguageData | null,
  secondaryLanguages: LanguageData | null,
  repositories: Repository[]
): AggregatedData {
  console.log('🔄 Starting aggregation...');
  
  const result = {
    profile: aggregateProfiles(primaryProfile, secondaryProfile),
    activity: aggregateActivity(primaryActivity, secondaryActivity),
    languages: aggregateLanguages(primaryLanguages, secondaryLanguages),
    repositories,
  };
  
  console.log('✅ Aggregation complete');
  return result;
}
