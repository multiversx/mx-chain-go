package factory

import "github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks"

type benchmarkCoordinator interface {
	RunAllTests() benchmarks.TestResults
	IsInterfaceNil() bool
}
