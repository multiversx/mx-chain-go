package storagePruningManager

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewInactiveStoragePruningManager(t *testing.T) {
	t.Parallel()

	ispm := NewInactiveStoragePruningManager()
	assert.False(t, check.IfNil(ispm))
}

func TestStoragePruningManager_MarkForEviction(t *testing.T) {
	t.Parallel()

	ispm := NewInactiveStoragePruningManager()
	err := ispm.MarkForEviction(nil, nil, nil, nil)
	assert.Nil(t, err)
}

func TestInactiveStoragePruningManager_PruneTrie(t *testing.T) {
	t.Parallel()

	ispm := NewInactiveStoragePruningManager()
	ispm.PruneTrie(nil, 0, nil)
}

func TestInactiveStoragePruningManager_CancelPrune(t *testing.T) {
	t.Parallel()

	ispm := NewInactiveStoragePruningManager()
	ispm.CancelPrune(nil, 0, nil)
}
