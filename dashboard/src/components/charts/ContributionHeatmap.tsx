'use client';

import CalendarHeatmap from 'react-calendar-heatmap';
import 'react-calendar-heatmap/dist/styles.css';
import { Activity } from 'lucide-react';

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
  const today = endDate || new Date();
  const start = startDate || new Date(today.getFullYear(), today.getMonth() - 3, today.getDate());

  const getColorClass = (count: number | undefined) => {
    if (!count || count === 0) return 'color-empty';
    if (count < 3) return 'color-scale-1';
    if (count < 6) return 'color-scale-2';
    if (count < 9) return 'color-scale-3';
    return 'color-scale-4';
  };

  const totalContributions = data.reduce((sum, day) => sum + day.count, 0);
  const maxDay = data.reduce((max, day) => day.count > max.count ? day : max, { date: '', count: 0 });

  return (
    <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Activity className="w-6 h-6 text-emerald-400" />
          <h2 className="text-2xl font-bold text-white">Contribution Heatmap</h2>
        </div>
        <div className="text-right">
          <p className="text-2xl font-bold text-emerald-400">{totalContributions}</p>
          <p className="text-sm text-slate-400">contributions</p>
        </div>
      </div>

      <div className="w-full overflow-x-auto">
        <div className="min-w-[800px]">
          <CalendarHeatmap
            startDate={start}
            endDate={today}
            values={data}
            classForValue={(value) => getColorClass(value?.count)}
            showWeekdayLabels={true}
            tooltipDataAttrs={(value: any) => {
              if (!value || !value.date) {
                return {};
              }
              return {
                'data-tip': `${value.count || 0} contributions on ${new Date(value.date).toLocaleDateString()}`
              };
            }}
          />
        </div>
      </div>

      {/* Legend and Stats */}
      <div className="mt-6 flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-center gap-2 text-sm text-slate-400">
          <span>Less</span>
          <div className="flex gap-1">
            <div className="w-3 h-3 rounded-sm bg-slate-700/50" />
            <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: '#9be9a8' }} />
            <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: '#40c463' }} />
            <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: '#30a14e' }} />
            <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: '#216e39' }} />
          </div>
          <span>More</span>
        </div>

        {maxDay.count > 0 && (
          <div className="text-sm text-slate-400">
            Most active: <span className="text-emerald-400 font-semibold">{maxDay.count}</span> on{' '}
            {new Date(maxDay.date).toLocaleDateString()}
          </div>
        )}
      </div>

      <style jsx global>{`
        .react-calendar-heatmap {
          width: 100%;
        }
        
        .react-calendar-heatmap text {
          font-size: 10px;
          fill: #94a3b8;
        }
        
        .react-calendar-heatmap .color-empty {
          fill: #1e293b;
        }
        
        .react-calendar-heatmap .color-scale-1 {
          fill: #9be9a8;
        }
        
        .react-calendar-heatmap .color-scale-2 {
          fill: #40c463;
        }
        
        .react-calendar-heatmap .color-scale-3 {
          fill: #30a14e;
        }
        
        .react-calendar-heatmap .color-scale-4 {
          fill: #216e39;
        }
        
        .react-calendar-heatmap rect:hover {
          stroke: #3b82f6;
          stroke-width: 2px;
          opacity: 0.8;
        }
      `}</style>
    </div>
  );
}
