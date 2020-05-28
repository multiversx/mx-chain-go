package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ dataRetriever.RequestedItemsHandler = (*BlacklistHandler)(nil)
var _ process.BlackListHandler = (*BlacklistHandler)(nil)

// BlacklistHandler is a mock implementation of BlacklistHandler that does not manage black listed keys
// (all keys [peers] are whitelisted)
type BlacklistHandler struct {
}

// Add does nothing
func (dbh *BlacklistHandler) Add(_ string) error {
	return nil
}

// AddWithSpan does nothing
func (dbh *BlacklistHandler) AddWithSpan(_ string, _ time.Duration) error {
	return nil
}

// Sweep does nothing
func (dbh *BlacklistHandler) Sweep() {
}

// Has outputs false (all peers are white listed)
func (dbh *BlacklistHandler) Has(_ string) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (dbh *BlacklistHandler) IsInterfaceNil() bool {
	return dbh == nil
}
