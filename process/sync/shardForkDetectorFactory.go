package sync

import "github.com/multiversx/mx-chain-go/process"

type shardForkDetectorFactory struct {
}

// NewShardForkDetectorFactory creates a new shard fork detector factory
func NewShardForkDetectorFactory() (*shardForkDetectorFactory, error) {
	return &shardForkDetectorFactory{}, nil
}

// CreateForkDetector creates a new fork detector
func (s *shardForkDetectorFactory) CreateForkDetector(args ForkDetectorFactoryArgs) (process.ForkDetector, error) {
	return NewShardForkDetector(args.RoundHandler, args.HeaderBlackList, args.BlockTracker, args.GenesisTime)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardForkDetectorFactory) IsInterfaceNil() bool {
	return s == nil
}
