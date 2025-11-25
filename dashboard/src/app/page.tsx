import { fetchAllData } from '@/lib/api/fetchers';
import { aggregateAllData } from '@/lib/api/aggregators';
import Hero from '@/components/Hero';
import StatsGrid from '@/components/StatsGrid';
import ContributionHeatmap from '@/components/charts/ContributionHeatmap';
import ActivityTimeline from '@/components/ActivityTimeline';
import LanguageChart from '@/components/LanguageChart';
import RepositoryGrid from '@/components/RepositoryGrid';

export default async function Home() {
  const [
    primaryProfile,
    secondaryProfile,
    primaryActivity,
    secondaryActivity,
    primaryLanguages,
    secondaryLanguages,
    repositories,
  ] = await fetchAllData();

  const { profile: profileData, activity: activityData, languages: languageData, repositories: repos } = aggregateAllData(
    primaryProfile,
    secondaryProfile,
    primaryActivity,
    secondaryActivity,
    primaryLanguages,
    secondaryLanguages,
    repositories
  );

  // Calculate stats (prefer the pre-calculated summary if present)
  const totalCommits = activityData
    ? (activityData.summary?.total_commits ?? Object.values(activityData.daily_metrics).reduce((sum, day) => sum + day.commits, 0))
    : 0;

  const totalPRs = activityData
    ? Object.values(activityData.daily_metrics).reduce((sum, day) => sum + day.prs, 0)
    : 0;

  const totalIssues = activityData
    ? Object.values(activityData.daily_metrics).reduce((sum, day) => sum + day.issues, 0)
    : 0;

  const activeDays = activityData
    ? Object.values(activityData.daily_metrics).filter(day => day.commits > 0).length
    : 0;

  const periodDays = activityData
    ? Math.ceil((new Date(activityData.metadata.end_date).getTime() - new Date(activityData.metadata.start_date).getTime()) / (1000 * 60 * 60 * 24))
    : 90;

  // Prepare heatmap data
  const heatmapData = activityData
    ? Object.entries(activityData.daily_metrics).map(([date, metrics]) => ({
        date,
        count: metrics.commits + metrics.prs + metrics.issues
      }))
    : [];

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        {/* Hero Section */}
        <Hero profile={profileData} />

        {/* Stats Grid */}
        <StatsGrid
          totalCommits={totalCommits}
          totalPRs={totalPRs}
          totalIssues={totalIssues}
          activeDays={activeDays}
          totalRepos={repos.length}
          periodDays={periodDays}
        />

        {/* Contribution Heatmap */}
        {activityData && (
          <div className="mb-8">
            <ContributionHeatmap
              data={heatmapData}
              startDate={new Date(activityData.metadata.start_date)}
              endDate={new Date(activityData.metadata.end_date)}
            />
          </div>
        )}

        {/* Activity Timeline & Language Distribution (stacked, full-width) */}
        <div className="grid grid-cols-1 gap-8 mb-8">
          {activityData && (
            <ActivityTimeline dailyMetrics={activityData.daily_metrics} />
          )}
          {languageData && (
            <LanguageChart languages={languageData.languages} />
          )}
        </div>

        {/* Repository Grid */}
        <div className="mb-8">
          <div className="mb-6">
            <h2 className="text-3xl font-bold text-white mb-2">Repositories</h2>
            <p className="text-slate-400">
              Browse and filter through all {repos.length} repositories
            </p>
          </div>
          <RepositoryGrid repositories={repos} />
        </div>

        {/* Footer */}
        <footer className="mt-12 pt-8 border-t border-slate-800 text-center">
          <div className="space-y-2">
            <p className="text-slate-400 text-sm">
              🚀 Built with Next.js, TypeScript, and Tailwind CSS
            </p>
            <p className="text-slate-500 text-xs">
              Data aggregated from multiple GitHub accounts and collected automatically every 6 hours
            </p>
            {activityData && (
              <p className="text-slate-600 text-xs">
                Last sync: {new Date(activityData.metadata.generated_at).toLocaleString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit'
                })}
              </p>
            )}
          </div>
        </footer>
      </div>
    </main>
  );
}
