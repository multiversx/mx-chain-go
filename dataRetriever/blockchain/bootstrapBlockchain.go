package blockchain

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

type bootstrapBlockchain struct {
	currentBlockHeader data.HeaderHandler
	mut                sync.RWMutex
}

// NewBootstrapBlockchain returns a new instance of bootstrapBlockchain
// It should be used for bootstrap only!
func NewBootstrapBlockchain() *bootstrapBlockchain {
	return &bootstrapBlockchain{}
}

// GetCurrentBlockHeader returns the current block header
func (bbc *bootstrapBlockchain) GetCurrentBlockHeader() data.HeaderHandler {
	bbc.mut.RLock()
	defer bbc.mut.RUnlock()
	return bbc.currentBlockHeader
}

// SetCurrentBlockHeaderAndRootHash returns nil always and saves the current block header
func (bbc *bootstrapBlockchain) SetCurrentBlockHeaderAndRootHash(bh data.HeaderHandler, _ []byte) error {
	bbc.mut.Lock()
	defer bbc.mut.Unlock()

	bbc.currentBlockHeader = bh
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bbc *bootstrapBlockchain) IsInterfaceNil() bool {
	return bbc == nil
}
