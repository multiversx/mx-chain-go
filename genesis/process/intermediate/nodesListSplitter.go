package intermediate

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type nodesListSplitter struct {
	allNodes       []sharding.GenesisNodeInfoHandler
	accountsParser genesis.AccountsParser
}

// NewNodesListSplitter returns an instance able to split the nodes by some criterias
func NewNodesListSplitter(
	initialNodesSetup genesis.InitialNodesHandler,
	accountsParser genesis.AccountsParser,
) (*nodesListSplitter, error) {

	if check.IfNil(initialNodesSetup) {
		return nil, genesis.ErrNilNodesSetup
	}
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

	return &nodesListSplitter{
		allNodes:       allNodes,
		accountsParser: accountsParser,
	}, nil
}

func (nls *nodesListSplitter) isDelegated(address []byte) bool {
	accounts := nls.accountsParser.InitialAccounts()
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
func (nls *nodesListSplitter) GetAllNodes() []sharding.GenesisNodeInfoHandler {
	return nls.allNodes
}

// GetDelegatedNodes returns the initial nodes that were delegated by the provided delegation SC address
func (nls *nodesListSplitter) GetDelegatedNodes(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler {
	delegatedNodes := make([]sharding.GenesisNodeInfoHandler, 0)
	for _, node := range nls.allNodes {
		if !nls.isDelegated(node.AddressBytes()) {
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
func (nls *nodesListSplitter) IsInterfaceNil() bool {
	return nls == nil
}
