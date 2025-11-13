import { fetchRepositories, fetchActivity, fetchLanguages, fetchProfile } from '@/lib/api';

export default async function Home() {
  const repos = await fetchRepositories();
  const activityData = await fetchActivity();
  const languageData = await fetchLanguages();
  const profileData = await fetchProfile();

  const totalCommits = activityData 
    ? Object.values(activityData.daily_metrics).reduce((sum, day) => sum + day.commits, 0)
    : 0;

  const totalPRs = activityData
    ? Object.values(activityData.daily_metrics).reduce((sum, day) => sum + day.prs, 0)
    : 0;

  const activeDays = activityData
    ? Object.values(activityData.daily_metrics).filter(day => day.commits > 0).length
    : 0;

  const topLanguages = languageData?.top_languages.slice(0, 5) || [];

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">
            GitHub Analytics Dashboard
          </h1>
          <p className="text-slate-400">
            {profileData?.name || 'Felipe Macedo'} • @{profileData?.login || 'felipemacedo1'}
          </p>
        </div>

        {/* KPI Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">Total Repositórios</span>
              <span className="text-2xl">📦</span>
            </div>
            <p className="text-3xl font-bold text-white">{repos.length}</p>
          </div>

          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">Total Commits (365d)</span>
              <span className="text-2xl">💻</span>
            </div>
            <p className="text-3xl font-bold text-emerald-400">{totalCommits}</p>
          </div>

          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">Pull Requests</span>
              <span className="text-2xl">🔀</span>
            </div>
            <p className="text-3xl font-bold text-blue-400">{totalPRs}</p>
          </div>

          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">Dias Ativos</span>
              <span className="text-2xl">🔥</span>
            </div>
            <p className="text-3xl font-bold text-amber-400">{activeDays}/365</p>
          </div>
        </div>

        {/* Top Languages */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6 mb-8">
          <h2 className="text-xl font-semibold text-white mb-4">Top 5 Linguagens</h2>
          <div className="space-y-3">
            {topLanguages.map((lang, index) => {
              const data = languageData?.languages[lang];
              return (
                <div key={lang} className="flex items-center gap-4">
                  <span className="text-2xl font-bold text-slate-600 w-8">{index + 1}</span>
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-white font-medium">{lang}</span>
                      <span className="text-slate-400 text-sm">{data?.percentage.toFixed(1)}%</span>
                    </div>
                    <div className="w-full bg-slate-700 rounded-full h-2">
                      <div 
                        className="bg-gradient-to-r from-blue-500 to-emerald-500 h-2 rounded-full transition-all"
                        style={{ width: `${data?.percentage || 0}%` }}
                      ></div>
                    </div>
                  </div>
                  <span className="text-slate-400 text-sm">{data?.repos} repos</span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Recent Repositories */}
        <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700 rounded-lg p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Repositórios Recentes</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {repos.slice(0, 6).map((repo) => (
              <a
                key={repo.name}
                href={repo.url}
                target="_blank"
                rel="noopener noreferrer"
                className="block bg-slate-700/50 border border-slate-600 rounded-lg p-4 hover:border-blue-500 transition-all hover:scale-105"
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="text-white font-semibold truncate">{repo.name}</h3>
                  {repo.language && (
                    <span className="text-xs bg-blue-500/20 text-blue-300 px-2 py-1 rounded-full ml-2 shrink-0">
                      {repo.language}
                    </span>
                  )}
                </div>
                {repo.description && (
                  <p className="text-slate-400 text-sm line-clamp-2">{repo.description}</p>
                )}
                <div className="mt-3 text-xs text-slate-500">
                  Atualizado: {new Date(repo.updated_at).toLocaleDateString('pt-BR')}
                </div>
              </a>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-slate-500 text-sm">
          <p>Última atualização: {new Date().toLocaleDateString('pt-BR', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
          })}</p>
          <p className="mt-2">🚀 Dados coletados automaticamente via GitHub API</p>
        </div>
      </div>
    </main>
  );
}
