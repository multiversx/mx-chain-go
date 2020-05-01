package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type bootstrapDataProvider struct {
	marshalizer marshal.Marshalizer
}

// NewBootstrapDataProvider returns a new instance of bootstrapDataProvider
func NewBootstrapDataProvider(marshalizer marshal.Marshalizer) (*bootstrapDataProvider, error) {
	if check.IfNil(marshalizer) {
		return nil, storage.ErrNilMarshalizer
	}
	return &bootstrapDataProvider{
		marshalizer: marshalizer,
	}, nil
}

// LoadForPath returns the bootstrap data and the storer for the given persister factory and path
func (bdp *bootstrapDataProvider) LoadForPath(
	persisterFactory storage.PersisterFactory,
	path string,
) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
	persister, err := persisterFactory.Create(path)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			errClose := persister.Close()
			if errClose != nil {
				log.Debug("encountered a non-critical error closing bootstrap persister",
					"error", errClose,
				)
			}
		}
	}()

	cacher, err := lrucache.NewCache(10)
	if err != nil {
		return nil, nil, err
	}

	storer, err := storageUnit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, nil, err
	}

	bootStorer, err := bdp.GetStorer(storer)
	if err != nil {
		return nil, nil, err
	}

	highestRound := bootStorer.GetHighestRound()
	bootstrapData, err := bootStorer.Get(highestRound)
	if err != nil {
		return nil, nil, err
	}

	return &bootstrapData, storer, nil
}

// GetStorer returns the bootstorer
func (bdp *bootstrapDataProvider) GetStorer(storer storage.Storer) (process.BootStorer, error) {
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(bdp.marshalizer, storer)
	if err != nil {
		return nil, err
	}

	return bootStorer, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bdp *bootstrapDataProvider) IsInterfaceNil() bool {
	return bdp == nil
}
