package mock

import "time"

// BenchmarkStub -
type BenchmarkStub struct {
	NameCalled func() string
	RunCalled  func() (time.Duration, error)
}

// Run -
func (bs *BenchmarkStub) Run() (time.Duration, error) {
	if bs.RunCalled != nil {
		return bs.RunCalled()
	}

	return 0, nil
}

// Name -
func (bs *BenchmarkStub) Name() string {
	if bs.NameCalled != nil {
		return bs.NameCalled()
	}

	return ""
}

// IsInterfaceNil -
func (bs *BenchmarkStub) IsInterfaceNil() bool {
	return bs == nil
}
