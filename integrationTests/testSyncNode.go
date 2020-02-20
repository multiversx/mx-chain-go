package integrationTests

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NewTestSyncNode returns a new TestProcessorNode instance with sync capabilities
func NewTestSyncNode(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	initialNodeAddr string,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {
			validator := mock.NewValidatorMock(big.NewInt(0), 0, []byte("add"), []byte("add"))
			return []sharding.Validator{validator}, nil
		},
	}

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)

	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
		NodesCoordinator: nodesCoordinator,
		BootstrapStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		StorageBootstrapper: &mock.StorageBootstrapperMock{},
		HeaderSigVerifier:   &mock.HeaderSigVerifierStub{},
		ChainID:             ChainID,
		EpochStartTrigger:   &mock.EpochStartTriggerStub{},
	}

	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()
	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}

	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initTestNodeWithSync()

	return tpn
}

func (tpn *TestProcessorNode) initTestNodeWithSync() {
	tpn.initChainHandler()
	tpn.initHeaderValidator()
	tpn.initRounder()
	tpn.initStorage()
	tpn.initAccountDBs()
	tpn.GenesisBlocks = CreateSimpleGenesisBlocks(tpn.ShardCoordinator)
	tpn.initEconomicsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initBlockTracker()
	tpn.initInterceptors()
	tpn.initInnerProcessors()
	tpn.initBlockProcessorWithSync()
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.SingleSigner,
	)
	tpn.initBootstrapper()
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()
}

func (tpn *TestProcessorNode) addGenesisBlocksIntoStorage() {
	for shardId, header := range tpn.GenesisBlocks {
		buffHeader, _ := TestMarshalizer.Marshal(header)
		headerHash := TestHasher.Compute(string(buffHeader))

		if shardId == sharding.MetachainShardId {
			metablockStorer := tpn.Storage.GetStorer(dataRetriever.MetaBlockUnit)
			_ = metablockStorer.Put(headerHash, buffHeader)
		} else {
			shardblockStorer := tpn.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
			_ = shardblockStorer.Put(headerHash, buffHeader)
		}
	}
}

func (tpn *TestProcessorNode) initBlockProcessorWithSync() {
	var err error

	argumentsBase := block.ArgBaseProcessor{
		Accounts:                     tpn.AccntState,
		ForkDetector:                 nil,
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		Store:                        tpn.Storage,
		ShardCoordinator:             tpn.ShardCoordinator,
		NodesCoordinator:             tpn.NodesCoordinator,
		FeeHandler:                   tpn.FeeAccumulator,
		Uint64Converter:              TestUint64Converter,
		RequestHandler:               tpn.RequestHandler,
		Core:                         nil,
		BlockChainHook:               &mock.BlockChainHookHandlerMock{},
		ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		HeaderValidator:              tpn.HeaderValidator,
		Rounder:                      &mock.RounderMock{},
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker: tpn.BlockTracker,
		DataPool:     tpn.DataPool,
		BlockChain:   tpn.BlockChain,
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.ForkDetector, _ = sync.NewMetaForkDetector(tpn.Rounder, tpn.BlackListHandler, tpn.BlockTracker, 0)
		argumentsBase.Core = &mock.ServiceContainerMock{}
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:         argumentsBase,
			SCDataGetter:             &mock.ScQueryMock{},
			SCToProtocol:             &mock.SCToProtocolStub{},
			PendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
			EpochStartDataCreator:    &mock.EpochStartDataCreatorStub{},
			EpochEconomics:           &mock.EpochEconomicsStub{},
			EpochRewardsCreator:      &mock.EpochRewardsCreatorStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		tpn.ForkDetector, _ = sync.NewShardForkDetector(tpn.Rounder, tpn.BlackListHandler, tpn.BlockTracker, 0)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.BlockChainHook = tpn.BlockchainHook
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		arguments := block.ArgShardProcessor{
			ArgBaseProcessor:       argumentsBase,
			TxsPoolsCleaner:        &mock.TxPoolsCleanerMock{},
			StateCheckpointModulus: stateCheckpointModulus,
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		fmt.Printf("Error creating blockprocessor: %s\n", err.Error())
	}
}

func (tpn *TestProcessorNode) createShardBootstrapper() (TestBootstrapper, error) {
	accountsStateWrapper, err := state.NewAccountsDbWrapperSync(tpn.AccntState)
	if err != nil {
		return nil, err
	}

	resolver, err := tpn.ResolverFinder.IntraShardResolver(factory.MiniBlocksTopic)
	if err != nil {
		return nil, err
	}

	miniBlocksResolver, ok := resolver.(process.MiniBlocksResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:         tpn.DataPool,
		Store:               tpn.Storage,
		ChainHandler:        tpn.BlockChain,
		Rounder:             tpn.Rounder,
		BlockProcessor:      tpn.BlockProcessor,
		WaitTime:            tpn.Rounder.TimeDuration(),
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		ForkDetector:        tpn.ForkDetector,
		RequestHandler:      tpn.RequestHandler,
		ShardCoordinator:    tpn.ShardCoordinator,
		Accounts:            accountsStateWrapper,
		BlackListHandler:    tpn.BlackListHandler,
		NetworkWatcher:      tpn.Messenger,
		BootStorer:          tpn.BootstrapStorer,
		StorageBootstrapper: tpn.StorageBootstrapper,
		EpochHandler:        tpn.EpochStartTrigger,
		MiniBlocksResolver:  miniBlocksResolver,
		Uint64Converter:     TestUint64Converter,
	}

	argsShardBootstrapper := sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
	}

	bootstrap, err := sync.NewShardBootstrap(argsShardBootstrapper)
	if err != nil {
		return nil, err
	}

	return &sync.TestShardBootstrap{
		ShardBootstrap: bootstrap,
	}, nil
}

func (tpn *TestProcessorNode) createMetaChainBootstrapper() (TestBootstrapper, error) {
	resolver, err := tpn.ResolverFinder.IntraShardResolver(factory.MiniBlocksTopic)
	if err != nil {
		return nil, err
	}

	miniBlocksResolver, ok := resolver.(process.MiniBlocksResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:         tpn.DataPool,
		Store:               tpn.Storage,
		ChainHandler:        tpn.BlockChain,
		Rounder:             tpn.Rounder,
		BlockProcessor:      tpn.BlockProcessor,
		WaitTime:            tpn.Rounder.TimeDuration(),
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		ForkDetector:        tpn.ForkDetector,
		RequestHandler:      tpn.RequestHandler,
		ShardCoordinator:    tpn.ShardCoordinator,
		Accounts:            tpn.AccntState,
		BlackListHandler:    tpn.BlackListHandler,
		NetworkWatcher:      tpn.Messenger,
		BootStorer:          tpn.BootstrapStorer,
		StorageBootstrapper: tpn.StorageBootstrapper,
		EpochHandler:        tpn.EpochStartTrigger,
		MiniBlocksResolver:  miniBlocksResolver,
		Uint64Converter:     TestUint64Converter,
	}

	argsMetaBootstrapper := sync.ArgMetaBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
		EpochBootstrapper:   tpn.EpochStartTrigger,
	}

	bootstrap, err := sync.NewMetaBootstrap(argsMetaBootstrapper)
	if err != nil {
		return nil, err
	}

	return &sync.TestMetaBootstrap{
		MetaBootstrap: bootstrap,
	}, nil
}

func (tpn *TestProcessorNode) initBootstrapper() {
	if tpn.ShardCoordinator.SelfId() < tpn.ShardCoordinator.NumberOfShards() {
		tpn.Bootstrapper, _ = tpn.createShardBootstrapper()
	} else {
		tpn.Bootstrapper, _ = tpn.createMetaChainBootstrapper()
	}
}
