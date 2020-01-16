package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

// StorageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapperHandler interface {
	getHeader(hash []byte) (data.HeaderHandler, error)
	getBlockBody(header data.HeaderHandler) (data.BodyHandler, error)
	applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error
	applySelfNotarizedHeaders(selfNotarizedHeadersHashes [][]byte) ([]data.HeaderHandler, error)
	cleanupNotarizedStorage(hash []byte)
	IsInterfaceNil() bool
}
