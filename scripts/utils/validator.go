package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// ValidateJSON valida se o conteúdo é um JSON válido
func ValidateJSON(data []byte) error {
	var js interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ValidateJSONFile valida se um arquivo contém JSON válido
func ValidateJSONFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	
	return ValidateJSON(data)
}

// ValidateStructure valida se o JSON tem a estrutura esperada
func ValidateStructure(data []byte, expectedKeys []string) error {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("not a JSON object: %w", err)
	}
	
	for _, key := range expectedKeys {
		if _, exists := obj[key]; !exists {
			return fmt.Errorf("missing required key: %s", key)
		}
	}
	
	return nil
}

// ValidateAndSaveJSON valida JSON antes de salvar
func ValidateAndSaveJSON(filePath string, data interface{}) error {
	// Converter para JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	
	// Validar JSON
	if err := ValidateJSON(jsonData); err != nil {
		return err
	}
	
	// Salvar em arquivo temporário primeiro
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing temp file: %w", err)
	}
	
	// Validar arquivo temporário
	if err := ValidateJSONFile(tmpFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Renomear para arquivo final
	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("error renaming file: %w", err)
	}
	
	return nil
}

// CheckJSONSize verifica se o JSON não está muito grande
func CheckJSONSize(data []byte, maxSizeMB int) error {
	sizeMB := float64(len(data)) / (1024 * 1024)
	if sizeMB > float64(maxSizeMB) {
		return fmt.Errorf("JSON size (%.2f MB) exceeds maximum allowed (%d MB)", sizeMB, maxSizeMB)
	}
	return nil
}
