package sync

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/tree"
)

type hardforkExclusionHandler struct {
	tree tree.IntervalTree
}

// NewHardforkExclusionHandler returns a new instance of hardforkExclusionHandler
func NewHardforkExclusionHandler(exclusionTree tree.IntervalTree) (*hardforkExclusionHandler, error) {
	if check.IfNil(exclusionTree) {
		return nil, ErrNilExclusionTree
	}

	return &hardforkExclusionHandler{
		tree: exclusionTree,
	}, nil
}

// IsRoundExcluded returns true if the provided round is excluded
func (handler *hardforkExclusionHandler) IsRoundExcluded(round uint64) bool {
	return handler.tree.Contains(round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *hardforkExclusionHandler) IsInterfaceNil() bool {
	return handler == nil
}
