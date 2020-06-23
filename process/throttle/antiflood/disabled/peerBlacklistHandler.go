package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.PeerBlackListHandler = (*PeerBlacklistHandler)(nil)

// PeerBlacklistHandler is a mock implementation of PeerBlacklistHandler that does not manage black listed keys
// (all keys [peers] are whitelisted)
type PeerBlacklistHandler struct {
}

// Add does nothing
func (pdbh *PeerBlacklistHandler) Add(_ core.PeerID) error {
	return nil
}

// AddWithSpan does nothing
func (pdbh *PeerBlacklistHandler) AddWithSpan(_ core.PeerID, _ time.Duration) error {
	return nil
}

// Update does nothing
func (pdbh *PeerBlacklistHandler) Update(_ core.PeerID, _ time.Duration) error {
	return nil
}

// Sweep does nothing
func (pdbh *PeerBlacklistHandler) Sweep() {
}

// Has outputs false (all peers are white listed)
func (pdbh *PeerBlacklistHandler) Has(_ core.PeerID) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (pdbh *PeerBlacklistHandler) IsInterfaceNil() bool {
	return pdbh == nil
}
