package evictionWaitingList

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type inactiveEvictionWaitingList struct {
}

// NewInactiveEvictionWaitingList creates a new inactiveEvictionWaitingList
func NewInactiveEvictionWaitingList() *inactiveEvictionWaitingList {
	return &inactiveEvictionWaitingList{}
}

// Put does nothing for this implementation
func (i *inactiveEvictionWaitingList) Put(_ []byte, _ data.ModifiedHashes) error {
	return nil
}

// Evict returns an empty ModifiedHashes map
func (i *inactiveEvictionWaitingList) Evict(_ []byte) (data.ModifiedHashes, error) {
	return map[string]struct{}{}, nil
}

// ShouldKeepHash always returns true for this implementation
func (i *inactiveEvictionWaitingList) ShouldKeepHash(_ string, _ data.TriePruningIdentifier) (bool, error) {
	return true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (i *inactiveEvictionWaitingList) IsInterfaceNil() bool {
	return i == nil
}
