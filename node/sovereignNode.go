package node

import (
	"context"
	"github.com/multiversx/mx-chain-core-go/data/api"
)

type SovereignNode struct {
	*Node
}

func NewSovereignNode(node *Node) *SovereignNode {
	return &SovereignNode{
		node,
	}
}

func (sn *SovereignNode) GetAllIssuedESDTs(tokenType string, ctx context.Context) ([]string, error) {
	return sn.baseGetAllIssuedESDTs(tokenType, ctx)
}

func (sn *SovereignNode) baseGetTokensIDsWithFilter(
	f filter,
	options api.AccountQueryOptions,
	ctx context.Context,
) ([]string, api.BlockInfo, error) {
	return sn.baseGetTokensIDsWithFilter(f, options, ctx)
}
