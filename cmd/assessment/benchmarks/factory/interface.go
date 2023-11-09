package factory

import "github.com/multiversx/mx-chain-go/cmd/assessment/benchmarks"

type benchmarkCoordinator interface {
	RunAllTests() *benchmarks.TestResults
	IsInterfaceNil() bool
}
