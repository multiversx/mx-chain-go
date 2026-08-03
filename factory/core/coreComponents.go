package core

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
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
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/chainparametersnotifier"
	commonConfigs "github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/common/enablers"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/common/fieldsChecker"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/common/graceperiod"
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
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
)

// supernovaFarAwayActivationEpoch marks the convention of disabling Supernova by setting its
// activation epoch far away; alignment checks are then replaced by the disabled-coherence ones
const supernovaFarAwayActivationEpoch = uint32(999999)

// supernovaFarAwayActivationRound must exceed any plausible real round on long-running
// chains (mainnet is past 31 million; this is ~1900 years of 600ms rounds)
const supernovaFarAwayActivationRound = uint64(99_999_999_999)

const supernovaHeaderVersion = "3"

var log = logger.GetOrCreate("factory")

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config                config.Config
	ConfigPathsHolder     config.ConfigurationPathsHolder
	EpochConfig           config.EpochConfig
	RoundConfig           config.RoundConfig
	RatingsConfig         config.RatingsConfig
	EconomicsConfig       config.EconomicsConfig
	ImportDbConfig        config.ImportDbConfig
	NodesConfig           config.NodesConfig
	WorkingDirectory      string
	ChanStopNodeProcess   chan endProcess.ArgEndProcess
	PrintPrettifiedHeader bool
}

// coreComponentsFactory is responsible for creating the core components
type coreComponentsFactory struct {
	config                config.Config
	configPathsHolder     config.ConfigurationPathsHolder
	epochConfig           config.EpochConfig
	roundConfig           config.RoundConfig
	ratingsConfig         config.RatingsConfig
	economicsConfig       config.EconomicsConfig
	importDbConfig        config.ImportDbConfig
	nodesSetupConfig      config.NodesConfig
	workingDir            string
	chanStopNodeProcess   chan endProcess.ArgEndProcess
	printPrettifiedHeader bool
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
	supernovaGenesisTime          time.Time
	chainID                       string
	minTransactionVersion         uint32
	epochNotifier                 process.EpochNotifier
	roundNotifier                 process.RoundNotifier
	chainParametersSubscriber     process.ChainParametersSubscriber
	enableRoundsHandler           common.EnableRoundsHandler
	epochStartNotifierWithConfirm factory.EpochStartNotifierWithConfirm
	chanStopNodeProcess           chan endProcess.ArgEndProcess
	nodeTypeProvider              core.NodeTypeProviderHandler
	encodedAddressLen             uint32
	wasmVMChangeLocker            common.Locker
	processStatusHandler          common.ProcessStatusHandler
	hardforkTriggerPubKey         []byte
	enableEpochsHandler           common.EnableEpochsHandler
	chainParametersHandler        process.ChainParametersHandler
	fieldsSizeChecker             common.FieldsSizeChecker
	epochChangeGracePeriodHandler common.EpochChangeGracePeriodHandler
	processConfigsHandler         common.ProcessConfigsHandler
	epochStartConfigsHandler      common.CommonConfigsHandler
	antifloodConfigsHandler       common.AntifloodConfigsHandler
	closingNodeStarted            *atomic.Bool
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) (*coreComponentsFactory, error) {
	return &coreComponentsFactory{
		config:                args.Config,
		configPathsHolder:     args.ConfigPathsHolder,
		epochConfig:           args.EpochConfig,
		roundConfig:           args.RoundConfig,
		ratingsConfig:         args.RatingsConfig,
		importDbConfig:        args.ImportDbConfig,
		economicsConfig:       args.EconomicsConfig,
		workingDir:            args.WorkingDirectory,
		chanStopNodeProcess:   args.ChanStopNodeProcess,
		nodesSetupConfig:      args.NodesConfig,
		printPrettifiedHeader: args.PrintPrettifiedHeader,
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

	epochChangeGracePeriodHandler, err := graceperiod.NewEpochChangeGracePeriod(ccf.config.GeneralSettings.EpochChangeGracePeriodByEpoch)
	if err != nil {
		return nil, fmt.Errorf("%w for epochChangeGracePeriod", err)
	}

	commonConfigsHandler, err := commonConfigs.NewCommonConfigsHandler(
		ccf.config.GeneralSettings.EpochStartConfigsByEpoch,
		ccf.config.GeneralSettings.EpochStartConfigsByRound,
		ccf.config.GeneralSettings.ConsensusConfigsByEpoch,
		ccf.config.GeneralSettings.ConsensusConfigsByRound,
		ccf.printPrettifiedHeader,
	)
	if err != nil {
		return nil, fmt.Errorf("%w for commonConfigsHandler", err)
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

	epochNotifier := forking.NewGenericEpochNotifier()
	epochStartHandlerWithConfirm := notifier.NewEpochStartSubscriptionHandler()
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(ccf.epochConfig.EnableEpochs, epochNotifier)
	if err != nil {
		return nil, err
	}

	chainParametersNotifier := chainparametersnotifier.NewChainParametersNotifier()
	argsChainParametersHandler := sharding.ArgsChainParametersHolder{
		EpochStartEventNotifier: epochStartHandlerWithConfirm,
		ChainParameters:         ccf.config.GeneralSettings.ChainParametersByEpoch,
		ChainParametersNotifier: chainParametersNotifier,
	}
	chainParametersHandler, err := sharding.NewChainParametersHolder(argsChainParametersHandler)
	if err != nil {
		return nil, err
	}

	roundNotifier := forking.NewGenericRoundNotifier()
	enableRoundsHandler, err := enablers.NewEnableRoundsHandler(ccf.roundConfig, roundNotifier)
	if err != nil {
		return nil, err
	}

	processConfigs, err := commonConfigs.NewProcessConfigsHandler(
		ccf.config.GeneralSettings.ProcessConfigsByEpoch,
		ccf.config.GeneralSettings.ProcessConfigsByRound,
		roundNotifier,
	)
	if err != nil {
		return nil, fmt.Errorf("%w for processConfigsByEpoch", err)
	}

	antifloodConfigsHandler, err := commonConfigs.NewAntifloodConfigsHandler(
		ccf.config.Antiflood,
		roundNotifier,
	)
	if err != nil {
		return nil, fmt.Errorf("%w for antifloodConfigsHandler", err)
	}

	genesisNodesConfig, err := sharding.NewNodesSetup(
		ccf.nodesSetupConfig,
		chainParametersHandler,
		addressPubkeyConverter,
		validatorPubkeyConverter,
		ccf.config.GeneralSettings.GenesisMaxNumberOfShards,
	)
	if err != nil {
		return nil, err
	}

	genesisRoundDuration := time.Millisecond * time.Duration(genesisNodesConfig.GetRoundDuration())
	supernovaRoundDuration, err := getSupernovaRoundDuration(enableEpochsHandler, chainParametersHandler)
	if err != nil {
		return nil, err
	}

	syncer := ntp.NewSyncTime(ccf.config.NTPConfig, nil)
	syncer.StartSyncingTime()
	log.Debug("NTP average clock offset", "value", syncer.ClockOffset())

	startRound := int64(0)
	supernovaStartRound := int64(enableRoundsHandler.GetActivationRound(common.SupernovaRoundFlag))

	if ccf.config.Hardfork.AfterHardFork {
		log.Debug("changed genesis time after hardfork",
			"old genesis time", genesisNodesConfig.StartTime,
			"new genesis time", ccf.config.Hardfork.GenesisTime)
		genesisNodesConfig.StartTime = ccf.config.Hardfork.GenesisTime
		startRound = int64(ccf.config.Hardfork.StartRound)
	}

	if genesisNodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)

		startTime := common.RoundToNearestMinute(syncer.CurrentTime())

		genesisNodesConfig.StartTime = common.GetGenesisUnixTimestampFromStartTime(startTime, enableEpochsHandler)
	}

	startTime := common.GetGenesisStartTimeFromUnixTimestamp(genesisNodesConfig.GetStartTime(), enableEpochsHandler)

	genesisTime := common.GetGenesisStartTimeFromUnixTimestamp(genesisNodesConfig.GetStartTime(), enableEpochsHandler)

	if genesisRoundDuration <= 0 {
		return nil, fmt.Errorf("invalid genesis round duration %d", genesisRoundDuration)
	}
	// saturate instead of overflowing for far-away activation rounds used to disable Supernova
	supernovaOffset := time.Duration(math.MaxInt64)
	if supernovaStartRound <= math.MaxInt64/genesisRoundDuration.Nanoseconds() {
		supernovaOffset = time.Duration(supernovaStartRound * genesisRoundDuration.Nanoseconds())
	}
	supernovaGenesisTime := genesisTime.Add(supernovaOffset)

	if supernovaStartRound < startRound {
		return nil, fmt.Errorf("supernovaStartRound %d lower then startRound %d",
			supernovaStartRound,
			startRound,
		)
	}

	if supernovaGenesisTime.Compare(genesisTime) < 0 {
		return nil, fmt.Errorf("supernovaGenesisTime %d lower then genesisTime %d",
			supernovaGenesisTime.UnixMilli(),
			genesisTime.UnixMilli(),
		)
	}

	err = validateSupernovaActivationTuple(
		ccf.config,
		enableEpochsHandler.GetActivationEpoch(common.SupernovaFlag),
		uint64(supernovaStartRound),
	)
	if err != nil {
		return nil, err
	}

	log.Info("start time",
		"formatted", startTime.Format("Mon Jan 2 15:04:05 MST 2006"),
		"unix timestamp", common.GetGenesisUnixTimestampFromStartTime(startTime, enableEpochsHandler),
		"supernova unix timestamp", common.GetGenesisUnixTimestampFromStartTime(supernovaGenesisTime, enableEpochsHandler),
		"round duration", genesisRoundDuration,
		"supernova round duration", supernovaRoundDuration,
	)

	roundArgs := round.ArgsRound{
		GenesisTimeStamp:          genesisTime,
		SupernovaGenesisTimeStamp: supernovaGenesisTime,
		CurrentTimeStamp:          syncer.CurrentTime(),
		RoundTimeDuration:         genesisRoundDuration,
		SupernovaTimeDuration:     supernovaRoundDuration,
		SyncTimer:                 syncer,
		StartRound:                startRound,
		SupernovaStartRound:       supernovaStartRound,
		EnableRoundsHandler:       enableRoundsHandler,
		ImportDBMode:              ccf.importDbConfig.IsImportDBMode,
	}
	roundHandler, err := round.NewRound(roundArgs)
	if err != nil {
		return nil, err
	}

	alarmScheduler := alarm.NewAlarmScheduler()
	// TODO: disable watchdog if block processing cutoff is enabled
	watchdogTimer, err := watchdog.NewWatchdog(alarmScheduler, ccf.chanStopNodeProcess, log)
	if err != nil {
		return nil, err
	}

	wasmVMChangeLocker := &sync.RWMutex{}

	txVersionChecker := versioning.NewTxVersionChecker(ccf.config.GeneralSettings.MinTransactionVersion)

	// This shard coordinator uses a hardcoded selfId of 0 as it does not know its selfId.
	// Its main purpose is to validate the rewards config (protocol sustainability address shard against meta),
	// inside economics data and should not be used for another scope.
	// The real component will be created later on, as part of bootstrap components.
	shardCoordinator, err := sharding.NewMultiShardCoordinator(genesisNodesConfig.NumberOfShards(), 0)
	if err != nil {
		return nil, err
	}

	log.Trace("creating economics data components")
	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics:           &ccf.economicsConfig,
		ChainParamsHandler:  chainParametersHandler,
		EpochNotifier:       epochNotifier,
		EnableEpochsHandler: enableEpochsHandler,
		TxVersionChecker:    txVersionChecker,
		PubkeyConverter:     addressPubkeyConverter,
		ShardCoordinator:    shardCoordinator,
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
		Config:                ccf.ratingsConfig,
		ChainParametersHolder: chainParametersHandler,
		EpochNotifier:         epochNotifier,
	}
	ratingsData, err := rating.NewRatingsData(ratingDataArgs)
	if err != nil {
		return nil, err
	}

	rater, err := rating.NewBlockSigningRater(ratingsData, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	argsNodesShuffler := &nodesCoordinator.NodesShufflerArgs{
		ShuffleBetweenShards: true,
		MaxNodesEnableConfig: ccf.epochConfig.EnableEpochs.MaxNodesChangeEnableEpoch,
		EnableEpochsHandler:  enableEpochsHandler,
		EnableEpochs:         ccf.epochConfig.EnableEpochs,
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

	fieldsSizeChecker, err := fieldsChecker.NewFieldsSizeChecker(chainParametersHandler, hasher)
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
		supernovaGenesisTime:          supernovaGenesisTime,
		chainID:                       ccf.config.GeneralSettings.ChainID,
		minTransactionVersion:         ccf.config.GeneralSettings.MinTransactionVersion,
		epochNotifier:                 epochNotifier,
		roundNotifier:                 roundNotifier,
		chainParametersSubscriber:     chainParametersNotifier,
		enableRoundsHandler:           enableRoundsHandler,
		epochStartNotifierWithConfirm: epochStartHandlerWithConfirm,
		chanStopNodeProcess:           ccf.chanStopNodeProcess,
		encodedAddressLen:             encodedAddressLen,
		nodeTypeProvider:              nodeTypeProvider,
		wasmVMChangeLocker:            wasmVMChangeLocker,
		processStatusHandler:          statusHandler.NewProcessStatusHandler(),
		hardforkTriggerPubKey:         pubKeyBytes,
		enableEpochsHandler:           enableEpochsHandler,
		chainParametersHandler:        chainParametersHandler,
		fieldsSizeChecker:             fieldsSizeChecker,
		epochChangeGracePeriodHandler: epochChangeGracePeriodHandler,
		processConfigsHandler:         processConfigs,
		epochStartConfigsHandler:      commonConfigsHandler,
		antifloodConfigsHandler:       antifloodConfigsHandler,
		closingNodeStarted:            &atomic.Bool{},
	}, nil
}

func getSupernovaRoundDuration(
	enableEpochsHandler common.EnableEpochsHandler,
	chainParametersHandler common.ChainParametersHandler,
) (time.Duration, error) {
	activationEpoch := enableEpochsHandler.GetActivationEpoch(common.SupernovaFlag)
	chainParams, err := chainParametersHandler.ChainParametersForEpoch(activationEpoch)
	if err != nil {
		return 0, err
	}

	return time.Duration(chainParams.RoundDuration) * time.Millisecond, nil
}

func hasConfigEntry[T any](entries []T, matches func(T) bool) bool {
	for _, entry := range entries {
		if matches(entry) {
			return true
		}
	}

	return false
}

// validateSupernovaActivationTuple checks that every config list carrying the Supernova
// activation boundary agrees with the activation flags
func validateSupernovaActivationTuple(cfg config.Config, supernovaEpoch uint32, supernovaRound uint64) error {
	if supernovaEpoch >= supernovaFarAwayActivationEpoch {
		return checkDisabledSupernovaCoherence(cfg)
	}

	err := checkSupernovaVersionEntry(cfg.Versions.VersionsByEpochs, supernovaEpoch, supernovaRound)
	if err != nil {
		return err
	}

	gs := cfg.GeneralSettings
	missingLists := make([]string, 0)
	checkList := func(listName string, found bool) {
		if !found {
			missingLists = append(missingLists, listName)
		}
	}

	checkList("GeneralSettings.ChainParametersByEpoch", hasConfigEntry(gs.ChainParametersByEpoch, func(e config.ChainParametersByEpochConfig) bool {
		return e.EnableEpoch == supernovaEpoch
	}))
	checkList("GeneralSettings.EpochChangeGracePeriodByEpoch", hasConfigEntry(gs.EpochChangeGracePeriodByEpoch, func(e config.EpochChangeGracePeriodByEpoch) bool {
		return e.EnableEpoch == supernovaEpoch
	}))
	checkList("GeneralSettings.ProcessConfigsByEpoch", hasConfigEntry(gs.ProcessConfigsByEpoch, func(e config.ProcessConfigByEpoch) bool {
		return e.EnableEpoch == supernovaEpoch
	}))
	checkList("GeneralSettings.EpochStartConfigsByEpoch", hasConfigEntry(gs.EpochStartConfigsByEpoch, func(e config.EpochStartConfigByEpoch) bool {
		return e.EnableEpoch == supernovaEpoch
	}))
	checkList("GeneralSettings.ConsensusConfigsByEpoch", hasConfigEntry(gs.ConsensusConfigsByEpoch, func(e config.ConsensusConfigByEpoch) bool {
		return e.EnableEpoch == supernovaEpoch
	}))
	checkList("GeneralSettings.ProcessConfigsByRound", hasConfigEntry(gs.ProcessConfigsByRound, func(e config.ProcessConfigByRound) bool {
		return e.EnableRound == supernovaRound
	}))
	checkList("GeneralSettings.EpochStartConfigsByRound", hasConfigEntry(gs.EpochStartConfigsByRound, func(e config.EpochStartConfigByRound) bool {
		return e.EnableRound == supernovaRound
	}))
	checkList("GeneralSettings.ConsensusConfigsByRound", hasConfigEntry(gs.ConsensusConfigsByRound, func(e config.ConsensusConfigByRound) bool {
		return e.EnableRound == supernovaRound
	}))

	if len(missingLists) > 0 {
		return fmt.Errorf("%w: no entry at supernova activation epoch %d or round %d in: %s",
			errors.ErrSupernovaActivationConfigMismatch,
			supernovaEpoch,
			supernovaRound,
			strings.Join(missingLists, ", "),
		)
	}

	return nil
}

// checkDisabledSupernovaCoherence ensures a disabled Supernova leaves no near-boundary
// leftovers that fire on their own: V3 header stamping and round-keyed config switches
func checkDisabledSupernovaCoherence(cfg config.Config) error {
	for _, version := range cfg.Versions.VersionsByEpochs {
		if version.Version != supernovaHeaderVersion {
			continue
		}
		if version.StartEpoch < supernovaFarAwayActivationEpoch {
			return fmt.Errorf("%w: supernova is disabled but [Versions] entry %q has StartEpoch %d, expected at least %d",
				errors.ErrSupernovaActivationConfigMismatch,
				supernovaHeaderVersion,
				version.StartEpoch,
				supernovaFarAwayActivationEpoch,
			)
		}
	}

	isNearRound := func(round uint64) bool {
		return round != 0 && round < supernovaFarAwayActivationRound
	}

	gs := cfg.GeneralSettings
	nearLists := make([]string, 0)
	checkList := func(listName string, hasNearEntry bool) {
		if hasNearEntry {
			nearLists = append(nearLists, listName)
		}
	}

	checkList("GeneralSettings.ProcessConfigsByRound", hasConfigEntry(gs.ProcessConfigsByRound, func(e config.ProcessConfigByRound) bool {
		return isNearRound(e.EnableRound)
	}))
	checkList("GeneralSettings.EpochStartConfigsByRound", hasConfigEntry(gs.EpochStartConfigsByRound, func(e config.EpochStartConfigByRound) bool {
		return isNearRound(e.EnableRound)
	}))
	checkList("GeneralSettings.ConsensusConfigsByRound", hasConfigEntry(gs.ConsensusConfigsByRound, func(e config.ConsensusConfigByRound) bool {
		return isNearRound(e.EnableRound)
	}))

	if len(nearLists) > 0 {
		return fmt.Errorf("%w: supernova is disabled but round-keyed entries below round %d exist in: %s",
			errors.ErrSupernovaActivationConfigMismatch,
			supernovaFarAwayActivationRound,
			strings.Join(nearLists, ", "),
		)
	}

	return nil
}

func checkSupernovaVersionEntry(versions []config.VersionByEpochs, supernovaEpoch uint32, supernovaRound uint64) error {
	for _, version := range versions {
		if version.Version != supernovaHeaderVersion {
			continue
		}
		if version.StartEpoch != supernovaEpoch || version.StartRound != supernovaRound {
			return fmt.Errorf("%w: [Versions] entry %q has StartEpoch %d and StartRound %d, expected %d and %d",
				errors.ErrSupernovaActivationConfigMismatch,
				supernovaHeaderVersion,
				version.StartEpoch,
				version.StartRound,
				supernovaEpoch,
				supernovaRound,
			)
		}

		return nil
	}

	return fmt.Errorf("%w: no [Versions] entry with Version %q",
		errors.ErrSupernovaActivationConfigMismatch,
		supernovaHeaderVersion,
	)
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
