package processingOnlyNode

import (
	"bytes"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/alarm"
	"github.com/multiversx/mx-chain-core-go/core/nodetype"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/core/watchdog"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing"
	hashingFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	marshalFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	factoryPubKey "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
)

type coreComponentsHolder struct {
	internalMarshaller            marshal.Marshalizer
	txMarshaller                  marshal.Marshalizer
	vmMarshaller                  marshal.Marshalizer
	hasher                        hashing.Hasher
	txSignHasher                  hashing.Hasher
	uint64SliceConverter          typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter        core.PubkeyConverter
	validatorPubKeyConverter      core.PubkeyConverter
	pathHandler                   storage.PathManagerHandler
	watchdog                      core.WatchdogTimer
	alarmScheduler                core.TimersScheduler
	syncTimer                     ntp.SyncTimer
	roundHandler                  consensus.RoundHandler
	economicsData                 process.EconomicsDataHandler
	apiEconomicsData              process.EconomicsDataHandler
	ratingsData                   process.RatingsInfoHandler
	rater                         sharding.PeerAccountListAndRatingHandler
	genesisNodesSetup             sharding.GenesisNodesSetupHandler
	nodesShuffler                 nodesCoordinator.NodesShuffler
	epochNotifier                 process.EpochNotifier
	enableRoundsHandler           process.EnableRoundsHandler
	roundNotifier                 process.RoundNotifier
	epochStartNotifierWithConfirm factory.EpochStartNotifierWithConfirm
	chanStopNodeProcess           chan endProcess.ArgEndProcess
	genesisTime                   time.Time
	chainID                       string
	minTransactionVersion         uint32
	txVersionChecker              process.TxVersionCheckerHandler
	encodedAddressLen             uint32
	nodeTypeProvider              core.NodeTypeProviderHandler
	wasmVMChangeLocker            common.Locker
	processStatusHandler          common.ProcessStatusHandler
	hardforkTriggerPubKey         []byte
	enableEpochsHandler           common.EnableEpochsHandler
}

type ArgsCoreComponentsHolder struct {
	Cfg                 config.Config
	EnableEpochsConfig  config.EnableEpochs
	RoundsConfig        config.RoundConfig
	EconomicsConfig     config.EconomicsConfig
	ChanStopNodeProcess chan endProcess.ArgEndProcess
	NodesSetupPath      string
	GasScheduleFilename string
	NumShards           uint32
	WorkingDir          string
}

func CreateCoreComponentsHolder(args ArgsCoreComponentsHolder) (factory.CoreComponentsHolder, error) {
	var err error
	instance := &coreComponentsHolder{}

	instance.internalMarshaller, err = marshalFactory.NewMarshalizer(args.Cfg.Marshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.txMarshaller, err = marshalFactory.NewMarshalizer(args.Cfg.TxSignMarshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.vmMarshaller, err = marshalFactory.NewMarshalizer(args.Cfg.VmMarshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.hasher, err = hashingFactory.NewHasher(args.Cfg.Hasher.Type)
	if err != nil {
		return nil, err
	}
	instance.txSignHasher, err = hashingFactory.NewHasher(args.Cfg.TxSignHasher.Type)
	if err != nil {
		return nil, err
	}
	instance.uint64SliceConverter = uint64ByteSlice.NewBigEndianConverter()
	instance.addressPubKeyConverter, err = factoryPubKey.NewPubkeyConverter(args.Cfg.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}
	instance.validatorPubKeyConverter, err = factoryPubKey.NewPubkeyConverter(args.Cfg.ValidatorPubkeyConverter)
	if err != nil {
		return nil, err
	}

	instance.pathHandler, err = storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: args.WorkingDir,
			ChainID:    args.Cfg.GeneralSettings.ChainID,
		},
	)
	if err != nil {
		return nil, err
	}

	// TODO check if we need the real watchdog
	instance.watchdog = &watchdog.DisabledWatchdog{}
	// TODO check if we need the real alarm scheduler
	instance.alarmScheduler = alarm.NewAlarmScheduler()
	// TODO check if we need the real sync time also this component need to be started
	instance.syncTimer = ntp.NewSyncTime(args.Cfg.NTPConfig, nil)
	// TODO discuss with Iulian about the round handler
	//instance.roundHandler

	instance.wasmVMChangeLocker = &sync.RWMutex{}
	instance.txVersionChecker = versioning.NewTxVersionChecker(args.Cfg.GeneralSettings.MinTransactionVersion)
	instance.epochNotifier = forking.NewGenericEpochNotifier()
	instance.enableEpochsHandler, err = enablers.NewEnableEpochsHandler(args.EnableEpochsConfig, instance.epochNotifier)
	if err != nil {
		return nil, err
	}

	argsGasSchedule := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig: config.GasScheduleConfig{
			GasScheduleByEpochs: []config.GasScheduleByEpochs{
				{
					StartEpoch: 0,
					FileName:   args.GasScheduleFilename,
				},
			},
		},
		ConfigDir:          "",
		EpochNotifier:      instance.epochNotifier,
		WasmVMChangeLocker: instance.wasmVMChangeLocker,
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasSchedule)
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

	argsEconomicsHandler := economics.ArgsNewEconomicsData{
		TxVersionChecker:            instance.txVersionChecker,
		BuiltInFunctionsCostHandler: builtInCostHandler,
		Economics:                   &args.EconomicsConfig,
		EpochNotifier:               instance.epochNotifier,
		EnableEpochsHandler:         instance.enableEpochsHandler,
	}

	instance.economicsData, err = economics.NewEconomicsData(argsEconomicsHandler)
	if err != nil {
		return nil, err
	}
	instance.apiEconomicsData = instance.economicsData

	// TODO check if we need this
	instance.ratingsData = nil
	instance.rater = nil

	instance.genesisNodesSetup, err = sharding.NewNodesSetup(args.NodesSetupPath, instance.addressPubKeyConverter, instance.validatorPubKeyConverter, args.NumShards)
	if err != nil {
		return nil, err
	}
	// TODO check if we need nodes shuffler
	instance.nodesShuffler = nil

	instance.roundNotifier = forking.NewGenericRoundNotifier()
	instance.enableRoundsHandler, err = enablers.NewEnableRoundsHandler(args.RoundsConfig, instance.roundNotifier)
	if err != nil {
		return nil, err
	}

	instance.epochStartNotifierWithConfirm = notifier.NewEpochStartSubscriptionHandler()
	instance.chanStopNodeProcess = args.ChanStopNodeProcess
	instance.genesisTime = time.Unix(instance.genesisNodesSetup.GetStartTime(), 0)
	instance.chainID = args.Cfg.GeneralSettings.ChainID
	instance.minTransactionVersion = args.Cfg.GeneralSettings.MinTransactionVersion
	instance.encodedAddressLen, err = computeEncodedAddressLen(instance.addressPubKeyConverter)
	if err != nil {
		return nil, err
	}

	instance.nodeTypeProvider = nodetype.NewNodeTypeProvider(core.NodeTypeObserver)
	instance.processStatusHandler = statusHandler.NewProcessStatusHandler()

	pubKeyBytes, err := instance.validatorPubKeyConverter.Decode(args.Cfg.Hardfork.PublicKeyToListenFrom)
	if err != nil {
		return nil, err
	}
	instance.hardforkTriggerPubKey = pubKeyBytes

	return instance, nil
}

func computeEncodedAddressLen(converter core.PubkeyConverter) (uint32, error) {
	emptyAddress := bytes.Repeat([]byte{0}, converter.Len())
	encodedEmptyAddress, err := converter.Encode(emptyAddress)
	if err != nil {
		return 0, err
	}

	return uint32(len(encodedEmptyAddress)), nil
}

func (c *coreComponentsHolder) InternalMarshalizer() marshal.Marshalizer {
	return c.internalMarshaller
}

func (c *coreComponentsHolder) SetInternalMarshalizer(marshalizer marshal.Marshalizer) error {
	c.internalMarshaller = marshalizer
	return nil
}

func (c *coreComponentsHolder) TxMarshalizer() marshal.Marshalizer {
	return c.txMarshaller
}

func (c *coreComponentsHolder) VmMarshalizer() marshal.Marshalizer {
	return c.vmMarshaller
}

func (c *coreComponentsHolder) Hasher() hashing.Hasher {
	return c.hasher
}

func (c *coreComponentsHolder) TxSignHasher() hashing.Hasher {
	return c.txSignHasher
}

func (c *coreComponentsHolder) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	return c.uint64SliceConverter
}

func (c *coreComponentsHolder) AddressPubKeyConverter() core.PubkeyConverter {
	return c.addressPubKeyConverter
}

func (c *coreComponentsHolder) ValidatorPubKeyConverter() core.PubkeyConverter {
	return c.validatorPubKeyConverter
}

func (c *coreComponentsHolder) PathHandler() storage.PathManagerHandler {
	return c.pathHandler
}

func (c *coreComponentsHolder) Watchdog() core.WatchdogTimer {
	return c.watchdog
}

func (c *coreComponentsHolder) AlarmScheduler() core.TimersScheduler {
	return c.alarmScheduler
}

func (c *coreComponentsHolder) SyncTimer() ntp.SyncTimer {
	return c.syncTimer
}

func (c *coreComponentsHolder) RoundHandler() consensus.RoundHandler {
	return c.roundHandler
}

func (c *coreComponentsHolder) EconomicsData() process.EconomicsDataHandler {
	return c.economicsData
}

func (c *coreComponentsHolder) APIEconomicsData() process.EconomicsDataHandler {
	return c.apiEconomicsData
}

func (c *coreComponentsHolder) RatingsData() process.RatingsInfoHandler {
	return c.ratingsData
}

func (c *coreComponentsHolder) Rater() sharding.PeerAccountListAndRatingHandler {
	return c.rater
}

func (c *coreComponentsHolder) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	return c.genesisNodesSetup
}

func (c *coreComponentsHolder) NodesShuffler() nodesCoordinator.NodesShuffler {
	return c.nodesShuffler
}

func (c *coreComponentsHolder) EpochNotifier() process.EpochNotifier {
	return c.epochNotifier
}

func (c *coreComponentsHolder) EnableRoundsHandler() process.EnableRoundsHandler {
	return c.enableRoundsHandler
}

func (c *coreComponentsHolder) RoundNotifier() process.RoundNotifier {
	return c.roundNotifier
}

func (c *coreComponentsHolder) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	return c.epochStartNotifierWithConfirm
}

func (c *coreComponentsHolder) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	return c.chanStopNodeProcess
}

func (c *coreComponentsHolder) GenesisTime() time.Time {
	return c.genesisTime
}

func (c *coreComponentsHolder) ChainID() string {
	return c.chainID
}

func (c *coreComponentsHolder) MinTransactionVersion() uint32 {
	return c.minTransactionVersion
}

func (c *coreComponentsHolder) TxVersionChecker() process.TxVersionCheckerHandler {
	return c.txVersionChecker
}

func (c *coreComponentsHolder) EncodedAddressLen() uint32 {
	return c.encodedAddressLen
}

func (c *coreComponentsHolder) NodeTypeProvider() core.NodeTypeProviderHandler {
	return c.nodeTypeProvider
}

func (c *coreComponentsHolder) WasmVMChangeLocker() common.Locker {
	return c.wasmVMChangeLocker
}

func (c *coreComponentsHolder) ProcessStatusHandler() common.ProcessStatusHandler {
	return c.processStatusHandler
}

func (c *coreComponentsHolder) HardforkTriggerPubKey() []byte {
	return c.hardforkTriggerPubKey
}

func (c *coreComponentsHolder) EnableEpochsHandler() common.EnableEpochsHandler {
	return c.enableEpochsHandler
}

func (c *coreComponentsHolder) IsInterfaceNil() bool {
	return c == nil
}
