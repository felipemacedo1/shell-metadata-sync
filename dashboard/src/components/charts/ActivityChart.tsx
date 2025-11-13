'use client';

import { Card, Title, AreaChart as TremorAreaChart } from '@tremor/react';

interface ActivityChartProps {
  data: Array<{
    date: string;
    commits: number;
    prs?: number;
    issues?: number;
  }>;
}

export default function ActivityChart({ data }: ActivityChartProps) {
  // Get last 90 days for better visualization
  const recentData = data.slice(-90);

  return (
    <Card>
      <Title>Atividade de Commits (Últimos 90 dias)</Title>
      <TremorAreaChart
        className="h-72 mt-4"
        data={recentData}
        index="date"
        categories={["commits"]}
        colors={["blue"]}
        showLegend={false}
        showGridLines={true}
        showAnimation={true}
        curveType="monotone"
      />
    </Card>
  );
}
