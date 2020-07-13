package factory

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion"
	factorySoftwareVersion "github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

// CreateSoftwareVersionChecker will create a new software version checker and will start check if a new software version
// is available
func CreateSoftwareVersionChecker(
	statusHandler core.AppStatusHandler,
	config config.SoftwareVersionConfig,
) (*softwareVersion.SoftwareVersionChecker, error) {
	softwareVersionCheckerFactory, err := factorySoftwareVersion.NewSoftwareVersionFactory(statusHandler, config)
	if err != nil {
		return nil, err
	}

	softwareVersionChecker, err := softwareVersionCheckerFactory.Create()
	if err != nil {
		return nil, err
	}

	return softwareVersionChecker, nil
}

// CreateLatestStorageDataProvider will create a latest storage data provider handler
func CreateLatestStorageDataProvider(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	generalConfig config.Config,
	chainID string,
	workingDir string,
	defaultDBPath string,
	defaultEpochString string,
	defaultShardString string,
) (storage.LatestStorageDataProviderHandler, error) {
	directoryReader := storageFactory.NewDirectoryReader()

	latestStorageDataArgs := storageFactory.ArgsLatestDataProvider{
		GeneralConfig:         generalConfig,
		Marshalizer:           marshalizer,
		Hasher:                hasher,
		BootstrapDataProvider: bootstrapDataProvider,
		DirectoryReader:       directoryReader,
		WorkingDir:            workingDir,
		ChainID:               chainID,
		DefaultDBPath:         defaultDBPath,
		DefaultEpochString:    defaultEpochString,
		DefaultShardString:    defaultShardString,
	}
	return storageFactory.NewLatestDataProvider(latestStorageDataArgs)
}

// CreateUnitOpener will create a new unit opener handler
func CreateUnitOpener(
	bootstrapDataProvider storageFactory.BootstrapDataProviderHandler,
	latestDataFromStorageProvider storage.LatestStorageDataProviderHandler,
	internalMarshalizer marshal.Marshalizer,
	generalConfig config.Config,
	chainID string,
	workingDir string,
	defaultDBPath string,
	defaultEpochString string,
	defaultShardString string,
) (storage.UnitOpenerHandler, error) {
	argsStorageUnitOpener := storageFactory.ArgsNewOpenStorageUnits{
		GeneralConfig:             generalConfig,
		Marshalizer:               internalMarshalizer,
		BootstrapDataProvider:     bootstrapDataProvider,
		LatestStorageDataProvider: latestDataFromStorageProvider,
		WorkingDir:                workingDir,
		ChainID:                   chainID,
		DefaultDBPath:             defaultDBPath,
		DefaultEpochString:        defaultEpochString,
		DefaultShardString:        defaultShardString,
	}

	return storageFactory.NewStorageUnitOpenHandler(argsStorageUnitOpener)
}

// PrepareOpenTopics will set to the anti flood handler the topics for which
// the node can receive messages from others than validators
func PrepareOpenTopics(
	antiflood mainFactory.P2PAntifloodHandler,
	shardCoordinator sharding.Coordinator,
) {
	selfID := shardCoordinator.SelfId()
	if selfID == core.MetachainShardId {
		antiflood.SetTopicsForAll(core.HeartbeatTopic)
		return
	}

	selfShardTxTopic := factory.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(core.HeartbeatTopic, selfShardTxTopic)
}
