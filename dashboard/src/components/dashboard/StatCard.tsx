'use client';

import { LucideIcon } from 'lucide-react';
import { AnimatedCounter } from '@/components/ui/AnimatedCounter';

interface StatCardProps {
  icon: LucideIcon;
  label: string;
  value: number | string;
  color: 'emerald' | 'blue' | 'orange' | 'purple' | 'cyan' | 'amber';
}

const colorConfig = {
  emerald: { gradient: 'from-emerald-500 to-green-600', bg: 'bg-emerald-500/10', icon: 'text-emerald-400' },
  blue: { gradient: 'from-blue-500 to-indigo-600', bg: 'bg-blue-500/10', icon: 'text-blue-400' },
  orange: { gradient: 'from-orange-500 to-red-600', bg: 'bg-orange-500/10', icon: 'text-orange-400' },
  purple: { gradient: 'from-purple-500 to-pink-600', bg: 'bg-purple-500/10', icon: 'text-purple-400' },
  cyan: { gradient: 'from-cyan-500 to-blue-600', bg: 'bg-cyan-500/10', icon: 'text-cyan-400' },
  amber: { gradient: 'from-amber-500 to-yellow-600', bg: 'bg-amber-500/10', icon: 'text-amber-400' },
};

export function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
  const colors = colorConfig[color];
  const numValue = typeof value === 'string' ? parseFloat(value) || 0 : value;

  return (
    <div className="group relative overflow-hidden rounded-xl bg-slate-800/50 backdrop-blur-sm border border-slate-700 p-6 hover:border-slate-600 transition-all hover:scale-105">
      <div className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-br ${colors.gradient} opacity-0 group-hover:opacity-10 blur-2xl transition-opacity`} />
      <div className="relative flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <div className={`p-2 rounded-lg ${colors.bg}`}>
              <Icon className={`w-5 h-5 ${colors.icon}`} />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">
            <AnimatedCounter value={numValue} />
          </p>
          <p className="text-sm text-slate-400">{label}</p>
        </div>
      </div>
    </div>
  );
}
