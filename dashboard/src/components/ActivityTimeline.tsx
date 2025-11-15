'use client';

import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { TrendingUp } from 'lucide-react';

interface ActivityTimelineProps {
  dailyMetrics: Record<string, {
    commits: number;
    prs: number;
    issues: number;
  }>;
}

export default function ActivityTimeline({ dailyMetrics }: ActivityTimelineProps) {
  // Convert daily metrics to chart data
  const chartData = Object.entries(dailyMetrics)
    .map(([date, metrics]) => ({
      date,
      commits: metrics.commits,
      prs: metrics.prs || 0,
      issues: metrics.issues || 0,
      total: metrics.commits + (metrics.prs || 0) + (metrics.issues || 0)
    }))
    .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
    .slice(-90); // Last 90 days

  // Calculate weekly data
  const weeklyData = [];
  for (let i = 0; i < chartData.length; i += 7) {
    const week = chartData.slice(i, i + 7);
    const weekStart = week[0]?.date;
    const commits = week.reduce((sum, day) => sum + day.commits, 0);
    const prs = week.reduce((sum, day) => sum + day.prs, 0);
    const issues = week.reduce((sum, day) => sum + day.issues, 0);
    
    if (weekStart) {
      weeklyData.push({
        week: new Date(weekStart).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        commits,
        prs,
        issues,
        total: commits + prs + issues
      });
    }
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-3 shadow-xl">
          <p className="text-white font-semibold mb-2">{label}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {entry.name}: {entry.value}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
      <div className="flex items-center gap-2 mb-6">
        <TrendingUp className="w-6 h-6 text-blue-400" />
        <h2 className="text-2xl font-bold text-white">Activity Timeline</h2>
      </div>
      
      <div className="mb-4">
        <p className="text-slate-400 text-sm">Weekly contribution activity over the last 90 days</p>
      </div>

      <ResponsiveContainer width="100%" height={300}>
        <AreaChart data={weeklyData}>
          <defs>
            <linearGradient id="colorCommits" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#10b981" stopOpacity={0.8}/>
              <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorPRs" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8}/>
              <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorIssues" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.8}/>
              <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis 
            dataKey="week" 
            stroke="#94a3b8"
            tick={{ fill: '#94a3b8', fontSize: 12 }}
          />
          <YAxis 
            stroke="#94a3b8"
            tick={{ fill: '#94a3b8', fontSize: 12 }}
          />
          <Tooltip content={<CustomTooltip />} />
          <Legend 
            wrapperStyle={{ color: '#94a3b8' }}
            iconType="circle"
          />
          <Area 
            type="monotone" 
            dataKey="commits" 
            stroke="#10b981" 
            fillOpacity={1} 
            fill="url(#colorCommits)"
            name="Commits"
          />
          <Area 
            type="monotone" 
            dataKey="prs" 
            stroke="#3b82f6" 
            fillOpacity={1} 
            fill="url(#colorPRs)"
            name="Pull Requests"
          />
          <Area 
            type="monotone" 
            dataKey="issues" 
            stroke="#f59e0b" 
            fillOpacity={1} 
            fill="url(#colorIssues)"
            name="Issues"
          />
        </AreaChart>
      </ResponsiveContainer>

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-4 mt-6 pt-6 border-t border-slate-700">
        <div className="text-center">
          <p className="text-3xl font-bold text-emerald-400">
            {weeklyData.reduce((sum, week) => sum + week.commits, 0)}
          </p>
          <p className="text-sm text-slate-400 mt-1">Total Commits</p>
        </div>
        <div className="text-center">
          <p className="text-3xl font-bold text-blue-400">
            {weeklyData.reduce((sum, week) => sum + week.prs, 0)}
          </p>
          <p className="text-sm text-slate-400 mt-1">Total PRs</p>
        </div>
        <div className="text-center">
          <p className="text-3xl font-bold text-amber-400">
            {weeklyData.reduce((sum, week) => sum + week.issues, 0)}
          </p>
          <p className="text-sm text-slate-400 mt-1">Total Issues</p>
        </div>
      </div>
    </div>
  );
}
