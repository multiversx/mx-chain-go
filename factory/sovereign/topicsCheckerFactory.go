package sovereign

import (
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
)

type topicsCheckerFactory struct {
}

// NewTopicsCheckerFactory creates a new topics checker factory
func NewTopicsCheckerFactory() *topicsCheckerFactory {
	return &topicsCheckerFactory{}
}

// CreateTopicsChecker creates a new topics checker for the chain run type normal
func (tcf *topicsCheckerFactory) CreateTopicsChecker() sovereign.TopicsCheckerHandler {
	return disabled.NewDisabledTopicsChecker()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tcf *topicsCheckerFactory) IsInterfaceNil() bool {
	return tcf == nil
}
