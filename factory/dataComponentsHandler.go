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
	*DataComponents
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
		DataComponents:        nil,
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
	mdc.DataComponents = dc
	mdc.mutDataComponents.Unlock()

	return nil
}

// Close closes the data components
func (mdc *managedDataComponents) Close() error {
	mdc.mutDataComponents.Lock()
	defer mdc.mutDataComponents.Unlock()

	err := mdc.DataComponents.Store.CloseAll()
	if err != nil {
		return err
	}

	mdc.DataComponents = nil

	return nil
}

// Blockchain returns the blockchain handler
func (mdc *managedDataComponents) Blockchain() data.ChainHandler {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.Blkc
}

func (mdc *managedDataComponents) SetBlockchain(chain data.ChainHandler) {
	mdc.mutDataComponents.Lock()
	mdc.Blkc = chain
	mdc.mutDataComponents.Unlock()
}

// StorageService returns the storage service
func (mdc *managedDataComponents) StorageService() dataRetriever.StorageService {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.Store
}

// Datapool returns the Datapool
func (mdc *managedDataComponents) Datapool() dataRetriever.PoolsHolder {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Datapool
}

// Clone creates a shallow clone of a managedDataComponents
func (mdc *managedDataComponents) Clone() interface{} {
	dataComponents := (*DataComponents)(nil)
	if mdc.DataComponents != nil {
		dataComponents = &DataComponents{
			Blkc:     mdc.Blockchain(),
			Store:    mdc.StorageService(),
			Datapool: mdc.Datapool(),
		}
	}

	return &managedDataComponents{
		DataComponents:        dataComponents,
		dataComponentsFactory: mdc.dataComponentsFactory,
		mutDataComponents:     sync.RWMutex{},
	}
}

// IsInterfaceNil returns true if the interface is nil
func (mdc *managedDataComponents) IsInterfaceNil() bool {
	return mdc == nil
}
