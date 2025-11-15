'use client';

import { Github, MapPin, Users, Star } from 'lucide-react';
import { ProfileData } from '@/lib/types';

interface HeroProps {
  profile: ProfileData | null;
}

export default function Hero({ profile }: HeroProps) {
  if (!profile) {
    return (
      <div className="text-center py-20">
        <div className="w-32 h-32 bg-slate-700 rounded-full mx-auto animate-pulse"></div>
      </div>
    );
  }

  return (
    <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-slate-800 via-slate-900 to-slate-800 border border-slate-700 p-8 mb-8">
      {/* Background decoration */}
      <div className="absolute inset-0 bg-grid-white/[0.02] bg-[size:50px_50px]"></div>
      <div className="absolute top-0 right-0 -mt-4 -mr-4 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl"></div>
      <div className="absolute bottom-0 left-0 -mb-4 -ml-4 w-72 h-72 bg-emerald-500/10 rounded-full blur-3xl"></div>
      
      <div className="relative flex flex-col md:flex-row items-center gap-8">
        {/* Avatar */}
        <div className="relative group">
          <div className="absolute -inset-1 bg-gradient-to-r from-blue-500 to-emerald-500 rounded-full blur opacity-75 group-hover:opacity-100 transition"></div>
          <img
            src={profile.avatar_url}
            alt={profile.name || profile.login}
            className="relative w-32 h-32 rounded-full border-4 border-slate-800 object-cover"
          />
        </div>

        {/* Profile Info */}
        <div className="flex-1 text-center md:text-left">
          <h1 className="text-4xl md:text-5xl font-bold text-white mb-2">
            {profile.name || profile.login}
          </h1>
          <div className="flex items-center justify-center md:justify-start gap-2 text-slate-400 mb-4">
            <Github className="w-5 h-5" />
            <a
              href={`https://github.com/${profile.login}`}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-blue-400 transition"
            >
              @{profile.login}
            </a>
          </div>
          
          {profile.bio && (
            <p className="text-lg text-slate-300 mb-4 max-w-2xl">
              {profile.bio}
            </p>
          )}

          {/* Quick Stats */}
          <div className="flex flex-wrap gap-6 justify-center md:justify-start text-sm">
            <div className="flex items-center gap-2">
              <Users className="w-4 h-4 text-slate-400" />
              <span className="text-white font-semibold">{profile.followers}</span>
              <span className="text-slate-400">followers</span>
            </div>
            <div className="flex items-center gap-2">
              <Users className="w-4 h-4 text-slate-400" />
              <span className="text-white font-semibold">{profile.following}</span>
              <span className="text-slate-400">following</span>
            </div>
            <div className="flex items-center gap-2">
              <Github className="w-4 h-4 text-slate-400" />
              <span className="text-white font-semibold">{profile.public_repos}</span>
              <span className="text-slate-400">repositories</span>
            </div>
            {profile.total_stars_received !== undefined && profile.total_stars_received > 0 && (
              <div className="flex items-center gap-2">
                <Star className="w-4 h-4 text-yellow-500" />
                <span className="text-white font-semibold">{profile.total_stars_received}</span>
                <span className="text-slate-400">stars</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Last updated */}
      <div className="relative mt-6 pt-6 border-t border-slate-700/50 text-center md:text-right">
        <p className="text-xs text-slate-500">
          Last updated: {new Date(profile.generated_at).toLocaleString('en-US', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          })}
        </p>
      </div>
    </div>
  );
}
