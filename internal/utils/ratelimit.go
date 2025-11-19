package utils

import (
"context"
"fmt"
"log"
"math"
"net/http"
"strconv"
"time"
)

// RateLimitHandler gerencia rate limits da API do GitHub
type RateLimitHandler struct {
maxRetries int
baseDelay  time.Duration
}

// NewRateLimitHandler cria um novo handler de rate limit
func NewRateLimitHandler(maxRetries int, baseDelay time.Duration) *RateLimitHandler {
return &RateLimitHandler{
maxRetries: maxRetries,
baseDelay:  baseDelay,
}
}

// RateLimitInfo contém informações sobre rate limit
type RateLimitInfo struct {
Limit     int
Remaining int
Reset     time.Time
Used      int
}

// ParseRateLimitHeaders extrai informações de rate limit dos headers HTTP
func ParseRateLimitHeaders(resp *http.Response) *RateLimitInfo {
info := &RateLimitInfo{}

if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
info.Limit, _ = strconv.Atoi(limit)
}

if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
info.Remaining, _ = strconv.Atoi(remaining)
}

if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
info.Reset = time.Unix(timestamp, 0)
}
}

if used := resp.Header.Get("X-RateLimit-Used"); used != "" {
info.Used, _ = strconv.Atoi(used)
}

return info
}

// HandleRateLimit verifica e trata erros de rate limit
func (h *RateLimitHandler) HandleRateLimit(resp *http.Response) error {
if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
return nil
}

info := ParseRateLimitHeaders(resp)

if info.Remaining == 0 {
waitDuration := time.Until(info.Reset)
if waitDuration > 0 {
log.Printf("⏳ Rate limit atingido. Aguardando até %s (%v)", 
info.Reset.Format(time.RFC3339), waitDuration)
return fmt.Errorf("rate limit exceeded, reset at %s", info.Reset.Format(time.RFC3339))
}
}

return fmt.Errorf("rate limit error: status=%d remaining=%d", resp.StatusCode, info.Remaining)
}

// RetryWithBackoff executa uma requisição com retry e backoff exponencial
func (h *RateLimitHandler) RetryWithBackoff(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
var resp *http.Response
var err error

for attempt := 0; attempt <= h.maxRetries; attempt++ {
if attempt > 0 {
delay := h.calculateBackoff(attempt)
log.Printf("🔄 Retry attempt %d/%d após %v", attempt, h.maxRetries, delay)

select {
case <-time.After(delay):
case <-ctx.Done():
return nil, ctx.Err()
}
}

resp, err = fn()
if err != nil {
if attempt == h.maxRetries {
return nil, fmt.Errorf("max retries exceeded: %w", err)
}
continue
}

// Verificar rate limit
if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
info := ParseRateLimitHeaders(resp)

if info.Remaining == 0 {
waitDuration := time.Until(info.Reset)
if waitDuration > 0 && waitDuration < 5*time.Minute {
log.Printf("⏳ Aguardando reset do rate limit: %v", waitDuration)
select {
case <-time.After(waitDuration + 1*time.Second):
continue
case <-ctx.Done():
return nil, ctx.Err()
}
}
}

if attempt == h.maxRetries {
return resp, h.HandleRateLimit(resp)
}
continue
}

// Sucesso
if resp.StatusCode >= 200 && resp.StatusCode < 300 {
return resp, nil
}

// Erro não relacionado a rate limit
if attempt == h.maxRetries {
return resp, fmt.Errorf("request failed with status %d", resp.StatusCode)
}
}

return resp, err
}

// calculateBackoff calcula o delay de backoff exponencial
func (h *RateLimitHandler) calculateBackoff(attempt int) time.Duration {
delay := float64(h.baseDelay) * math.Pow(2, float64(attempt-1))

// Cap máximo de 1 minuto
if delay > float64(60*time.Second) {
delay = float64(60 * time.Second)
}

return time.Duration(delay)
}

// LogRateLimitInfo loga informações sobre o rate limit atual
func LogRateLimitInfo(resp *http.Response) {
info := ParseRateLimitHeaders(resp)
if info.Limit > 0 {
percentage := float64(info.Remaining) / float64(info.Limit) * 100
log.Printf("📊 Rate Limit: %d/%d (%.1f%% restante, reset em %s)",
info.Remaining, info.Limit, percentage, info.Reset.Format("15:04:05"))
}
}
