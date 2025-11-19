package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// StructuredLog representa um log estruturado em JSON
type StructuredLog struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger interface para logging estruturado e tradicional
type Logger struct {
	structured bool
}

// NewLogger cria um novo logger
func NewLogger(structured bool) *Logger {
	return &Logger{structured: structured}
}

// Info loga mensagem de informação
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log("INFO", message, fields...)
}

// Warning loga mensagem de aviso
func (l *Logger) Warning(message string, fields ...map[string]interface{}) {
	l.log("WARNING", message, fields...)
}

// Error loga mensagem de erro
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log("ERROR", message, fields...)
}

// Success loga mensagem de sucesso
func (l *Logger) Success(message string, fields ...map[string]interface{}) {
	l.log("SUCCESS", message, fields...)
}

func (l *Logger) log(level, message string, fields ...map[string]interface{}) {
	if l.structured {
		logEntry := StructuredLog{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     level,
			Message:   message,
		}
		
		if len(fields) > 0 {
			logEntry.Fields = fields[0]
		}
		
		jsonLog, _ := json.Marshal(logEntry)
		fmt.Println(string(jsonLog))
	} else {
		prefix := l.getPrefix(level)
		if len(fields) > 0 {
			log.Printf("%s %s %v", prefix, message, fields[0])
		} else {
			log.Printf("%s %s", prefix, message)
		}
	}
}

func (l *Logger) getPrefix(level string) string {
	switch level {
	case "INFO":
		return "ℹ️ "
	case "WARNING":
		return "⚠️ "
	case "ERROR":
		return "❌"
	case "SUCCESS":
		return "✅"
	default:
		return ""
	}
}

// Metrics estrutura para métricas de execução
type Metrics struct {
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	ItemsTotal    int
	ItemsSuccess  int
	ItemsFailed   int
	APICallsTotal int
	RateLimitHits int
}

// NewMetrics cria nova instância de métricas
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// Finish finaliza métricas
func (m *Metrics) Finish() {
	m.EndTime = time.Now()
	m.Duration = m.EndTime.Sub(m.StartTime)
}

// ToMap converte métricas para map
func (m *Metrics) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"start_time":      m.StartTime.Format(time.RFC3339),
		"end_time":        m.EndTime.Format(time.RFC3339),
		"duration_sec":    m.Duration.Seconds(),
		"items_total":     m.ItemsTotal,
		"items_success":   m.ItemsSuccess,
		"items_failed":    m.ItemsFailed,
		"api_calls_total": m.APICallsTotal,
		"rate_limit_hits": m.RateLimitHits,
	}
}

// String retorna representação em string das métricas
func (m *Metrics) String() string {
	successRate := float64(0)
	if m.ItemsTotal > 0 {
		successRate = float64(m.ItemsSuccess) / float64(m.ItemsTotal) * 100
	}
	
	return fmt.Sprintf(
		"Duration: %v | Items: %d/%d (%.1f%% success) | API Calls: %d | Rate Limit Hits: %d",
		m.Duration.Round(time.Millisecond),
		m.ItemsSuccess,
		m.ItemsTotal,
		successRate,
		m.APICallsTotal,
		m.RateLimitHits,
	)
}
