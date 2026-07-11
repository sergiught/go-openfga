package openfga

import (
	"context"
	"sync"
)

const (
	defaultMaxParallel       = 10
	defaultMaxPerChunk       = 1
	defaultMaxChecksPerBatch = 50
)

// resolvePositive returns v when positive, otherwise def.
func resolvePositive(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

// runParallel invokes fn for each index in [0,n) with at most maxWorkers
// running concurrently. It returns a length-n slice whose element i holds
// fn(i)'s error. If ctx is cancelled before index i starts, element i is set
// to ctx.Err() and fn(i) is not called. All scheduled fn calls are allowed to
// finish; a sibling error never cancels others.
func runParallel(ctx context.Context, n, maxWorkers int, fn func(i int) error) []error {
	errs := make([]error, n)
	if n == 0 {
		return errs
	}
	sem := make(chan struct{}, resolvePositive(maxWorkers, 1))
	var wg sync.WaitGroup
	for i := range n {
		if err := ctx.Err(); err != nil {
			errs[i] = err
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			errs[i] = fn(i)
		}(i)
	}
	wg.Wait()
	return errs
}
