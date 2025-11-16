import { ProfileData, ActivityData, LanguageData, Repository } from '@/lib/types';
import { readFileSync } from 'fs';
import { join } from 'path';

const basePath = process.env.NODE_ENV === 'production' ? '/shell-metadata-sync' : '';

async function fetchStaticData<T>(filename: string): Promise<T | null> {
  try {
    // During build time, read from file system
    if (typeof window === 'undefined' && process.env.NODE_ENV === 'production') {
      const filePath = join(process.cwd(), 'public', 'data', `${filename}.json`);
      const fileContents = readFileSync(filePath, 'utf8');
      const data = JSON.parse(fileContents);
      console.log(`✅ Loaded ${filename} from filesystem`);
      return data;
    }
    
    // In browser, fetch normally
    const url = `${basePath}/data/${filename}.json`;
    const response = await fetch(url, { 
      cache: 'force-cache',
      headers: { 'Content-Type': 'application/json' }
    });
    
    if (!response.ok) {
      console.error(`Failed to fetch ${filename}: ${response.status}`);
      return null;
    }
    
    const data = await response.json();
    console.log(`✅ Loaded ${filename}:`, Object.keys(data));
    return data;
  } catch (error) {
    console.error(`Error fetching ${filename}:`, error);
    return null;
  }
}

export async function fetchProfile(): Promise<ProfileData | null> {
  return await fetchStaticData<ProfileData>('profile');
}

export async function fetchProfileSecondary(): Promise<ProfileData | null> {
  return await fetchStaticData<ProfileData>('profile-secondary');
}

export async function fetchActivity(): Promise<ActivityData | null> {
  return await fetchStaticData<ActivityData>('activity-daily');
}

export async function fetchActivitySecondary(): Promise<ActivityData | null> {
  return await fetchStaticData<ActivityData>('activity-daily-secondary');
}

export async function fetchLanguages(): Promise<LanguageData | null> {
  return await fetchStaticData<LanguageData>('languages');
}

export async function fetchLanguagesSecondary(): Promise<LanguageData | null> {
  return await fetchStaticData<LanguageData>('languages-secondary');
}

export async function fetchRepositories(): Promise<Repository[]> {
  const data = await fetchStaticData<Repository[]>('projects');
  return data || [];
}

export async function fetchAllData() {
  const results = await Promise.all([
    fetchProfile(),
    fetchProfileSecondary(),
    fetchActivity(),
    fetchActivitySecondary(),
    fetchLanguages(),
    fetchLanguagesSecondary(),
    fetchRepositories(),
  ]);
  
  console.log('📊 Fetched data:', {
    profile: !!results[0],
    profileSecondary: !!results[1],
    activity: !!results[2],
    activitySecondary: !!results[3],
    languages: !!results[4],
    languagesSecondary: !!results[5],
    repositories: results[6]?.length || 0
  });
  
  return results;
}
