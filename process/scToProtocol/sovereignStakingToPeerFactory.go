package scToProtocol

import "github.com/multiversx/mx-chain-go/process"

type sovereignStakingToPeerFactory struct {
}

// NewSovereignStakingToPeerFactory creates a new staking to peer factory for sovereign chain run type
func NewSovereignStakingToPeerFactory() *sovereignStakingToPeerFactory {
	return &sovereignStakingToPeerFactory{}
}

// CreateStakingToPeer creates a new staking to peer for sovereign chain run type
func (f *sovereignStakingToPeerFactory) CreateStakingToPeer(args ArgStakingToPeer) (process.SmartContractToProtocolHandler, error) {
	sp, err := NewStakingToPeer(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignStakingToPeer(sp)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignStakingToPeerFactory) IsInterfaceNil() bool {
	return f == nil
}
