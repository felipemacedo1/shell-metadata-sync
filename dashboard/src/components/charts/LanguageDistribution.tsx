'use client';

import { Card, Title, DonutChart } from '@tremor/react';

interface LanguageDistributionProps {
  data: Array<{
    name: string;
    value: number;
    percentage: number;
  }>;
}

export default function LanguageDistribution({ data }: LanguageDistributionProps) {
  return (
    <Card>
      <Title>Distribuição de Linguagens</Title>
      <DonutChart
        className="h-72 mt-4"
        data={data}
        category="value"
        index="name"
        colors={["slate", "violet", "indigo", "rose", "cyan", "amber", "emerald", "blue"]}
        showAnimation={true}
        valueFormatter={(value: number) => `${(value / 1000).toFixed(1)}KB`}
      />
    </Card>
  );
}
