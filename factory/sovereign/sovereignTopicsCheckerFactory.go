package sovereign

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignTopicsCheckerFactory struct {
	topicsChecker sovereign.TopicsCheckerHandler
}

// NewSovereignTopicsCheckerFactory creates a new sovereign topics checker factory
func NewSovereignTopicsCheckerFactory(tc sovereign.TopicsCheckerHandler) (*sovereignTopicsCheckerFactory, error) {
	if check.IfNil(tc) {
		return nil, errors.ErrNilTopicsChecker
	}

	return &sovereignTopicsCheckerFactory{
		topicsChecker: tc,
	}, nil
}

// CreateTopicsChecker creates a new topics checker for the chain run type sovereign
func (stcf *sovereignTopicsCheckerFactory) CreateTopicsChecker() sovereign.TopicsCheckerHandler {
	return stcf.topicsChecker
}

// IsInterfaceNil returns true if there is no value under the interface
func (stcf *sovereignTopicsCheckerFactory) IsInterfaceNil() bool {
	return stcf == nil
}
