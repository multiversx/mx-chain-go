package components

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	Config                   config.Config
	EpochConfig              config.EpochConfig
	EconomicsConfig          config.EconomicsConfig
	RoundsConfig             config.RoundConfig
	PreferencesConfig        config.Preferences
	ImportDBConfig           config.ImportDbConfig
	ContextFlagsConfig       config.ContextFlagsConfig
	SystemSCConfig           config.SystemSmartContractsConfig
	ConfigurationPathsHolder config.ConfigurationPathsHolder

	ChanStopNodeProcess    chan endProcess.ArgEndProcess
	SyncedBroadcastNetwork SyncedBroadcastNetworkHandler

	GasScheduleFilename string

	NumShards uint32
	ShardID   uint32
	SkIndex   int
}

type testOnlyProcessingNode struct {
	CoreComponentsHolder      factory.CoreComponentsHolder
	StatusCoreComponents      factory.StatusCoreComponentsHolder
	StateComponentsHolder     factory.StateComponentsHolder
	StatusComponentsHolder    factory.StatusComponentsHolder
	CryptoComponentsHolder    factory.CryptoComponentsHolder
	NetworkComponentsHolder   factory.NetworkComponentsHolder
	BootstrapComponentsHolder factory.BootstrapComponentsHolder
	ProcessComponentsHolder   factory.ProcessComponentsHolder
	DataComponentsHolder      factory.DataComponentsHolder

	NodesCoordinator            nodesCoordinator.NodesCoordinator
	ChainHandler                chainData.ChainHandler
	ShardCoordinator            sharding.Coordinator
	ArgumentsParser             process.ArgumentsParser
	TransactionFeeHandler       process.TransactionFeeHandler
	StoreService                dataRetriever.StorageService
	BuiltinFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	DataPool                    dataRetriever.PoolsHolder
	TxLogsProcessor             process.TransactionLogProcessor
}

// NewTestOnlyProcessingNode creates a new instance of a node that is able to only process transactions
func NewTestOnlyProcessingNode(args ArgsTestOnlyProcessingNode) (*testOnlyProcessingNode, error) {
	instance := &testOnlyProcessingNode{
		ArgumentsParser: smartContract.NewArgumentParser(),
		StoreService:    CreateStore(args.NumShards),
	}
	err := instance.createBasicComponents(args.NumShards, args.ShardID)
	if err != nil {
		return nil, err
	}

	instance.CoreComponentsHolder, err = CreateCoreComponentsHolder(ArgsCoreComponentsHolder{
		Config:              args.Config,
		EnableEpochsConfig:  args.EpochConfig.EnableEpochs,
		RoundsConfig:        args.RoundsConfig,
		EconomicsConfig:     args.EconomicsConfig,
		ChanStopNodeProcess: args.ChanStopNodeProcess,
		NumShards:           args.NumShards,
		WorkingDir:          args.ContextFlagsConfig.WorkingDir,
		GasScheduleFilename: args.GasScheduleFilename,
		NodesSetupPath:      args.ConfigurationPathsHolder.Nodes,
	})
	if err != nil {
		return nil, err
	}

	instance.StatusCoreComponents, err = CreateStatusCoreComponentsHolder(args.Config, instance.CoreComponentsHolder)
	if err != nil {
		return nil, err
	}

	err = instance.createBlockChain(args.ShardID)
	if err != nil {
		return nil, err
	}

	instance.StateComponentsHolder, err = CreateStateComponents(ArgsStateComponents{
		Config:         args.Config,
		CoreComponents: instance.CoreComponentsHolder,
		StatusCore:     instance.StatusCoreComponents,
		StoreService:   instance.StoreService,
		ChainHandler:   instance.ChainHandler,
	})
	if err != nil {
		return nil, err
	}
	instance.StatusComponentsHolder, err = CreateStatusComponentsHolder(args.ShardID)
	if err != nil {
		return nil, err
	}
	instance.CryptoComponentsHolder, err = CreateCryptoComponentsHolder(ArgsCryptoComponentsHolder{
		Config:                  args.Config,
		EnableEpochsConfig:      args.EpochConfig.EnableEpochs,
		Preferences:             args.PreferencesConfig,
		CoreComponentsHolder:    instance.CoreComponentsHolder,
		ValidatorKeyPemFileName: args.ConfigurationPathsHolder.ValidatorKey,
		SkIndex:                 args.SkIndex,
	})
	if err != nil {
		return nil, err
	}

	instance.NetworkComponentsHolder, err = CreateNetworkComponentsHolder(args.SyncedBroadcastNetwork)
	if err != nil {
		return nil, err
	}

	instance.BootstrapComponentsHolder, err = CreateBootstrapComponentHolder(ArgsBootstrapComponentsHolder{
		CoreComponents:       instance.CoreComponentsHolder,
		CryptoComponents:     instance.CryptoComponentsHolder,
		NetworkComponents:    instance.NetworkComponentsHolder,
		StatusCoreComponents: instance.StatusCoreComponents,
		WorkingDir:           args.ContextFlagsConfig.WorkingDir,
		FlagsConfig:          args.ContextFlagsConfig,
		ImportDBConfig:       args.ImportDBConfig,
		PrefsConfig:          args.PreferencesConfig,
		Config:               args.Config,
	})
	if err != nil {
		return nil, err
	}

	err = instance.createDataPool(args)
	if err != nil {
		return nil, err
	}
	err = instance.createTransactionLogProcessor()
	if err != nil {
		return nil, err
	}

	err = instance.createNodesCoordinator(args.PreferencesConfig.Preferences, args.Config)
	if err != nil {
		return nil, err
	}

	instance.DataComponentsHolder, err = CreateDataComponentsHolder(ArgsDataComponentsHolder{
		Chain:              instance.ChainHandler,
		StorageService:     instance.StoreService,
		DataPool:           instance.DataPool,
		InternalMarshaller: instance.CoreComponentsHolder.InternalMarshalizer(),
	})
	if err != nil {
		return nil, err
	}

	instance.ProcessComponentsHolder, err = CreateProcessComponentsHolder(ArgsProcessComponentsHolder{
		CoreComponents:           instance.CoreComponentsHolder,
		CryptoComponents:         instance.CryptoComponentsHolder,
		NetworkComponents:        instance.NetworkComponentsHolder,
		BootstrapComponents:      instance.BootstrapComponentsHolder,
		StateComponents:          instance.StateComponentsHolder,
		StatusComponents:         instance.StatusComponentsHolder,
		StatusCoreComponents:     instance.StatusCoreComponents,
		FlagsConfig:              args.ContextFlagsConfig,
		ImportDBConfig:           args.ImportDBConfig,
		PrefsConfig:              args.PreferencesConfig,
		Config:                   args.Config,
		EconomicsConfig:          args.EconomicsConfig,
		SystemSCConfig:           args.SystemSCConfig,
		EpochConfig:              args.EpochConfig,
		ConfigurationPathsHolder: args.ConfigurationPathsHolder,
		NodesCoordinator:         instance.NodesCoordinator,
		DataComponents:           instance.DataComponentsHolder,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (node *testOnlyProcessingNode) createBasicComponents(numShards, selfShardID uint32) error {
	var err error

	node.TransactionFeeHandler, err = postprocess.NewFeeAccumulator()
	if err != nil {
		return err
	}
	node.ShardCoordinator, err = sharding.NewMultiShardCoordinator(numShards, selfShardID)
	if err != nil {
		return err
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

func (node *testOnlyProcessingNode) createDataPool(args ArgsTestOnlyProcessingNode) error {
	var err error

	argsDataPool := dataRetrieverFactory.ArgsDataPool{
		Config:           &args.Config,
		EconomicsData:    node.CoreComponentsHolder.EconomicsData(),
		ShardCoordinator: node.ShardCoordinator,
		Marshalizer:      node.CoreComponentsHolder.InternalMarshalizer(),
		PathManager:      node.CoreComponentsHolder.PathHandler(),
	}

	node.DataPool, err = dataRetrieverFactory.NewDataPoolFromConfig(argsDataPool)

	return err
}

func (node *testOnlyProcessingNode) createTransactionLogProcessor() error {
	logsStorer, err := node.StoreService.GetStorer(dataRetriever.TxLogsUnit)
	if err != nil {
		return err
	}
	argsTxLogProcessor := transactionLog.ArgTxLogProcessor{
		Storer:               logsStorer,
		Marshalizer:          node.CoreComponentsHolder.InternalMarshalizer(),
		SaveInStorageEnabled: true,
	}

	node.TxLogsProcessor, err = transactionLog.NewTxLogProcessor(argsTxLogProcessor)

	return err
}

func (node *testOnlyProcessingNode) createNodesCoordinator(pref config.PreferencesConfig, generalConfig config.Config) error {
	nodesShufflerOut, err := bootstrapComp.CreateNodesShuffleOut(
		node.CoreComponentsHolder.GenesisNodesSetup(),
		generalConfig.EpochStartConfig,
		node.CoreComponentsHolder.ChanStopNodeProcess(),
	)
	if err != nil {
		return err
	}

	bootstrapStorer, err := node.StoreService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

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
		node.ShardCoordinator.SelfId(),
		node.BootstrapComponentsHolder.EpochBootstrapParams(),
		node.BootstrapComponentsHolder.EpochBootstrapParams().Epoch(),
		node.CoreComponentsHolder.ChanStopNodeProcess(),
		node.CoreComponentsHolder.NodeTypeProvider(),
		node.CoreComponentsHolder.EnableEpochsHandler(),
		node.DataPool.CurrentEpochValidatorInfo(),
	)
	if err != nil {
		return err
	}

	return nil
}

// CreateNewBlock create and process a new block
func (node *testOnlyProcessingNode) CreateNewBlock(nonce uint64, round uint64) error {
	bp := node.ProcessComponentsHolder.BlockProcessor()
	newHeader, err := node.prepareHeader(nonce, round)
	if err != nil {
		return err
	}

	header, block, err := bp.CreateBlock(newHeader, func() bool {
		return true
	})
	if err != nil {
		return err
	}

	err = bp.ProcessBlock(header, block, func() time.Duration {
		// TODO fix this
		return 1000
	})
	if err != nil {
		return err
	}

	err = bp.CommitBlock(header, block)
	if err != nil {
		return err
	}

	return nil
}

func (node *testOnlyProcessingNode) prepareHeader(nonce uint64, round uint64) (chainData.HeaderHandler, error) {
	bp := node.ProcessComponentsHolder.BlockProcessor()
	newHeader, err := bp.CreateNewHeader(round, nonce)
	if err != nil {
		return nil, err
	}
	err = newHeader.SetShardID(node.ShardCoordinator.SelfId())
	if err != nil {
		return nil, err
	}

	return newHeader, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (node *testOnlyProcessingNode) IsInterfaceNil() bool {
	return node == nil
}
