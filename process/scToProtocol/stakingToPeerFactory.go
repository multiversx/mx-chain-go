package scToProtocol

import "github.com/multiversx/mx-chain-go/process"

type stakingToPeerFactory struct {
}

// NewStakingToPeerFactory creates a new staking to peer factory for normal chain run type
func NewStakingToPeerFactory() *stakingToPeerFactory {
	return &stakingToPeerFactory{}
}

// CreateStakingToPeer creates a new staking to peer for normal chain run type
func (f *stakingToPeerFactory) CreateStakingToPeer(args ArgStakingToPeer) (process.SmartContractToProtocolHandler, error) {
	return NewStakingToPeer(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *stakingToPeerFactory) IsInterfaceNil() bool {
	return f == nil
}
