'use client';

import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Code2 } from 'lucide-react';

interface LanguageChartProps {
  languages: Record<string, {
    bytes: number;
    repos: number;
    percentage: number;
  }>;
}

const COLORS = [
  '#3b82f6', // blue
  '#8b5cf6', // violet
  '#10b981', // emerald
  '#f59e0b', // amber
  '#ef4444', // red
  '#06b6d4', // cyan
  '#ec4899', // pink
  '#14b8a6', // teal
  '#f97316', // orange
  '#6366f1', // indigo
];

export default function LanguageChart({ languages }: LanguageChartProps) {
  const chartData = Object.entries(languages)
    .map(([name, stats]) => ({
      name,
      value: stats.bytes,
      percentage: stats.percentage,
      repos: stats.repos
    }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 10); // Top 10 languages

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-3 shadow-xl">
          <p className="text-white font-semibold mb-1">{data.name}</p>
          <p className="text-sm text-slate-300">
            {(data.value / 1024).toFixed(1)} KB
          </p>
          <p className="text-sm text-slate-300">
            {data.percentage.toFixed(1)}%
          </p>
          <p className="text-sm text-slate-400">
            {data.repos} {data.repos === 1 ? 'repository' : 'repositories'}
          </p>
        </div>
      );
    }
    return null;
  };

  const CustomLegend = ({ payload }: any) => {
    return (
      <div className="grid grid-cols-2 gap-2 mt-4">
        {payload.map((entry: any, index: number) => (
          <div key={index} className="flex items-center gap-2 text-sm">
            <div 
              className="w-3 h-3 rounded-full shrink-0" 
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-slate-300 truncate">{entry.value}</span>
            <span className="text-slate-500 ml-auto">
              {entry.payload.percentage.toFixed(1)}%
            </span>
          </div>
        ))}
      </div>
    );
  };

  return (
    <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
      <div className="flex items-center gap-2 mb-6">
        <Code2 className="w-6 h-6 text-purple-400" />
        <h2 className="text-2xl font-bold text-white">Language Distribution</h2>
      </div>

      <ResponsiveContainer width="100%" height={300}>
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            labelLine={false}
            outerRadius={100}
            innerRadius={60}
            fill="#8884d8"
            dataKey="value"
            animationBegin={0}
            animationDuration={800}
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend content={<CustomLegend />} />
        </PieChart>
      </ResponsiveContainer>

      {/* Language Bars */}
      <div className="mt-6 space-y-3">
        {chartData.slice(0, 5).map((lang, index) => (
          <div key={lang.name}>
            <div className="flex items-center justify-between mb-1">
              <div className="flex items-center gap-2">
                <div 
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-white font-medium">{lang.name}</span>
              </div>
              <span className="text-slate-400 text-sm">
                {lang.percentage.toFixed(1)}% · {lang.repos} repos
              </span>
            </div>
            <div className="w-full bg-slate-700/50 rounded-full h-2">
              <div 
                className="h-2 rounded-full transition-all"
                style={{ 
                  width: `${lang.percentage}%`,
                  backgroundColor: COLORS[index % COLORS.length]
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
