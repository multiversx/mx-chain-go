package chainSimulator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

type chainSimulatorFacade struct {
	chainSimulator ChainSimulator
	metaNode       process.NodeHandler
}

// NewChainSimulatorFacade returns the chain Simulator facade
func NewChainSimulatorFacade(chainSimulator ChainSimulator) (*chainSimulatorFacade, error) {
	if check.IfNil(chainSimulator) {
		return nil, errNilChainSimulator
	}

	metaNode := chainSimulator.GetNodeHandler(common.MetachainShardId)
	if check.IfNil(metaNode) {
		return nil, errNilMetachainNode
	}

	return &chainSimulatorFacade{
		chainSimulator: chainSimulator,
		metaNode:       metaNode,
	}, nil
}

// GetExistingAccountFromBech32AddressString will return the existing account for the provided address in bech32 format
func (f *chainSimulatorFacade) GetExistingAccountFromBech32AddressString(address string) (vmcommon.UserAccountHandler, error) {
	addressBytes, err := f.metaNode.GetCoreComponents().AddressPubKeyConverter().Decode(address)
	if err != nil {
		return nil, err
	}

	shardID := f.metaNode.GetShardCoordinator().ComputeId(addressBytes)

	shardNodeHandler := f.chainSimulator.GetNodeHandler(shardID)
	if check.IfNil(shardNodeHandler) {
		return nil, fmt.Errorf("%w missing node handler for shard %d", errShardSetupError, shardID)
	}

	account, err := shardNodeHandler.GetStateComponents().AccountsAdapter().GetExistingAccount(addressBytes)
	if err != nil {
		return nil, err
	}

	return account.(vmcommon.UserAccountHandler), nil
}
