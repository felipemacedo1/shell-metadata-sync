package utils

import (
	"context"
	"sync"
)

// WorkerPool gerencia um pool de workers para processamento concorrente
type WorkerPool struct {
	workerCount int
	jobs        chan func() error
	results     chan error
	wg          sync.WaitGroup
}

// NewWorkerPool cria um novo pool de workers
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func() error, workerCount*2),
		results:     make(chan error, workerCount*2),
	}
}

// Start inicia os workers
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

// Submit adiciona um job ao pool
func (wp *WorkerPool) Submit(job func() error) {
	wp.jobs <- job
}

// Close fecha o pool e aguarda todos os workers finalizarem
func (wp *WorkerPool) Close() []error {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
	
	var errors []error
	for err := range wp.results {
		if err != nil {
			errors = append(errors, err)
		}
	}
	
	return errors
}

// worker processa jobs do canal
func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	
	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			err := job()
			wp.results <- err
			
		case <-ctx.Done():
			return
		}
	}
}

// ProcessBatch processa um lote de itens concorrentemente
func ProcessBatch[T any](ctx context.Context, items []T, workerCount int, processFunc func(T) error) []error {
	pool := NewWorkerPool(workerCount)
	pool.Start(ctx)
	
	for _, item := range items {
		item := item // capture for closure
		pool.Submit(func() error {
			return processFunc(item)
		})
	}
	
	return pool.Close()
}
