package factory

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

var _ ComponentHandler = (*managedDataComponents)(nil)
var _ DataComponentsHolder = (*managedDataComponents)(nil)
var _ DataComponentsHandler = (*managedDataComponents)(nil)

// DataComponentsHandlerArgs holds the arguments required to create a crypto components handler
type DataComponentsHandlerArgs DataComponentsFactoryArgs

// managedDataComponents creates the data components handler that can create, close and access the data components
type managedDataComponents struct {
	*dataComponents
	dataComponentsFactory *dataComponentsFactory
	mutDataComponents     sync.RWMutex
}

// NewManagedDataComponents creates a new data components handler
func NewManagedDataComponents(args DataComponentsHandlerArgs) (*managedDataComponents, error) {
	dcf, err := NewDataComponentsFactory(DataComponentsFactoryArgs(args))
	if err != nil {
		return nil, err
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
		return err
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

// Blockchain returns the blockchain handler
func (mdc *managedDataComponents) Blockchain() data.ChainHandler {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.dataComponents == nil {
		return nil
	}

	return mdc.blkc
}

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

// Clone creates a shallow clone of a managedDataComponents
func (mdc *managedDataComponents) Clone() interface{} {
	dataComps := (*dataComponents)(nil)
	if mdc.dataComponents != nil {
		dataComps = &dataComponents{
			blkc:     mdc.Blockchain(),
			store:    mdc.StorageService(),
			datapool: mdc.Datapool(),
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
