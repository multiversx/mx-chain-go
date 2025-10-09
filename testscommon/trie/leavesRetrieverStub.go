package trie

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
)

// LeavesRetrieverStub -
type LeavesRetrieverStub struct {
	GetLeavesCalled func(numLeaves int, iteratorState [][]byte, leavesParser common.TrieLeafParser, ctx context.Context) (map[string]string, [][]byte, error)
}

// GetLeaves -
func (lrs *LeavesRetrieverStub) GetLeaves(numLeaves int, iteratorState [][]byte, leavesParser common.TrieLeafParser, ctx context.Context) (map[string]string, [][]byte, error) {
	if lrs.GetLeavesCalled != nil {
		return lrs.GetLeavesCalled(numLeaves, iteratorState, leavesParser, ctx)
	}
	return nil, nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lrs *LeavesRetrieverStub) IsInterfaceNil() bool {
	return lrs == nil
}
