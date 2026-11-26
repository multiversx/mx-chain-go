package leavesRetriever

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
)

type disabledLeavesRetriever struct{}

// NewDisabledLeavesRetriever creates a new disabled leaves retriever
func NewDisabledLeavesRetriever() *disabledLeavesRetriever {
	return &disabledLeavesRetriever{}
}

func (dlr *disabledLeavesRetriever) GetLeaves(_ int, _ [][]byte, _ common.TrieLeafParser, _ context.Context) (map[string]string, [][]byte, error) {
	return nil, nil, common.ErrTrieLeavesRetrieverDisabled
}

// IsInterfaceNil returns true if there is no value under the interface
func (dlr *disabledLeavesRetriever) IsInterfaceNil() bool {
	return dlr == nil
}
