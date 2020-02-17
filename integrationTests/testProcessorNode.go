package integrationTests

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"math/big"
	"sort"
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
	factory3 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
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
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	scToProtocol2 "github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/track"
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

// TestAddressConverterBLS represents an address converter from BLS public keys
var TestAddressConverterBLS, _ = addressConverters.NewPlainAddressConverter(128, "0x")

// TestMultiSig represents a mock multisig
var TestMultiSig = mock.NewMultiSigner(1)

// TestKeyGenForAccounts represents a mock key generator for balances
var TestKeyGenForAccounts = signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())

// TestUint64Converter represents an uint64 to byte slice converter
var TestUint64Converter = uint64ByteSlice.NewBigEndianConverter()

// MinTxGasPrice defines minimum gas price required by a transaction
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

// ChainID is the chain ID identifier used in integration tests, processing nodes
var ChainID = []byte("integration tests chain ID")

// sizeCheckDelta the maximum allowed bufer overhead (p2p unmarshalling)
const sizeCheckDelta = 100

const stateCheckpointModulus = 100

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
	ShardCoordinator sharding.Coordinator
	NodesCoordinator sharding.NodesCoordinator
	Messenger        p2p.Messenger

	OwnAccount *TestWalletAccount
	NodeKeys   *TestKeyPair

	DataPool      dataRetriever.PoolsHolder
	Storage       dataRetriever.StorageService
	PeerState     state.AccountsAdapter
	AccntState    state.AccountsAdapter
	TrieContainer state.TriesHolder
	BlockChain    data.ChainHandler
	GenesisBlocks map[uint32]data.HeaderHandler

	EconomicsData *economics.TestEconomicsData

	BlackListHandler      process.BlackListHandler
	HeaderValidator       process.HeaderConstructionValidator
	BlockTracker          process.BlockTracker
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
	FeeAccumulator         process.TransactionFeeHandler

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

	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor

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

	pkBytes := make([]byte, 128)
	address := make([]byte, 32)
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkBytes, address)
			return []sharding.Validator{v}, nil
		},
	}

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
		ChainID:           ChainID,
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
		ChainID:           ChainID,
	}

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	if tpn.ShardCoordinator.SelfId() != sharding.MetachainShardId {
		tpn.DataPool = dPool
	} else {
		tpn.initDataPools()
	}
	tpn.initTestNode()

	return tpn
}

func (tpn *TestProcessorNode) initAccountDBs() {
	tpn.TrieContainer = state.NewDataTriesHolder()
	var stateTrie data.Trie
	tpn.AccntState, stateTrie, _ = CreateAccountsDB(factory2.UserAccount)
	tpn.TrieContainer.Put([]byte(factory3.UserAccountTrie), stateTrie)

	var peerTrie data.Trie
	tpn.PeerState, peerTrie, _ = CreateAccountsDB(factory2.ValidatorAccount)
	tpn.TrieContainer.Put([]byte(factory3.PeerAccountTrie), peerTrie)
}

func (tpn *TestProcessorNode) initTestNode() {
	tpn.initHeaderValidator()
	tpn.initRounder()
	tpn.initStorage()
	tpn.initAccountDBs()
	tpn.initChainHandler()
	tpn.initEconomicsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initValidatorStatistics()
	rootHash, _ := tpn.ValidatorStatisticsProcessor.RootHash()
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
		tpn.DataPool,
		tpn.EconomicsData.EconomicsData,
		rootHash,
	)
	tpn.initBlockTracker()
	tpn.initInterceptors()
	tpn.initInnerProcessors()
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(tpn.VMContainer, tpn.EconomicsData.MaxGasLimitPerBlock())
	tpn.initBlockProcessor(stateCheckpointModulus)
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
	tpn.DataPool = CreateTestDataPool(nil)
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
				LeaderPercentage: 0.10,
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:  maxGasLimitPerBlock,
				MinGasPrice:          minGasPrice,
				MinGasLimit:          minGasLimit,
				GasPerDataByte:       "1",
				DataLimitForBaseCalc: "10000",
			},
			ValidatorSettings: config.ValidatorSettings{
				StakeValue:    "500",
				UnBoundPeriod: "5",
			},
			RatingSettings: config.RatingSettings{
				StartRating:                 500000,
				MaxRating:                   1000000,
				MinRating:                   1,
				ProposerDecreaseRatingStep:  3858,
				ProposerIncreaseRatingStep:  1929,
				ValidatorDecreaseRatingStep: 61,
				ValidatorIncreaseRatingStep: 31,
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

		argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime: tpn.Rounder.TimeStamp(),
			Settings: &config.EpochStartConfig{
				MinRoundsBetweenEpochs: 1000,
				RoundsPerEpoch:         10000,
			},
			Epoch:              0,
			EpochStartNotifier: &mock.EpochStartNotifierStub{},
			Storage:            tpn.Storage,
			Marshalizer:        TestMarshalizer,
		}
		epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
		tpn.EpochStartTrigger = &metachain.TestTrigger{}
		tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)

		interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
			tpn.ShardCoordinator,
			tpn.NodesCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			TestHasher,
			TestMultiSig,
			tpn.DataPool,
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
			tpn.BlockTracker,
			tpn.EpochStartTrigger,
		)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:        TestMarshalizer,
			Hasher:             TestHasher,
			HeaderValidator:    tpn.HeaderValidator,
			Uint64Converter:    TestUint64Converter,
			DataPool:           tpn.DataPool,
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
			tpn.DataPool,
			TestAddressConverter,
			maxTxNonceDeltaAllowed,
			tpn.EconomicsData,
			tpn.BlackListHandler,
			tpn.HeaderSigVerifier,
			tpn.ChainID,
			sizeCheckDelta,
			tpn.BlockTracker,
			tpn.EpochStartTrigger,
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
			tpn.DataPool,
			TestUint64Converter,
			dataPacker,
			tpn.TrieContainer,
			100,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewMetaResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
			100,
		)
	} else {
		resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			tpn.DataPool,
			TestUint64Converter,
			dataPacker,
			tpn.TrieContainer,
			100,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewShardResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
			100,
			tpn.ShardCoordinator.SelfId(),
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

	if tpn.ValidatorStatisticsProcessor == nil {
		tpn.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorMock{}
	}

	interimProcFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressConverter,
		tpn.Storage,
		tpn.DataPool,
		tpn.EconomicsData.EconomicsData,
	)

	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = tpn.InterimProcContainer.Get(dataBlock.SmartContractResultBlock)

	tpn.RewardsProcessor, _ = rewardTransaction.NewRewardTxProcessor(
		tpn.AccntState,
		TestAddressConverter,
		tpn.ShardCoordinator,
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

	tpn.FeeAccumulator, _ = postprocess.NewFeeAccumulator()
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
		tpn.FeeAccumulator,
		tpn.EconomicsData,
		txTypeHandler,
		tpn.GasHandler,
	)

	receiptsHandler, _ := tpn.InterimProcContainer.Get(dataBlock.ReceiptBlock)
	badBlocskHandler, _ := tpn.InterimProcContainer.Get(dataBlock.InvalidBlock)
	tpn.TxProcessor, _ = transaction.NewTxProcessor(
		tpn.AccntState,
		TestHasher,
		TestAddressConverter,
		TestMarshalizer,
		tpn.ShardCoordinator,
		tpn.ScProcessor,
		tpn.FeeAccumulator,
		txTypeHandler,
		tpn.EconomicsData,
		receiptsHandler,
		badBlocskHandler,
	)

	tpn.MiniBlocksCompacter, _ = preprocess.NewMiniBlocksCompaction(tpn.EconomicsData, tpn.ShardCoordinator, tpn.GasHandler)

	fact, _ := shard.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.DataPool,
		TestAddressConverter,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		tpn.ScProcessor,
		tpn.ScProcessor.(process.SmartContractResultProcessor),
		tpn.RewardsProcessor,
		tpn.EconomicsData,
		tpn.MiniBlocksCompacter,
		tpn.GasHandler,
		tpn.BlockTracker,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		TestHasher,
		TestMarshalizer,
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.DataPool.MiniBlocks(),
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
		tpn.DataPool,
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

	tpn.FeeAccumulator, _ = postprocess.NewFeeAccumulator()
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
		tpn.FeeAccumulator,
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
		tpn.EconomicsData,
	)

	tpn.MiniBlocksCompacter, _ = preprocess.NewMiniBlocksCompaction(tpn.EconomicsData, tpn.ShardCoordinator, tpn.GasHandler)

	fact, _ := metaProcess.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.DataPool,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		scProcessor,
		tpn.EconomicsData.EconomicsData,
		tpn.MiniBlocksCompacter,
		tpn.GasHandler,
		tpn.BlockTracker,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		TestHasher,
		TestMarshalizer,
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.DataPool.MiniBlocks(),
		tpn.RequestHandler,
		tpn.PreProcessorsContainer,
		tpn.InterimProcContainer,
		tpn.GasHandler,
	)
}

func (tpn *TestProcessorNode) initValidatorStatistics() {
	initialNodes := make([]*sharding.InitialNode, 0)
	nodesMap := tpn.NodesCoordinator.GetAllValidatorsPublicKeys()
	for _, pks := range nodesMap {
		for _, pk := range pks {
			validator, _, _ := tpn.NodesCoordinator.GetValidatorWithPublicKey(pk)
			n := &sharding.InitialNode{
				PubKey:   core.ToHex(validator.PubKey()),
				Address:  core.ToHex(validator.Address()),
				NodeInfo: sharding.NodeInfo{},
			}
			initialNodes = append(initialNodes, n)
		}
	}

	sort.Slice(initialNodes, func(i, j int) bool {
		return bytes.Compare([]byte(initialNodes[i].PubKey), []byte(initialNodes[j].PubKey)) > 0
	})

	rater, _ := rating.NewBlockSigningRater(tpn.EconomicsData.RatingsData())

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes:        initialNodes,
		PeerAdapter:         tpn.PeerState,
		AdrConv:             TestAddressConverterBLS,
		NodesCoordinator:    tpn.NodesCoordinator,
		ShardCoordinator:    tpn.ShardCoordinator,
		DataPool:            tpn.DataPool,
		StorageService:      tpn.Storage,
		Marshalizer:         TestMarshalizer,
		StakeValue:          big.NewInt(500),
		Rater:               rater,
		MaxComputableRounds: 1000,
	}

	tpn.ValidatorStatisticsProcessor, _ = peer.NewValidatorStatisticsProcessor(arguments)
}

func (tpn *TestProcessorNode) addMockVm(blockchainHook vmcommon.BlockchainHook) {
	mockVM, _ := mock.NewOneSCExecutorMockVM(blockchainHook, TestHasher)
	mockVM.GasForOperation = OpGasValueForMockVm

	_ = tpn.VMContainer.Add(factory.InternalTestingVM, mockVM)
}

func (tpn *TestProcessorNode) initBlockProcessor(stateCheckpointModulus uint) {
	var err error

	tpn.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return nil
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
		FeeHandler:                   tpn.FeeAccumulator,
		Uint64Converter:              TestUint64Converter,
		RequestHandler:               tpn.RequestHandler,
		Core:                         nil,
		BlockChainHook:               tpn.BlockchainHook,
		ValidatorStatisticsProcessor: tpn.ValidatorStatisticsProcessor,
		HeaderValidator:              tpn.HeaderValidator,
		Rounder:                      tpn.Rounder,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker: tpn.BlockTracker,
		DataPool:     tpn.DataPool,
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
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
			CurrTxs:     tpn.DataPool.CurrentBlockTxs(),
			ScQuery:     tpn.SCQueryService,
		}
		scToProtocol, _ := scToProtocol2.NewStakingToPeer(argsStakingToPeer)
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:         argumentsBase,
			SCDataGetter:             tpn.SCQueryService,
			SCToProtocol:             scToProtocol,
			PeerChangesHandler:       scToProtocol,
			PendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
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
		node.WithDataPool(tpn.DataPool),
	)
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
	hdrHandlers := func(header data.HeaderHandler, key []byte) {
		atomic.AddInt32(&tpn.CounterHdrRecv, 1)
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.DataPool.Headers().RegisterHandler(hdrHandlers)
	} else {
		txHandler := func(key []byte) {
			atomic.AddInt32(&tpn.CounterTxRecv, 1)
		}
		mbHandlers := func(key []byte) {
			atomic.AddInt32(&tpn.CounterMbRecv, 1)
		}

		tpn.DataPool.UnsignedTransactions().RegisterHandler(txHandler)
		tpn.DataPool.Transactions().RegisterHandler(txHandler)
		tpn.DataPool.RewardTransactions().RegisterHandler(txHandler)
		tpn.DataPool.Headers().RegisterHandler(hdrHandlers)
		tpn.DataPool.MiniBlocks().RegisterHandler(mbHandlers)
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
	maxTime := time.Second * 2

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
	blockBody, err = tpn.BlockProcessor.ApplyBodyToHeader(tpn.BlockChain, blockHeader, blockBody)
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
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.Headers() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	headerObjects, _, err := tpn.DataPool.Headers().GetHeadersByNonceAndShardId(nonce, tpn.ShardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("no headers found for nonce and shard id %d %d %s", nonce, tpn.ShardCoordinator.SelfId(), err.Error()))
	}

	headerObject := headerObjects[len(headerObjects)-1]

	header, ok := headerObject.(*dataBlock.Header)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.Header stored in headers found for nonce and shard id %d %d", nonce, tpn.ShardCoordinator.SelfId()))
	}

	return header, nil
}

// GetBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetBlockBody(header *dataBlock.Header) (dataBlock.Body, error) {
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := dataBlock.Body{}
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		mbObject, ok := tpn.DataPool.MiniBlocks().Get(miniBlockHash)
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
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := dataBlock.Body{}
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		mbObject, ok := tpn.DataPool.MiniBlocks().Get(miniBlockHash)
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
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.Headers() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	headerObjects, _, err := tpn.DataPool.Headers().GetHeadersByNonceAndShardId(nonce, sharding.MetachainShardId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("no headers found for nonce and shard id %d %d %s", nonce, sharding.MetachainShardId, err.Error()))
	}

	headerObject := headerObjects[len(headerObjects)-1]

	header, ok := headerObject.(*dataBlock.MetaBlock)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.MetaBlock stored in headers found for nonce and shard id %d %d", nonce, sharding.MetachainShardId))
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
	mbCacher := tpn.DataPool.MiniBlocks()
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

func (tpn *TestProcessorNode) initBlockTracker() {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:           TestHasher,
		HeaderValidator:  tpn.HeaderValidator,
		Marshalizer:      TestMarshalizer,
		RequestHandler:   tpn.RequestHandler,
		Rounder:          tpn.Rounder,
		ShardCoordinator: tpn.ShardCoordinator,
		Store:            tpn.Storage,
		StartHeaders:     tpn.GenesisBlocks,
	}

	if tpn.ShardCoordinator.SelfId() != sharding.MetachainShardId {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
			PoolsHolder:    tpn.DataPool,
		}

		tpn.BlockTracker, _ = track.NewShardBlockTrack(arguments)
	} else {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
			PoolsHolder:    tpn.DataPool,
		}

		tpn.BlockTracker, _ = track.NewMetaBlockTrack(arguments)
	}
}

func (tpn *TestProcessorNode) initHeaderValidator() {
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      TestHasher,
		Marshalizer: TestMarshalizer,
	}

	tpn.HeaderValidator, _ = block.NewHeaderValidator(argsHeaderValidator)
}
