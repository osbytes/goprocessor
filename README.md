# goprocessor

<a href="https://codecov.io/gh/osbytes/goprocessor">
    <img src="https://codecov.io/gh/osbytes/goprocessor/branch/main/graph/badge.svg" alt="codecov" />
</a>
<a href="https://github.com/osbytes/goprocessor/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/osbytes/goprocessor.svg" alt="License" />
</a>

a go job processor which manages the orchestration of concurrent jobs with a generic interface for easy portability

## Requirements

[go 1.18](https://tip.golang.org/doc/go1.18) for generics

## Installation

```sh
go get github.com/osbytes/goprocessor
```

## Usage

```go
package main

import (
	"context"
	"log"

	"github.com/osbytes/goprocessor"
)

func main() {
	errs := goprocessor.Process(
		context.Background(),
		[]int{1, 2, 3, 4, 5},
		func(item int) error {

			// item processing code here...

			return nil
		},
		&goprocessor.Options{
			GracefulShutdown:   false,
			MaxConcurrentItems: 0,
			RetryLastErrorOnly: false,
			RetryMaxPerItem:    0,
		},
	)
	if len(errs) > 0 {
		log.Fatal(errs)
	}

	errs = goprocessor.Process(
		context.Background(),
		[]string{"a", "b", "c", "d", "e"},
		func(item string) error {

			// item processing code here...

			return nil
		},
		nil,
	)
	if len(errs) > 0 {
		log.Fatal(errs)
	}
}
```