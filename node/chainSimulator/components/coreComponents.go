package components

import (
	"bytes"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
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
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type coreComponentsHolder struct {
	closeHandler                  *closeHandler
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

// ArgsCoreComponentsHolder will hold arguments needed for the core components holder
type ArgsCoreComponentsHolder struct {
	Config              config.Config
	EnableEpochsConfig  config.EnableEpochs
	RoundsConfig        config.RoundConfig
	EconomicsConfig     config.EconomicsConfig
	ChanStopNodeProcess chan endProcess.ArgEndProcess
	InitialRound        int64
	NodesSetupPath      string
	GasScheduleFilename string
	NumShards           uint32
	WorkingDir          string

	MinNodesPerShard uint32
	MinNodesMeta     uint32
}

// CreateCoreComponents will create a new instance of factory.CoreComponentsHolder
func CreateCoreComponents(args ArgsCoreComponentsHolder) (factory.CoreComponentsHandler, error) {
	var err error
	instance := &coreComponentsHolder{
		closeHandler: NewCloseHandler(),
	}

	instance.internalMarshaller, err = marshalFactory.NewMarshalizer(args.Config.Marshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.txMarshaller, err = marshalFactory.NewMarshalizer(args.Config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.vmMarshaller, err = marshalFactory.NewMarshalizer(args.Config.VmMarshalizer.Type)
	if err != nil {
		return nil, err
	}
	instance.hasher, err = hashingFactory.NewHasher(args.Config.Hasher.Type)
	if err != nil {
		return nil, err
	}
	instance.txSignHasher, err = hashingFactory.NewHasher(args.Config.TxSignHasher.Type)
	if err != nil {
		return nil, err
	}
	instance.uint64SliceConverter = uint64ByteSlice.NewBigEndianConverter()
	instance.addressPubKeyConverter, err = factoryPubKey.NewPubkeyConverter(args.Config.AddressPubkeyConverter)
	if err != nil {
		return nil, err
	}
	instance.validatorPubKeyConverter, err = factoryPubKey.NewPubkeyConverter(args.Config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, err
	}

	instance.pathHandler, err = storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: args.WorkingDir,
			ChainID:    args.Config.GeneralSettings.ChainID,
		},
	)
	if err != nil {
		return nil, err
	}

	instance.watchdog = &watchdog.DisabledWatchdog{}
	instance.alarmScheduler = &mock.AlarmSchedulerStub{}
	instance.syncTimer = &testscommon.SyncTimerStub{}

	instance.genesisNodesSetup, err = sharding.NewNodesSetup(args.NodesSetupPath, instance.addressPubKeyConverter, instance.validatorPubKeyConverter, args.NumShards)
	if err != nil {
		return nil, err
	}

	roundDuration := time.Millisecond * time.Duration(instance.genesisNodesSetup.GetRoundDuration())
	instance.roundHandler = NewManualRoundHandler(instance.genesisNodesSetup.GetStartTime(), roundDuration, args.InitialRound)

	instance.wasmVMChangeLocker = &sync.RWMutex{}
	instance.txVersionChecker = versioning.NewTxVersionChecker(args.Config.GeneralSettings.MinTransactionVersion)
	instance.epochNotifier = forking.NewGenericEpochNotifier()
	instance.enableEpochsHandler, err = enablers.NewEnableEpochsHandler(args.EnableEpochsConfig, instance.epochNotifier)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	argsEconomicsHandler := economics.ArgsNewEconomicsData{
		TxVersionChecker:    instance.txVersionChecker,
		Economics:           &args.EconomicsConfig,
		EpochNotifier:       instance.epochNotifier,
		EnableEpochsHandler: instance.enableEpochsHandler,
	}

	instance.economicsData, err = economics.NewEconomicsData(argsEconomicsHandler)
	if err != nil {
		return nil, err
	}
	instance.apiEconomicsData = instance.economicsData

	// TODO check if we need this
	instance.ratingsData = &testscommon.RatingsInfoMock{}
	instance.rater = &testscommon.RaterMock{}

	instance.nodesShuffler, err = nodesCoordinator.NewHashValidatorsShuffler(&nodesCoordinator.NodesShufflerArgs{
		NodesShard:           args.MinNodesPerShard,
		NodesMeta:            args.MinNodesMeta,
		Hysteresis:           0,
		Adaptivity:           false,
		ShuffleBetweenShards: true,
		MaxNodesEnableConfig: args.EnableEpochsConfig.MaxNodesChangeEnableEpoch,
		EnableEpochsHandler:  instance.enableEpochsHandler,
	})
	if err != nil {
		return nil, err
	}

	instance.roundNotifier = forking.NewGenericRoundNotifier()
	instance.enableRoundsHandler, err = enablers.NewEnableRoundsHandler(args.RoundsConfig, instance.roundNotifier)
	if err != nil {
		return nil, err
	}

	instance.epochStartNotifierWithConfirm = notifier.NewEpochStartSubscriptionHandler()
	instance.chanStopNodeProcess = args.ChanStopNodeProcess
	instance.genesisTime = time.Unix(instance.genesisNodesSetup.GetStartTime(), 0)
	instance.chainID = args.Config.GeneralSettings.ChainID
	instance.minTransactionVersion = args.Config.GeneralSettings.MinTransactionVersion
	instance.encodedAddressLen, err = computeEncodedAddressLen(instance.addressPubKeyConverter)
	if err != nil {
		return nil, err
	}

	instance.nodeTypeProvider = nodetype.NewNodeTypeProvider(core.NodeTypeObserver)
	instance.processStatusHandler = statusHandler.NewProcessStatusHandler()

	pubKeyBytes, err := instance.validatorPubKeyConverter.Decode(args.Config.Hardfork.PublicKeyToListenFrom)
	if err != nil {
		return nil, err
	}
	instance.hardforkTriggerPubKey = pubKeyBytes

	instance.collectClosableComponents()

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

// InternalMarshalizer will return the internal marshaller
func (c *coreComponentsHolder) InternalMarshalizer() marshal.Marshalizer {
	return c.internalMarshaller
}

// SetInternalMarshalizer will set the internal marshaller
func (c *coreComponentsHolder) SetInternalMarshalizer(marshalizer marshal.Marshalizer) error {
	c.internalMarshaller = marshalizer
	return nil
}

// TxMarshalizer will return the transaction marshaller
func (c *coreComponentsHolder) TxMarshalizer() marshal.Marshalizer {
	return c.txMarshaller
}

// VmMarshalizer will return the vm marshaller
func (c *coreComponentsHolder) VmMarshalizer() marshal.Marshalizer {
	return c.vmMarshaller
}

// Hasher will return the hasher
func (c *coreComponentsHolder) Hasher() hashing.Hasher {
	return c.hasher
}

// TxSignHasher will return the transaction sign hasher
func (c *coreComponentsHolder) TxSignHasher() hashing.Hasher {
	return c.txSignHasher
}

// Uint64ByteSliceConverter will return the uint64 to slice converter
func (c *coreComponentsHolder) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	return c.uint64SliceConverter
}

// AddressPubKeyConverter will return the address pub key converter
func (c *coreComponentsHolder) AddressPubKeyConverter() core.PubkeyConverter {
	return c.addressPubKeyConverter
}

// ValidatorPubKeyConverter will return the validator pub key converter
func (c *coreComponentsHolder) ValidatorPubKeyConverter() core.PubkeyConverter {
	return c.validatorPubKeyConverter
}

// PathHandler will return the path handler
func (c *coreComponentsHolder) PathHandler() storage.PathManagerHandler {
	return c.pathHandler
}

// Watchdog will return the watch dog
func (c *coreComponentsHolder) Watchdog() core.WatchdogTimer {
	return c.watchdog
}

// AlarmScheduler will return the alarm scheduler
func (c *coreComponentsHolder) AlarmScheduler() core.TimersScheduler {
	return c.alarmScheduler
}

// SyncTimer will return the sync timer
func (c *coreComponentsHolder) SyncTimer() ntp.SyncTimer {
	return c.syncTimer
}

// RoundHandler will return the round handler
func (c *coreComponentsHolder) RoundHandler() consensus.RoundHandler {
	return c.roundHandler
}

// EconomicsData will return the economics data handler
func (c *coreComponentsHolder) EconomicsData() process.EconomicsDataHandler {
	return c.economicsData
}

// APIEconomicsData will return the api economics data handler
func (c *coreComponentsHolder) APIEconomicsData() process.EconomicsDataHandler {
	return c.apiEconomicsData
}

// RatingsData will return the ratings data handler
func (c *coreComponentsHolder) RatingsData() process.RatingsInfoHandler {
	return c.ratingsData
}

// Rater will return the rater handler
func (c *coreComponentsHolder) Rater() sharding.PeerAccountListAndRatingHandler {
	return c.rater
}

// GenesisNodesSetup will return the genesis nodes setup handler
func (c *coreComponentsHolder) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	return c.genesisNodesSetup
}

// NodesShuffler will return the nodes shuffler
func (c *coreComponentsHolder) NodesShuffler() nodesCoordinator.NodesShuffler {
	return c.nodesShuffler
}

// EpochNotifier will return the epoch notifier
func (c *coreComponentsHolder) EpochNotifier() process.EpochNotifier {
	return c.epochNotifier
}

// EnableRoundsHandler will return the enable rounds handler
func (c *coreComponentsHolder) EnableRoundsHandler() process.EnableRoundsHandler {
	return c.enableRoundsHandler
}

// RoundNotifier will return the round notifier
func (c *coreComponentsHolder) RoundNotifier() process.RoundNotifier {
	return c.roundNotifier
}

// EpochStartNotifierWithConfirm will return the epoch start notifier with confirm
func (c *coreComponentsHolder) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	return c.epochStartNotifierWithConfirm
}

// ChanStopNodeProcess will return the channel for stop node process
func (c *coreComponentsHolder) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	return c.chanStopNodeProcess
}

// GenesisTime will return the genesis time
func (c *coreComponentsHolder) GenesisTime() time.Time {
	return c.genesisTime
}

// ChainID will return the chain id
func (c *coreComponentsHolder) ChainID() string {
	return c.chainID
}

// MinTransactionVersion will return the min transaction version
func (c *coreComponentsHolder) MinTransactionVersion() uint32 {
	return c.minTransactionVersion
}

// TxVersionChecker will return the tx version checker
func (c *coreComponentsHolder) TxVersionChecker() process.TxVersionCheckerHandler {
	return c.txVersionChecker
}

// EncodedAddressLen will return the len of encoded address
func (c *coreComponentsHolder) EncodedAddressLen() uint32 {
	return c.encodedAddressLen
}

// NodeTypeProvider will return the node type provider
func (c *coreComponentsHolder) NodeTypeProvider() core.NodeTypeProviderHandler {
	return c.nodeTypeProvider
}

// WasmVMChangeLocker will return the wasm vm change locker
func (c *coreComponentsHolder) WasmVMChangeLocker() common.Locker {
	return c.wasmVMChangeLocker
}

// ProcessStatusHandler will return the process status handler
func (c *coreComponentsHolder) ProcessStatusHandler() common.ProcessStatusHandler {
	return c.processStatusHandler
}

// HardforkTriggerPubKey will return the pub key for the hard fork trigger
func (c *coreComponentsHolder) HardforkTriggerPubKey() []byte {
	return c.hardforkTriggerPubKey
}

// EnableEpochsHandler will return the enable epoch handler
func (c *coreComponentsHolder) EnableEpochsHandler() common.EnableEpochsHandler {
	return c.enableEpochsHandler
}

func (c *coreComponentsHolder) collectClosableComponents() {
	c.closeHandler.AddComponent(c.alarmScheduler)
	c.closeHandler.AddComponent(c.syncTimer)
}

// Close will call the Close methods on all inner components
func (c *coreComponentsHolder) Close() error {
	return c.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *coreComponentsHolder) IsInterfaceNil() bool {
	return c == nil
}

// Create will do nothing
func (c *coreComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (c *coreComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (c *coreComponentsHolder) String() string {
	return ""
}
