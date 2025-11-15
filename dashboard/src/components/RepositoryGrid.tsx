'use client';

import { useState, useMemo } from 'react';
import { ExternalLink, Star, GitFork, AlertCircle, Search, Filter, Calendar } from 'lucide-react';
import { Repository } from '@/lib/types';

interface RepositoryGridProps {
  repositories: Repository[];
}

type SortOption = 'updated' | 'name' | 'stars' | 'language';

export default function RepositoryGrid({ repositories }: RepositoryGridProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLanguage, setSelectedLanguage] = useState<string>('all');
  const [sortBy, setSortBy] = useState<SortOption>('updated');

  // Extract unique languages
  const languages = useMemo(() => {
    const langs = new Set<string>();
    repositories.forEach(repo => {
      if (repo.language) langs.add(repo.language);
    });
    return Array.from(langs).sort();
  }, [repositories]);

  // Filter and sort repositories
  const filteredRepos = useMemo(() => {
    let filtered = repositories;

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(repo =>
        repo.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (repo.description && repo.description.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    // Language filter
    if (selectedLanguage !== 'all') {
      filtered = filtered.filter(repo => repo.language === selectedLanguage);
    }

    // Sort
    const sorted = [...filtered].sort((a, b) => {
      switch (sortBy) {
        case 'updated':
          return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
        case 'name':
          return a.name.localeCompare(b.name);
        case 'stars':
          return (b.stargazers_count || 0) - (a.stargazers_count || 0);
        case 'language':
          return (a.language || '').localeCompare(b.language || '');
        default:
          return 0;
      }
    });

    return sorted;
  }, [repositories, searchTerm, selectedLanguage, sortBy]);

  const getRelativeTime = (date: string) => {
    const now = new Date();
    const updated = new Date(date);
    const diffInDays = Math.floor((now.getTime() - updated.getTime()) / (1000 * 60 * 60 * 24));
    
    if (diffInDays === 0) return 'Today';
    if (diffInDays === 1) return 'Yesterday';
    if (diffInDays < 7) return `${diffInDays} days ago`;
    if (diffInDays < 30) return `${Math.floor(diffInDays / 7)} weeks ago`;
    if (diffInDays < 365) return `${Math.floor(diffInDays / 30)} months ago`;
    return `${Math.floor(diffInDays / 365)} years ago`;
  };

  return (
    <div className="space-y-6">
      {/* Filters Section */}
      <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6">
        <div className="flex flex-col lg:flex-row gap-4">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              type="text"
              placeholder="Search repositories..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-slate-900/50 border border-slate-700 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-blue-500 transition"
            />
          </div>

          {/* Language Filter */}
          <div className="relative">
            <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
            <select
              value={selectedLanguage}
              onChange={(e) => setSelectedLanguage(e.target.value)}
              className="pl-10 pr-8 py-2 bg-slate-900/50 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-blue-500 transition appearance-none cursor-pointer min-w-[180px]"
            >
              <option value="all">All Languages</option>
              {languages.map(lang => (
                <option key={lang} value={lang}>{lang}</option>
              ))}
            </select>
          </div>

          {/* Sort By */}
          <div className="relative">
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as SortOption)}
              className="px-4 py-2 bg-slate-900/50 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-blue-500 transition appearance-none cursor-pointer min-w-[150px]"
            >
              <option value="updated">Recently Updated</option>
              <option value="name">Name</option>
              <option value="stars">Most Stars</option>
              <option value="language">Language</option>
            </select>
          </div>
        </div>

        {/* Active filters info */}
        <div className="mt-4 flex items-center gap-2 text-sm text-slate-400">
          <span>Showing {filteredRepos.length} of {repositories.length} repositories</span>
        </div>
      </div>

      {/* Repository Grid */}
      {filteredRepos.length === 0 ? (
        <div className="text-center py-12 bg-slate-800/30 rounded-xl border border-slate-700">
          <Search className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-slate-400">No repositories found matching your criteria</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredRepos.map((repo) => (
            <a
              key={`${repo.owner}/${repo.name}`}
              href={repo.url}
              target="_blank"
              rel="noopener noreferrer"
              className="group block bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-xl p-6 hover:border-blue-500 transition-all hover:scale-105"
            >
              <div className="flex items-start justify-between mb-3">
                <h3 className="text-lg font-semibold text-white group-hover:text-blue-400 transition truncate flex-1">
                  {repo.name}
                </h3>
                <ExternalLink className="w-5 h-5 text-slate-400 group-hover:text-blue-400 transition shrink-0 ml-2" />
              </div>

              {repo.description && (
                <p className="text-slate-400 text-sm mb-4 line-clamp-2 min-h-[40px]">
                  {repo.description}
                </p>
              )}

              <div className="flex items-center gap-4 text-sm text-slate-400 mb-4">
                {repo.language && (
                  <div className="flex items-center gap-1.5">
                    <span className="w-3 h-3 rounded-full bg-blue-500"></span>
                    <span>{repo.language}</span>
                  </div>
                )}
                {repo.stargazers_count !== undefined && repo.stargazers_count > 0 && (
                  <div className="flex items-center gap-1">
                    <Star className="w-4 h-4 fill-yellow-500 text-yellow-500" />
                    <span>{repo.stargazers_count}</span>
                  </div>
                )}
                {repo.forks_count !== undefined && repo.forks_count > 0 && (
                  <div className="flex items-center gap-1">
                    <GitFork className="w-4 h-4" />
                    <span>{repo.forks_count}</span>
                  </div>
                )}
              </div>

              <div className="flex items-center gap-2 text-xs text-slate-500 pt-4 border-t border-slate-700/50">
                <Calendar className="w-3.5 h-3.5" />
                <span>Updated {getRelativeTime(repo.updated_at)}</span>
              </div>
            </a>
          ))}
        </div>
      )}
    </div>
  );
}
