package factory

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var _ ComponentHandler = (*managedDataComponents)(nil)
var _ DataComponentsHolder = (*managedDataComponents)(nil)
var _ DataComponentsHandler = (*managedDataComponents)(nil)

// managedDataComponents creates the data components handler that can create, close and access the data components
type managedDataComponents struct {
	*dataComponents
	dataComponentsFactory *dataComponentsFactory
	mutDataComponents     sync.RWMutex
}

// NewManagedDataComponents creates a new data components handler
func NewManagedDataComponents(dcf *dataComponentsFactory) (*managedDataComponents, error) {
	if dcf == nil {
		return nil, errors.ErrNilDataComponentsFactory
	}

	return &managedDataComponents{
		dataComponents:        nil,
		dataComponentsFactory: dcf,
	}, nil
}

// Create creates the data components
func (mdc *managedDataComponents) Create() error {
	dc, err := mdc.dataComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDataComponentsFactoryCreate, err)
	}

	mdc.mutDataComponents.Lock()
	mdc.dataComponents = dc
	mdc.mutDataComponents.Unlock()

	return nil
}

// Close closes the data components
func (mdc *managedDataComponents) Close() error {
	mdc.mutDataComponents.Lock()
	defer mdc.mutDataComponents.Unlock()

	if mdc.dataComponents != nil {
		err := mdc.dataComponents.Close()
		if err != nil {
			return err
		}
		mdc.dataComponents = nil
	}

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mdc *managedDataComponents) CheckSubcomponents() error {
	mdc.mutDataComponents.Lock()
	defer mdc.mutDataComponents.Unlock()

	if mdc.dataComponents == nil {
		return errors.ErrNilDataComponents
	}
	if check.IfNil(mdc.blkc) {
		return errors.ErrNilBlockChainHandler
	}
	if check.IfNil(mdc.store) {
		return errors.ErrNilStorageService
	}
	if check.IfNil(mdc.datapool) {
		return errors.ErrNilPoolsHolder
	}
	if check.IfNil(mdc.miniBlocksProvider) {
		return errors.ErrNilMiniBlocksProvider
	}

	return nil
}

// Blockchain returns the blockchain handler
func (mdc *managedDataComponents) Blockchain() data.ChainHandler {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.dataComponents == nil {
		return nil
	}

	return mdc.blkc
}

// SetBlockchain sets the blockchain subcomponent
func (mdc *managedDataComponents) SetBlockchain(chain data.ChainHandler) {
	mdc.mutDataComponents.Lock()
	mdc.blkc = chain
	mdc.mutDataComponents.Unlock()
}

// StorageService returns the storage service
func (mdc *managedDataComponents) StorageService() dataRetriever.StorageService {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.dataComponents == nil {
		return nil
	}

	return mdc.store
}

// Datapool returns the Datapool
func (mdc *managedDataComponents) Datapool() dataRetriever.PoolsHolder {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.dataComponents == nil {
		return nil
	}

	return mdc.dataComponents.datapool
}

// MiniBlocksProvider returns the MiniBlockProvider
func (mdc *managedDataComponents) MiniBlocksProvider() MiniBlockProvider {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.dataComponents == nil {
		return nil
	}

	return mdc.dataComponents.miniBlocksProvider
}

// Clone creates a shallow clone of a managedDataComponents
func (mdc *managedDataComponents) Clone() interface{} {
	dataComps := (*dataComponents)(nil)
	if mdc.dataComponents != nil {
		dataComps = &dataComponents{
			blkc:               mdc.Blockchain(),
			store:              mdc.StorageService(),
			datapool:           mdc.Datapool(),
			miniBlocksProvider: mdc.MiniBlocksProvider(),
		}
	}

	return &managedDataComponents{
		dataComponents:        dataComps,
		dataComponentsFactory: mdc.dataComponentsFactory,
		mutDataComponents:     sync.RWMutex{},
	}
}

// IsInterfaceNil returns true if the interface is nil
func (mdc *managedDataComponents) IsInterfaceNil() bool {
	return mdc == nil
}

// String returns the name of the component
func (mdc *managedDataComponents) String() string {
	return "managedDataComponents"
}
