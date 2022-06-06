package state

import (
	"bytes"
	"fmt"
	"sync"
)

type accountsDBApiTrieController struct {
	innerAccountsAdapter AccountsAdapter
	latestRootHash       []byte
	mutex                sync.RWMutex
}

func newAccountsDBApiTrieController(innerAccountsAdapter AccountsAdapter) *accountsDBApiTrieController {
	return &accountsDBApiTrieController{
		innerAccountsAdapter: innerAccountsAdapter,
	}
}

func (controller *accountsDBApiTrieController) recreateTrieIfNecessary(targetRootHash []byte) error {
	if len(targetRootHash) == 0 {
		return fmt.Errorf("%w in accountsDBApiTrieController", ErrNilRootHash)
	}

	controller.mutex.RLock()
	lastRootHash := controller.latestRootHash
	controller.mutex.RUnlock()

	if bytes.Equal(lastRootHash, targetRootHash) {
		return nil
	}

	return controller.doRecreateTrie(targetRootHash)
}

func (controller *accountsDBApiTrieController) doRecreateTrie(targetRootHash []byte) error {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	// early exit for possible multiple re-entrances here
	lastRootHash := controller.latestRootHash
	if bytes.Equal(lastRootHash, targetRootHash) {
		return nil
	}

	err := controller.innerAccountsAdapter.RecreateTrie(targetRootHash)
	if err != nil {
		controller.latestRootHash = nil
		return err
	}

	controller.latestRootHash = targetRootHash

	return nil
}

func (controller *accountsDBApiTrieController) getLatestRootHash() ([]byte, error) {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	if controller.latestRootHash == nil {
		return nil, ErrNilRootHash
	}

	return controller.latestRootHash, nil
}
