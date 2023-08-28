package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

// ForkDetectorFactoryStub -
type ForkDetectorFactoryStub struct {
	CreateForkDetectorCalled func(args sync.ForkDetectorFactoryArgs) (process.ForkDetector, error)
}

// NewForkDetectorFactoryStub -
func NewForkDetectorFactoryStub() *ForkDetectorFactoryStub {
	return &ForkDetectorFactoryStub{}
}

// CreateForkDetector -
func (f *ForkDetectorFactoryStub) CreateForkDetector(args sync.ForkDetectorFactoryArgs) (process.ForkDetector, error) {
	if f.CreateForkDetectorCalled != nil {
		return f.CreateForkDetectorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (f *ForkDetectorFactoryStub) IsInterfaceNil() bool {
	return false
}
