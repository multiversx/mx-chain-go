package factory

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// BootstrapDataProviderHandler defines which actions should be done for loading bootstrap data from the boot storer
type BootstrapDataProviderHandler interface {
	LoadForPath(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error)
	GetStorer(storer storage.Storer) (process.BootStorer, error)
	IsInterfaceNil() bool
}
