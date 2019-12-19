package integrationTests

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/memp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	procFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	scToProtocol2 "github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
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

// TestKeyGenForAccounts represents a mock key generator for balances
var TestKeyGenForAccounts = signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())

// TestUint64Converter represents an uint64 to byte slice converter
var TestUint64Converter = uint64ByteSlice.NewBigEndianConverter()

// MinTxGasPrice defines minimum gas price required by a transaction
//TODO refactor all tests to pass with a non zero value
var MinTxGasPrice = uint64(10)

// MinTxGasLimit defines minimum gas limit required by a transaction
var MinTxGasLimit = uint64(1000)

// MaxGasLimitPerBlock defines maximum gas limit allowed per one block
const MaxGasLimitPerBlock = uint64(300000)

const maxTxNonceDeltaAllowed = 8000
const minConnectedPeers = 0

// OpGasValueForMockVm represents the gas value that it consumed by each operation called on the mock VM
// By operation, we mean each go function that is called on the VM implementation
const OpGasValueForMockVm = uint64(50)

// TimeSpanForBadHeaders is the expiry time for an added block header hash
var TimeSpanForBadHeaders = time.Second * 30

// roundDuration defines the duration of the round
const roundDuration = 5 * time.Second

// IntegrationTestsChainID is the chain ID identifier used in integration tests, processing nodes
var IntegrationTestsChainID = []byte("integration tests chain ID")

// sizeCheckDelta the maximum allowed bufer overhead (p2p unmarshalling)
const sizeCheckDelta = 100

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
	PeerState     state.AccountsAdapter
	AccntState    state.AccountsAdapter
	BlockChain    data.ChainHandler
	GenesisBlocks map[uint32]data.HeaderHandler

	EconomicsData *economics.TestEconomicsData

	BlackListHandler      process.BlackListHandler
	InterceptorsContainer process.InterceptorsContainer
	ResolversContainer    dataRetriever.ResolversContainer
	ResolverFinder        dataRetriever.ResolversFinder
	RequestHandler        process.RequestHandler

	InterimProcContainer   process.IntermediateProcessorContainer
	TxProcessor            process.TransactionProcessor
	TxCoordinator          process.TransactionCoordinator
	ScrForwarder           process.IntermediateTransactionHandler
	BlockchainHook         *hooks.BlockChainHookImpl
	VMContainer            process.VirtualMachinesContainer
	ArgsParser             process.ArgumentsParser
	ScProcessor            process.SmartContractProcessor
	RewardsProcessor       process.RewardTransactionProcessor
	PreProcessorsContainer process.PreProcessorsContainer
	MiniBlocksCompacter    process.MiniBlocksCompacter
	GasHandler             process.GasHandler

	ForkDetector          process.ForkDetector
	BlockProcessor        process.BlockProcessor
	BroadcastMessenger    consensus.BroadcastMessenger
	Bootstrapper          TestBootstrapper
	Rounder               *mock.RounderMock
	BootstrapStorer       *mock.BoostrapStorerMock
	StorageBootstrapper   *mock.StorageBootstrapperMock
	RequestedItemsHandler dataRetriever.RequestedItemsHandler

	EpochStartTrigger TestEpochStartTrigger

	MultiSigner       crypto.MultiSigner
	HeaderSigVerifier process.InterceptedHeaderSigVerifier

	//Node is used to call the functionality already implemented in it
	Node           *node.Node
	SCQueryService external.SCQueryService

	CounterHdrRecv int32
	CounterMbRecv  int32
	CounterTxRecv  int32
	CounterMetaRcv int32

	ChainID []byte
}

// NewTestProcessorNode returns a new TestProcessorNode instance with a libp2p messenger
func NewTestProcessorNode(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	initialNodeAddr string,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	pkAddr := []byte("aaa00000000000000000000000000000")
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {

			address := pkAddr
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkAddr, address)

			return []sharding.Validator{v}, nil
		},
	}

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
		ChainID:           IntegrationTestsChainID,
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initTestNode()

	tpn.StorageBootstrapper = &mock.StorageBootstrapperMock{}
	tpn.BootstrapStorer = &mock.BoostrapStorerMock{}

	return tpn
}

// NewTestProcessorNodeWithMemP2P returns a new TestProcessorNode instance with a memory-based messenger
func NewTestProcessorNodeWithMemP2P(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	network *memp2p.Network,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	pkAddr := []byte("aaa00000000000000000000000000000")
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {

			address := pkAddr
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkAddr, address)

			return []sharding.Validator{v}, nil
		},
	}

	messenger, _ := memp2p.NewMessenger(network)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initTestNode()

	tpn.StorageBootstrapper = &mock.StorageBootstrapperMock{}
	tpn.BootstrapStorer = &mock.BoostrapStorerMock{}

	return tpn
}

// NewTestProcessorNodeWithCustomDataPool returns a new TestProcessorNode instance with the given data pool
func NewTestProcessorNodeWithCustomDataPool(maxShards uint32, nodeShardId uint32, txSignPrivKeyShardId uint32, initialNodeAddr string, dPool dataRetriever.PoolsHolder) *TestProcessorNode {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	_ = messenger.SetThresholdMinConnectedPeers(minConnectedPeers)
	nodesCoordinator := &mock.NodesCoordinatorMock{}
	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
		ChainID:           IntegrationTestsChainID,
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
	tpn.PeerState, _, _ = CreateAccountsDB(factory2.ValidatorAccount)
	tpn.initChainHandler()
	tpn.initEconomicsData()
	tpn.initInterceptors()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initInnerProcessors()
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
	tpn.GenesisBlocks = CreateGenesisBlocks(
		tpn.AccntState,
		TestAddressConverter,
		&sharding.NodesSetup{},
		tpn.ShardCoordinator,
		tpn.Storage,
		tpn.BlockChain,
		TestMarshalizer,
		TestHasher,
		TestUint64Converter,
		tpn.MetaDataPool,
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
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
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
	maxGasLimitPerBlock := strconv.FormatUint(MaxGasLimitPerBlock, 10)
	minGasPrice := strconv.FormatUint(MinTxGasPrice, 10)
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
				MaxGasLimitPerBlock: maxGasLimitPerBlock,
				MinGasPrice:         minGasPrice,
				MinGasLimit:         minGasLimit,
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
	tpn.BlackListHandler = timecache.NewTimeCache(TimeSpanForBadHeaders)

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
			tpn.OwnAccount.BlockSingleSigner,
			tpn.OwnAccount.KeygenTxSign,
			tpn.OwnAccount.KeygenBlockSign,
			maxTxNonceDeltaAllowed,
			tpn.EconomicsData,
			tpn.BlackListHandler,
			tpn.HeaderSigVerifier,
			tpn.ChainID,
			sizeCheckDelta,
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
			tpn.OwnAccount.KeygenBlockSign,
			tpn.OwnAccount.SingleSigner,
			tpn.OwnAccount.BlockSingleSigner,
			TestMultiSig,
			tpn.ShardDataPool,
			TestAddressConverter,
			maxTxNonceDeltaAllowed,
			tpn.EconomicsData,
			tpn.BlackListHandler,
			tpn.HeaderSigVerifier,
			tpn.ChainID,
			sizeCheckDelta,
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
			100,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewMetaResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
			factory.ShardHeadersForMetachainTopic,
			factory.MetachainBlocksTopic,
			factory.TransactionTopic,
			factory.UnsignedTransactionTopic,
			factory.MiniBlocksTopic,
			100,
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
			100,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewShardResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
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

func createAndAddIeleVM(
	vmContainer process.VirtualMachinesContainer,
	blockChainHook vmcommon.BlockchainHook,
) {
	cryptoHook := hooks.NewVMCryptoHook()
	ieleVM := endpoint.NewElrondIeleVM(factory.IELEVirtualMachine, endpoint.ElrondTestnet, blockChainHook, cryptoHook)
	_ = vmContainer.Add(factory.IELEVirtualMachine, ieleVM)
}

func (tpn *TestProcessorNode) initInnerProcessors() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.initMetaInnerProcessors()
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
	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)
	gasSchedule := arwenConfig.MakeGasMap(1)
	vmFactory, _ := shard.NewVMContainerFactory(maxGasLimitPerBlock, gasSchedule, argsHook)

	tpn.VMContainer, _ = vmFactory.Create()
	tpn.BlockchainHook, _ = vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	createAndAddIeleVM(tpn.VMContainer, tpn.BlockchainHook)

	mockVM, _ := mock.NewOneSCExecutorMockVM(tpn.BlockchainHook, TestHasher)
	mockVM.GasForOperation = OpGasValueForMockVm
	_ = tpn.VMContainer.Add(procFactory.InternalTestingVM, mockVM)

	tpn.ArgsParser, _ = vmcommon.NewAtArgumentParser()
	txTypeHandler, _ := coordinator.NewTxTypeHandler(TestAddressConverter, tpn.ShardCoordinator, tpn.AccntState)
	tpn.GasHandler, _ = preprocess.NewGasComputation(tpn.EconomicsData)
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
		tpn.EconomicsData,
		txTypeHandler,
		tpn.GasHandler,
	)

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

	tpn.MiniBlocksCompacter, _ = preprocess.NewMiniBlocksCompaction(tpn.EconomicsData, tpn.ShardCoordinator, tpn.GasHandler)

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
		tpn.GasHandler,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.ShardDataPool.MiniBlocks(),
		tpn.RequestHandler,
		tpn.PreProcessorsContainer,
		tpn.InterimProcContainer,
		tpn.GasHandler,
	)
}

func (tpn *TestProcessorNode) initMetaInnerProcessors() {
	interimProcFactory, _ := metaProcess.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressConverter,
		tpn.Storage,
		tpn.MetaDataPool,
	)

	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = tpn.InterimProcContainer.Get(dataBlock.SmartContractResultBlock)

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         tpn.AccntState,
		AddrConv:         TestAddressConverter,
		StorageService:   tpn.Storage,
		BlockChain:       tpn.BlockChain,
		ShardCoordinator: tpn.ShardCoordinator,
		Marshalizer:      TestMarshalizer,
		Uint64Converter:  TestUint64Converter,
	}

	vmFactory, _ := metaProcess.NewVMContainerFactory(argsHook, tpn.EconomicsData.EconomicsData)

	tpn.VMContainer, _ = vmFactory.Create()
	tpn.BlockchainHook, _ = vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)

	tpn.addMockVm(tpn.BlockchainHook)

	txTypeHandler, _ := coordinator.NewTxTypeHandler(TestAddressConverter, tpn.ShardCoordinator, tpn.AccntState)
	tpn.ArgsParser, _ = vmcommon.NewAtArgumentParser()
	tpn.GasHandler, _ = preprocess.NewGasComputation(tpn.EconomicsData)
	scProcessor, _ := smartContract.NewSmartContractProcessor(
		tpn.VMContainer,
		tpn.ArgsParser,
		TestHasher,
		TestMarshalizer,
		tpn.AccntState,
		vmFactory.BlockChainHookImpl(),
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.ScrForwarder,
		&metaProcess.TransactionFeeHandler{},
		tpn.EconomicsData,
		txTypeHandler,
		tpn.GasHandler,
	)
	tpn.ScProcessor = scProcessor
	tpn.TxProcessor, _ = transaction.NewMetaTxProcessor(
		tpn.AccntState,
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.ScProcessor,
		txTypeHandler,
	)

	tpn.MiniBlocksCompacter, _ = preprocess.NewMiniBlocksCompaction(tpn.EconomicsData, tpn.ShardCoordinator, tpn.GasHandler)

	fact, _ := metaProcess.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.MetaDataPool,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		scProcessor,
		tpn.EconomicsData.EconomicsData,
		tpn.MiniBlocksCompacter,
		tpn.GasHandler,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.MetaDataPool.MiniBlocks(),
		tpn.RequestHandler,
		tpn.PreProcessorsContainer,
		tpn.InterimProcContainer,
		tpn.GasHandler,
	)
}

func (tpn *TestProcessorNode) addMockVm(blockchainHook vmcommon.BlockchainHook) {
	mockVM, _ := mock.NewOneSCExecutorMockVM(blockchainHook, TestHasher)
	mockVM.GasForOperation = OpGasValueForMockVm

	_ = tpn.VMContainer.Add(factory.InternalTestingVM, mockVM)
}

func (tpn *TestProcessorNode) initBlockProcessor() {
	var err error

	tpn.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      TestHasher,
		Marshalizer: TestMarshalizer,
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

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
		BlockChainHook:               tpn.BlockchainHook,
		ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
		HeaderValidator:              headerValidator,
		Rounder:                      &mock.RounderMock{},
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {

		argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime: argumentsBase.Rounder.TimeStamp(),
			Settings: &config.EpochStartConfig{
				MinRoundsBetweenEpochs: 1000,
				RoundsPerEpoch:         10000,
			},
			Epoch:              0,
			EpochStartNotifier: &mock.EpochStartNotifierStub{},
		}
		epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
		tpn.EpochStartTrigger = &metachain.TestTrigger{}
		tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)

		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.TxCoordinator = tpn.TxCoordinator

		blsKeyedAddressConverter, _ := addressConverters.NewPlainAddressConverter(
			128,
			"0x",
		)
		argsStakingToPeer := scToProtocol2.ArgStakingToPeer{
			AdrConv:     blsKeyedAddressConverter,
			Hasher:      TestHasher,
			Marshalizer: TestMarshalizer,
			PeerState:   tpn.PeerState,
			BaseState:   tpn.AccntState,
			ArgParser:   tpn.ArgsParser,
			CurrTxs:     tpn.MetaDataPool.CurrentBlockTxs(),
			ScQuery:     tpn.SCQueryService,
		}
		scToProtocol, _ := scToProtocol2.NewStakingToPeer(argsStakingToPeer)
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:   argumentsBase,
			DataPool:           tpn.MetaDataPool,
			SCDataGetter:       tpn.SCQueryService,
			SCToProtocol:       scToProtocol,
			PeerChangesHandler: scToProtocol,
			PendingMiniBlocks:  &mock.PendingMiniBlocksHandlerStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)

	} else {
		argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:        TestMarshalizer,
			Hasher:             TestHasher,
			HeaderValidator:    headerValidator,
			Uint64Converter:    TestUint64Converter,
			DataPool:           tpn.ShardDataPool,
			Storage:            tpn.Storage,
			RequestHandler:     tpn.RequestHandler,
			Epoch:              0,
			Validity:           1,
			Finality:           1,
			EpochStartNotifier: &mock.EpochStartNotifierStub{},
		}
		epochStartTrigger, _ := shardchain.NewEpochStartTrigger(argsShardEpochStart)
		tpn.EpochStartTrigger = &shardchain.TestTrigger{}
		tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)

		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.BlockChainHook = tpn.BlockchainHook
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
		node.WithMarshalizer(TestMarshalizer, 100),
		node.WithHasher(TestHasher),
		node.WithHasher(TestHasher),
		node.WithAddressConverter(TestAddressConverter),
		node.WithAccountsAdapter(tpn.AccntState),
		node.WithKeyGen(tpn.OwnAccount.KeygenTxSign),
		node.WithKeyGenForAccounts(TestKeyGenForAccounts),
		node.WithTxFeeHandler(tpn.EconomicsData),
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
		node.WithBlackListHandler(tpn.BlackListHandler),
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
		tx.Value.String(),
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
	maxTime := time.Second * 200000

	haveTime := func() bool {
		elapsedTime := time.Since(startTime)
		remainingTime := maxTime - elapsedTime
		return remainingTime > 0
	}

	blockHeader := tpn.BlockProcessor.CreateNewHeader()

	blockHeader.SetShardID(tpn.ShardCoordinator.SelfId())
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
	blockHeader.SetLeaderSignature([]byte("leader sign"))
	blockHeader.SetChainID(tpn.ChainID)

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

// GetMetaBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetMetaBlockBody(header *dataBlock.MetaBlock) (dataBlock.Body, error) {
	invalidCachers := tpn.MetaDataPool == nil || tpn.MetaDataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := dataBlock.Body{}
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		mbObject, ok := tpn.MetaDataPool.MiniBlocks().Get(miniBlockHash)
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

	body, err := tpn.GetMetaBlockBody(header)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.ProcessBlock(
		tpn.BlockChain,
		header,
		body,
		func() time.Duration {
			return time.Second * 2000
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

func (tpn *TestProcessorNode) initRequestedItemsHandler() {
	tpn.RequestedItemsHandler = timecache.NewTimeCache(roundDuration)
}
