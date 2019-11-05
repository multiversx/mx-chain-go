package integrationTests

import (
	"context"
	"encoding/hex"
	"fmt"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/pkg/errors"
)

// TestHasher represents a Sha256 hasher
var TestHasher = sha256.Sha256{}

// TestMarshalizer represents a JSON marshalizer
var TestMarshalizer = &marshal.JsonMarshalizer{}

// TestAddressConverter represents a plain address converter
var TestAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")

// TestMultiSig represents a mock multisig
var TestMultiSig = mock.NewMultiSigner(1)

// TestUint64Converter represents an uint64 to byte slice converter
var TestUint64Converter = uint64ByteSlice.NewBigEndianConverter()

// MinTxGasPrice minimum gas price required by a transaction
//TODO refactor all tests to pass with a non zero value
var MinTxGasPrice = uint64(0)

// MinTxGasLimit minimum gas limit required by a transaction
var MinTxGasLimit = uint64(4)

const maxTxNonceDeltaAllowed = 8000

// TestKeyPair holds a pair of private/public Keys
type TestKeyPair struct {
	Sk crypto.PrivateKey
	Pk crypto.PublicKey
}

//CryptoParams holds crypto parametres
type CryptoParams struct {
	KeyGen       crypto.KeyGenerator
	Keys         map[uint32][]*TestKeyPair
	SingleSigner crypto.SingleSigner
}

// TestProcessorNode represents a container type of class used in integration tests
// with all its fields exported
type TestProcessorNode struct {
	ShardCoordinator      sharding.Coordinator
	NodesCoordinator      sharding.NodesCoordinator
	SpecialAddressHandler process.SpecialAddressHandler
	Messenger             p2p.Messenger

	OwnAccount *TestWalletAccount
	NodeKeys   *TestKeyPair

	ShardDataPool dataRetriever.PoolsHolder
	MetaDataPool  dataRetriever.MetaPoolsHolder
	Storage       dataRetriever.StorageService
	AccntState    state.AccountsAdapter
	BlockChain    data.ChainHandler
	GenesisBlocks map[uint32]data.HeaderHandler

	EconomicsData *economics.TestEconomicsData

	InterceptorsContainer process.InterceptorsContainer
	ResolversContainer    dataRetriever.ResolversContainer
	ResolverFinder        dataRetriever.ResolversFinder
	RequestHandler        process.RequestHandler

	InterimProcContainer   process.IntermediateProcessorContainer
	TxProcessor            process.TransactionProcessor
	TxCoordinator          process.TransactionCoordinator
	ScrForwarder           process.IntermediateTransactionHandler
	VMContainer            process.VirtualMachinesContainer
	ArgsParser             process.ArgumentsParser
	ScProcessor            process.SmartContractProcessor
	RewardsProcessor       process.RewardTransactionProcessor
	PreProcessorsContainer process.PreProcessorsContainer
	MiniBlocksCompacter    process.MiniBlocksCompacter
	BlockChainHookImpl     process.BlockChainHookHandler

	ForkDetector       process.ForkDetector
	BlockProcessor     process.BlockProcessor
	BroadcastMessenger consensus.BroadcastMessenger
	Bootstrapper       TestBootstrapper
	Rounder            *mock.RounderMock

	MultiSigner crypto.MultiSigner

	//Node is used to call the functionality already implemented in it
	Node         *node.Node
	ScDataGetter external.ScDataGetter

	CounterHdrRecv int32
	CounterMbRecv  int32
	CounterTxRecv  int32
	CounterMetaRcv int32
}

// NewTestProcessorNode returns a new TestProcessorNode instance
func NewTestProcessorNode(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	initialNodeAddr string,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)
	nodesCoordinator := &mock.NodesCoordinatorMock{}
	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
		NodesCoordinator: nodesCoordinator,
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initTestNode()

	return tpn
}

// NewTestProcessorNodeWithCustomDataPool returns a new TestProcessorNode instance with the given data pool
func NewTestProcessorNodeWithCustomDataPool(maxShards uint32, nodeShardId uint32, txSignPrivKeyShardId uint32, initialNodeAddr string, dPool dataRetriever.PoolsHolder) *TestProcessorNode {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	nodesCoordinator := &mock.NodesCoordinatorMock{}
	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
		NodesCoordinator: nodesCoordinator,
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	if tpn.ShardCoordinator.SelfId() != sharding.MetachainShardId {
		tpn.ShardDataPool = dPool
	} else {
		tpn.initDataPools()
	}
	tpn.initTestNode()

	return tpn
}

func (tpn *TestProcessorNode) initTestNode() {
	tpn.SpecialAddressHandler = mock.NewSpecialAddressHandlerMock(
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.NodesCoordinator,
	)
	tpn.initStorage()
	tpn.AccntState, _, _ = CreateAccountsDB(factory2.UserAccount)
	tpn.initChainHandler()
	tpn.initEconomicsData()
	tpn.initInterceptors()
	tpn.initResolvers()
	tpn.initInnerProcessors()
	tpn.GenesisBlocks = CreateGenesisBlocks(
		tpn.AccntState,
		TestAddressConverter,
		&sharding.NodesSetup{},
		tpn.ShardCoordinator,
		CreateMetaStore(tpn.ShardCoordinator),
		tpn.BlockChain,
		TestMarshalizer,
		TestHasher,
		TestUint64Converter,
		CreateTestMetaDataPool(),
		tpn.EconomicsData.EconomicsData,
	)
	tpn.initBlockProcessor()
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.SingleSigner,
	)
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.ScDataGetter, _ = smartContract.NewSCDataGetter(TestAddressConverter, tpn.VMContainer)
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()
}

func (tpn *TestProcessorNode) initDataPools() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.MetaDataPool = CreateTestMetaDataPool()
	} else {
		tpn.ShardDataPool = CreateTestShardDataPool(nil)
	}
}

func (tpn *TestProcessorNode) initStorage() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.Storage = CreateMetaStore(tpn.ShardCoordinator)
	} else {
		tpn.Storage = CreateShardStore(tpn.ShardCoordinator.NumberOfShards())
	}
}

func (tpn *TestProcessorNode) initChainHandler() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.BlockChain = CreateMetaChain()
	} else {
		tpn.BlockChain = CreateShardChain()
	}
}

func (tpn *TestProcessorNode) initEconomicsData() {
	mingGasPrice := strconv.FormatUint(MinTxGasPrice, 10)
	minGasLimit := strconv.FormatUint(MinTxGasLimit, 10)

	economicsData, _ := economics.NewEconomicsData(
		&config.ConfigEconomics{
			EconomicsAddresses: config.EconomicsAddresses{
				CommunityAddress: "addr1",
				BurnAddress:      "addr2",
			},
			RewardsSettings: config.RewardsSettings{
				RewardsValue:        "1000",
				CommunityPercentage: 0.10,
				LeaderPercentage:    0.50,
				BurnPercentage:      0.40,
			},
			FeeSettings: config.FeeSettings{
				MinGasPrice: mingGasPrice,
				MinGasLimit: minGasLimit,
			},
			ValidatorSettings: config.ValidatorSettings{
				StakeValue:    "500",
				UnBoundPeriod: "5",
			},
		},
	)

	tpn.EconomicsData = &economics.TestEconomicsData{
		EconomicsData: economicsData,
	}
}

func (tpn *TestProcessorNode) initInterceptors() {
	var err error
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
			tpn.ShardCoordinator,
			tpn.NodesCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			TestHasher,
			TestMultiSig,
			tpn.MetaDataPool,
			tpn.AccntState,
			TestAddressConverter,
			tpn.OwnAccount.SingleSigner,
			tpn.OwnAccount.KeygenTxSign,
			maxTxNonceDeltaAllowed,
			tpn.EconomicsData,
		)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
			tpn.AccntState,
			tpn.ShardCoordinator,
			tpn.NodesCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			TestHasher,
			tpn.OwnAccount.KeygenTxSign,
			tpn.OwnAccount.SingleSigner,
			TestMultiSig,
			tpn.ShardDataPool,
			TestAddressConverter,
			maxTxNonceDeltaAllowed,
			tpn.EconomicsData,
		)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (tpn *TestProcessorNode) initResolvers() {
	dataPacker, _ := partitioning.NewSimpleDataPacker(TestMarshalizer)

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			tpn.MetaDataPool,
			TestUint64Converter,
			dataPacker,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewMetaResolverRequestHandler(
			tpn.ResolverFinder,
			factory.ShardHeadersForMetachainTopic,
			factory.MetachainBlocksTopic,
			factory.TransactionTopic,
			factory.UnsignedTransactionTopic,
			factory.MiniBlocksTopic,
		)
	} else {
		resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			tpn.ShardDataPool,
			TestUint64Converter,
			dataPacker,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewShardResolverRequestHandler(
			tpn.ResolverFinder,
			factory.TransactionTopic,
			factory.UnsignedTransactionTopic,
			factory.RewardsTransactionTopic,
			factory.MiniBlocksTopic,
			factory.HeadersTopic,
			factory.MetachainBlocksTopic,
			100,
		)
	}
}

func (tpn *TestProcessorNode) initInnerProcessors() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		return
	}

	interimProcFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressConverter,
		tpn.SpecialAddressHandler,
		tpn.Storage,
		tpn.ShardDataPool,
		tpn.EconomicsData.EconomicsData,
	)

	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = tpn.InterimProcContainer.Get(dataBlock.SmartContractResultBlock)
	rewardsInter, _ := tpn.InterimProcContainer.Get(dataBlock.RewardsBlock)
	rewardsHandler, _ := rewardsInter.(process.TransactionFeeHandler)
	internalTxProducer, _ := rewardsInter.(process.InternalTransactionProducer)

	tpn.RewardsProcessor, _ = rewardTransaction.NewRewardTxProcessor(
		tpn.AccntState,
		TestAddressConverter,
		tpn.ShardCoordinator,
		rewardsInter,
	)

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         tpn.AccntState,
		AddrConv:         TestAddressConverter,
		StorageService:   tpn.Storage,
		BlockChain:       tpn.BlockChain,
		ShardCoordinator: tpn.ShardCoordinator,
		Marshalizer:      TestMarshalizer,
		Uint64Converter:  TestUint64Converter,
	}

	var vmFactory process.VirtualMachinesContainerFactory
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		vmFactory, _ = metaProcess.NewVMContainerFactory(argsHook, tpn.EconomicsData.EconomicsData)
	} else {
		vmFactory, _ = shard.NewVMContainerFactory(argsHook)
	}

	tpn.VMContainer, _ = vmFactory.Create()
	tpn.BlockChainHookImpl = vmFactory.BlockChainHookImpl()

	tpn.ArgsParser, _ = smartContract.NewAtArgumentParser()
	tpn.ScProcessor, _ = smartContract.NewSmartContractProcessor(
		tpn.VMContainer,
		tpn.ArgsParser,
		TestHasher,
		TestMarshalizer,
		tpn.AccntState,
		vmFactory.BlockChainHookImpl(),
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.ScrForwarder,
		rewardsHandler,
	)

	txTypeHandler, _ := coordinator.NewTxTypeHandler(TestAddressConverter, tpn.ShardCoordinator, tpn.AccntState)

	tpn.TxProcessor, _ = transaction.NewTxProcessor(
		tpn.AccntState,
		TestHasher,
		TestAddressConverter,
		TestMarshalizer,
		tpn.ShardCoordinator,
		tpn.ScProcessor,
		rewardsHandler,
		txTypeHandler,
		tpn.EconomicsData,
	)

	tpn.MiniBlocksCompacter, _ = preprocess.NewMiniBlocksCompaction(tpn.EconomicsData, tpn.ShardCoordinator)

	fact, _ := shard.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.ShardDataPool,
		TestAddressConverter,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		tpn.ScProcessor,
		tpn.ScProcessor.(process.SmartContractResultProcessor),
		tpn.RewardsProcessor,
		internalTxProducer,
		tpn.EconomicsData,
		tpn.MiniBlocksCompacter,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.ShardDataPool.MiniBlocks(),
		tpn.RequestHandler,
		tpn.PreProcessorsContainer,
		tpn.InterimProcContainer,
	)
}

func (tpn *TestProcessorNode) initBlockProcessor() {
	var err error

	tpn.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}

	argumentsBase := block.ArgBaseProcessor{
		Accounts:                     tpn.AccntState,
		ForkDetector:                 tpn.ForkDetector,
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
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		argumentsBase.Core = &mock.ServiceContainerMock{}
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:   argumentsBase,
			DataPool:           tpn.MetaDataPool,
			SCDataGetter:       &mock.ScDataGetterMock{},
			SCToProtocol:       &mock.SCToProtocolStub{},
			PeerChangesHandler: &mock.PeerChangesHandler{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		argumentsBase.BlockChainHook = tpn.BlockChainHookImpl
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

func (tpn *TestProcessorNode) setGenesisBlock() {
	genesisBlock := tpn.GenesisBlocks[tpn.ShardCoordinator.SelfId()]
	_ = tpn.BlockChain.SetGenesisHeader(genesisBlock)
	hash, _ := core.CalculateHash(TestMarshalizer, TestHasher, genesisBlock)
	tpn.BlockChain.SetGenesisHeaderHash(hash)
}

func (tpn *TestProcessorNode) initNode() {
	var err error

	tpn.Node, err = node.NewNode(
		node.WithMessenger(tpn.Messenger),
		node.WithMarshalizer(TestMarshalizer),
		node.WithHasher(TestHasher),
		node.WithHasher(TestHasher),
		node.WithAddressConverter(TestAddressConverter),
		node.WithAccountsAdapter(tpn.AccntState),
		node.WithKeyGen(tpn.OwnAccount.KeygenTxSign),
		node.WithShardCoordinator(tpn.ShardCoordinator),
		node.WithNodesCoordinator(tpn.NodesCoordinator),
		node.WithBlockChain(tpn.BlockChain),
		node.WithUint64ByteSliceConverter(TestUint64Converter),
		node.WithMultiSigner(tpn.MultiSigner),
		node.WithSingleSigner(tpn.OwnAccount.SingleSigner),
		node.WithTxSignPrivKey(tpn.OwnAccount.SkTxSign),
		node.WithTxSignPubKey(tpn.OwnAccount.PkTxSign),
		node.WithPrivKey(tpn.NodeKeys.Sk),
		node.WithPubKey(tpn.NodeKeys.Pk),
		node.WithInterceptorsContainer(tpn.InterceptorsContainer),
		node.WithResolversFinder(tpn.ResolverFinder),
		node.WithBlockProcessor(tpn.BlockProcessor),
		node.WithTxSingleSigner(tpn.OwnAccount.SingleSigner),
		node.WithDataStore(tpn.Storage),
		node.WithSyncer(&mock.SyncTimerMock{}),
	)
	if err != nil {
		fmt.Printf("Error creating node: %s\n", err.Error())
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		err = tpn.Node.ApplyOptions(
			node.WithMetaDataPool(tpn.MetaDataPool),
		)
	} else {
		err = tpn.Node.ApplyOptions(
			node.WithDataPool(tpn.ShardDataPool),
		)
	}

	if err != nil {
		fmt.Printf("Error creating node: %s\n", err.Error())
	}
}

// SendTransaction can send a transaction (it does the dispatching)
func (tpn *TestProcessorNode) SendTransaction(tx *dataTransaction.Transaction) (string, error) {
	txHash, err := tpn.Node.SendTransaction(
		tx.Nonce,
		hex.EncodeToString(tx.SndAddr),
		hex.EncodeToString(tx.RcvAddr),
		tx.Value,
		tx.GasPrice,
		tx.GasLimit,
		tx.Data,
		tx.Signature,
	)
	return txHash, err
}

func (tpn *TestProcessorNode) addHandlersForCounters() {
	metaHandlers := func(key []byte) {
		atomic.AddInt32(&tpn.CounterMetaRcv, 1)
	}
	hdrHandlers := func(key []byte) {
		atomic.AddInt32(&tpn.CounterHdrRecv, 1)
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.MetaDataPool.ShardHeaders().RegisterHandler(hdrHandlers)
		tpn.MetaDataPool.MetaBlocks().RegisterHandler(metaHandlers)
	} else {
		txHandler := func(key []byte) {
			atomic.AddInt32(&tpn.CounterTxRecv, 1)
		}
		mbHandlers := func(key []byte) {
			atomic.AddInt32(&tpn.CounterMbRecv, 1)
		}

		tpn.ShardDataPool.UnsignedTransactions().RegisterHandler(txHandler)
		tpn.ShardDataPool.Transactions().RegisterHandler(txHandler)
		tpn.ShardDataPool.RewardTransactions().RegisterHandler(txHandler)
		tpn.ShardDataPool.Headers().RegisterHandler(hdrHandlers)
		tpn.ShardDataPool.MetaBlocks().RegisterHandler(metaHandlers)
		tpn.ShardDataPool.MiniBlocks().RegisterHandler(mbHandlers)
	}
}

// StartSync calls Bootstrapper.StartSync. Errors if bootstrapper is not set
func (tpn *TestProcessorNode) StartSync() error {
	if tpn.Bootstrapper == nil {
		return errors.New("no bootstrapper available")
	}

	tpn.Bootstrapper.StartSync()

	return nil
}

// LoadTxSignSkBytes alters the already generated sk/pk pair
func (tpn *TestProcessorNode) LoadTxSignSkBytes(skBytes []byte) {
	tpn.OwnAccount.LoadTxSignSkBytes(skBytes)
}

// ProposeBlock proposes a new block
func (tpn *TestProcessorNode) ProposeBlock(round uint64, nonce uint64) (data.BodyHandler, data.HeaderHandler, [][]byte) {
	startTime := time.Now()
	maxTime := time.Second

	haveTime := func() bool {
		elapsedTime := time.Since(startTime)
		remainingTime := maxTime - elapsedTime
		return remainingTime > 0
	}

	var blockHeader data.HeaderHandler
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		blockHeader = &dataBlock.MetaBlock{}
	} else {
		blockHeader = &dataBlock.Header{}
	}

	blockHeader.SetRound(round)
	blockHeader.SetNonce(nonce)
	blockHeader.SetPubKeysBitmap([]byte{1})
	currHdr := tpn.BlockChain.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = tpn.BlockChain.GetGenesisHeader()
	}

	buff, _ := TestMarshalizer.Marshal(currHdr)
	blockHeader.SetPrevHash(TestHasher.Compute(string(buff)))
	blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	sig, _ := TestMultiSig.AggregateSigs(nil)
	blockHeader.SetSignature(sig)
	blockHeader.SetRandSeed(sig)

	blockBody, err := tpn.BlockProcessor.CreateBlockBody(blockHeader, haveTime)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, nil
	}
	err = tpn.BlockProcessor.ApplyBodyToHeader(blockHeader, blockBody)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, nil
	}

	shardBlockBody, ok := blockBody.(dataBlock.Body)
	txHashes := make([][]byte, 0)
	if !ok {
		return blockBody, blockHeader, txHashes
	}

	for _, mb := range shardBlockBody {
		for _, hash := range mb.TxHashes {
			copiedHash := make([]byte, len(hash))
			copy(copiedHash, hash)
			txHashes = append(txHashes, copiedHash)
		}
	}

	return blockBody, blockHeader, txHashes
}

// BroadcastBlock broadcasts the block and body to the connected peers
func (tpn *TestProcessorNode) BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) {
	_ = tpn.BroadcastMessenger.BroadcastBlock(body, header)
	_ = tpn.BroadcastMessenger.BroadcastShardHeader(header)
	miniBlocks, transactions, _ := tpn.BlockProcessor.MarshalizedDataToBroadcast(header, body)
	_ = tpn.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = tpn.BroadcastMessenger.BroadcastTransactions(transactions)
}

// CommitBlock commits the block and body
func (tpn *TestProcessorNode) CommitBlock(body data.BodyHandler, header data.HeaderHandler) {
	_ = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, body)
}

// GetShardHeader returns the first *dataBlock.Header stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetShardHeader(nonce uint64) (*dataBlock.Header, error) {
	invalidCachers := tpn.ShardDataPool == nil || tpn.ShardDataPool.Headers() == nil || tpn.ShardDataPool.HeadersNonces() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	syncMapHashNonce, ok := tpn.ShardDataPool.HeadersNonces().Get(nonce)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce link in HeadersNonces for nonce %d", nonce))
	}

	headerHash, ok := syncMapHashNonce.Load(tpn.ShardCoordinator.SelfId())
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce hash in HeadersNonces for nonce %d", nonce))
	}

	headerObject, ok := tpn.ShardDataPool.Headers().Get(headerHash)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no header found for hash %s", hex.EncodeToString(headerHash)))
	}

	header, ok := headerObject.(*dataBlock.Header)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.Header stored in headers found for hash %s", hex.EncodeToString(headerHash)))
	}

	return header, nil
}

// GetBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetBlockBody(header *dataBlock.Header) (dataBlock.Body, error) {
	invalidCachers := tpn.ShardDataPool == nil || tpn.ShardDataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := dataBlock.Body{}
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		mbObject, ok := tpn.ShardDataPool.MiniBlocks().Get(miniBlockHash)
		if !ok {
			return nil, errors.New(fmt.Sprintf("no miniblock found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		mb, ok := mbObject.(*dataBlock.MiniBlock)
		if !ok {
			return nil, errors.New(fmt.Sprintf("not a *dataBlock.MiniBlock stored in miniblocks found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		body = append(body, mb)
	}

	return body, nil
}

// GetMetaHeader returns the first *dataBlock.MetaBlock stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetMetaHeader(nonce uint64) (*dataBlock.MetaBlock, error) {
	invalidCachers := tpn.MetaDataPool == nil || tpn.MetaDataPool.MetaBlocks() == nil || tpn.MetaDataPool.HeadersNonces() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	syncMapHashNonce, ok := tpn.MetaDataPool.HeadersNonces().Get(nonce)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce link in HeadersNonces for nonce %d", nonce))
	}

	headerHash, ok := syncMapHashNonce.Load(tpn.ShardCoordinator.SelfId())
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce hash in HeadersNonces for nonce %d", nonce))
	}

	headerObject, ok := tpn.MetaDataPool.MetaBlocks().Get(headerHash)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no header found for hash %s", hex.EncodeToString(headerHash)))
	}

	header, ok := headerObject.(*dataBlock.MetaBlock)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.MetaBlock stored in headers found for hash %s", hex.EncodeToString(headerHash)))
	}

	return header, nil
}

// SyncNode tries to process and commit a block already stored in data pool with provided nonce
func (tpn *TestProcessorNode) SyncNode(nonce uint64) error {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		return tpn.syncMetaNode(nonce)
	} else {
		return tpn.syncShardNode(nonce)
	}
}

func (tpn *TestProcessorNode) syncShardNode(nonce uint64) error {
	header, err := tpn.GetShardHeader(nonce)
	if err != nil {
		return err
	}

	body, err := tpn.GetBlockBody(header)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.ProcessBlock(
		tpn.BlockChain,
		header,
		body,
		func() time.Duration {
			return time.Second * 2
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, body)
	if err != nil {
		return err
	}

	return nil
}

func (tpn *TestProcessorNode) syncMetaNode(nonce uint64) error {
	header, err := tpn.GetMetaHeader(nonce)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.ProcessBlock(
		tpn.BlockChain,
		header,
		dataBlock.Body{},
		func() time.Duration {
			return time.Second * 2
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, dataBlock.Body{})
	if err != nil {
		return err
	}

	return nil
}

// SetAccountNonce sets the account nonce with journal
func (tpn *TestProcessorNode) SetAccountNonce(nonce uint64) error {
	nodeAccount, _ := tpn.AccntState.GetAccountWithJournal(tpn.OwnAccount.Address)
	err := nodeAccount.(*state.Account).SetNonceWithJournal(nonce)
	if err != nil {
		return err
	}

	_, err = tpn.AccntState.Commit()
	if err != nil {
		return err
	}

	return nil
}

// MiniBlocksPresent checks if the all the miniblocks are present in the pool
func (tpn *TestProcessorNode) MiniBlocksPresent(hashes [][]byte) bool {
	mbCacher := tpn.ShardDataPool.MiniBlocks()
	for i := 0; i < len(hashes); i++ {
		ok := mbCacher.Has(hashes[i])
		if !ok {
			return false
		}
	}

	return true
}

func (tpn *TestProcessorNode) initRounder() {
	tpn.Rounder = &mock.RounderMock{}
}
