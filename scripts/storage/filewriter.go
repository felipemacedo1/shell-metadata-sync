package storage

import (
"encoding/json"
"os"
"path/filepath"
)

// SaveToMultiplePaths salva o JSON em múltiplos paths
func SaveToMultiplePaths(data interface{}, paths ...string) error {
for _, path := range paths {
if err := SaveJSON(path, data); err != nil {
return err
}
}
return nil
}

// SaveJSON salva dados em formato JSON
func SaveJSON(path string, v interface{}) error {
// Criar diretório se não existir
if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
return err
}

tmp := path + ".tmp"
f, err := os.Create(tmp)
if err != nil {
return err
}

enc := json.NewEncoder(f)
enc.SetIndent("", "  ")
if err := enc.Encode(v); err != nil {
f.Close()
return err
}

if err := f.Close(); err != nil {
return err
}

return os.Rename(tmp, path)
}

// GetDefaultPaths retorna os paths padrão (data/ e dashboard/public/data/)
func GetDefaultPaths(filename string) []string {
return []string{
filepath.Join("data", filename),
filepath.Join("dashboard", "public", "data", filename),
}
}
