package scToProtocol

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

type modifiedMBShardIDCheckerHandler interface {
	isModifiedStateMBValid(miniBlock *block.MiniBlock) bool
}

// StakingToPeerFactoryHandler defines the factory interface to create sc to protocol handler
type StakingToPeerFactoryHandler interface {
	CreateStakingToPeer(args ArgStakingToPeer) (process.SmartContractToProtocolHandler, error)
	IsInterfaceNil() bool
}
