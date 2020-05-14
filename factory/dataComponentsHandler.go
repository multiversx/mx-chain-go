package factory

import (
	"context"
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
	cancelFunc            func()
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
	_, mdc.cancelFunc = context.WithCancel(context.Background())
	mdc.mutDataComponents.Unlock()

	return nil
}

// Close closes the data components
func (mdc *managedDataComponents) Close() error {
	mdc.mutDataComponents.Lock()
	mdc.cancelFunc()
	mdc.cancelFunc = nil
	mdc.DataComponents = nil
	mdc.mutDataComponents.Unlock()

	return nil
}

// Blockchain returns the blockchain handler
func (mdc *managedDataComponents) Blockchain() data.ChainHandler {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Blkc
}

// StorageService returns the storage service
func (mdc *managedDataComponents) StorageService() dataRetriever.StorageService {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Store
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

// IsInterfaceNil returns true if the interface is nil
func (mdc *managedDataComponents) IsInterfaceNil() bool {
	return mdc == nil
}
