package node

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory"
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

// GetBlockHeaderByHash -
func (n *Node) GetBlockHeaderByHash(headerHash []byte) (data.HeaderHandler, error) {
	return n.getBlockHeaderByHash(headerHash)
}

// MergeAccountQueryOptionsIntoBlockInfo -
func MergeAccountQueryOptionsIntoBlockInfo(options api.AccountQueryOptions, info common.BlockInfo) common.BlockInfo {
	return mergeAccountQueryOptionsIntoBlockInfo(options, info)
}

// ExtractApiBlockInfoIfErrAccountNotFoundAtBlock -
func ExtractApiBlockInfoIfErrAccountNotFoundAtBlock(err error) (api.BlockInfo, bool) {
	return extractApiBlockInfoIfErrAccountNotFoundAtBlock(err)
}
