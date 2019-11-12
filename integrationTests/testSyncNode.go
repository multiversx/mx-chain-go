package integrationTests

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/process"

	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/block"
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
	nodesCoordinator := &mock.NodesCoordinatorMock{}

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)

	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
		NodesCoordinator: nodesCoordinator,
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
	tpn.initRounder()
	tpn.initStorage()
	tpn.AccntState, _, _ = CreateAccountsDB(0)
	tpn.initChainHandler()
	tpn.GenesisBlocks = CreateSimpleGenesisBlocks(tpn.ShardCoordinator)
	tpn.SpecialAddressHandler = mock.NewSpecialAddressHandlerMock(
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.NodesCoordinator,
	)
	tpn.initEconomicsData()
	tpn.initInterceptors()
	tpn.initResolvers()
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
	tpn.ScDataGetter, _ = smartContract.NewSCDataGetter(TestAddressConverter, tpn.VMContainer)
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

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      TestHasher,
		Marshalizer: TestMarshalizer,
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

	argumentsBase := block.ArgBaseProcessor{
		Accounts:                     tpn.AccntState,
		ForkDetector:                 nil,
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		Store:                        tpn.Storage,
		ShardCoordinator:             tpn.ShardCoordinator,
		NodesCoordinator:             tpn.NodesCoordinator,
		SpecialAddressHandler:        tpn.SpecialAddressHandler,
		Uint64Converter:              TestUint64Converter,
		StartHeaders:                 tpn.GenesisBlocks,
		RequestHandler:               tpn.RequestHandler,
		Core:                         nil,
		BlockChainHook:               &mock.BlockChainHookHandlerMock{},
		ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
		EndOfEpochTrigger:            &mock.EndOfEpochTriggerStub{},
		HeaderValidator:              headerValidator,
		Rounder:                      &mock.RounderMock{},
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.ForkDetector, _ = sync.NewMetaForkDetector(tpn.Rounder, tpn.BlackListHandler)
		argumentsBase.Core = &mock.ServiceContainerMock{}
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:   argumentsBase,
			DataPool:           tpn.MetaDataPool,
			SCDataGetter:       &mock.ScDataGetterMock{},
			SCToProtocol:       &mock.SCToProtocolStub{},
			PeerChangesHandler: &mock.PeerChangesHandler{},
			PendingMiniBlocks:  &mock.PendingMiniBlocksHandlerStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)

	} else {
		tpn.ForkDetector, _ = sync.NewShardForkDetector(tpn.Rounder, tpn.BlackListHandler)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.BlockChainHook = tpn.BlockChainHookImpl.(process.BlockChainHookHandler)
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
			DataPool:         tpn.ShardDataPool,
			TxsPoolsCleaner:  &mock.TxPoolsCleanerMock{},
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		fmt.Printf("Error creating blockprocessor: %s\n", err.Error())
	}
}

func (tpn *TestProcessorNode) createShardBootstrapper() (TestBootstrapper, error) {
	bootstrap, err := sync.NewShardBootstrap(
		tpn.ShardDataPool,
		tpn.Storage,
		tpn.BlockChain,
		tpn.Rounder,
		tpn.BlockProcessor,
		tpn.Rounder.TimeDuration(),
		TestHasher,
		TestMarshalizer,
		tpn.ForkDetector,
		tpn.ResolverFinder,
		tpn.ShardCoordinator,
		tpn.AccntState,
		1,
		tpn.BlackListHandler,
		tpn.Messenger,
	)
	if err != nil {
		return nil, err
	}

	return &sync.TestShardBootstrap{
		ShardBootstrap: bootstrap,
	}, nil
}

func (tpn *TestProcessorNode) createMetaChainBootstrapper() (TestBootstrapper, error) {
	bootstrap, err := sync.NewMetaBootstrap(
		tpn.MetaDataPool,
		tpn.Storage,
		tpn.BlockChain,
		tpn.Rounder,
		tpn.BlockProcessor,
		tpn.Rounder.TimeDuration(),
		TestHasher,
		TestMarshalizer,
		tpn.ForkDetector,
		tpn.ResolverFinder,
		tpn.ShardCoordinator,
		tpn.AccntState,
		1,
		tpn.BlackListHandler,
		tpn.Messenger,
	)

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
