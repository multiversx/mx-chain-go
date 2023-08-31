package processingOnlyNode

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	hashingFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	marshalFactory "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	Config              config.Config
	EnableEpochsConfig  config.EnableEpochs
	EconomicsConfig     config.EconomicsConfig
	GasScheduleFilename string
	WorkingDir          string
	NumShards           uint32
	ShardID             uint32
}

type testOnlyProcessingNode struct {
	RoundNotifier      process.RoundNotifier
	EpochNotifier      process.EpochNotifier
	WasmerChangeLocker common.Locker
	ArgumentsParser    process.ArgumentsParser
	TxVersionChecker   process.TxVersionCheckerHandler

	Marshaller               marshal.Marshalizer
	Hasher                   coreData.Hasher
	ShardCoordinator         sharding.Coordinator
	TransactionFeeHandler    process.TransactionFeeHandler
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
	EnableEpochsHandler      common.EnableEpochsHandler
	PathHandler              storage.PathManagerHandler

	GasScheduleNotifier         core.GasScheduleNotifier
	BuiltinFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	EconomicsData               process.EconomicsDataHandler
	DataPool                    dataRetriever.PoolsHolder
}

// NewTestOnlyProcessingNode creates a new instance of a node that is able to only process transactions
func NewTestOnlyProcessingNode(args ArgsTestOnlyProcessingNode) (*testOnlyProcessingNode, error) {
	instance := &testOnlyProcessingNode{
		RoundNotifier:      forking.NewGenericRoundNotifier(),
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		WasmerChangeLocker: &sync.RWMutex{},
		ArgumentsParser:    smartContract.NewArgumentParser(),
		TxVersionChecker:   versioning.NewTxVersionChecker(args.Config.GeneralSettings.MinTransactionVersion),
	}

	err := instance.createBasicComponents(args)
	if err != nil {
		return nil, err
	}

	err = instance.createGasScheduleNotifier(args)
	if err != nil {
		return nil, err
	}

	err = instance.createBuiltinFunctionsCostHandler()
	if err != nil {
		return nil, err
	}

	err = instance.createEconomicsHandler(args)
	if err != nil {
		return nil, err
	}

	err = instance.createDataPool(args)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (node *testOnlyProcessingNode) createBasicComponents(args ArgsTestOnlyProcessingNode) error {
	var err error

	node.Marshaller, err = marshalFactory.NewMarshalizer(args.Config.Marshalizer.Type)
	if err != nil {
		return err
	}

	node.Hasher, err = hashingFactory.NewHasher(args.Config.Hasher.Type)
	if err != nil {
		return err
	}

	node.ShardCoordinator, err = sharding.NewMultiShardCoordinator(args.ShardID, args.NumShards)
	if err != nil {
		return err
	}

	node.TransactionFeeHandler, err = postprocess.NewFeeAccumulator()
	if err != nil {
		return err
	}

	node.ValidatorPubKeyConverter, err = factory.NewPubkeyConverter(args.Config.ValidatorPubkeyConverter)
	if err != nil {
		return err
	}

	node.AddressPubKeyConverter, err = factory.NewPubkeyConverter(args.Config.AddressPubkeyConverter)
	if err != nil {
		return err
	}

	node.EnableEpochsHandler, err = enablers.NewEnableEpochsHandler(args.EnableEpochsConfig, node.EpochNotifier)
	if err != nil {
		return err
	}

	node.PathHandler, err = storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: args.WorkingDir,
			ChainID:    args.Config.GeneralSettings.ChainID,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (node *testOnlyProcessingNode) createGasScheduleNotifier(args ArgsTestOnlyProcessingNode) error {
	var err error

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
		EpochNotifier:      node.EpochNotifier,
		WasmVMChangeLocker: node.WasmerChangeLocker,
	}
	node.GasScheduleNotifier, err = forking.NewGasScheduleNotifier(argsGasSchedule)

	return err
}

func (node *testOnlyProcessingNode) createBuiltinFunctionsCostHandler() error {
	var err error

	args := &economics.ArgsBuiltInFunctionCost{
		GasSchedule: node.GasScheduleNotifier,
		ArgsParser:  node.ArgumentsParser,
	}

	node.BuiltinFunctionsCostHandler, err = economics.NewBuiltInFunctionsCost(args)

	return err
}

func (node *testOnlyProcessingNode) createEconomicsHandler(args ArgsTestOnlyProcessingNode) error {
	var err error

	argsEconomicsHandler := economics.ArgsNewEconomicsData{
		TxVersionChecker:            node.TxVersionChecker,
		BuiltInFunctionsCostHandler: node.BuiltinFunctionsCostHandler,
		Economics:                   &args.EconomicsConfig,
		EpochNotifier:               node.EpochNotifier,
		EnableEpochsHandler:         node.EnableEpochsHandler,
	}

	node.EconomicsData, err = economics.NewEconomicsData(argsEconomicsHandler)

	return err
}

func (node *testOnlyProcessingNode) createDataPool(args ArgsTestOnlyProcessingNode) error {
	var err error

	argsDataPool := dataRetrieverFactory.ArgsDataPool{
		Config:           &args.Config,
		EconomicsData:    node.EconomicsData,
		ShardCoordinator: node.ShardCoordinator,
		Marshalizer:      node.Marshaller,
		PathManager:      node.PathHandler,
	}

	node.DataPool, err = dataRetrieverFactory.NewDataPoolFromConfig(argsDataPool)

	return err
}
