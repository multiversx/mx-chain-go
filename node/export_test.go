package node

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/state"
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

// SetTxGuardianData -
func (n *Node) SetTxGuardianData(guardian string, guardianSigHex string, tx *transaction.Transaction) error {
	return n.setTxGuardianData(guardian, guardianSigHex, tx)
}

func (n *Node) GetPendingAndActiveGuardians(
	userAccount state.UserAccountHandler,
) (activeGuardian *api.Guardian, pendingGuardian *api.Guardian, err error) {
	return n.getPendingAndActiveGuardians(userAccount)
}
