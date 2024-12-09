package persister

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/storage"
)

// PersistentStatusHandler defines a persistent status handler
type PersistentStatusHandler interface {
	core.AppStatusHandler
	SetStorage(store storage.Storer) error
}
