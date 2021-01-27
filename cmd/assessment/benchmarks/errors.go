package benchmarks

import "errors"

// ErrEmptyBenchmarksSlice signals that the provided benchmarks slice was empty
var ErrEmptyBenchmarksSlice = errors.New("empty benchmarks slice provided")

// ErrNilBenchmark signals that a nil benchmark was provided
var ErrNilBenchmark = errors.New("nil benchmark")

// ErrFileDoesNotExists signals that the required file does not exists
var ErrFileDoesNotExists = errors.New("file does not exists")
