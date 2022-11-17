package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

func (st *storageBootstrapper) sovereignChainGetScheduledRootHash(headerFromStorage data.HeaderHandler, _ []byte) []byte {
	return headerFromStorage.GetRootHash()
}

func (st *storageBootstrapper) sovereignChainSetScheduledInfo(headerInfo bootstrapStorage.BootstrapData) {
}
