import { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface CardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  gradient?: boolean;
  onClick?: () => void;
}

export function Card({ children, className, hover = false, gradient = false, onClick }: CardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-xl border border-slate-700 p-6 bg-slate-800/50 backdrop-blur-sm',
        hover && 'hover:border-slate-600 hover:scale-105 transition-all cursor-pointer',
        gradient && 'relative overflow-hidden group',
        className
      )}
    >
      {gradient && <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10 opacity-0 group-hover:opacity-100 transition-opacity" />}
      <div className="relative">{children}</div>
    </div>
  );
}
