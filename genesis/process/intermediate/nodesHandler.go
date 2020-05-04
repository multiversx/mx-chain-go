package intermediate

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type nodesHandler struct {
	allNodes       []sharding.GenesisNodeInfoHandler
	accountsParser genesis.AccountsParser
}

// NewNodesHandler returns a wrapper instance over the initial nodes setup in order to add some helper functions
// used in genesis delegation process
func NewNodesHandler(
	initialNodesSetup genesis.InitialNodesHandler,
	accountsParser genesis.AccountsParser,
) (*nodesHandler, error) {

	if check.IfNil(accountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}

	eligible, waiting := initialNodesSetup.InitialNodesInfo()

	allNodes := make([]sharding.GenesisNodeInfoHandler, 0)
	keys := make([]uint32, 0)
	for shard := range eligible {
		keys = append(keys, shard)
	}

	//it is important that the processing is done in a deterministic way
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, shardID := range keys {
		allNodes = append(allNodes, eligible[shardID]...)
		allNodes = append(allNodes, waiting[shardID]...)
	}

	return &nodesHandler{
		allNodes:       allNodes,
		accountsParser: accountsParser,
	}, nil
}

func (nh *nodesHandler) isStaked(address []byte) bool {
	accounts := nh.accountsParser.InitialAccounts()
	for _, ac := range accounts {
		if !bytes.Equal(ac.AddressBytes(), address) {
			continue
		}

		return ac.GetStakingValue().Cmp(zero) > 0
	}

	return false
}

func (nh *nodesHandler) isDelegated(address []byte) bool {
	accounts := nh.accountsParser.InitialAccounts()
	for _, ac := range accounts {
		dh := ac.GetDelegationHandler()
		if check.IfNil(dh) {
			continue
		}

		if !bytes.Equal(dh.AddressBytes(), address) {
			continue
		}

		return dh.GetValue().Cmp(zero) > 0
	}

	return false
}

// GetAllNodes returns all initial nodes that (directly staked or delegated)
func (nh *nodesHandler) GetAllNodes() []sharding.GenesisNodeInfoHandler {
	return nh.allNodes
}

// GetDelegatedNodes returns the initial nodes that were delegated by the provided delegation SC address
func (nh *nodesHandler) GetDelegatedNodes(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler {
	delegatedNodes := make([]sharding.GenesisNodeInfoHandler, 0)
	for _, node := range nh.allNodes {
		if !nh.isDelegated(node.AddressBytes()) {
			continue
		}
		if !bytes.Equal(node.AddressBytes(), delegationScAddress) {
			continue
		}

		delegatedNodes = append(delegatedNodes, node)
	}

	return delegatedNodes
}

// IsInterfaceNil returns if underlying object is true
func (nh *nodesHandler) IsInterfaceNil() bool {
	return nh == nil
}
