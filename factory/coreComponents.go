package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/watchdog"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/hashing"
	hasherFactory "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	marshalizerFactory "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config              config.Config
	WorkingDirectory    string
	GenesisTime         time.Time
	ChanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponentsFactory is responsible for creating the core components
type coreComponentsFactory struct {
	config              config.Config
	workingDir          string
	genesisTime         time.Time
	chanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponents is the DTO used for core components
type coreComponents struct {
	hasher                   hashing.Hasher
	internalMarshalizer      marshal.Marshalizer
	vmMarshalizer            marshal.Marshalizer
	txSignMarshalizer        marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter   core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
	statusHandler            core.AppStatusHandler
	pathHandler              storage.PathManagerHandler
	syncTimer                ntp.SyncTimer
	alarmScheduler           core.TimersScheduler
	watchdog                 core.WatchdogTimer
	genesisTime              time.Time
	chainID                  string
	minTransactionVersion    uint32
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) *coreComponentsFactory {
	return &coreComponentsFactory{
		config:              args.Config,
		workingDir:          args.WorkingDirectory,
		chanStopNodeProcess: args.ChanStopNodeProcess,
	}
}

// Create creates the core components
func (ccf *coreComponentsFactory) Create() (*coreComponents, error) {
	hasher, err := hasherFactory.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", ErrMarshalizerCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	addressPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}
	validatorPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}

	pruningStorerPathTemplate, staticStorerPathTemplate := ccf.createStorerTemplatePaths()
	pathHandler, err := pathmanager.NewPathManager(pruningStorerPathTemplate, staticStorerPathTemplate)
	if err != nil {
		return nil, err
	}

	syncer := ntp.NewSyncTime(ccf.config.NTPConfig, nil)
	syncer.StartSyncingTime()
	log.Debug("NTP average clock offset", "value", syncer.ClockOffset())

	alarmScheduler := alarm.NewAlarmScheduler()
	watchdogTimer, err := watchdog.NewWatchdog(alarmScheduler, ccf.chanStopNodeProcess)
	if err != nil {
		return nil, err
	}

	return &coreComponents{
		hasher:                   hasher,
		internalMarshalizer:      internalMarshalizer,
		vmMarshalizer:            vmMarshalizer,
		txSignMarshalizer:        txSignMarshalizer,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		addressPubKeyConverter:   addressPubkeyConverter,
		validatorPubKeyConverter: validatorPubkeyConverter,
		statusHandler:            statusHandler.NewNilStatusHandler(),
		pathHandler:              pathHandler,
		syncTimer:                syncer,
		alarmScheduler:           alarmScheduler,
		watchdog:                 watchdogTimer,
		genesisTime:              ccf.genesisTime,
		chainID:                  ccf.config.GeneralSettings.ChainID,
		minTransactionVersion:    ccf.config.GeneralSettings.MinTransactionVersion,
	}, nil
}

func (ccf *coreComponentsFactory) createStorerTemplatePaths() (string, string) {
	pathTemplateForPruningStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		fmt.Sprintf("%s_%s", core.DefaultEpochString, core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		core.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	return pathTemplateForPruningStorer, pathTemplateForStaticStorer
}

// Closes all underlying components that need closing
func (cc *coreComponents) Close() error {
	if cc.statusHandler != nil {
		cc.statusHandler.Close()
	}

	return nil
}
