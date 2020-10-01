package factory

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
)

//TODO remove this
var log = logger.GetOrCreate("main")

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	RegisterForEpochChangeConfirmed(handler func(epoch uint32))
	NotifyEpochChangeConfirmed(epoch uint32)
	IsInterfaceNil() bool
}

// CreateLatestStorageDataProvider will create a latest storage data provider handler
func CreateLatestStorageDataProvider(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	generalConfig config.Config,
	parentDir string,
	defaultEpochString string,
	defaultShardString string,
) (storage.LatestStorageDataProviderHandler, error) {
	directoryReader := storageFactory.NewDirectoryReader()

	latestStorageDataArgs := storageFactory.ArgsLatestDataProvider{
		GeneralConfig:         generalConfig,
		BootstrapDataProvider: bootstrapDataProvider,
		DirectoryReader:       directoryReader,
		ParentDir:             parentDir,
		DefaultEpochString:    defaultEpochString,
		DefaultShardString:    defaultShardString,
	}
	return storageFactory.NewLatestDataProvider(latestStorageDataArgs)
}

// CreateUnitOpener will create a new unit opener handler
func CreateUnitOpener(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	latestDataFromStorageProvider storage.LatestStorageDataProviderHandler,
	generalConfig config.Config,
	defaultEpochString string,
	defaultShardString string,
) (storage.UnitOpenerHandler, error) {
	argsStorageUnitOpener := storageFactory.ArgsNewOpenStorageUnits{
		GeneralConfig:             generalConfig,
		BootstrapDataProvider:     bootstrapDataProvider,
		LatestStorageDataProvider: latestDataFromStorageProvider,
		DefaultEpochString:        defaultEpochString,
		DefaultShardString:        defaultShardString,
	}

	return storageFactory.NewStorageUnitOpenHandler(argsStorageUnitOpener)
}
