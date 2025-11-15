import { ProfileData, ActivityData, LanguageData, Repository } from '@/lib/types';

const isServer = typeof window === 'undefined';
const basePath = process.env.NODE_ENV === 'production' ? '/dev-metadata-sync' : '';

async function fetchStaticData<T>(filename: string): Promise<T | null> {
  try {
    if (isServer) {
      const fs = await import('fs/promises');
      const path = await import('path');
      const filePath = path.join(process.cwd(), 'public', 'data', `${filename}.json`);
      const data = await fs.readFile(filePath, 'utf-8');
      return JSON.parse(data);
    } else {
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
    throw error;
  }
}

export async function fetchProfile(): Promise<ProfileData> {
  const data = await fetchStaticData<ProfileData>('profile');
  if (!data) throw new Error('Profile data not found');
  return data;
}

export async function fetchProfileSecondary(): Promise<ProfileData | null> {
  try {
    return await fetchStaticData<ProfileData>('profile-secondary');
  } catch {
    return null;
  }
}

export async function fetchActivity(): Promise<ActivityData> {
  const data = await fetchStaticData<ActivityData>('activity-daily');
  if (!data) throw new Error('Activity data not found');
  return data;
}

export async function fetchActivitySecondary(): Promise<ActivityData | null> {
  try {
    return await fetchStaticData<ActivityData>('activity-daily-secondary');
  } catch {
    return null;
  }
}

export async function fetchLanguages(): Promise<LanguageData> {
  const data = await fetchStaticData<LanguageData>('languages');
  if (!data) throw new Error('Language data not found');
  return data;
}

export async function fetchLanguagesSecondary(): Promise<LanguageData | null> {
  try {
    return await fetchStaticData<LanguageData>('languages-secondary');
  } catch {
    return null;
  }
}

export async function fetchRepositories(): Promise<Repository[]> {
  const data = await fetchStaticData<Repository[]>('projects');
  return data || [];
}

export async function fetchAllData() {
  return await Promise.all([
    fetchProfile(),
    fetchProfileSecondary(),
    fetchActivity(),
    fetchActivitySecondary(),
    fetchLanguages(),
    fetchLanguagesSecondary(),
    fetchRepositories(),
  ]);
}
