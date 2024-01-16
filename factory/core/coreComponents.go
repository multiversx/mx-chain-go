package core

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/alarm"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/nodetype"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/core/watchdog"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing"
	hasherFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	marshalizerFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("factory")

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config              config.Config
	ConfigPathsHolder   config.ConfigurationPathsHolder
	EpochConfig         config.EpochConfig
	RoundConfig         config.RoundConfig
	RatingsConfig       config.RatingsConfig
	EconomicsConfig     config.EconomicsConfig
	ImportDbConfig      config.ImportDbConfig
	NodesFilename       string
	WorkingDirectory    string
	ChanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponentsFactory is responsible for creating the core components
type coreComponentsFactory struct {
	config              config.Config
	configPathsHolder   config.ConfigurationPathsHolder
	epochConfig         config.EpochConfig
	roundConfig         config.RoundConfig
	ratingsConfig       config.RatingsConfig
	economicsConfig     config.EconomicsConfig
	importDbConfig      config.ImportDbConfig
	nodesFilename       string
	workingDir          string
	chanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponents is the DTO used for core components
type coreComponents struct {
	hasher                        hashing.Hasher
	txSignHasher                  hashing.Hasher
	internalMarshalizer           marshal.Marshalizer
	vmMarshalizer                 marshal.Marshalizer
	txSignMarshalizer             marshal.Marshalizer
	uint64ByteSliceConverter      typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter        core.PubkeyConverter
	validatorPubKeyConverter      core.PubkeyConverter
	pathHandler                   storage.PathManagerHandler
	syncTimer                     ntp.SyncTimer
	roundHandler                  consensus.RoundHandler
	alarmScheduler                core.TimersScheduler
	watchdog                      core.WatchdogTimer
	nodesSetupHandler             sharding.GenesisNodesSetupHandler
	economicsData                 process.EconomicsDataHandler
	apiEconomicsData              process.EconomicsDataHandler
	ratingsData                   process.RatingsInfoHandler
	rater                         sharding.PeerAccountListAndRatingHandler
	nodesShuffler                 nodesCoordinator.NodesShuffler
	txVersionChecker              process.TxVersionCheckerHandler
	genesisTime                   time.Time
	chainID                       string
	minTransactionVersion         uint32
	epochNotifier                 process.EpochNotifier
	roundNotifier                 process.RoundNotifier
	enableRoundsHandler           process.EnableRoundsHandler
	epochStartNotifierWithConfirm factory.EpochStartNotifierWithConfirm
	chanStopNodeProcess           chan endProcess.ArgEndProcess
	nodeTypeProvider              core.NodeTypeProviderHandler
	encodedAddressLen             uint32
	wasmVMChangeLocker            common.Locker
	processStatusHandler          common.ProcessStatusHandler
	hardforkTriggerPubKey         []byte
	enableEpochsHandler           common.EnableEpochsHandler
	increaseRoundChan             chan uint64
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) (*coreComponentsFactory, error) {
	return &coreComponentsFactory{
		config:              args.Config,
		configPathsHolder:   args.ConfigPathsHolder,
		epochConfig:         args.EpochConfig,
		roundConfig:         args.RoundConfig,
		ratingsConfig:       args.RatingsConfig,
		importDbConfig:      args.ImportDbConfig,
		economicsConfig:     args.EconomicsConfig,
		workingDir:          args.WorkingDirectory,
		chanStopNodeProcess: args.ChanStopNodeProcess,
		nodesFilename:       args.NodesFilename,
	}, nil
}

// Create creates the core components
func (ccf *coreComponentsFactory) Create() (*coreComponents, error) {
	hasher, err := hasherFactory.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	txSignHasher, err := hasherFactory.NewHasher(ccf.config.TxSignHasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrHasherCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	addressPubkeyConverter, err := commonFactory.NewPubkeyConverter(ccf.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}

	validatorPubkeyConverter, err := commonFactory.NewPubkeyConverter(ccf.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}

	pathHandler, err := storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: ccf.workingDir,
			ChainID:    ccf.config.GeneralSettings.ChainID,
		},
	)
	if err != nil {
		return nil, err
	}

	genesisTimeStamp := "1596117600"
	i, err := strconv.ParseInt(genesisTimeStamp, 10, 64)
	if err != nil {
		panic(err)
	}

	currentTimeMilis := time.Now().UnixMilli()

	time.Sleep(time.Duration(6000-currentTimeMilis%6000) * time.Millisecond)

	log.Info("AfterSleep", "time", time.Now())
	tm := time.Unix(i, 0)
	log.Info("tm", "time", tm)
	increaseChan := make(chan uint64)
	var syncTimer ntp.SyncTimer
	syncTimer = ntp.NewShadowForkSyncTimer(tm, increaseChan)

	//go func() {
	//	log.Info("sleeping for 100 seconds")
	//	time.Sleep(100 * time.Second)
	//	sleepMiliseconds := int64(4000)
	//	round := uint64(0)
	//	for {
	//		currentTimeMilis = time.Now().UnixMilli()
	//		sleepDuration := time.Duration(sleepMiliseconds-currentTimeMilis%sleepMiliseconds) * time.Millisecond
	//		time.Sleep(sleepDuration)
	//		log.Debug("sleeping for", "miliseconds", sleepDuration)
	//		increaseChan <- round
	//		round++
	//	}
	//}()

	var syncer ntp.SyncTimer

	//syncer = ntp.NewSyncTime(ccf.config.NTPConfig, nil)
	//syncer.StartSyncingTime()

	syncer = syncTimer

	log.Debug("NTP average clock offset", "value", syncer.ClockOffset())

	genesisNodesConfig, err := sharding.NewNodesSetup(
		ccf.nodesFilename,
		addressPubkeyConverter,
		validatorPubkeyConverter,
		ccf.config.GeneralSettings.GenesisMaxNumberOfShards,
	)
	if err != nil {
		return nil, err
	}

	startRound := int64(0)
	if ccf.config.Hardfork.AfterHardFork {
		log.Debug("changed genesis time after hardfork",
			"old genesis time", genesisNodesConfig.StartTime,
			"new genesis time", ccf.config.Hardfork.GenesisTime)
		genesisNodesConfig.StartTime = ccf.config.Hardfork.GenesisTime
		startRound = int64(ccf.config.Hardfork.StartRound)
	}

	if genesisNodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime()
		genesisNodesConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(genesisNodesConfig.StartTime, 0)

	log.Info("start time",
		"formatted", startTime.Format("Mon Jan 2 15:04:05 MST 2006"),
		"seconds", startTime.Unix())

	log.Debug("config", "file", ccf.nodesFilename)

	genesisTime := time.Unix(genesisNodesConfig.StartTime, 0)
	roundHandler, err := round.NewRound(
		genesisTime,
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(genesisNodesConfig.RoundDuration),
		syncer,
		startRound,
	)
	if err != nil {
		return nil, err
	}

	alarmScheduler := alarm.NewAlarmScheduler()
	// TODO: disable watchdog if block processing cutoff is enabled
	watchdogTimer, err := watchdog.NewWatchdog(alarmScheduler, ccf.chanStopNodeProcess, log)
	if err != nil {
		return nil, err
	}

	roundNotifier := forking.NewGenericRoundNotifier()
	enableRoundsHandler, err := enablers.NewEnableRoundsHandler(ccf.roundConfig, roundNotifier)
	if err != nil {
		return nil, err
	}

	epochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(ccf.epochConfig.EnableEpochs, epochNotifier)
	if err != nil {
		return nil, err
	}

	wasmVMChangeLocker := &sync.RWMutex{}
	gasScheduleConfigurationFolderName := ccf.configPathsHolder.GasScheduleDirectoryName
	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig:  ccf.epochConfig.GasSchedule,
		ConfigDir:          gasScheduleConfigurationFolderName,
		EpochNotifier:      epochNotifier,
		WasmVMChangeLocker: wasmVMChangeLocker,
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	if err != nil {
		return nil, err
	}

	builtInCostHandler, err := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		ArgsParser:  smartContract.NewArgumentParser(),
		GasSchedule: gasScheduleNotifier,
	})
	if err != nil {
		return nil, err
	}

	txVersionChecker := versioning.NewTxVersionChecker(ccf.config.GeneralSettings.MinTransactionVersion)

	log.Trace("creating economics data components")
	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics:                   &ccf.economicsConfig,
		EpochNotifier:               epochNotifier,
		EnableEpochsHandler:         enableEpochsHandler,
		BuiltInFunctionsCostHandler: builtInCostHandler,
		TxVersionChecker:            txVersionChecker,
	}
	economicsData, err := economics.NewEconomicsData(argsNewEconomicsData)
	if err != nil {
		return nil, err
	}

	apiEconomicsData, err := economics.NewAPIEconomicsData(economicsData)
	if err != nil {
		return nil, err
	}

	log.Trace("creating ratings data")
	ratingDataArgs := rating.RatingsDataArg{
		Config:                   ccf.ratingsConfig,
		ShardConsensusSize:       genesisNodesConfig.ConsensusGroupSize,
		MetaConsensusSize:        genesisNodesConfig.MetaChainConsensusGroupSize,
		ShardMinNodes:            genesisNodesConfig.MinNodesPerShard,
		MetaMinNodes:             genesisNodesConfig.MetaChainMinNodes,
		RoundDurationMiliseconds: genesisNodesConfig.RoundDuration,
	}
	ratingsData, err := rating.NewRatingsData(ratingDataArgs)
	if err != nil {
		return nil, err
	}

	rater, err := rating.NewBlockSigningRater(ratingsData)
	if err != nil {
		return nil, err
	}

	argsNodesShuffler := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           genesisNodesConfig.MinNumberOfShardNodes(),
		NodesMeta:            genesisNodesConfig.MinNumberOfMetaNodes(),
		Hysteresis:           genesisNodesConfig.GetHysteresis(),
		Adaptivity:           genesisNodesConfig.GetAdaptivity(),
		ShuffleBetweenShards: true,
		MaxNodesEnableConfig: ccf.epochConfig.EnableEpochs.MaxNodesChangeEnableEpoch,
		EnableEpochsHandler:  enableEpochsHandler,
	}

	nodesShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(argsNodesShuffler)
	if err != nil {
		return nil, err
	}

	// set as observer at first - it will be updated when creating the nodes coordinator
	nodeTypeProvider := nodetype.NewNodeTypeProvider(core.NodeTypeObserver)

	pubKeyStr := ccf.config.Hardfork.PublicKeyToListenFrom
	pubKeyBytes, err := validatorPubkeyConverter.Decode(pubKeyStr)
	if err != nil {
		return nil, err
	}

	encodedAddressLen, err := computeEncodedAddressLen(addressPubkeyConverter)
	if err != nil {
		return nil, err
	}

	return &coreComponents{
		hasher:                        hasher,
		txSignHasher:                  txSignHasher,
		internalMarshalizer:           internalMarshalizer,
		vmMarshalizer:                 vmMarshalizer,
		txSignMarshalizer:             txSignMarshalizer,
		uint64ByteSliceConverter:      uint64ByteSliceConverter,
		addressPubKeyConverter:        addressPubkeyConverter,
		validatorPubKeyConverter:      validatorPubkeyConverter,
		pathHandler:                   pathHandler,
		syncTimer:                     syncer,
		roundHandler:                  roundHandler,
		alarmScheduler:                alarmScheduler,
		watchdog:                      watchdogTimer,
		nodesSetupHandler:             genesisNodesConfig,
		economicsData:                 economicsData,
		apiEconomicsData:              apiEconomicsData,
		ratingsData:                   ratingsData,
		rater:                         rater,
		nodesShuffler:                 nodesShuffler,
		txVersionChecker:              txVersionChecker,
		genesisTime:                   genesisTime,
		chainID:                       ccf.config.GeneralSettings.ChainID,
		minTransactionVersion:         ccf.config.GeneralSettings.MinTransactionVersion,
		epochNotifier:                 epochNotifier,
		roundNotifier:                 roundNotifier,
		enableRoundsHandler:           enableRoundsHandler,
		epochStartNotifierWithConfirm: notifier.NewEpochStartSubscriptionHandler(),
		chanStopNodeProcess:           ccf.chanStopNodeProcess,
		encodedAddressLen:             encodedAddressLen,
		nodeTypeProvider:              nodeTypeProvider,
		wasmVMChangeLocker:            wasmVMChangeLocker,
		processStatusHandler:          statusHandler.NewProcessStatusHandler(),
		hardforkTriggerPubKey:         pubKeyBytes,
		enableEpochsHandler:           enableEpochsHandler,
		increaseRoundChan:             increaseChan,
	}, nil
}

// Close closes all underlying components
func (cc *coreComponents) Close() error {
	if !check.IfNil(cc.alarmScheduler) {
		cc.alarmScheduler.Close()
	}
	if !check.IfNil(cc.syncTimer) {
		err := cc.syncTimer.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func computeEncodedAddressLen(converter core.PubkeyConverter) (uint32, error) {
	emptyAddress := bytes.Repeat([]byte{0}, converter.Len())
	encodedEmptyAddress, err := converter.Encode(emptyAddress)
	if err != nil {
		return 0, err
	}

	return uint32(len(encodedEmptyAddress)), nil
}
