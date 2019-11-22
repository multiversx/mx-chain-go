package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/sync"
)

// StorageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapperHandler interface {
	getHeader(hash []byte) (data.HeaderHandler, error)
	getBlockBody(header data.HeaderHandler) (data.BodyHandler, error)
	applyNotarizedBlocks(lastNotarized map[uint32]*sync.HdrInfo) error
	cleanupNotarizedStorage(lastNotarized map[uint32]*sync.HdrInfo)
	IsInterfaceNil() bool
}
