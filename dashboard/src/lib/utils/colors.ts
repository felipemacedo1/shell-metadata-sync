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

export const CHART_COLORS = [
  '#3b82f6', '#8b5cf6', '#10b981', '#f59e0b',
  '#ef4444', '#06b6d4', '#ec4899', '#14b8a6',
  '#f97316', '#6366f1'
];

export function getLanguageColor(language: string): string {
  return LANGUAGE_COLORS[language] || '#6b7280';
}

export function getHeatmapColor(value: number, max: number): string {
  if (value === 0) return '#1e293b';
  const percentage = value / max;
  if (percentage < 0.25) return '#9be9a8';
  if (percentage < 0.50) return '#40c463';
  if (percentage < 0.75) return '#30a14e';
  return '#216e39';
}

export function hexToRgba(hex: string, alpha: number): string {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}
