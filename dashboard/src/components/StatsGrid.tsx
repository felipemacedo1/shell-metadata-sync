'use client';

import { GitCommit, GitPullRequest, Flame, Calendar, TrendingUp, FolderGit2 } from 'lucide-react';
import { AnimatedCounter } from '@/components/ui/AnimatedCounter';

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
    {
      icon: GitCommit,
      label: 'Total Commits',
      value: totalCommits.toLocaleString(),
      color: 'from-emerald-500 to-green-600',
      bgColor: 'bg-emerald-500/10',
      iconColor: 'text-emerald-400'
    },
    {
      icon: GitPullRequest,
      label: 'Pull Requests',
      value: totalPRs.toLocaleString(),
      color: 'from-blue-500 to-indigo-600',
      bgColor: 'bg-blue-500/10',
      iconColor: 'text-blue-400'
    },
    {
      icon: Flame,
      label: 'Active Days',
      value: `${activeDays}/${periodDays}`,
      color: 'from-orange-500 to-red-600',
      bgColor: 'bg-orange-500/10',
      iconColor: 'text-orange-400'
    },
    {
      icon: FolderGit2,
      label: 'Repositories',
      value: totalRepos.toLocaleString(),
      color: 'from-purple-500 to-pink-600',
      bgColor: 'bg-purple-500/10',
      iconColor: 'text-purple-400'
    },
    {
      icon: TrendingUp,
      label: 'Avg. Commits/Week',
      value: avgCommitsPerWeek,
      color: 'from-cyan-500 to-blue-600',
      bgColor: 'bg-cyan-500/10',
      iconColor: 'text-cyan-400'
    },
    {
      icon: Calendar,
      label: 'Activity Rate',
      value: `${activityRate}%`,
      color: 'from-amber-500 to-yellow-600',
      bgColor: 'bg-amber-500/10',
      iconColor: 'text-amber-400'
    }
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
      {stats.map((stat, index) => {
        const Icon = stat.icon;
        return (
          <div
            key={index}
            className="group relative overflow-hidden rounded-xl bg-slate-800/50 backdrop-blur-sm border border-slate-700 p-6 hover:border-slate-600 transition-all hover:scale-105"
          >
            {/* Background gradient */}
            <div className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-br ${stat.color} opacity-0 group-hover:opacity-10 blur-2xl transition-opacity`}></div>
            
            <div className="relative flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <div className={`p-2 rounded-lg ${stat.bgColor}`}>
                    <Icon className={`w-5 h-5 ${stat.iconColor}`} />
                  </div>
                </div>
                <p className="text-3xl font-bold text-white mb-1">{stat.value}</p>
                <p className="text-sm text-slate-400">{stat.label}</p>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
