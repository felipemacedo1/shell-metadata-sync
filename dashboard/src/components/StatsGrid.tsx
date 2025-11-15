'use client';

import { GitCommit, GitPullRequest, Flame, Calendar, TrendingUp, FolderGit2 } from 'lucide-react';
import { StatCard } from '@/components/dashboard/StatCard';

interface StatsGridProps {
  totalCommits: number;
  totalPRs: number;
  totalIssues: number;
  activeDays: number;
  totalRepos: number;
  periodDays: number;
}

export default function StatsGrid({
  totalCommits,
  totalPRs,
  totalIssues,
  activeDays,
  totalRepos,
  periodDays = 90
}: StatsGridProps) {
  const avgCommitsPerWeek = ((totalCommits / periodDays) * 7).toFixed(1);
  const activityRate = ((activeDays / periodDays) * 100).toFixed(0);

  const stats = [
    { icon: GitCommit, label: 'Total Commits', value: totalCommits, color: 'emerald' as const },
    { icon: GitPullRequest, label: 'Pull Requests', value: totalPRs, color: 'blue' as const },
    { icon: Flame, label: 'Active Days', value: `${activeDays}/${periodDays}`, color: 'orange' as const },
    { icon: FolderGit2, label: 'Repositories', value: totalRepos, color: 'purple' as const },
    { icon: TrendingUp, label: 'Avg. Commits/Week', value: avgCommitsPerWeek, color: 'cyan' as const },
    { icon: Calendar, label: 'Activity Rate', value: `${activityRate}%`, color: 'amber' as const }
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
      {stats.map((stat, index) => (
        <StatCard key={index} {...stat} />
      ))}
    </div>
  );
}
