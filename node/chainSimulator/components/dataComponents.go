package components

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/provider"
	"github.com/multiversx/mx-chain-go/factory"
)

// ArgsDataComponentsHolder will hold the components needed for data components
type ArgsDataComponentsHolder struct {
	Chain              data.ChainHandler
	StorageService     dataRetriever.StorageService
	DataPool           dataRetriever.PoolsHolder
	InternalMarshaller marshal.Marshalizer
}

type dataComponentsHolder struct {
	closeHandler      *closeHandler
	chain             data.ChainHandler
	storageService    dataRetriever.StorageService
	dataPool          dataRetriever.PoolsHolder
	miniBlockProvider factory.MiniBlockProvider
}

// CreateDataComponents will create the data components holder
func CreateDataComponents(args ArgsDataComponentsHolder) (*dataComponentsHolder, error) {
	miniBlockStorer, err := args.StorageService.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    args.DataPool.MiniBlocks(),
		MiniBlockStorage: miniBlockStorer,
		Marshalizer:      args.InternalMarshaller,
	}

	miniBlocksProvider, err := provider.NewMiniBlockProvider(arg)
	if err != nil {
		return nil, err
	}

	instance := &dataComponentsHolder{
		closeHandler:      NewCloseHandler(),
		chain:             args.Chain,
		storageService:    args.StorageService,
		dataPool:          args.DataPool,
		miniBlockProvider: miniBlocksProvider,
	}

	instance.collectClosableComponents()

	return instance, nil
}

// Blockchain will return the blockchain handler
func (d *dataComponentsHolder) Blockchain() data.ChainHandler {
	return d.chain
}

// SetBlockchain will set the blockchain handler
func (d *dataComponentsHolder) SetBlockchain(chain data.ChainHandler) error {
	d.chain = chain

	return nil
}

// StorageService will return the storage service
func (d *dataComponentsHolder) StorageService() dataRetriever.StorageService {
	return d.storageService
}

// Datapool will return the data pool
func (d *dataComponentsHolder) Datapool() dataRetriever.PoolsHolder {
	return d.dataPool
}

// MiniBlocksProvider will return the mini blocks provider
func (d *dataComponentsHolder) MiniBlocksProvider() factory.MiniBlockProvider {
	return d.miniBlockProvider
}

// Clone will clone the data components holder
func (d *dataComponentsHolder) Clone() interface{} {
	return &dataComponentsHolder{
		chain:             d.chain,
		storageService:    d.storageService,
		dataPool:          d.dataPool,
		miniBlockProvider: d.miniBlockProvider,
		closeHandler:      d.closeHandler,
	}
}

func (d *dataComponentsHolder) collectClosableComponents() {
	d.closeHandler.AddComponent(d.storageService)
	d.closeHandler.AddComponent(d.dataPool)
}

// Close will call the Close methods on all inner components
func (d *dataComponentsHolder) Close() error {
	return d.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *dataComponentsHolder) IsInterfaceNil() bool {
	return d == nil
}

// Create will do nothing
func (d *dataComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (d *dataComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (d *dataComponentsHolder) String() string {
	return ""
}
