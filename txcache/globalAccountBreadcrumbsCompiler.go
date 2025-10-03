package txcache

import "sync"

type globalAccountBreadcrumbsCompiler struct {
	// TODO analyze if this mutex is needed
	mutCompiler              sync.RWMutex
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb
}

func newGlobalAccountBreadcrumbsCompiler() *globalAccountBreadcrumbsCompiler {
	return &globalAccountBreadcrumbsCompiler{
		mutCompiler:              sync.RWMutex{},
		globalAccountBreadcrumbs: make(map[string]*globalAccountBreadcrumb),
	}
}

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
