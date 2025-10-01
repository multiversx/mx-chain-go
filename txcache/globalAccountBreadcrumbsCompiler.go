package txcache

import (
	"sync"
)

type globalAccountBreadcrumbsCompiler struct {
	mutCompiler              sync.RWMutex
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb
}

func newGlobalAccountBreadcrumbsCompiler() *globalAccountBreadcrumbsCompiler {
	return &globalAccountBreadcrumbsCompiler{
		mutCompiler:              sync.RWMutex{},
		globalAccountBreadcrumbs: make(map[string]*globalAccountBreadcrumb),
	}
}

func (gabc *globalAccountBreadcrumbsCompiler) updateGlobalBreadcrumbsOnAddedBlockOnProposed(tb *trackedBlock) {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			globalBreadcrumb = newGlobalAccountBreadcrumb()
			gabc.globalAccountBreadcrumbs[account] = globalBreadcrumb
		}

		globalBreadcrumb.updateOnAddedAccountBreadcrumb(breadcrumb)
	}
}

func (gabc *globalAccountBreadcrumbsCompiler) updateGlobalBreadcrumbsOnRemovedBlockOnProposed(tb *trackedBlock) error {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			return errGlobalBreadcrumbDoesNotExist
		}

		shouldBeDeleted, err := globalBreadcrumb.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb)
		if err != nil {
			return err
		}

		if shouldBeDeleted {
			delete(gabc.globalAccountBreadcrumbs, account)
		}
	}

	return nil
}

func (gabc *globalAccountBreadcrumbsCompiler) updateGlobalBreadcrumbsOnRemovedBlockOnExecuted(tb *trackedBlock) error {
	gabc.mutCompiler.Lock()
	defer gabc.mutCompiler.Unlock()

	breadcrumbsOfTrackedBlock := tb.breadcrumbsByAddress
	for account, breadcrumb := range breadcrumbsOfTrackedBlock {
		globalBreadcrumb, ok := gabc.globalAccountBreadcrumbs[account]
		if !ok {
			return errGlobalBreadcrumbDoesNotExist
		}

		shouldBeDeleted, err := globalBreadcrumb.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb)
		if err != nil {
			return err
		}

		if shouldBeDeleted {
			delete(gabc.globalAccountBreadcrumbs, account)
		}
	}

	return nil
}

func (gabc *globalAccountBreadcrumbsCompiler) getGlobalBreadcrumbByAddress(address string) (*globalAccountBreadcrumb, error) {
	gabc.mutCompiler.RLock()
	defer gabc.mutCompiler.RUnlock()

	_, ok := gabc.globalAccountBreadcrumbs[address]
	if !ok {
		return nil, errGlobalBreadcrumbDoesNotExist
	}

	return gabc.globalAccountBreadcrumbs[address].createCopy(), nil
}

func (gabc *globalAccountBreadcrumbsCompiler) getGlobalBreadcrumbs() map[string]*globalAccountBreadcrumb {
	gabc.mutCompiler.RLock()
	defer gabc.mutCompiler.RUnlock()

	globalBreadcrumbsCopy := make(map[string]*globalAccountBreadcrumb)
	for account, globalBreadcrumb := range gabc.globalAccountBreadcrumbs {
		globalBreadcrumbsCopy[account] = globalBreadcrumb.createCopy()
	}

	return globalBreadcrumbsCopy
}
