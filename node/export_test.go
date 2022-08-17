package node

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/factory"
)

// GetClosableComponentName -
func (n *Node) GetClosableComponentName(component factory.Closer, index int) string {
	return n.getClosableComponentName(component, index)
}

// ComputeProof -
func (n *Node) ComputeProof(rootHash []byte, key []byte) (*common.GetProofResponse, error) {
	return n.getProof(rootHash, key)
}

// AddClosableComponents -
func (n *Node) AddClosableComponents(components ...factory.Closer) {
	n.closableComponents = append(n.closableComponents, components...)
}

// AddBlockCoordinatesToAccountQueryOptions -
func (n *Node) AddBlockCoordinatesToAccountQueryOptions(options api.AccountQueryOptions) (api.AccountQueryOptions, error) {
	return n.addBlockCoordinatesToAccountQueryOptions(options)
}

// MergeAccountQueryOptionsIntoBlockInfo -
func MergeAccountQueryOptionsIntoBlockInfo(options api.AccountQueryOptions, info common.BlockInfo) common.BlockInfo {
	return mergeAccountQueryOptionsIntoBlockInfo(options, info)
}
