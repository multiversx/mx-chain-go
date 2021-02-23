package benchmarks

import "time"

// BenchmarkRunner defines a benchmark test able to measure in some way the host
type BenchmarkRunner interface {
	Run() (time.Duration, error)
	Name() string
	IsInterfaceNil() bool
}
