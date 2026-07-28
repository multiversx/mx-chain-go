package components

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"

	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/debug/handler"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	Configs      config.Configs
	APIInterface APIConfigurator

	ChanStopNodeProcess    chan endProcess.ArgEndProcess
	SyncedBroadcastNetwork SyncedBroadcastNetworkHandler
	Monitor                factory.HeartbeatV2Monitor

	InitialRound                int64
	InitialNonce                uint64
	GasScheduleFilename         string
	NumShards                   uint32
	ShardIDStr                  string
	BypassTxSignatureCheck      bool
	BypassBlockSignatureCheck   bool
	MinNodesPerShard            uint32
	ConsensusGroupSize          uint32
	MinNodesMeta                uint32
	MetaChainConsensusGroupSize uint32
	RoundDurationInMillis       uint64
	VmQueryDelayAfterStartInMs  uint64
	GenesisTime                 time.Time
	BypassCreateBlockTimeCheck  bool
	CreateBlockMaxTimePercent   float64
	// EnableConsensus, when true, builds the node with a manual sync timer and the
	// production round handler so its chronology/SPoS subrounds can be driven manually, and
	// makes the simulator produce blocks through the consensus stack instead of the direct
	// block-processor invocation
	EnableConsensus bool
	// EnableFastConsensusCrypto replaces only the simulator node's BLS signing operations with
	// deterministic hash-based signatures while preserving consensus signature handling.
	EnableFastConsensusCrypto bool
	// ValidatorKeysPemFileOverride, when set, makes the node load its managed validator keys
	// from this PEM file instead of the all-validators PEM. Used to build single-key consensus
	// nodes (one physical node per validator) for real multi-party consensus (S5 phase B).
	ValidatorKeysPemFileOverride string
}

type testOnlyProcessingNode struct {
	closeHandler              *closeHandler
	CoreComponentsHolder      factory.CoreComponentsHandler
	StatusCoreComponents      factory.StatusCoreComponentsHandler
	StateComponentsHolder     factory.StateComponentsHandler
	StatusComponentsHolder    factory.StatusComponentsHandler
	CryptoComponentsHolder    factory.CryptoComponentsHandler
	NetworkComponentsHolder   factory.NetworkComponentsHandler
	BootstrapComponentsHolder factory.BootstrapComponentsHandler
	ProcessComponentsHolder   factory.ProcessComponentsHandler
	DataComponentsHolder      factory.DataComponentsHandler

	NodesCoordinator       nodesCoordinator.NodesCoordinator
	ChainHandler           chainData.ChainHandler
	ArgumentsParser        process.ArgumentsParser
	TransactionFeeHandler  process.TransactionFeeHandler
	StoreService           dataRetriever.StorageService
	DataPool               dataRetriever.PoolsHolder
	broadcastMessenger     consensus.BroadcastMessenger
	syncedBroadcastNetwork SyncedBroadcastNetworkHandler
	enableConsensus        bool

	httpServer    shared.UpgradeableHttpServerHandler
	facadeHandler shared.FacadeHandler

	basePeers map[uint32]core.PeerID

	// consensusDrive is set only in consensus-path execution mode; it lets the simulator step
	// this node's chronology and SPoS subrounds forward one round at a time
	consensusDrive *nodeConsensusDrive
}

// NewTestOnlyProcessingNode creates a new instance of a node that is able to only process transactions
func NewTestOnlyProcessingNode(args ArgsTestOnlyProcessingNode) (*testOnlyProcessingNode, error) {
	instance := &testOnlyProcessingNode{
		ArgumentsParser:        smartContract.NewArgumentParser(),
		StoreService:           CreateStore(args.NumShards),
		closeHandler:           NewCloseHandler(),
		syncedBroadcastNetwork: args.SyncedBroadcastNetwork,
		enableConsensus:        args.EnableConsensus,
	}

	var err error
	instance.TransactionFeeHandler = postprocess.NewFeeAccumulator()

	instance.CoreComponentsHolder, err = CreateCoreComponents(ArgsCoreComponentsHolder{
		Config:                      *args.Configs.GeneralConfig,
		EnableEpochsConfig:          args.Configs.EpochConfig.EnableEpochs,
		RoundsConfig:                *args.Configs.RoundConfig,
		EconomicsConfig:             *args.Configs.EconomicsConfig,
		ChanStopNodeProcess:         args.ChanStopNodeProcess,
		NumShards:                   args.NumShards,
		WorkingDir:                  args.Configs.FlagsConfig.WorkingDir,
		GasScheduleFilename:         args.GasScheduleFilename,
		NodesSetupPath:              args.Configs.ConfigurationPathsHolder.Nodes,
		InitialRound:                args.InitialRound,
		MinNodesPerShard:            args.MinNodesPerShard,
		ConsensusGroupSize:          args.ConsensusGroupSize,
		MinNodesMeta:                args.MinNodesMeta,
		MetaChainConsensusGroupSize: args.MetaChainConsensusGroupSize,
		RoundDurationInMs:           args.RoundDurationInMillis,
		RatingConfig:                *args.Configs.RatingsConfig,
		GenesisTime:                 args.GenesisTime,
		PrintPrettifiedHeader:       args.Configs.FlagsConfig.PrintPrettifiedHeader,
		EnableConsensus:             args.EnableConsensus,
	})
	if err != nil {
		return nil, err
	}

	instance.StatusCoreComponents, err = CreateStatusCoreComponents(args.Configs, instance.CoreComponentsHolder)
	if err != nil {
		return nil, err
	}

	allValidatorKeysPemFile := args.Configs.ConfigurationPathsHolder.AllValidatorKeys
	validatorKeyPemFile := ""
	if len(args.ValidatorKeysPemFileOverride) > 0 {
		// consensus mode with group size > 1 builds N single-key nodes per shard: each node is a
		// single-key validator loading only its own BLS key, instead of a multikey node managing
		// all of them. A single-key node stamps its consensus messages with the node's physical
		// p2p identity (multikey would use per-key virtual identities the in-memory network has no
		// way to deliver as), so leader selection and BLS multi-signing run across distinct nodes.
		validatorKeyPemFile = args.ValidatorKeysPemFileOverride
		allValidatorKeysPemFile = "missing.pem"
	}

	instance.CryptoComponentsHolder, err = CreateCryptoComponents(ArgsCryptoComponentsHolder{
		Config:                      *args.Configs.GeneralConfig,
		EnableEpochsConfig:          args.Configs.EpochConfig.EnableEpochs,
		Preferences:                 *args.Configs.PreferencesConfig,
		CoreComponentsHolder:        instance.CoreComponentsHolder,
		BypassTxSignatureCheck:      args.BypassTxSignatureCheck,
		BypassBlockSignatureCheck:   args.BypassBlockSignatureCheck,
		EnableFastConsensusCrypto:   args.EnableFastConsensusCrypto,
		AllValidatorKeysPemFileName: allValidatorKeysPemFile,
		ValidatorKeyPemFileName:     validatorKeyPemFile,
	})
	if err != nil {
		return nil, err
	}

	if len(args.ValidatorKeysPemFileOverride) > 0 {
		// a single-key consensus node must broadcast under the same peer ID the keys handler
		// associates with its validator key (the crypto p2p identity); otherwise consensus
		// messages fail the originator check (the in-memory network stamps the envelope pid from
		// the messenger, and the worker compares it against the message's keys-handler pid)
		var pid core.PeerID
		pid, err = p2pFactory.NewP2PKeyConverter().ConvertPublicKeyToPeerID(instance.CryptoComponentsHolder.P2pPublicKey())
		if err != nil {
			return nil, err
		}
		instance.NetworkComponentsHolder, err = CreateNetworkComponentsWithPeerID(args.SyncedBroadcastNetwork, pid)
	} else {
		instance.NetworkComponentsHolder, err = CreateNetworkComponents(args.SyncedBroadcastNetwork)
	}
	if err != nil {
		return nil, err
	}

	instance.BootstrapComponentsHolder, err = CreateBootstrapComponents(ArgsBootstrapComponentsHolder{
		CoreComponents:       instance.CoreComponentsHolder,
		CryptoComponents:     instance.CryptoComponentsHolder,
		NetworkComponents:    instance.NetworkComponentsHolder,
		StatusCoreComponents: instance.StatusCoreComponents,
		WorkingDir:           args.Configs.FlagsConfig.WorkingDir,
		FlagsConfig:          *args.Configs.FlagsConfig,
		ImportDBConfig:       *args.Configs.ImportDbConfig,
		PrefsConfig:          *args.Configs.PreferencesConfig,
		Config:               *args.Configs.GeneralConfig,
		ShardIDStr:           args.ShardIDStr,
	})
	if err != nil {
		return nil, err
	}

	selfShardID := instance.GetShardCoordinator().SelfId()

	statusComponentsH, err := CreateStatusComponents(
		selfShardID,
		instance.StatusCoreComponents.AppStatusHandler(),
		args.Configs.GeneralConfig.GeneralSettings.StatusPollingIntervalSec,
		*args.Configs.ExternalConfig,
		instance.CoreComponentsHolder,
	)
	if err != nil {
		return nil, err
	}

	instance.StatusComponentsHolder = statusComponentsH

	err = instance.createBlockChain(selfShardID)
	if err != nil {
		return nil, err
	}

	instance.StateComponentsHolder, err = CreateStateComponents(ArgsStateComponents{
		Config:         *args.Configs.GeneralConfig,
		CoreComponents: instance.CoreComponentsHolder,
		StatusCore:     instance.StatusCoreComponents,
		StoreService:   instance.StoreService,
		ChainHandler:   instance.ChainHandler,
	})
	if err != nil {
		return nil, err
	}

	instance.DataPool, err = dataRetrieverFactory.NewDataPoolFromConfig(dataRetrieverFactory.ArgsDataPool{
		Config:           args.Configs.GeneralConfig,
		EconomicsData:    instance.CoreComponentsHolder.EconomicsData(),
		ShardCoordinator: instance.BootstrapComponentsHolder.ShardCoordinator(),
		Marshalizer:      instance.CoreComponentsHolder.InternalMarshalizer(),
		PathManager:      instance.CoreComponentsHolder.PathHandler(),
	})
	if err != nil {
		return nil, err
	}
	if instance.enableConsensus {
		instance.DataPool = newPoolsHolderWithSyncHeaders(instance.DataPool)
	}

	err = instance.createNodesCoordinator(args.Configs.PreferencesConfig.Preferences, *args.Configs.GeneralConfig)
	if err != nil {
		return nil, err
	}

	statusComponentsH.SetNodesCoordinator(instance.NodesCoordinator)

	instance.DataComponentsHolder, err = CreateDataComponents(ArgsDataComponentsHolder{
		Chain:              instance.ChainHandler,
		StorageService:     instance.StoreService,
		DataPool:           instance.DataPool,
		InternalMarshaller: instance.CoreComponentsHolder.InternalMarshalizer(),
	})
	if err != nil {
		return nil, err
	}

	processComponentsHolder, err := CreateProcessComponents(ArgsProcessComponentsHolder{
		CoreComponents:           instance.CoreComponentsHolder,
		CryptoComponents:         instance.CryptoComponentsHolder,
		NetworkComponents:        instance.NetworkComponentsHolder,
		BootstrapComponents:      instance.BootstrapComponentsHolder,
		StateComponents:          instance.StateComponentsHolder,
		StatusComponents:         instance.StatusComponentsHolder,
		StatusCoreComponents:     instance.StatusCoreComponents,
		FlagsConfig:              *args.Configs.FlagsConfig,
		ImportDBConfig:           *args.Configs.ImportDbConfig,
		PrefsConfig:              *args.Configs.PreferencesConfig,
		Config:                   *args.Configs.GeneralConfig,
		EconomicsConfig:          *args.Configs.EconomicsConfig,
		SystemSCConfig:           *args.Configs.SystemSCConfig,
		EpochConfig:              *args.Configs.EpochConfig,
		RoundConfig:              *args.Configs.RoundConfig,
		ConfigurationPathsHolder: *args.Configs.ConfigurationPathsHolder,
		NodesCoordinator:         instance.NodesCoordinator,
		DataComponents:           instance.DataComponentsHolder,
		GenesisNonce:             args.InitialNonce,
		GenesisRound:             uint64(args.InitialRound),
	})
	if err != nil {
		return nil, err
	}
	if args.EnableConsensus && !args.BypassCreateBlockTimeCheck {
		processComponentsHolder.blockProcessor = newCreateBlockTimeBoundProcessor(
			processComponentsHolder.blockProcessor,
			instance.CoreComponentsHolder.RoundHandler(),
			args.CreateBlockMaxTimePercent,
		)
	}
	instance.ProcessComponentsHolder = processComponentsHolder

	err = instance.StatusComponentsHolder.SetForkDetector(instance.ProcessComponentsHolder.ForkDetector())
	if err != nil {
		return nil, err
	}

	err = instance.StatusComponentsHolder.StartPolling()
	if err != nil {
		return nil, err
	}

	err = instance.createBroadcastMessenger()
	if err != nil {
		return nil, err
	}

	if args.EnableConsensus {
		instance.consensusDrive, err = instance.createConsensusComponents(*args.Configs.GeneralConfig)
		if err != nil {
			return nil, err
		}
	}

	err = instance.createFacade(args.Configs, args.APIInterface, args.VmQueryDelayAfterStartInMs, args.Monitor)
	if err != nil {
		return nil, err
	}

	err = instance.createHttpServer(args.Configs)
	if err != nil {
		return nil, err
	}

	err = instance.createInterceptorDebugHandler(args.Configs)
	if err != nil {
		return nil, err
	}

	instance.collectClosableComponents()

	return instance, nil
}

func (node *testOnlyProcessingNode) createInterceptorDebugHandler(configs config.Configs) error {
	debugHandler, err := handler.NewInterceptorDebugHandler(configs.GeneralConfig.Debug.InterceptorResolver, node.CoreComponentsHolder.SyncTimer())
	if err != nil {
		return err
	}

	node.CoreComponentsHolder.EpochStartNotifierWithConfirm().RegisterHandler(debugHandler.EpochStartEventHandler())

	var errFound error
	node.ProcessComponentsHolder.InterceptorsContainer().Iterate(func(key string, interceptor process.Interceptor) bool {
		err = interceptor.SetInterceptedDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on interceptors", errFound)
	}

	return nil
}

func (node *testOnlyProcessingNode) createBlockChain(selfShardID uint32) error {
	var err error
	if selfShardID == core.MetachainShardId {
		node.ChainHandler, err = blockchain.NewMetaChain(node.StatusCoreComponents.AppStatusHandler())
	} else {
		node.ChainHandler, err = blockchain.NewBlockChain(node.StatusCoreComponents.AppStatusHandler())
	}

	return err
}

func (node *testOnlyProcessingNode) createNodesCoordinator(pref config.PreferencesConfig, generalConfig config.Config) error {
	nodesShufflerOut, err := bootstrapComp.CreateNodesShuffleOut(
		node.CoreComponentsHolder.GenesisNodesSetup(),
		generalConfig.EpochStartConfig,
		node.CoreComponentsHolder.ChanStopNodeProcess(),
		node.CoreComponentsHolder.ChainParametersHandler(),
	)
	if err != nil {
		return err
	}

	bootstrapStorer, err := node.StoreService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

	shardID := node.BootstrapComponentsHolder.ShardCoordinator().SelfId()
	shardIDStr := fmt.Sprintf("%d", shardID)
	if shardID == core.MetachainShardId {
		shardIDStr = "metachain"
	}

	pref.DestinationShardAsObserver = shardIDStr

	node.NodesCoordinator, err = bootstrapComp.CreateNodesCoordinator(
		nodesShufflerOut,
		node.CoreComponentsHolder.GenesisNodesSetup(),
		pref,
		node.CoreComponentsHolder.EpochStartNotifierWithConfirm(),
		node.CryptoComponentsHolder.PublicKey(),
		node.CoreComponentsHolder.InternalMarshalizer(),
		node.CoreComponentsHolder.Hasher(),
		node.CoreComponentsHolder.Rater(),
		bootstrapStorer,
		node.CoreComponentsHolder.NodesShuffler(),
		node.BootstrapComponentsHolder.ShardCoordinator().SelfId(),
		node.BootstrapComponentsHolder.EpochBootstrapParams(),
		node.BootstrapComponentsHolder.EpochBootstrapParams().Epoch(),
		node.CoreComponentsHolder.ChanStopNodeProcess(),
		node.CoreComponentsHolder.NodeTypeProvider(),
		node.CoreComponentsHolder.EnableEpochsHandler(),
		node.DataPool.CurrentEpochValidatorInfo(),
		node.BootstrapComponentsHolder.NodesCoordinatorRegistryFactory(),
		node.CoreComponentsHolder.ChainParametersHandler(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (node *testOnlyProcessingNode) createBroadcastMessenger() error {
	broadcastMessenger, err := sposFactory.GetBroadcastMessenger(
		node.CoreComponentsHolder.InternalMarshalizer(),
		node.CoreComponentsHolder.Hasher(),
		node.NetworkComponentsHolder.NetworkMessenger(),
		node.ProcessComponentsHolder.ShardCoordinator(),
		node.CryptoComponentsHolder.PeerSignatureHandler(),
		node.DataComponentsHolder.Datapool().Headers(),
		node.DataComponentsHolder.Datapool().Proofs(),
		node.CoreComponentsHolder.EnableEpochsHandler(),
		node.ProcessComponentsHolder.InterceptorsContainer(),
		node.CoreComponentsHolder.AlarmScheduler(),
		node.CryptoComponentsHolder.KeysHandler(),
	)
	if err != nil {
		return err
	}

	instantMessenger, err := NewInstantBroadcastMessenger(broadcastMessenger, node.BootstrapComponentsHolder.ShardCoordinator())
	if err != nil {
		return err
	}

	if node.enableConsensus {
		deliveryTracker, _ := node.syncedBroadcastNetwork.(blockBodyDeliveryTracker)
		instantMessenger.setBlockBodyDeliveryTracker(deliveryTracker)
		headerDeliveryTracker, _ := node.syncedBroadcastNetwork.(blockHeaderDeliveryTracker)
		instantMessenger.setBlockHeaderDeliveryTracker(headerDeliveryTracker)

		shardID := node.BootstrapComponentsHolder.ShardCoordinator().SelfId()
		node.syncedBroadcastNetwork.RegisterHeaderNotifier(shardID, node.CoreComponentsHolder.EpochNotifier().CheckEpoch)
		instantMessenger.setBeforeBroadcastHeader(func(header chainData.HeaderHandler) {
			node.syncedBroadcastNetwork.NotifyHeader(shardID, header)
		})
		instantMessenger.setProposalDataHandler(func(
			header chainData.HeaderHandler,
			bodyBytes []byte,
			pkBytes []byte,
		) error {
			if header.IsHeaderV3() {
				return nil
			}

			blockProcessor := node.ProcessComponentsHolder.BlockProcessor()
			body := blockProcessor.DecodeBlockBody(bodyBytes)
			if body == nil {
				return errors.New("cannot decode consensus proposal body")
			}

			headerHash, errHash := core.CalculateHash(
				node.CoreComponentsHolder.InternalMarshalizer(),
				node.CoreComponentsHolder.Hasher(),
				header,
			)
			if errHash != nil {
				return errHash
			}

			miniBlocks, transactions, errPrepare := blockProcessor.MarshalizedDataToBroadcast(
				headerHash,
				header,
				body,
			)
			if errPrepare != nil {
				return errPrepare
			}

			selfBodyBytes, errMarshal := node.CoreComponentsHolder.InternalMarshalizer().Marshal(body)
			if errMarshal != nil {
				return errMarshal
			}
			if miniBlocks == nil {
				miniBlocks = make(map[uint32][]byte)
			}
			miniBlocks[shardID] = selfBodyBytes

			return instantMessenger.broadcastMiniblockData(miniBlocks, transactions, pkBytes)
		})
	}
	node.broadcastMessenger = instantMessenger

	return nil
}

// GetProcessComponents will return the process components
func (node *testOnlyProcessingNode) GetProcessComponents() factory.ProcessComponentsHolder {
	return node.ProcessComponentsHolder
}

// GetChainHandler will return the chain handler
func (node *testOnlyProcessingNode) GetChainHandler() chainData.ChainHandler {
	return node.ChainHandler
}

// GetBroadcastMessenger will return the broadcast messenger
func (node *testOnlyProcessingNode) GetBroadcastMessenger() consensus.BroadcastMessenger {
	return node.broadcastMessenger
}

// AdvanceConsensusClock bumps the node's manual round clock by one round, starting a new consensus
// round. It is only valid in consensus-path execution mode.
func (node *testOnlyProcessingNode) AdvanceConsensusClock() error {
	if node.consensusDrive == nil {
		return errNodeNotInConsensusMode
	}

	node.consensusDrive.advanceClock()
	return nil
}

// RearmConsensusRound prepares the chronology to retry the current manual-clock round. It is only
// valid in consensus-path execution mode and does not advance the clock.
func (node *testOnlyProcessingNode) RearmConsensusRound() error {
	if node.consensusDrive == nil {
		return errNodeNotInConsensusMode
	}

	node.consensusDrive.rearmCurrentRound()
	return nil
}

// StepConsensusSubround steps the node's chronology forward by one subround. The simulator
// interleaves steps across a shard's nodes so their consensus messages flow and quorum is reached.
func (node *testOnlyProcessingNode) StepConsensusSubround() error {
	if node.consensusDrive == nil {
		return errNodeNotInConsensusMode
	}

	return node.consensusDrive.step()
}

// WaitConsensusSubround waits for the subround started by StepConsensusSubround to return.
func (node *testOnlyProcessingNode) WaitConsensusSubround() error {
	if node.consensusDrive == nil {
		return errNodeNotInConsensusMode
	}

	return node.consensusDrive.waitStep()
}

// ConsensusDriveState returns the current chronology subround and restart generation.
func (node *testOnlyProcessingNode) ConsensusDriveState() (int, uint64, error) {
	if node.consensusDrive == nil {
		return 0, 0, errNodeNotInConsensusMode
	}

	subround, generation := node.consensusDrive.state()
	return subround, generation, nil
}

// GetShardCoordinator will return the shard coordinator
func (node *testOnlyProcessingNode) GetShardCoordinator() sharding.Coordinator {
	return node.BootstrapComponentsHolder.ShardCoordinator()
}

// GetCryptoComponents will return the crypto components
func (node *testOnlyProcessingNode) GetCryptoComponents() factory.CryptoComponentsHolder {
	return node.CryptoComponentsHolder
}

// GetCoreComponents will return the core components
func (node *testOnlyProcessingNode) GetCoreComponents() factory.CoreComponentsHolder {
	return node.CoreComponentsHolder
}

// GetDataComponents will return the data components
func (node *testOnlyProcessingNode) GetDataComponents() factory.DataComponentsHolder {
	return node.DataComponentsHolder
}

// GetStateComponents will return the state components
func (node *testOnlyProcessingNode) GetStateComponents() factory.StateComponentsHolder {
	return node.StateComponentsHolder
}

// GetFacadeHandler will return the facade handler
func (node *testOnlyProcessingNode) GetFacadeHandler() shared.FacadeHandler {
	return node.facadeHandler
}

// GetStatusCoreComponents will return the status core components
func (node *testOnlyProcessingNode) GetStatusCoreComponents() factory.StatusCoreComponentsHolder {
	return node.StatusCoreComponents
}

// GetNetworkComponents will return the network components
func (node *testOnlyProcessingNode) GetNetworkComponents() factory.NetworkComponentsHolder {
	return node.NetworkComponentsHolder
}

func (node *testOnlyProcessingNode) collectClosableComponents() {
	node.closeHandler.AddComponent(node.ProcessComponentsHolder)
	node.closeHandler.AddComponent(node.DataComponentsHolder)
	node.closeHandler.AddComponent(node.StateComponentsHolder)
	node.closeHandler.AddComponent(node.StatusComponentsHolder)
	node.closeHandler.AddComponent(node.BootstrapComponentsHolder)
	node.closeHandler.AddComponent(node.NetworkComponentsHolder)
	node.closeHandler.AddComponent(node.StatusCoreComponents)
	node.closeHandler.AddComponent(node.CoreComponentsHolder)
	node.closeHandler.AddComponent(node.facadeHandler)

	if facade.DefaultRestPortOff != node.facadeHandler.RestApiInterface() {
		node.closeHandler.AddComponent(node.httpServer)
	}
}

// SetKeyValueForAddress will set the provided state for the given address
func (node *testOnlyProcessingNode) SetKeyValueForAddress(address []byte, keyValueMap map[string]string) error {
	userAccount, err := node.getUserAccount(address)
	if err != nil {
		return err
	}

	err = setKeyValueMap(userAccount, keyValueMap)
	if err != nil {
		return err
	}

	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	err = accountsAdapter.SaveAccount(userAccount)
	if err != nil {
		return err
	}

	_, err = accountsAdapter.Commit()

	return err
}

func setKeyValueMap(userAccount state.UserAccountHandler, keyValueMap map[string]string) error {
	for keyHex, valueHex := range keyValueMap {
		keyDecoded, err := hex.DecodeString(keyHex)
		if err != nil {
			return fmt.Errorf("cannot decode key, error: %w", err)
		}
		valueDecoded, err := hex.DecodeString(valueHex)
		if err != nil {
			return fmt.Errorf("cannot decode value, error: %w", err)
		}

		err = userAccount.SaveKeyValue(keyDecoded, valueDecoded)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetStateForAddress will set the state for the give address
func (node *testOnlyProcessingNode) SetStateForAddress(address []byte, addressState *dtos.AddressState) error {
	userAccount, err := node.getUserAccount(address)
	if err != nil {
		return err
	}

	err = setNonceAndBalanceForAccount(userAccount, addressState.Nonce, addressState.Balance)
	if err != nil {
		return err
	}

	err = setKeyValueMap(userAccount, addressState.Pairs)
	if err != nil {
		return err
	}

	err = node.setScDataIfNeeded(address, userAccount, addressState)
	if err != nil {
		return err
	}

	rootHash, err := base64.StdEncoding.DecodeString(addressState.RootHash)
	if err != nil {
		return err
	}
	if len(rootHash) != 0 {
		userAccount.SetRootHash(rootHash)
	}

	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	err = accountsAdapter.SaveAccount(userAccount)
	if err != nil {
		return err
	}

	newRootHash, err := accountsAdapter.Commit()
	node.setBlockchainRootHashIfSupernovaIsActive(newRootHash)

	return err
}

func (node *testOnlyProcessingNode) setBlockchainRootHashIfSupernovaIsActive(
	rootHash []byte,
) {
	if !node.CoreComponentsHolder.EnableRoundsHandler().IsFlagEnabled(common.SupernovaRoundFlag) {
		return
	}

	header := node.ChainHandler.GetLastExecutedBlockHeader()
	_, hash, _ := node.ChainHandler.GetLastExecutedBlockInfo()
	node.ChainHandler.SetLastExecutedBlockHeaderAndRootHash(header, hash, rootHash)

	lastExecutionResult := node.ChainHandler.GetLastExecutionResult()

	metaResult, isMeta := lastExecutionResult.(*block.MetaExecutionResult)
	if isMeta {
		metaResult.ExecutionResult.BaseExecutionResult.RootHash = rootHash
		node.ChainHandler.SetLastExecutionInfo(header, metaResult)
		return
	}

	shardResult, isShard := lastExecutionResult.(*block.ExecutionResult)
	if isShard {
		shardResult.BaseExecutionResult.RootHash = rootHash
		node.ChainHandler.SetLastExecutionInfo(header, shardResult)
		return
	}

	updatedLastExecutionResult := &block.BaseExecutionResult{
		HeaderHash:  lastExecutionResult.GetHeaderHash(),
		HeaderNonce: lastExecutionResult.GetHeaderNonce(),
		HeaderRound: lastExecutionResult.GetHeaderRound(),
		HeaderEpoch: lastExecutionResult.GetHeaderEpoch(),
		RootHash:    rootHash,
		GasUsed:     lastExecutionResult.GetGasUsed(),
	}
	node.ChainHandler.SetLastExecutionInfo(header, updatedLastExecutionResult)
}

// RemoveAccount will remove the account for the given address
func (node *testOnlyProcessingNode) RemoveAccount(address []byte) error {
	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	err := accountsAdapter.RemoveAccount(address)
	if err != nil {
		return err
	}

	_, err = accountsAdapter.Commit()
	return err
}

// ForceChangeOfEpoch will force change of epoch
func (node *testOnlyProcessingNode) ForceChangeOfEpoch() error {
	currentHeader := node.DataComponentsHolder.Blockchain().GetCurrentBlockHeader()
	if currentHeader == nil {
		currentHeader = node.DataComponentsHolder.Blockchain().GetGenesisHeader()
	}

	node.ProcessComponentsHolder.EpochStartTrigger().ForceEpochStart(currentHeader.GetRound() + 1)

	return nil
}

func setNonceAndBalanceForAccount(userAccount state.UserAccountHandler, nonce *uint64, balance string) error {
	if nonce != nil {
		// set nonce to zero
		userAccount.IncreaseNonce(-userAccount.GetNonce())
		// set nonce with the provided value
		userAccount.IncreaseNonce(*nonce)
	}

	if balance == "" {
		return nil
	}

	providedBalance, ok := big.NewInt(0).SetString(balance, 10)
	if !ok {
		return errors.New("cannot convert string balance to *big.Int")
	}

	// set balance to zero
	userBalance := userAccount.GetBalance()
	err := userAccount.AddToBalance(userBalance.Neg(userBalance))
	if err != nil {
		return err
	}
	// set provided balance
	return userAccount.AddToBalance(providedBalance)
}

func (node *testOnlyProcessingNode) setScDataIfNeeded(address []byte, userAccount state.UserAccountHandler, addressState *dtos.AddressState) error {
	if !core.IsSmartContractAddress(address) {
		return nil
	}

	if addressState.Code != "" {
		decodedCode, err := hex.DecodeString(addressState.Code)
		if err != nil {
			return err
		}
		userAccount.SetCode(decodedCode)
	}

	if addressState.CodeHash != "" {
		codeHash, errD := base64.StdEncoding.DecodeString(addressState.CodeHash)
		if errD != nil {
			return errD
		}
		userAccount.SetCodeHash(codeHash)
	}

	if addressState.CodeMetadata != "" {
		decodedCodeMetadata, errD := base64.StdEncoding.DecodeString(addressState.CodeMetadata)
		if errD != nil {
			return errD
		}
		userAccount.SetCodeMetadata(decodedCodeMetadata)
	}

	if addressState.Owner != "" {
		ownerAddress, errD := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(addressState.Owner)
		if errD != nil {
			return errD
		}
		userAccount.SetOwnerAddress(ownerAddress)
	}

	if addressState.DeveloperRewards != "" {
		developerRewards, ok := big.NewInt(0).SetString(addressState.DeveloperRewards, 10)
		if !ok {
			return errors.New("cannot convert string developer rewards to *big.Int")
		}
		userAccount.AddToDeveloperReward(developerRewards)
	}

	return nil
}

func (node *testOnlyProcessingNode) getUserAccount(address []byte) (state.UserAccountHandler, error) {
	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	account, err := accountsAdapter.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, errors.New("cannot cast AccountHandler to UserAccountHandler")
	}

	return userAccount, nil
}

// GetBasePeers returns return network messenger ids for base nodes
func (node *testOnlyProcessingNode) GetBasePeers() map[uint32]core.PeerID {
	return node.basePeers
}

// SetBasePeers will set base network messenger id nodes per shard
func (node *testOnlyProcessingNode) SetBasePeers(basePeers map[uint32]core.PeerID) {
	node.basePeers = basePeers
}

// Close will call the Close methods on all inner components
func (node *testOnlyProcessingNode) Close() error {
	return node.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (node *testOnlyProcessingNode) IsInterfaceNil() bool {
	return node == nil
}
