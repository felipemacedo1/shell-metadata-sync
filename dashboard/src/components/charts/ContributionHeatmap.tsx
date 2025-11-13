'use client';

import CalendarHeatmap from 'react-calendar-heatmap';
import 'react-calendar-heatmap/dist/styles.css';

interface HeatmapData {
  date: string;
  count: number;
}

interface ContributionHeatmapProps {
  data: HeatmapData[];
}

export default function ContributionHeatmap({ data }: ContributionHeatmapProps) {
  const today = new Date();
  const oneYearAgo = new Date(today.getFullYear() - 1, today.getMonth(), today.getDate());

  const getColorClass = (count: number | undefined) => {
    if (!count || count === 0) return 'color-empty';
    if (count < 3) return 'color-scale-1';
    if (count < 6) return 'color-scale-2';
    if (count < 9) return 'color-scale-3';
    return 'color-scale-4';
  };

  return (
    <div className="w-full overflow-x-auto p-4 bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700">
      <h3 className="text-lg font-semibold mb-4 text-slate-900 dark:text-white">
        Contribution Heatmap (Últimos 365 dias)
      </h3>
      <div className="min-w-[800px]">
        <CalendarHeatmap
          startDate={oneYearAgo}
          endDate={today}
          values={data}
          classForValue={(value) => getColorClass(value?.count)}
          showWeekdayLabels={true}
        />
      </div>
      <style jsx global>{`
        .react-calendar-heatmap {
          width: 100%;
        }
        
        .react-calendar-heatmap text {
          font-size: 10px;
          fill: #64748b;
        }
        
        .react-calendar-heatmap .color-empty {
          fill: #ebedf0;
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
          stroke: #555;
          stroke-width: 1px;
        }

        @media (prefers-color-scheme: dark) {
          .react-calendar-heatmap .color-empty {
            fill: #161b22;
          }
          
          .react-calendar-heatmap text {
            fill: #8b949e;
          }
        }
      `}</style>
    </div>
  );
}
