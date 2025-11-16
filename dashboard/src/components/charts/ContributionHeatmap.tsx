'use client';

import { useState } from 'react';
import CalendarHeatmap from 'react-calendar-heatmap';
import 'react-calendar-heatmap/dist/styles.css';
import { Activity, TrendingUp, Calendar, Flame } from 'lucide-react';

interface HeatmapData {
  date: string;
  count: number;
}

interface ContributionHeatmapProps {
  data: HeatmapData[];
  startDate?: Date;
  endDate?: Date;
}

export default function ContributionHeatmap({ data, startDate, endDate }: ContributionHeatmapProps) {
  const [hoveredDay, setHoveredDay] = useState<HeatmapData | null>(null);
  
  const today = endDate || new Date();
  const start = startDate || new Date(today.getFullYear(), today.getMonth() - 3, today.getDate());

  const getColorClass = (count: number | undefined) => {
    if (!count || count === 0) return 'color-empty';
    if (count < 3) return 'color-scale-1';
    if (count < 6) return 'color-scale-2';
    if (count < 10) return 'color-scale-3';
    return 'color-scale-4';
  };

  // Calculate statistics
  const totalContributions = data.reduce((sum, day) => sum + day.count, 0);
  const maxDay = data.reduce((max, day) => day.count > max.count ? day : max, { date: '', count: 0 });
  const activeDays = data.filter(day => day.count > 0).length;
  const avgPerDay = activeDays > 0 ? (totalContributions / activeDays).toFixed(1) : '0';
  
  // Calculate current streak
  const sortedData = [...data].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
  let currentStreak = 0;
  for (const day of sortedData) {
    if (day.count > 0) currentStreak++;
    else break;
  }
  
  // Calculate longest streak
  let longestStreak = 0;
  let tempStreak = 0;
  const sortedAsc = [...data].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
  for (const day of sortedAsc) {
    if (day.count > 0) {
      tempStreak++;
      longestStreak = Math.max(longestStreak, tempStreak);
    } else {
      tempStreak = 0;
    }
  }

  return (
    <div className="bg-gradient-to-br from-slate-800/50 to-slate-900/50 backdrop-blur-sm border border-slate-700/50 rounded-xl p-6 shadow-xl hover:border-slate-600/50 transition-all duration-300">
      {/* Header with Stats */}
      <div className="flex items-start justify-between mb-6 flex-wrap gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-emerald-500/10 rounded-lg">
            <Activity className="w-6 h-6 text-emerald-400" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-white">Contribution Activity</h2>
            <p className="text-sm text-slate-400 mt-0.5">
              {start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} - {today.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
            </p>
          </div>
        </div>
        
        {/* Stats Cards */}
        <div className="flex gap-4 flex-wrap">
          <div className="bg-slate-900/50 border border-slate-700/50 rounded-lg px-4 py-2">
            <p className="text-xs text-slate-400 mb-1">Total</p>
            <p className="text-2xl font-bold text-emerald-400">{totalContributions}</p>
          </div>
          <div className="bg-slate-900/50 border border-slate-700/50 rounded-lg px-4 py-2">
            <div className="flex items-center gap-1 mb-1">
              <Flame className="w-3 h-3 text-orange-400" />
              <p className="text-xs text-slate-400">Streak</p>
            </div>
            <p className="text-2xl font-bold text-orange-400">{currentStreak}</p>
          </div>
          <div className="bg-slate-900/50 border border-slate-700/50 rounded-lg px-4 py-2">
            <div className="flex items-center gap-1 mb-1">
              <TrendingUp className="w-3 h-3 text-blue-400" />
              <p className="text-xs text-slate-400">Avg/Day</p>
            </div>
            <p className="text-2xl font-bold text-blue-400">{avgPerDay}</p>
          </div>
        </div>
      </div>

      {/* Heatmap */}
      <div className="relative w-full overflow-x-auto scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
        <div className="min-w-[800px] relative">
          <CalendarHeatmap
            startDate={start}
            endDate={today}
            values={data}
            classForValue={(value) => getColorClass(value?.count)}
            showWeekdayLabels={true}
            onMouseOver={(event, value) => {
              if (value) {
                setHoveredDay(value as HeatmapData);
              }
            }}
            onMouseLeave={() => setHoveredDay(null)}
          />
          
          {/* Inline Tooltip */}
          {hoveredDay && (
            <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-4 py-2 bg-slate-900 border border-emerald-500/30 rounded-lg shadow-xl text-sm pointer-events-none z-50 whitespace-nowrap">
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-emerald-400" />
                <span className="text-white font-semibold">
                  {new Date(hoveredDay.date).toLocaleDateString('en-US', { 
                    weekday: 'short', 
                    month: 'short', 
                    day: 'numeric',
                    year: 'numeric'
                  })}
                </span>
              </div>
              <p className="text-emerald-400 font-bold mt-1">
                {hoveredDay.count} {hoveredDay.count === 1 ? 'contribution' : 'contributions'}
              </p>
              <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 translate-y-full">
                <div className="border-8 border-transparent border-t-slate-900"></div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Legend and Additional Stats */}
      <div className="mt-6 pt-4 border-t border-slate-700/50">
        <div className="flex items-center justify-between flex-wrap gap-4">
          {/* Color Legend */}
          <div className="flex items-center gap-3">
            <span className="text-xs text-slate-400">Less</span>
            <div className="flex gap-1.5">
              <div className="w-4 h-4 rounded border border-slate-600/50 bg-slate-800/50 hover:scale-110 transition-transform cursor-pointer" title="No contributions" />
              <div className="w-4 h-4 rounded bg-emerald-900/50 hover:scale-110 transition-transform cursor-pointer" title="1-2 contributions" />
              <div className="w-4 h-4 rounded bg-emerald-700/70 hover:scale-110 transition-transform cursor-pointer" title="3-5 contributions" />
              <div className="w-4 h-4 rounded bg-emerald-600 hover:scale-110 transition-transform cursor-pointer" title="6-9 contributions" />
              <div className="w-4 h-4 rounded bg-emerald-500 hover:scale-110 transition-transform cursor-pointer" title="10+ contributions" />
            </div>
            <span className="text-xs text-slate-400">More</span>
          </div>

          {/* Stats Summary */}
          <div className="flex items-center gap-6 text-xs text-slate-400">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-emerald-400" />
              <span>Active: <span className="text-emerald-400 font-semibold">{activeDays}</span> days</span>
            </div>
            {maxDay.count > 0 && (
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-orange-400" />
                <span>Best: <span className="text-orange-400 font-semibold">{maxDay.count}</span> on {new Date(maxDay.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
              </div>
            )}
            {longestStreak > 0 && (
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-blue-400" />
                <span>Longest: <span className="text-blue-400 font-semibold">{longestStreak}</span> days</span>
              </div>
            )}
          </div>
        </div>
      </div>

      <style jsx global>{`
        .react-calendar-heatmap {
          width: 100%;
        }
        
        .react-calendar-heatmap text {
          font-size: 10px;
          fill: #94a3b8;
          font-weight: 500;
        }
        
        .react-calendar-heatmap .color-empty {
          fill: #1e293b;
          rx: 2;
        }
        
        .react-calendar-heatmap .color-scale-1 {
          fill: #064e3b;
          rx: 2;
        }
        
        .react-calendar-heatmap .color-scale-2 {
          fill: #047857;
          rx: 2;
        }
        
        .react-calendar-heatmap .color-scale-3 {
          fill: #10b981;
          rx: 2;
        }
        
        .react-calendar-heatmap .color-scale-4 {
          fill: #34d399;
          rx: 2;
        }
        
        .react-calendar-heatmap rect {
          transition: all 0.2s ease;
        }
        
        .react-calendar-heatmap rect:hover {
          stroke: #3b82f6;
          stroke-width: 2px;
          opacity: 0.9;
          transform: scale(1.1);
          rx: 3;
        }
        
        .scrollbar-thin::-webkit-scrollbar {
          height: 6px;
        }
        
        .scrollbar-thin::-webkit-scrollbar-track {
          background: transparent;
        }
        
        .scrollbar-thin::-webkit-scrollbar-thumb {
          background: #475569;
          border-radius: 3px;
        }
        
        .scrollbar-thin::-webkit-scrollbar-thumb:hover {
          background: #64748b;
        }
      `}</style>
    </div>
  );
}
