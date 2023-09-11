package processingOnlyNode

import (
	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	Config              config.Config
	EnableEpochsConfig  config.EnableEpochs
	EconomicsConfig     config.EconomicsConfig
	RoundsConfig        config.RoundConfig
	ChanStopNodeProcess chan endProcess.ArgEndProcess
	GasScheduleFilename string
	WorkingDir          string
	NodesSetupPath      string
	NumShards           uint32
	ShardID             uint32
}

type testOnlyProcessingNode struct {
	CoreComponentsHolder  factory.CoreComponentsHolder
	StatusCoreComponents  factory.StatusCoreComponentsHolder
	StateComponentsHolder factory.StateComponentsHolder

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
		Cfg:                 args.Config,
		EnableEpochsConfig:  args.EnableEpochsConfig,
		RoundsConfig:        args.RoundsConfig,
		EconomicsConfig:     args.EconomicsConfig,
		ChanStopNodeProcess: args.ChanStopNodeProcess,
		NumShards:           args.NumShards,
		WorkingDir:          args.WorkingDir,
		GasScheduleFilename: args.GasScheduleFilename,
		NodesSetupPath:      args.NodesSetupPath,
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
		Cfg:            args.Config,
		CoreComponents: instance.CoreComponentsHolder,
		StatusCore:     instance.StatusCoreComponents,
		StoreService:   instance.StoreService,
		ChainHandler:   instance.ChainHandler,
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
