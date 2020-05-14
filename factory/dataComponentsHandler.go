package factory

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

var _ ComponentHandler = (*ManagedDataComponents)(nil)
var _ DataComponentsHolder = (*ManagedDataComponents)(nil)
var _ DataComponentsHandler = (*ManagedDataComponents)(nil)

// DataComponentsHandlerArgs holds the arguments required to create a crypto components handler
type DataComponentsHandlerArgs DataComponentsFactoryArgs

// ManagedDataComponents creates the data components handler that can create, close and access the data components
type ManagedDataComponents struct {
	*DataComponents
	dataComponentsFactory *dataComponentsFactory
	cancelFunc            func()
	mutDataComponents     sync.RWMutex
}

// NewManagedDataComponents creates a new data components handler
func NewManagedDataComponents(args DataComponentsHandlerArgs) (*ManagedDataComponents, error) {
	dcf, err := NewDataComponentsFactory(DataComponentsFactoryArgs(args))
	if err != nil {
		return nil, err
	}

	return &ManagedDataComponents{
		DataComponents:        nil,
		dataComponentsFactory: dcf,
	}, nil
}

// Create creates the data components
func (mdc *ManagedDataComponents) Create() error {
	cryptoComponents, err := mdc.dataComponentsFactory.Create()
	if err != nil {
		return err
	}

	mdc.mutDataComponents.Lock()
	mdc.DataComponents = cryptoComponents
	_, mdc.cancelFunc = context.WithCancel(context.Background())
	mdc.mutDataComponents.Unlock()

	return nil
}

// Close closes the data components
func (mdc *ManagedDataComponents) Close() error {
	mdc.mutDataComponents.Lock()
	mdc.cancelFunc()
	mdc.cancelFunc = nil
	mdc.DataComponents = nil
	mdc.mutDataComponents.Unlock()

	return nil
}

// Blockchain returns the blockchain handler
func (mdc *ManagedDataComponents) Blockchain() data.ChainHandler {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Blkc
}

// StorageService returns the storage service
func (mdc *ManagedDataComponents) StorageService() dataRetriever.StorageService {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Store
}

// Datapool returns the Datapool
func (mdc *ManagedDataComponents) Datapool() dataRetriever.PoolsHolder {
	mdc.mutDataComponents.RLock()
	defer mdc.mutDataComponents.RUnlock()

	if mdc.DataComponents == nil {
		return nil
	}

	return mdc.DataComponents.Datapool
}

// IsInterfaceNil returns true if the interface is nil
func (mdc *ManagedDataComponents) IsInterfaceNil() bool {
	return mdc == nil
}
