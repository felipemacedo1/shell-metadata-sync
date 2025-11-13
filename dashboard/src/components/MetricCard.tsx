'use client';

import { Card, Title, Metric } from '@tremor/react';
import { LucideIcon } from 'lucide-react';

interface MetricCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  color?: 'blue' | 'violet' | 'emerald' | 'amber' | 'rose';
  trend?: {
    value: number;
    label: string;
  };
}

const colorClasses = {
  blue: 'text-blue-500',
  violet: 'text-violet-500',
  emerald: 'text-emerald-500',
  amber: 'text-amber-500',
  rose: 'text-rose-500',
};

export default function MetricCard({ 
  title, 
  value, 
  icon: Icon, 
  color = 'blue',
  trend 
}: MetricCardProps) {
  return (
    <Card decoration="top" decorationColor={color}>
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Icon className={`h-5 w-5 ${colorClasses[color]}`} />
            <Title>{title}</Title>
          </div>
          <Metric>{value.toLocaleString()}</Metric>
          {trend && (
            <p className="text-sm text-gray-500 mt-1">
              <span className={trend.value >= 0 ? 'text-emerald-500' : 'text-rose-500'}>
                {trend.value >= 0 ? '↑' : '↓'} {Math.abs(trend.value)}%
              </span>
              {' '}{trend.label}
            </p>
          )}
        </div>
      </div>
    </Card>
  );
}
