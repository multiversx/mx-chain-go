package persister

import (
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
)

// NewPersisterFactory -
func NewPersisterFactory() storage.PersisterFactoryHandler {
	return factory.NewPersisterFactoryHandler(2, 1)
}
