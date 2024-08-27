package scToProtocol

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignStakingToPeer struct {
	*stakingToPeer
}

// NewSovereignStakingToPeer creates the staking to peer protocol for sovereign chain
func NewSovereignStakingToPeer(sp *stakingToPeer) (*sovereignStakingToPeer, error) {
	if check.IfNil(sp) {
		return nil, process.ErrNilSCToProtocol
	}

	sp.modifiedMBShardIDCheckerHandler = &sovModifiedMBShardIDChecker{}

	return &sovereignStakingToPeer{
		sp,
	}, nil
}
