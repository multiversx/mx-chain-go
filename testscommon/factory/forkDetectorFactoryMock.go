package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

// ForkDetectorFactoryMock -
type ForkDetectorFactoryMock struct {
	CreateForkDetectorCalled func(args sync.ForkDetectorFactoryArgs) (process.ForkDetector, error)
}

// CreateForkDetector -
func (f *ForkDetectorFactoryMock) CreateForkDetector(args sync.ForkDetectorFactoryArgs) (process.ForkDetector, error) {
	if f.CreateForkDetectorCalled != nil {
		return f.CreateForkDetectorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (f *ForkDetectorFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
