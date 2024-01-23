package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

func (st *storageBootstrapper) sovereignChainGetScheduledRootHash(headerFromStorage data.HeaderHandler, _ []byte) []byte {
	return headerFromStorage.GetRootHash()
}

func (st *storageBootstrapper) sovereignChainSetScheduledInfo(_ bootstrapStorage.BootstrapData) {
}
