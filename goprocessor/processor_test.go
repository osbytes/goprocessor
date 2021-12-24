package goprocessor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	assert := assert.New(t)

	processed := map[string]bool{}
	mu := sync.Mutex{}

	errs := Process(
		context.Background(),
		[]string{"a", "b", "c"},
		func(item string) error {
			mu.Lock()
			processed[item] = true
			mu.Unlock()

			return nil
		},
		nil,
	)
	assert.Len(errs, 0)

	assert.Contains(processed, "a")
	assert.Contains(processed, "b")
	assert.Contains(processed, "c")
}

func TestProcess_errors(t *testing.T) {
	assert := assert.New(t)

	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err3 := errors.New("error3")

	errs := Process(
		context.Background(),
		[]string{"a", "b", "c"},
		func(item string) error {
			switch item {
			case "a":
				return err1
			case "b":
				return err2
			case "c":
				return err3
			}
			return nil
		},
		&Options{MaxConcurrentItems: 1},
	)
	assert.Len(errs, 1)
	assert.ErrorIs(err1, errs[0])
}

func TestProcess_errors_graceful(t *testing.T) {
	assert := assert.New(t)

	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err3 := errors.New("error3")

	errs := Process(
		context.Background(),
		[]string{"a", "b", "c"},
		func(item string) error {
			switch item {
			case "a":
				return err1
			case "b":
				return err2
			case "c":
				return err3
			}
			return nil
		},
		&Options{MaxConcurrentItems: 3, GracefulShutdown: true},
	)
	assert.Len(errs, 3)
	assert.Contains(errs, err1)
	assert.Contains(errs, err2)
	assert.Contains(errs, err3)
}

func TestProcess_retries_success(t *testing.T) {
	assert := assert.New(t)

	attemptNum := uint(1)
	maxAttempts := uint(3)

	errs := Process(
		context.Background(),
		[]string{"a"},
		func(item string) error {
			if attemptNum < maxAttempts {
				attemptNum++
				return errors.New("err")
			}

			return nil
		},
		&Options{RetryMaxPerItem: maxAttempts},
	)
	assert.Len(errs, 0)
}

func TestProcess_retries_error(t *testing.T) {
	assert := assert.New(t)

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := strconv.Itoa(seededRand.Int())

	err := fmt.Errorf(num)
	attemptNum := uint(1)

	errs := Process(
		context.Background(),
		[]string{"a"},
		func(item string) error {
			if attemptNum < 3 {
				attemptNum++
				return err
			}

			return nil
		},
		&Options{RetryMaxPerItem: 2, RetryLastErrorOnly: false},
	)
	assert.Len(errs, 1)
	assert.Contains(errs[0].Error(), num)
}

func TestProcess_context_cancelled(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errs := Process(
		ctx,
		[]string{"a"},
		func(item string) error {
			return nil
		},
		nil,
	)
	assert.Len(errs, 1)
	assert.ErrorIs(errs[0], context.Canceled)
}
