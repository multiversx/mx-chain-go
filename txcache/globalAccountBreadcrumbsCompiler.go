package txcache

import (
	"sync"
)

// globalAccountBreadcrumbsCompiler represents the global account breadcrumbs compiler used in the Selection Tracker.
// A globalAccountBreadcrumbsCompiler holds a globalAccountBreadcrumb for each account.
type globalAccountBreadcrumbsCompiler struct {
	// TODO analyze if this mutex is needed
	mutCompiler              sync.RWMutex
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb
}

// newGlobalAccountBreadcrumbsCompiler creates a new global account breadcrumb compiler
func newGlobalAccountBreadcrumbsCompiler() *globalAccountBreadcrumbsCompiler {
	return &globalAccountBreadcrumbsCompiler{
		mutCompiler:              sync.RWMutex{},
		globalAccountBreadcrumbs: make(map[string]*globalAccountBreadcrumb),
	}
}

// updateOnAddedBlock updates the global state of the account when a block is added on the OnProposedBlock flow
func (gabc *globalAccountBreadcrumbsCompiler) updateOnAddedBlock(tb *trackedBlock) {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			globalBreadcrumb = newGlobalAccountBreadcrumb()
			gabc.globalAccountBreadcrumbs[account] = globalBreadcrumb
		}

		globalBreadcrumb.updateOnAddedBreadcrumb(breadcrumb)
	}
}

// updateAfterRemovedBlockWithSameNonceOrAbove updates the global state of the account when a block is removed on the OnProposedBlock flow
func (gabc *globalAccountBreadcrumbsCompiler) updateAfterRemovedBlockWithSameNonceOrAbove(tb *trackedBlock) error {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			return errGlobalBreadcrumbDoesNotExist
		}

		shouldBeDeleted, err := globalBreadcrumb.updateOnRemoveBreadcrumbWithSameNonceOrAbove(breadcrumb)
		if err != nil {
			return err
		}

		if shouldBeDeleted {
			delete(gabc.globalAccountBreadcrumbs, account)
		}
	}

	return nil
}

// updateAfterRemovedBlockWithSameNonceOrBelow updates the global state of the account when a block is removed on the OnExecutedBlock flow
func (gabc *globalAccountBreadcrumbsCompiler) updateAfterRemovedBlockWithSameNonceOrBelow(tb *trackedBlock) error {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			return errGlobalBreadcrumbDoesNotExist
		}

		shouldBeDeleted, err := globalBreadcrumb.updateOnRemovedBreadcrumbWithSameNonceOrBelow(breadcrumb)
		if err != nil {
			return err
		}

		if shouldBeDeleted {
			delete(gabc.globalAccountBreadcrumbs, account)
		}
	}

	return nil
}

// getGlobalBreadcrumbByAddress returns a deep copy of the global breadcrumb of a certain address
func (gabc *globalAccountBreadcrumbsCompiler) getGlobalBreadcrumbByAddress(address string) (*globalAccountBreadcrumb, error) {
	gabc.mutCompiler.RLock()
	defer gabc.mutCompiler.RUnlock()

	_, ok := gabc.globalAccountBreadcrumbs[address]
	if !ok {
		return nil, errGlobalBreadcrumbDoesNotExist
	}

	return gabc.globalAccountBreadcrumbs[address].createCopy(), nil
}

// getGlobalBreadcrumbs returns a deep copy of the map of global accounts breadcrumbs
func (gabc *globalAccountBreadcrumbsCompiler) getGlobalBreadcrumbs() map[string]*globalAccountBreadcrumb {
	gabc.mutCompiler.RLock()
	defer gabc.mutCompiler.RUnlock()

	globalBreadcrumbsCopy := make(map[string]*globalAccountBreadcrumb)
	for account, globalBreadcrumb := range gabc.globalAccountBreadcrumbs {
		globalBreadcrumbsCopy[account] = globalBreadcrumb.createCopy()
	}

	return globalBreadcrumbsCopy
}

// cleanGlobalBreadcrumbs resets the global accounts breadcrumbs
func (gabc *globalAccountBreadcrumbsCompiler) cleanGlobalBreadcrumbs() {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	gabc.globalAccountBreadcrumbs = make(map[string]*globalAccountBreadcrumb)
}
