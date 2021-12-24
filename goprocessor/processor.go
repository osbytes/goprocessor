package goprocessor

import (
	"context"
	"sync"

	"github.com/avast/retry-go"
)

type Options struct {
	MaxConcurrentItems int
	RetryMaxPerItem    uint
	RetryLastErrorOnly bool
	GracefulShutdown   bool
}

func Process[T any](ctx context.Context, items []T, fn func(item T) error, opts *Options) []error {
	maxConcurrent := 5
	maxRetriesPerItem := uint(0)
	gracefulShutdown := false
	retryLastErrorOnly := true

	if opts != nil {
		if opts.MaxConcurrentItems > 0 {
			maxConcurrent = opts.MaxConcurrentItems
		}

		maxRetriesPerItem = opts.RetryMaxPerItem
		gracefulShutdown = opts.GracefulShutdown
		retryLastErrorOnly = opts.RetryLastErrorOnly
	}

	boundedCh := make(chan struct{}, maxConcurrent)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	errs := []error{}
	syncAddErr := func(err error) {
		mu.Lock()
		defer mu.Unlock()

		errs = append(errs, err)
	}

	for _, item := range items {
		boundedCh <- struct{}{}

		if len(errs) > 0 {
			break
		}

		it := item

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-boundedCh
			}()

			select {
			case <-ctx.Done():
				syncAddErr(ctx.Err())
				return

			default:
				if maxRetriesPerItem > 0 {
					err := retry.Do(func() error {
						return fn(it)
					}, retry.Attempts(maxRetriesPerItem), retry.LastErrorOnly(retryLastErrorOnly)) // TOOD: allow more configuration on this retry
					if err != nil {
						syncAddErr(err)
					}

				} else {

					if err := fn(it); err != nil {
						syncAddErr(err)
					}
				}
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 && !gracefulShutdown {
		return []error{errs[0]}
	}

	return errs
}
