# SonarCloud Issues Sync - Centralized

Workflow centralizado para sincronização de issues do SonarCloud com GitHub Issues.

## 📋 Características

### ✅ Cenários Suportados

1. **Execução Agendada**
   - 2x por dia (6h e 18h UTC)
   - Processa todos os repositórios automaticamente

2. **Execução Manual** (`workflow_dispatch`)
   - Filtrar repos específicos
   - Modo dry-run (teste sem alterações)
   - Limitar número de repos processados

3. **Descoberta Automática**
   - Lista repos pessoais (felipemacedo1)
   - Lista repos da organização (growthfolio)
   - Detecta quais têm SonarCloud configurado

4. **Gerenciamento Inteligente de Issues**
   - Cria issues para novos problemas do SonarCloud
   - Atualiza issues existentes
   - Fecha issues resolvidas (opcional)
   - Evita duplicatas

5. **Rate Limiting**
   - Sleep entre repos (evita throttling)
   - Sleep entre issues (respeita API limits)

6. **Tratamento de Erros**
   - Continue on error (um repo falhar não para os outros)
   - Logging detalhado
   - Relatórios de execução

7. **Observabilidade**
   - Logs JSON estruturados
   - Artifacts uploadados (retidos por 30 dias)
   - Métricas exportadas para MongoDB (opcional)

8. **Labels Automáticas**
   - `sonarcloud` - Identifica issues do Sonar
   - `severity:high|medium|low` - Severidade
   - `type:bug|code_smell|vulnerability` - Tipo

## 🚀 Setup

### 1. Secrets Necessários

Configure em `Settings > Secrets and variables > Actions`:

```yaml
GITHUB_TOKEN: <automático, não precisa configurar>
SONAR_TOKEN: <seu token do SonarCloud>
MONGODB_URI: <opcional, para métricas>
```

### 2. Estrutura de Diretórios

```
shell-metadata-sync/
├── .github/workflows/
│   └── sync-sonar-issues.yml
├── scripts/
│   └── sonar-issue-sync.sh
├── data/
│   ├── sonar-sync-TIMESTAMP.json  (relatórios)
│   └── sonar-sync-TIMESTAMP.log   (logs)
└── README-SONAR-SYNC.md
```

### 3. Permissões

O workflow precisa de:
- `contents: write` - Para commit de dados
- `issues: write` - Para gerenciar issues nos repos

## 📖 Uso

### Execução Automática (Agendada)

Roda automaticamente 2x por dia. Não precisa fazer nada!

### Execução Manual

#### Processar todos os repos:
```yaml
Inputs:
  repos_filter: all
  dry_run: false
  max_repos: 0
```

#### Testar com repos específicos (dry-run):
```yaml
Inputs:
  repos_filter: ktar,go-portifolio
  dry_run: true
  max_repos: 5
```

#### Processar apenas repos da org:
```yaml
Inputs:
  repos_filter: growthfolio
  dry_run: false
  max_repos: 0
```

## 🔍 Exemplo de Issue Criada

```markdown
**Issue Key:** `AY1234567890`
**Severity:** high
**Type:** bug
**File:** `src/main/App.java`
**Line:** 42

**Description:**
Remove this unused variable

---
🔗 [View in SonarCloud](https://sonarcloud.io/project/issues?id=...)
```

## 📊 Relatório Gerado

```json
{
  "execution": {
    "timestamp": "2025-11-18T00:00:00Z",
    "duration_seconds": 325,
    "dry_run": false
  },
  "summary": {
    "total_repos": 52,
    "processed": 52,
    "successful": 50,
    "failed": 2,
    "skipped": 15
  },
  "issues": {
    "created": 45,
    "updated": 12,
    "closed": 3
  }
}
```

## 🛠️ Troubleshooting

### Issue: "No SonarCloud project found"
**Causa**: Repo não tem SonarCloud configurado  
**Solução**: Normal, será pulado automaticamente

### Issue: "Failed to create issue"
**Causa**: Permissões insuficientes ou rate limit  
**Solução**: Verifique GITHUB_TOKEN e aguarde alguns minutos

### Issue: "SonarCloud API error"
**Causa**: SONAR_TOKEN inválido ou expirado  
**Solução**: Gere novo token em SonarCloud > My Account > Security

## 🔄 Integração com Dashboard

Os dados são salvos em:
- `/data/sonar-sync-*.json` - Métricas agregadas
- MongoDB (opcional) - Para consultas e dashboards

Adicione ao seu dashboard Next.js:

```typescript
// pages/sonar-quality.tsx
import sonarData from '../data/sonar-sync-latest.json';

export default function SonarQuality() {
  return (
    <div>
      <h1>Code Quality Metrics</h1>
      <p>Total Issues: {sonarData.issues.created}</p>
      <p>Success Rate: {sonarData.summary.successful / sonarData.summary.total_repos}%</p>
    </div>
  );
}
```

## ⚙️ Customização

### Alterar frequência de execução:
Edite o cron em `sync-sonar-issues.yml`:
```yaml
schedule:
  - cron: '0 */4 * * *'  # A cada 4 horas
```

### Adicionar notificações:
Edite o step "Send notification on failure":
```yaml
- name: Send notification on failure
  if: failure()
  run: |
    curl -X POST $SLACK_WEBHOOK \
      -d '{"text": "❌ Sonar sync failed!"}'
```

### Customizar labels:
Edite a função `create_github_issue` em `sonar-issue-sync.sh`:
```bash
local labels="sonarcloud,severity:${severity},type:${type},priority:high"
```

## 📝 TODO / Melhorias Futuras

- [ ] Fechar issues resolvidas automaticamente
- [ ] Suporte a múltiplas organizações
- [ ] Dashboard web para visualização
- [ ] Notificações por Slack/Discord
- [ ] Métricas de tendência (issues over time)
- [ ] Filtros por severidade mínima
- [ ] Integração com JIRA/Linear

## 🤝 Contribuindo

Para testar mudanças antes de aplicar:

1. Use `dry_run: true`
2. Limite com `max_repos: 5`
3. Filtre repos de teste: `repos_filter: test-repo`

## 📄 Licença

MIT
