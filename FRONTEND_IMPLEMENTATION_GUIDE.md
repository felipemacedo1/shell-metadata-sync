# 🚀 Guia Prático de Implementação - Dashboard Frontend

## Parte 1: Refatoração da Camada de Dados

### 1.1 Separar Fetchers (lib/api/fetchers.ts)

```typescript
// lib/api/fetchers.ts
import { ProfileData, ActivityData, LanguageData, Repository } from '@/lib/types';

const isServer = typeof window === 'undefined';
const basePath = process.env.NODE_ENV === 'production' ? '/dev-metadata-sync' : '';

/**
 * Fetch genérico para dados estáticos
 */
async function fetchStaticData<T>(filename: string): Promise<T | null> {
  try {
    if (isServer) {
      // Server-side: lê do filesystem
      const fs = await import('fs/promises');
      const path = await import('path');
      const filePath = path.join(process.cwd(), 'public', 'data', `${filename}.json`);
      const data = await fs.readFile(filePath, 'utf-8');
      return JSON.parse(data);
    } else {
      // Client-side: fetch HTTP
      const url = `${basePath}/data/${filename}.json`;
      const response = await fetch(url, { 
        cache: 'no-store',
        headers: { 'Content-Type': 'application/json' }
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      return await response.json();
    }
  } catch (error) {
    console.error(`Error fetching ${filename}:`, error);
    throw error; // Propaga erro em vez de retornar null
  }
}

/**
 * Fetch profile data
 */
export async function fetchProfile(): Promise<ProfileData> {
  const data = await fetchStaticData<ProfileData>('profile');
  if (!data) throw new Error('Profile data not found');
  return data;
}

/**
 * Fetch secondary profile data
 */
export async function fetchProfileSecondary(): Promise<ProfileData | null> {
  try {
    return await fetchStaticData<ProfileData>('profile-secondary');
  } catch {
    return null; // Secondary é opcional
  }
}

/**
 * Fetch activity data
 */
export async function fetchActivity(): Promise<ActivityData> {
  const data = await fetchStaticData<ActivityData>('activity-daily');
  if (!data) throw new Error('Activity data not found');
  return data;
}

/**
 * Fetch secondary activity data
 */
export async function fetchActivitySecondary(): Promise<ActivityData | null> {
  try {
    return await fetchStaticData<ActivityData>('activity-daily-secondary');
  } catch {
    return null;
  }
}

/**
 * Fetch language data
 */
export async function fetchLanguages(): Promise<LanguageData> {
  const data = await fetchStaticData<LanguageData>('languages');
  if (!data) throw new Error('Language data not found');
  return data;
}

/**
 * Fetch secondary language data
 */
export async function fetchLanguagesSecondary(): Promise<LanguageData | null> {
  try {
    return await fetchStaticData<LanguageData>('languages-secondary');
  } catch {
    return null;
  }
}

/**
 * Fetch repositories
 */
export async function fetchRepositories(): Promise<Repository[]> {
  const data = await fetchStaticData<Repository[]>('projects');
  return data || [];
}

/**
 * Fetch metadata
 */
export async function fetchMetadata(): Promise<any> {
  return await fetchStaticData('metadata');
}

/**
 * Fetch all data in parallel
 */
export async function fetchAllData() {
  return await Promise.all([
    fetchProfile(),
    fetchProfileSecondary(),
    fetchActivity(),
    fetchActivitySecondary(),
    fetchLanguages(),
    fetchLanguagesSecondary(),
    fetchRepositories(),
    fetchMetadata(),
  ]);
}
```

### 1.2 Criar Aggregators (lib/api/aggregators.ts)

```typescript
// lib/api/aggregators.ts
import { ProfileData, ActivityData, LanguageData, DailyMetric } from '@/lib/types';

/**
 * Agrega dados de perfil de múltiplas contas
 */
export function aggregateProfiles(
  primary: ProfileData,
  secondary: ProfileData | null
): ProfileData {
  if (!secondary) return primary;

  return {
    ...primary,
    // Mantém info do usuário primário
    login: primary.login,
    name: primary.name || 'Felipe Macedo',
    bio: primary.bio,
    avatar_url: primary.avatar_url,
    
    // Agrega estatísticas
    followers: primary.followers + secondary.followers,
    following: primary.following + secondary.following,
    public_repos: primary.public_repos + secondary.public_repos,
    total_stars_received: (primary.total_stars_received || 0) + (secondary.total_stars_received || 0),
    total_forks_received: (primary.total_forks_received || 0) + (secondary.total_forks_received || 0),
    
    // Combina organizações únicas
    organizations: [
      ...new Set([
        ...(primary.organizations || []),
        ...(secondary.organizations || [])
      ])
    ],
    
    generated_at: primary.generated_at,
  };
}

/**
 * Agrega métricas diárias de atividade
 */
export function aggregateActivity(
  primary: ActivityData,
  secondary: ActivityData | null
): ActivityData {
  if (!secondary) return primary;

  const mergedMetrics: Record<string, DailyMetric> = { ...primary.daily_metrics };

  // Merge daily metrics
  Object.entries(secondary.daily_metrics).forEach(([date, metrics]) => {
    if (mergedMetrics[date]) {
      mergedMetrics[date] = {
        commits: mergedMetrics[date].commits + metrics.commits,
        prs: mergedMetrics[date].prs + metrics.prs,
        issues: mergedMetrics[date].issues + metrics.issues,
      };
    } else {
      mergedMetrics[date] = metrics;
    }
  });

  return {
    metadata: primary.metadata,
    daily_metrics: mergedMetrics,
    summary: primary.summary, // Será recalculado depois
  };
}

/**
 * Agrega dados de linguagens
 */
export function aggregateLanguages(
  primary: LanguageData,
  secondary: LanguageData | null
): LanguageData {
  if (!secondary) return primary;

  const mergedLanguages: Record<string, {
    bytes: number;
    repos: number;
    percentage: number;
  }> = {};

  let totalBytes = 0;

  // Combina linguagens
  const allLanguages = [
    ...Object.keys(primary.languages),
    ...Object.keys(secondary.languages)
  ];

  [...new Set(allLanguages)].forEach(lang => {
    const primaryLang = primary.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };
    const secondaryLang = secondary.languages[lang] || { bytes: 0, repos: 0, percentage: 0 };

    mergedLanguages[lang] = {
      bytes: primaryLang.bytes + secondaryLang.bytes,
      repos: primaryLang.repos + secondaryLang.repos,
      percentage: 0, // Será recalculado
    };

    totalBytes += mergedLanguages[lang].bytes;
  });

  // Recalcula percentagens
  Object.keys(mergedLanguages).forEach(lang => {
    mergedLanguages[lang].percentage = (mergedLanguages[lang].bytes / totalBytes) * 100;
  });

  // Ordena e pega top languages
  const topLanguages = Object.entries(mergedLanguages)
    .sort(([, a], [, b]) => b.bytes - a.bytes)
    .slice(0, 10)
    .map(([name]) => name);

  return {
    metadata: {
      user: primary.metadata.user,
      generated_at: primary.metadata.generated_at,
    },
    languages: mergedLanguages,
    top_languages: topLanguages,
  };
}

/**
 * Função principal de agregação
 */
export interface AggregatedData {
  profile: ProfileData;
  activity: ActivityData;
  languages: LanguageData;
  repositories: Repository[];
}

export function aggregateAllData(
  primaryProfile: ProfileData,
  secondaryProfile: ProfileData | null,
  primaryActivity: ActivityData,
  secondaryActivity: ActivityData | null,
  primaryLanguages: LanguageData,
  secondaryLanguages: LanguageData | null,
  repositories: Repository[]
): AggregatedData {
  return {
    profile: aggregateProfiles(primaryProfile, secondaryProfile),
    activity: aggregateActivity(primaryActivity, secondaryActivity),
    languages: aggregateLanguages(primaryLanguages, secondaryLanguages),
    repositories,
  };
}
```

### 1.3 Criar Hook Principal (lib/hooks/useGitHubData.ts)

```typescript
// lib/hooks/useGitHubData.ts
'use client';

import { useState, useEffect } from 'react';
import { fetchAllData } from '@/lib/api/fetchers';
import { aggregateAllData, type AggregatedData } from '@/lib/api/aggregators';

interface UseGitHubDataReturn extends AggregatedData {
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useGitHubData(): UseGitHubDataReturn {
  const [data, setData] = useState<AggregatedData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [
        primaryProfile,
        secondaryProfile,
        primaryActivity,
        secondaryActivity,
        primaryLanguages,
        secondaryLanguages,
        repositories,
      ] = await fetchAllData();

      const aggregated = aggregateAllData(
        primaryProfile,
        secondaryProfile,
        primaryActivity,
        secondaryActivity,
        primaryLanguages,
        secondaryLanguages,
        repositories
      );

      setData(aggregated);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
      console.error('Error loading GitHub data:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  return {
    profile: data?.profile || null,
    activity: data?.activity || null,
    languages: data?.languages || null,
    repositories: data?.repositories || [],
    isLoading,
    error,
    refetch: loadData,
  };
}
```

---

## Parte 2: Componentes UI Base

### 2.1 Card Component (components/ui/Card.tsx)

```typescript
// components/ui/Card.tsx
import { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface CardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  gradient?: boolean;
  onClick?: () => void;
}

export function Card({ 
  children, 
  className, 
  hover = false,
  gradient = false,
  onClick 
}: CardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-xl border border-slate-700 p-6',
        'bg-slate-800/50 backdrop-blur-sm',
        hover && 'hover:border-slate-600 hover:scale-105 transition-all cursor-pointer',
        gradient && 'relative overflow-hidden',
        className
      )}
    >
      {gradient && (
        <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10 opacity-0 group-hover:opacity-100 transition-opacity" />
      )}
      <div className="relative">{children}</div>
    </div>
  );
}

export function CardHeader({ 
  children, 
  className 
}: { 
  children: ReactNode; 
  className?: string; 
}) {
  return (
    <div className={cn('mb-4', className)}>
      {children}
    </div>
  );
}

export function CardTitle({ 
  children, 
  icon: Icon,
  className 
}: { 
  children: ReactNode; 
  icon?: React.ComponentType<{ className?: string }>;
  className?: string; 
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      {Icon && <Icon className="w-6 h-6 text-blue-400" />}
      <h2 className="text-2xl font-bold text-white">{children}</h2>
    </div>
  );
}

export function CardContent({ 
  children, 
  className 
}: { 
  children: ReactNode; 
  className?: string; 
}) {
  return (
    <div className={cn('', className)}>
      {children}
    </div>
  );
}
```

### 2.2 Skeleton Component (components/ui/Skeleton.tsx)

```typescript
// components/ui/Skeleton.tsx
import { cn } from '@/lib/utils';

interface SkeletonProps {
  className?: string;
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
}

export function Skeleton({ 
  className,
  variant = 'rectangular',
  width,
  height 
}: SkeletonProps) {
  return (
    <div
      className={cn(
        'animate-pulse bg-slate-700/50',
        variant === 'circular' && 'rounded-full',
        variant === 'text' && 'rounded h-4',
        variant === 'rectangular' && 'rounded-lg',
        className
      )}
      style={{ width, height }}
    />
  );
}

// Skeletons específicos para componentes
export function HeroSkeleton() {
  return (
    <Card className="mb-8">
      <div className="flex flex-col md:flex-row items-center gap-8">
        <Skeleton variant="circular" width={128} height={128} />
        <div className="flex-1 space-y-4 w-full">
          <Skeleton variant="text" width="60%" height={40} />
          <Skeleton variant="text" width="40%" height={20} />
          <Skeleton variant="text" width="80%" height={20} />
          <div className="flex gap-4">
            <Skeleton variant="text" width={100} height={20} />
            <Skeleton variant="text" width={100} height={20} />
            <Skeleton variant="text" width={100} height={20} />
          </div>
        </div>
      </div>
    </Card>
  );
}

export function StatsGridSkeleton() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
      {[...Array(6)].map((_, i) => (
        <Card key={i}>
          <Skeleton variant="circular" width={40} height={40} className="mb-4" />
          <Skeleton variant="text" width="60%" height={32} className="mb-2" />
          <Skeleton variant="text" width="40%" height={16} />
        </Card>
      ))}
    </div>
  );
}

export function ChartSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton variant="text" width="40%" height={24} />
      </CardHeader>
      <Skeleton variant="rectangular" width="100%" height={300} />
    </Card>
  );
}
```

### 2.3 Error Boundary (components/layout/ErrorBoundary.tsx)

```typescript
// components/layout/ErrorBoundary.tsx
'use client';

import { Component, ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Card } from '@/components/ui/Card';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.props.onError?.(error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <Card className="max-w-2xl mx-auto my-8">
          <div className="text-center">
            <AlertTriangle className="w-16 h-16 text-red-500 mx-auto mb-4" />
            <h2 className="text-2xl font-bold text-white mb-2">
              Oops! Something went wrong
            </h2>
            <p className="text-slate-400 mb-6">
              {this.state.error?.message || 'An unexpected error occurred'}
            </p>
            <button
              onClick={this.handleReset}
              className="flex items-center gap-2 mx-auto px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition"
            >
              <RefreshCw className="w-4 h-4" />
              Reload Page
            </button>
          </div>
        </Card>
      );
    }

    return this.props.children;
  }
}

// Hook para usar error state
export function useError() {
  const [error, setError] = useState<Error | null>(null);

  const clearError = () => setError(null);

  return { error, setError, clearError };
}
```

---

## Parte 3: Utilities

### 3.1 Date Utils (lib/utils/dates.ts)

```typescript
// lib/utils/dates.ts

/**
 * Retorna tempo relativo (ex: "2 days ago")
 */
export function getRelativeTime(dateString: string): string {
  const now = new Date();
  const date = new Date(dateString);
  const diffInMs = now.getTime() - date.getTime();
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

  if (diffInDays === 0) return 'Today';
  if (diffInDays === 1) return 'Yesterday';
  if (diffInDays < 7) return `${diffInDays} days ago`;
  if (diffInDays < 30) {
    const weeks = Math.floor(diffInDays / 7);
    return `${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`;
  }
  if (diffInDays < 365) {
    const months = Math.floor(diffInDays / 30);
    return `${months} ${months === 1 ? 'month' : 'months'} ago`;
  }
  const years = Math.floor(diffInDays / 365);
  return `${years} ${years === 1 ? 'year' : 'years'} ago`;
}

/**
 * Formata data
 */
export function formatDate(
  dateString: string,
  options?: Intl.DateTimeFormatOptions
): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', options || {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

/**
 * Diferença em dias entre duas datas
 */
export function getDaysDifference(startDate: string, endDate: string): number {
  const start = new Date(startDate);
  const end = new Date(endDate);
  return Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
}

/**
 * Agrupa datas por semana
 */
export function groupByWeek(dates: string[]): Record<string, string[]> {
  const groups: Record<string, string[]> = {};

  dates.forEach(date => {
    const d = new Date(date);
    // Pega segunda-feira da semana
    const monday = new Date(d);
    monday.setDate(d.getDate() - d.getDay() + 1);
    const weekKey = monday.toISOString().split('T')[0];

    if (!groups[weekKey]) {
      groups[weekKey] = [];
    }
    groups[weekKey].push(date);
  });

  return groups;
}

/**
 * Agrupa datas por mês
 */
export function groupByMonth(dates: string[]): Record<string, string[]> {
  const groups: Record<string, string[]> = {};

  dates.forEach(date => {
    const d = new Date(date);
    const monthKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;

    if (!groups[monthKey]) {
      groups[monthKey] = [];
    }
    groups[monthKey].push(date);
  });

  return groups;
}
```

### 3.2 Format Utils (lib/utils/format.ts)

```typescript
// lib/utils/format.ts

/**
 * Formata número com separadores
 */
export function formatNumber(num: number): string {
  return new Intl.NumberFormat('en-US').format(num);
}

/**
 * Formata bytes para formato legível
 */
export function formatBytes(bytes: number, decimals = 1): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(decimals))} ${sizes[i]}`;
}

/**
 * Formata percentagem
 */
export function formatPercentage(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Trunca texto
 */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return `${text.substring(0, maxLength)}...`;
}

/**
 * Formata número compacto (1.2K, 3.4M)
 */
export function formatCompact(num: number): string {
  if (num < 1000) return num.toString();
  if (num < 1000000) return `${(num / 1000).toFixed(1)}K`;
  return `${(num / 1000000).toFixed(1)}M`;
}
```

### 3.3 Color Utils (lib/utils/colors.ts)

```typescript
// lib/utils/colors.ts

/**
 * Mapa de cores de linguagens de programação
 * Baseado em: https://github.com/ozh/github-colors
 */
export const LANGUAGE_COLORS: Record<string, string> = {
  'JavaScript': '#f1e05a',
  'TypeScript': '#3178c6',
  'Python': '#3572A5',
  'Go': '#00ADD8',
  'Rust': '#dea584',
  'Java': '#b07219',
  'C++': '#f34b7d',
  'C': '#555555',
  'C#': '#178600',
  'Ruby': '#701516',
  'PHP': '#4F5D95',
  'Swift': '#F05138',
  'Kotlin': '#A97BFF',
  'Dart': '#00B4AB',
  'CSS': '#563d7c',
  'HTML': '#e34c26',
  'Shell': '#89e051',
  'Vim Script': '#199f4b',
  'Dockerfile': '#384d54',
  'Makefile': '#427819',
  'CMake': '#DA3434',
};

/**
 * Cores para gráficos
 */
export const CHART_COLORS = [
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

/**
 * Retorna cor para uma linguagem
 */
export function getLanguageColor(language: string): string {
  return LANGUAGE_COLORS[language] || '#6b7280'; // gray default
}

/**
 * Gera escala de cores para heatmap
 */
export function getHeatmapColor(value: number, max: number): string {
  if (value === 0) return '#1e293b'; // empty
  
  const percentage = value / max;
  
  if (percentage < 0.25) return '#9be9a8'; // light green
  if (percentage < 0.50) return '#40c463'; // medium green
  if (percentage < 0.75) return '#30a14e'; // dark green
  return '#216e39'; // darkest green
}

/**
 * Converte hex para rgba
 */
export function hexToRgba(hex: string, alpha: number): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}
```

---

## Parte 4: Refatoração de Componentes Existentes

### 4.1 Refatorar StatsGrid (components/dashboard/StatsGrid/index.tsx)

```typescript
// components/dashboard/StatsGrid/index.tsx
'use client';

import { StatCard } from './StatCard';
import { GitCommit, GitPullRequest, Flame, Calendar, TrendingUp, FolderGit2 } from 'lucide-react';

interface StatsGridProps {
  totalCommits: number;
  totalPRs: number;
  totalIssues: number;
  activeDays: number;
  totalRepos: number;
  periodDays: number;
}

export function StatsGrid({
  totalCommits,
  totalPRs,
  totalIssues,
  activeDays,
  totalRepos,
  periodDays = 90
}: StatsGridProps) {
  const avgCommitsPerWeek = ((totalCommits / periodDays) * 7).toFixed(1);
  const activityRate = ((activeDays / periodDays) * 100).toFixed(0);

  const stats = [
    {
      icon: GitCommit,
      label: 'Total Commits',
      value: totalCommits,
      color: 'emerald',
      suffix: '',
    },
    {
      icon: GitPullRequest,
      label: 'Pull Requests',
      value: totalPRs,
      color: 'blue',
      suffix: '',
    },
    {
      icon: Flame,
      label: 'Active Days',
      value: activeDays,
      color: 'orange',
      suffix: `/${periodDays}`,
    },
    {
      icon: FolderGit2,
      label: 'Repositories',
      value: totalRepos,
      color: 'purple',
      suffix: '',
    },
    {
      icon: TrendingUp,
      label: 'Avg. Commits/Week',
      value: parseFloat(avgCommitsPerWeek),
      color: 'cyan',
      suffix: '',
    },
    {
      icon: Calendar,
      label: 'Activity Rate',
      value: parseFloat(activityRate),
      color: 'amber',
      suffix: '%',
    }
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
      {stats.map((stat, index) => (
        <StatCard key={index} {...stat} />
      ))}
    </div>
  );
}
```

```typescript
// components/dashboard/StatsGrid/StatCard.tsx
'use client';

import { LucideIcon } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { AnimatedCounter } from '@/components/ui/AnimatedCounter';

interface StatCardProps {
  icon: LucideIcon;
  label: string;
  value: number;
  color: 'emerald' | 'blue' | 'orange' | 'purple' | 'cyan' | 'amber';
  suffix?: string;
}

const colorConfig = {
  emerald: {
    gradient: 'from-emerald-500 to-green-600',
    bg: 'bg-emerald-500/10',
    icon: 'text-emerald-400',
  },
  blue: {
    gradient: 'from-blue-500 to-indigo-600',
    bg: 'bg-blue-500/10',
    icon: 'text-blue-400',
  },
  orange: {
    gradient: 'from-orange-500 to-red-600',
    bg: 'bg-orange-500/10',
    icon: 'text-orange-400',
  },
  purple: {
    gradient: 'from-purple-500 to-pink-600',
    bg: 'bg-purple-500/10',
    icon: 'text-purple-400',
  },
  cyan: {
    gradient: 'from-cyan-500 to-blue-600',
    bg: 'bg-cyan-500/10',
    icon: 'text-cyan-400',
  },
  amber: {
    gradient: 'from-amber-500 to-yellow-600',
    bg: 'bg-amber-500/10',
    icon: 'text-amber-400',
  },
};

export function StatCard({ icon: Icon, label, value, color, suffix = '' }: StatCardProps) {
  const colors = colorConfig[color];

  return (
    <Card hover gradient className="group">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <div className={`p-2 rounded-lg ${colors.bg}`}>
              <Icon className={`w-5 h-5 ${colors.icon}`} />
            </div>
          </div>
          <p className="text-3xl font-bold text-white mb-1">
            <AnimatedCounter value={value} />
            {suffix && <span className="text-2xl text-slate-400">{suffix}</span>}
          </p>
          <p className="text-sm text-slate-400">{label}</p>
        </div>
      </div>

      {/* Background gradient on hover */}
      <div className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-br ${colors.gradient} opacity-0 group-hover:opacity-10 blur-2xl transition-opacity`} />
    </Card>
  );
}
```

```typescript
// components/ui/AnimatedCounter.tsx
'use client';

import { useEffect, useState } from 'react';

interface AnimatedCounterProps {
  value: number;
  duration?: number;
}

export function AnimatedCounter({ value, duration = 1000 }: AnimatedCounterProps) {
  const [count, setCount] = useState(0);

  useEffect(() => {
    let startTime: number;
    let animationFrame: number;

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp;
      const progress = timestamp - startTime;
      const percentage = Math.min(progress / duration, 1);

      // Easing function (easeOutExpo)
      const eased = percentage === 1 ? 1 : 1 - Math.pow(2, -10 * percentage);

      setCount(Math.floor(eased * value));

      if (percentage < 1) {
        animationFrame = requestAnimationFrame(animate);
      }
    };

    animationFrame = requestAnimationFrame(animate);

    return () => cancelAnimationFrame(animationFrame);
  }, [value, duration]);

  return <>{count.toLocaleString()}</>;
}
```

---

## Parte 5: Exemplo Completo - Page.tsx Refatorado

```typescript
// app/page.tsx
import { fetchAllData } from '@/lib/api/fetchers';
import { aggregateAllData } from '@/lib/api/aggregators';
import { ErrorBoundary } from '@/components/layout/ErrorBoundary';
import { Hero } from '@/components/dashboard/Hero';
import { StatsGrid } from '@/components/dashboard/StatsGrid';
import { ContributionHeatmap } from '@/components/charts/ContributionHeatmap';
import { ActivityTimeline } from '@/components/dashboard/ActivityTimeline';
import { LanguageChart } from '@/components/dashboard/LanguageChart';
import { RepositoryGrid } from '@/components/dashboard/RepositoryGrid';
import { calculateActivityStats } from '@/lib/utils/calculations';

export default async function Home() {
  try {
    // Fetch all data
    const [
      primaryProfile,
      secondaryProfile,
      primaryActivity,
      secondaryActivity,
      primaryLanguages,
      secondaryLanguages,
      repositories,
    ] = await fetchAllData();

    // Aggregate data
    const {
      profile,
      activity,
      languages,
      repositories: repos
    } = aggregateAllData(
      primaryProfile,
      secondaryProfile,
      primaryActivity,
      secondaryActivity,
      primaryLanguages,
      secondaryLanguages,
      repositories
    );

    // Calculate stats
    const stats = calculateActivityStats(activity);

    // Prepare heatmap data
    const heatmapData = Object.entries(activity.daily_metrics).map(([date, metrics]) => ({
      date,
      count: metrics.commits + metrics.prs + metrics.issues
    }));

    return (
      <ErrorBoundary>
        <main className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            
            {/* Hero Section */}
            <Hero profile={profile} />

            {/* Stats Grid */}
            <StatsGrid {...stats} totalRepos={repos.length} />

            {/* Contribution Heatmap */}
            <div className="mb-8">
              <ContributionHeatmap
                data={heatmapData}
                startDate={new Date(activity.metadata.start_date)}
                endDate={new Date(activity.metadata.end_date)}
              />
            </div>

            {/* Activity Timeline & Language Distribution */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
              <ActivityTimeline dailyMetrics={activity.daily_metrics} />
              <LanguageChart languages={languages.languages} />
            </div>

            {/* Repository Grid */}
            <div className="mb-8">
              <div className="mb-6">
                <h2 className="text-3xl font-bold text-white mb-2">Repositories</h2>
                <p className="text-slate-400">
                  Browse and filter through all {repos.length} repositories
                </p>
              </div>
              <RepositoryGrid repositories={repos} />
            </div>

            {/* Footer */}
            <footer className="mt-12 pt-8 border-t border-slate-800 text-center">
              <div className="space-y-2">
                <p className="text-slate-400 text-sm">
                  🚀 Built with Next.js, TypeScript, and Tailwind CSS
                </p>
                <p className="text-slate-500 text-xs">
                  Data aggregated from multiple GitHub accounts
                </p>
                <p className="text-slate-600 text-xs">
                  Last sync: {new Date(activity.metadata.generated_at).toLocaleString()}
                </p>
              </div>
            </footer>
          </div>
        </main>
      </ErrorBoundary>
    );
  } catch (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-white mb-4">Failed to load data</h1>
          <p className="text-slate-400">{error instanceof Error ? error.message : 'Unknown error'}</p>
        </div>
      </div>
    );
  }
}
```

---

## 🎯 Próximos Passos

1. **Criar utility function para cn()** (classnames)
2. **Adicionar testes unitários** para utilities
3. **Implementar paginação** no RepositoryGrid
4. **Adicionar filtro de período** no ActivityTimeline
5. **Melhorar responsividade** do Heatmap
6. **Adicionar temas** (dark/light mode)
7. **Performance:** Implementar virtual scrolling para repos

Quer que eu implemente alguma parte específica?
