package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ dataRetriever.RequestedItemsHandler = (*disabledBlacklistHandler)(nil)
var _ process.BlackListHandler = (*disabledBlacklistHandler)(nil)

// disabledBlacklistHandler is a mock implementation of BlacklistHandler that does not manage black listed keys
// (all keys [peers] are whitelisted)
type disabledBlacklistHandler struct {
}

// Add does nothing
func (nbh *disabledBlacklistHandler) Add(_ string) error {
	return nil
}

// AddWithSpan does nothing
func (nbh *disabledBlacklistHandler) AddWithSpan(_ string, _ time.Duration) error {
	return nil
}

// Sweep does nothing
func (nbh *disabledBlacklistHandler) Sweep() {
}

// Has outputs false (all peers are white listed)
func (nbh *disabledBlacklistHandler) Has(_ string) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nbh *disabledBlacklistHandler) IsInterfaceNil() bool {
	return nbh == nil
}
