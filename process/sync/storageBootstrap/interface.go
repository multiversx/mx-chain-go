package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// StorageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapperHandler interface {
	getHeader(hash []byte) (data.HeaderHandler, error)
	getBlockBody(header data.HeaderHandler) (data.BodyHandler, error)
	applyCrossNotarizedHeaders(crossNotarizedHeadersHashes map[uint32][]byte) error
	applySelfNotarizedHeaders(selfNotarizedHeadersHashes [][]byte) ([]data.HeaderHandler, error)
	cleanupNotarizedStorage(hash []byte)
	IsInterfaceNil() bool
}
