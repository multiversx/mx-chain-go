package sync

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignForkDetectorFactory struct {
	shardForkDetectorFactory ForkDetectorCreator
}

// NewSovereignForkDetectorFactory creates a new shard fork detector factory
func NewSovereignForkDetectorFactory(fdc ForkDetectorCreator) (*sovereignForkDetectorFactory, error) {
	if check.IfNil(fdc) {
		return nil, process.ErrNilForkDetectorCreator
	}
	return &sovereignForkDetectorFactory{
		shardForkDetectorFactory: fdc,
	}, nil
}

// CreateForkDetector creates a new fork detector
func (s *sovereignForkDetectorFactory) CreateForkDetector(args ForkDetectorFactoryArgs) (process.ForkDetector, error) {
	sfd, err := s.shardForkDetectorFactory.CreateForkDetector(args)
	if err != nil {
		return nil, err
	}

	sharFD, ok := sfd.(*shardForkDetector)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	return NewSovereignChainShardForkDetector(sharFD)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignForkDetectorFactory) IsInterfaceNil() bool {
	return s == nil
}
