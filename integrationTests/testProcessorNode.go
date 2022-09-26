package integrationTests

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/core/versioning"
	"github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	mclsig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	nodeFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/enablers"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	hdrFactory "github.com/ElrondNetwork/elrond-go/factory/block"
	heartbeatComp "github.com/ElrondNetwork/elrond-go/factory/heartbeat"
	"github.com/ElrondNetwork/elrond-go/factory/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	p2pRating "github.com/ElrondNetwork/elrond-go/p2p/rating"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	procFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat/validator"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	processMock "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	processSync "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/ElrondNetwork/elrond-go/process/txsSender"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/blockInfoProviders"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/cache"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/bootstrapMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	dblookupextMock "github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/genesisMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/mainFactoryMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/outport"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/trie/keyBuilder"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmProcess "github.com/ElrondNetwork/elrond-go/vm/process"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

var zero = big.NewInt(0)

var hardforkPubKey = "153dae6cb3963260f309959bf285537b77ae16d82e9933147be7827f7394de8dc97d9d9af41e970bc72aecb44b77e819621081658c37f7000d21e2d0e8963df83233407bde9f46369ba4fcd03b57f40b80b06c191a428cfb5c447ec510e79307"

// TestHasher represents a sha256 hasher
var TestHasher = sha256.NewSha256()

// TestTxSignHasher represents a sha3 legacy keccak 256 hasher
var TestTxSignHasher = keccak.NewKeccak()

// TestMarshalizer represents the main marshalizer
var TestMarshalizer = &marshal.GogoProtoMarshalizer{}

// TestVmMarshalizer represents the marshalizer used in vm communication
var TestVmMarshalizer = &marshal.JsonMarshalizer{}

// TestTxSignMarshalizer represents the marshalizer used in vm communication
var TestTxSignMarshalizer = &marshal.JsonMarshalizer{}

// TestAddressPubkeyConverter represents an address public key converter
var TestAddressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)

// TestValidatorPubkeyConverter represents an address public key converter
var TestValidatorPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(96)

// TestMultiSig represents a mock multisig
var TestMultiSig = cryptoMocks.NewMultiSigner()

// TestKeyGenForAccounts represents a mock key generator for balances
var TestKeyGenForAccounts = signing.NewKeyGenerator(ed25519.NewEd25519())

// TestUint64Converter represents an uint64 to byte slice converter
var TestUint64Converter = uint64ByteSlice.NewBigEndianConverter()

// TestBuiltinFunctions is an additional map of builtin functions to be added
// to the scProcessor
var TestBuiltinFunctions = make(map[string]vmcommon.BuiltinFunction)

// TestBlockSizeThrottler represents a block size throttler used in adaptive block size computation
var TestBlockSizeThrottler = &mock.BlockSizeThrottlerStub{}

// TestBlockSizeComputation represents a block size computation handler
var TestBlockSizeComputationHandler, _ = preprocess.NewBlockSizeComputation(TestMarshalizer, TestBlockSizeThrottler, uint32(core.MegabyteSize*90/100))

// TestBalanceComputationHandler represents a balance computation handler
var TestBalanceComputationHandler, _ = preprocess.NewBalanceComputation()

// TestAppStatusHandler represents an AppStatusHandler
var TestAppStatusHandler = &statusHandlerMock.AppStatusHandlerStub{}

// MinTxGasPrice defines minimum gas price required by a transaction
var MinTxGasPrice = uint64(100)

// MinTxGasLimit defines minimum gas limit required by a transaction
var MinTxGasLimit = uint64(1000)

// MaxGasLimitPerBlock defines maximum gas limit allowed per one block
const MaxGasLimitPerBlock = uint64(3000000)

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

// MinTransactionVersion is the minimum transaction version used in integration tests, processing nodes
var MinTransactionVersion = uint32(1)

// SoftwareVersion is the software version identifier used in integration tests, processing nodes
var SoftwareVersion = []byte("intT")

var testProtocolSustainabilityAddress = "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"

// DelegationManagerConfigChangeAddress represents the address that can change the config parameters of the
// delegation manager system smartcontract
var DelegationManagerConfigChangeAddress = "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80"

// sizeCheckDelta the maximum allowed bufer overhead (p2p unmarshalling)
const sizeCheckDelta = 100

const stateCheckpointModulus = uint(100)

// UnreachableEpoch defines an unreachable epoch for integration tests
const UnreachableEpoch = uint32(1000000)

// TestSingleSigner defines a Ed25519Signer
var TestSingleSigner = &ed25519SingleSig.Ed25519Signer{}

// TestSingleBlsSigner defines a BlsSingleSigner
var TestSingleBlsSigner = &mclsig.BlsSingleSigner{}

// TestKeyPair holds a pair of private/public Keys
type TestKeyPair struct {
	Sk crypto.PrivateKey
	Pk crypto.PublicKey
}

// CryptoParams holds crypto parametres
type CryptoParams struct {
	KeyGen       crypto.KeyGenerator
	Keys         map[uint32][]*TestKeyPair
	SingleSigner crypto.SingleSigner
	TxKeyGen     crypto.KeyGenerator
	TxKeys       map[uint32][]*TestKeyPair
}

// Connectable defines the operations for a struct to become connectable by other struct
// In other words, all instances that implement this interface are able to connect with each other
type Connectable interface {
	ConnectTo(connectable Connectable) error
	GetConnectableAddress() string
	IsInterfaceNil() bool
}

// ArgTestProcessorNode represents the DTO used to create a new TestProcessorNode
type ArgTestProcessorNode struct {
	MaxShards               uint32
	NodeShardId             uint32
	TxSignPrivKeyShardId    uint32
	WithBLSSigVerifier      bool
	WithSync                bool
	GasScheduleMap          GasScheduleMap
	EpochsConfig            *config.EnableEpochs
	VMConfig                *config.VirtualMachineConfig
	EconomicsConfig         *config.EconomicsConfig
	DataPool                dataRetriever.PoolsHolder
	TrieStore               storage.Storer
	HardforkPk              crypto.PublicKey
	GenesisFile             string
	StateCheckpointModulus  *IntWrapper
	NodeKeys                *TestKeyPair
	NodesSetup              sharding.GenesisNodesSetupHandler
	NodesCoordinator        nodesCoordinator.NodesCoordinator
	MultiSigner             crypto.MultiSigner
	RatingsData             *rating.RatingsData
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
	OwnAccount              *TestWalletAccount
	EpochStartSubscriber    notifier.EpochStartNotifier
}

// TestProcessorNode represents a container type of class used in integration tests
// with all its fields exported
type TestProcessorNode struct {
	ShardCoordinator sharding.Coordinator
	NodesCoordinator nodesCoordinator.NodesCoordinator
	PeerShardMapper  process.PeerShardMapper
	NodesSetup       sharding.GenesisNodesSetupHandler
	Messenger        p2p.Messenger

	OwnAccount *TestWalletAccount
	NodeKeys   *TestKeyPair

	ExportFolder        string
	DataPool            dataRetriever.PoolsHolder
	Storage             dataRetriever.StorageService
	PeerState           state.AccountsAdapter
	AccntState          state.AccountsAdapter
	TrieStorageManagers map[string]common.StorageManager
	TrieContainer       common.TriesHolder
	BlockChain          data.ChainHandler
	GenesisBlocks       map[uint32]data.HeaderHandler

	EconomicsData *economics.TestEconomicsData
	RatingsData   *rating.RatingsData

	BlockBlackListHandler process.TimeCacher
	HeaderValidator       process.HeaderConstructionValidator
	BlockTracker          process.BlockTracker
	InterceptorsContainer process.InterceptorsContainer
	ResolversContainer    dataRetriever.ResolversContainer
	ResolverFinder        dataRetriever.ResolversFinder
	RequestHandler        process.RequestHandler
	ArwenChangeLocker     common.Locker

	InterimProcContainer   process.IntermediateProcessorContainer
	TxProcessor            process.TransactionProcessor
	TxCoordinator          process.TransactionCoordinator
	ScrForwarder           process.IntermediateTransactionHandler
	BlockchainHook         *hooks.BlockChainHookImpl
	VMContainer            process.VirtualMachinesContainer
	ArgsParser             process.ArgumentsParser
	ScProcessor            *smartContract.TestScProcessor
	RewardsProcessor       process.RewardTransactionProcessor
	PreProcessorsContainer process.PreProcessorsContainer
	GasHandler             process.GasHandler
	FeeAccumulator         process.TransactionFeeHandler
	SmartContractParser    genesis.InitialSmartContractParser
	SystemSCFactory        vm.SystemSCContainerFactory

	ForkDetector             process.ForkDetector
	BlockProcessor           process.BlockProcessor
	BroadcastMessenger       consensus.BroadcastMessenger
	MiniblocksProvider       process.MiniBlockProvider
	Bootstrapper             TestBootstrapper
	RoundHandler             *mock.RoundHandlerMock
	BootstrapStorer          *mock.BoostrapStorerMock
	StorageBootstrapper      *mock.StorageBootstrapperMock
	RequestedItemsHandler    dataRetriever.RequestedItemsHandler
	WhiteListHandler         process.WhiteListHandler
	WhiteListerVerifiedTxs   process.WhiteListHandler
	NetworkShardingCollector consensus.NetworkShardingCollector

	EpochStartTrigger  TestEpochStartTrigger
	EpochStartNotifier notifier.EpochStartNotifier

	MultiSigner             crypto.MultiSigner
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier

	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
	Rater                        sharding.PeerAccountListAndRatingHandler

	EpochStartSystemSCProcessor process.EpochStartSystemSCProcessor

	// Node is used to call the functionality already implemented in it
	Node           *node.Node
	SCQueryService external.SCQueryService

	CounterHdrRecv       int32
	CounterMbRecv        int32
	CounterTxRecv        int32
	CounterMetaRcv       int32
	ReceivedTransactions sync.Map

	InitialNodes []*sharding.InitialNode

	ChainID               []byte
	MinTransactionVersion uint32

	ExportHandler            update.ExportHandler
	WaitTime                 time.Duration
	HistoryRepository        dblookupext.HistoryRepository
	EpochNotifier            process.EpochNotifier
	EnableEpochs             config.EnableEpochs
	EnableEpochsHandler      common.EnableEpochsHandler
	UseValidVmBlsSigVerifier bool

	TransactionLogProcessor process.TransactionLogProcessor
	PeersRatingHandler      p2p.PeersRatingHandler
	HardforkTrigger         node.HardforkTrigger
}

// CreatePkBytes creates 'numShards' public key-like byte slices
func CreatePkBytes(numShards uint32) map[uint32][]byte {
	pk := []byte("afafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafaf")
	pksbytes := make(map[uint32][]byte, numShards+1)
	for i := uint32(0); i < numShards; i++ {
		pksbytes[i] = make([]byte, len(pk))
		copy(pksbytes[i], pk)
		pksbytes[i][0] = byte(i)
	}

	pksbytes[core.MetachainShardId] = make([]byte, 128)
	pksbytes[core.MetachainShardId] = pk
	pksbytes[core.MetachainShardId][0] = byte(numShards)

	return pksbytes
}

func newBaseTestProcessorNode(args ArgTestProcessorNode) *TestProcessorNode {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(args.MaxShards, args.NodeShardId)

	pksBytes := CreatePkBytes(args.MaxShards)
	address := []byte("afafafafafafafafafafafafafafafaf")
	numNodes := uint32(len(pksBytes))

	nodesSetup := args.NodesSetup
	if check.IfNil(nodesSetup) {
		nodesSetup = getDefaultNodesSetup(args.MaxShards, numNodes, address, pksBytes)
	}

	nodesCoordinatorInstance := args.NodesCoordinator
	if check.IfNil(nodesCoordinatorInstance) {
		nodesCoordinatorInstance = getDefaultNodesCoordinator(args.MaxShards, pksBytes)
	}

	peersRatingHandler, _ := p2pRating.NewPeersRatingHandler(
		p2pRating.ArgPeersRatingHandler{
			TopRatedCache: testscommon.NewCacherMock(),
			BadRatedCache: testscommon.NewCacherMock(),
		})

	messenger := CreateMessengerWithNoDiscoveryAndPeersRatingHandler(peersRatingHandler)

	genericEpochNotifier := forking.NewGenericEpochNotifier()
	epochsConfig := args.EpochsConfig
	if epochsConfig == nil {
		epochsConfig = GetDefaultEnableEpochsConfig()
	}
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(*epochsConfig, genericEpochNotifier)

	logsProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{Marshalizer: TestMarshalizer})
	tpn := &TestProcessorNode{
		ShardCoordinator:         shardCoordinator,
		Messenger:                messenger,
		NodesCoordinator:         nodesCoordinatorInstance,
		ChainID:                  ChainID,
		MinTransactionVersion:    MinTransactionVersion,
		NodesSetup:               nodesSetup,
		HistoryRepository:        &dblookupextMock.HistoryRepositoryStub{},
		EpochNotifier:            genericEpochNotifier,
		EnableEpochsHandler:      enableEpochsHandler,
		ArwenChangeLocker:        &sync.RWMutex{},
		TransactionLogProcessor:  logsProcessor,
		Bootstrapper:             mock.NewTestBootstrapperMock(),
		PeersRatingHandler:       peersRatingHandler,
		PeerShardMapper:          mock.NewNetworkShardingCollectorMock(),
		EnableEpochs:             *epochsConfig,
		UseValidVmBlsSigVerifier: args.WithBLSSigVerifier,
		StorageBootstrapper:      &mock.StorageBootstrapperMock{},
		BootstrapStorer:          &mock.BoostrapStorerMock{},
		RatingsData:              args.RatingsData,
		EpochStartNotifier:       args.EpochStartSubscriber,
	}

	tpn.NodeKeys = args.NodeKeys
	if tpn.NodeKeys == nil {
		kg := &mock.KeyGenMock{}
		tpn.NodeKeys = &TestKeyPair{}
		tpn.NodeKeys.Sk, tpn.NodeKeys.Pk = kg.GeneratePair()
	}

	tpn.MultiSigner = TestMultiSig
	if !check.IfNil(args.MultiSigner) {
		tpn.MultiSigner = args.MultiSigner
	}

	tpn.OwnAccount = args.OwnAccount
	if tpn.OwnAccount == nil {
		tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, args.TxSignPrivKeyShardId)
	}

	tpn.HeaderSigVerifier = args.HeaderSigVerifier
	if check.IfNil(tpn.HeaderSigVerifier) {
		tpn.HeaderSigVerifier = &mock.HeaderSigVerifierStub{}
	}

	tpn.HeaderIntegrityVerifier = args.HeaderIntegrityVerifier
	if check.IfNil(tpn.HeaderIntegrityVerifier) {
		tpn.HeaderIntegrityVerifier = CreateHeaderIntegrityVerifier()
	}

	tpn.initDataPools()

	if !check.IfNil(args.DataPool) {
		tpn.DataPool = args.DataPool
		_ = messenger.SetThresholdMinConnectedPeers(minConnectedPeers)
	}

	return tpn
}

// NewTestProcessorNode returns a new TestProcessorNode instance with a libp2p messenger and the provided arguments
func NewTestProcessorNode(args ArgTestProcessorNode) *TestProcessorNode {
	tpn := newBaseTestProcessorNode(args)
	tpn.initTestNodeWithArgs(args)

	return tpn
}

// ConnectTo will try to initiate a connection to the provided parameter
func (tpn *TestProcessorNode) ConnectTo(connectable Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return tpn.Messenger.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (tpn *TestProcessorNode) GetConnectableAddress() string {
	if tpn == nil {
		return "nil"
	}

	return GetConnectableAddress(tpn.Messenger)
}

// Close -
func (tpn *TestProcessorNode) Close() {
	_ = tpn.Messenger.Close()
	_ = tpn.VMContainer.Close()
}

func (tpn *TestProcessorNode) initAccountDBsWithPruningStorer() {
	if check.IfNil(tpn.EpochStartNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}
	trieStorageManager := CreateTrieStorageManagerWithPruningStorer(tpn.ShardCoordinator, tpn.EpochStartNotifier)
	tpn.TrieContainer = state.NewDataTriesHolder()
	var stateTrie common.Trie
	tpn.AccntState, stateTrie = CreateAccountsDB(UserAccount, trieStorageManager)
	tpn.TrieContainer.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	var peerTrie common.Trie
	tpn.PeerState, peerTrie = CreateAccountsDB(ValidatorAccount, trieStorageManager)
	tpn.TrieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)

	tpn.TrieStorageManagers = make(map[string]common.StorageManager)
	tpn.TrieStorageManagers[trieFactory.UserAccountTrie] = trieStorageManager
	tpn.TrieStorageManagers[trieFactory.PeerAccountTrie] = trieStorageManager
}

func (tpn *TestProcessorNode) initAccountDBs(store storage.Storer) {
	trieStorageManager, _ := CreateTrieStorageManager(store)
	tpn.TrieContainer = state.NewDataTriesHolder()
	var stateTrie common.Trie
	tpn.AccntState, stateTrie = CreateAccountsDB(UserAccount, trieStorageManager)
	tpn.TrieContainer.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	var peerTrie common.Trie
	tpn.PeerState, peerTrie = CreateAccountsDB(ValidatorAccount, trieStorageManager)
	tpn.TrieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)

	tpn.TrieStorageManagers = make(map[string]common.StorageManager)
	tpn.TrieStorageManagers[trieFactory.UserAccountTrie] = trieStorageManager
	tpn.TrieStorageManagers[trieFactory.PeerAccountTrie] = trieStorageManager
}

func (tpn *TestProcessorNode) initValidatorStatistics() {
	rater, _ := rating.NewBlockSigningRater(tpn.RatingsData)

	if check.IfNil(tpn.NodesSetup) {
		tpn.NodesSetup = &mock.NodesSetupStub{
			MinNumberOfNodesCalled: func() uint32 {
				return tpn.ShardCoordinator.NumberOfShards() * 2
			},
		}
	}

	arguments := peer.ArgValidatorStatisticsProcessor{
		PeerAdapter:                          tpn.PeerState,
		PubkeyConv:                           TestValidatorPubkeyConverter,
		NodesCoordinator:                     tpn.NodesCoordinator,
		ShardCoordinator:                     tpn.ShardCoordinator,
		DataPool:                             tpn.DataPool,
		StorageService:                       tpn.Storage,
		Marshalizer:                          TestMarshalizer,
		Rater:                                rater,
		MaxComputableRounds:                  1000,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		RewardsHandler:                       tpn.EconomicsData,
		NodesSetup:                           tpn.NodesSetup,
		GenesisNonce:                         tpn.BlockChain.GetGenesisHeader().GetNonce(),
		EnableEpochsHandler:                  tpn.EnableEpochsHandler,
	}

	tpn.ValidatorStatisticsProcessor, _ = peer.NewValidatorStatisticsProcessor(arguments)
}

func (tpn *TestProcessorNode) initGenesisBlocks(args ArgTestProcessorNode) {
	if args.GenesisFile != "" {
		tpn.SmartContractParser, _ = parsing.NewSmartContractsParser(
			args.GenesisFile,
			TestAddressPubkeyConverter,
			&mock.KeyGenMock{},
		)
		tpn.GenesisBlocks = CreateFullGenesisBlocks(
			tpn.AccntState,
			tpn.PeerState,
			tpn.TrieStorageManagers,
			tpn.NodesSetup,
			tpn.ShardCoordinator,
			tpn.Storage,
			tpn.BlockChain,
			tpn.DataPool,
			tpn.EconomicsData,
			&genesisMocks.AccountsParserStub{},
			tpn.SmartContractParser,
			tpn.EnableEpochs,
		)
		return
	}

	if args.WithSync {
		tpn.GenesisBlocks = CreateSimpleGenesisBlocks(tpn.ShardCoordinator)
		return
	}

	tpn.GenesisBlocks = CreateGenesisBlocks(
		tpn.AccntState,
		tpn.PeerState,
		tpn.TrieStorageManagers,
		TestAddressPubkeyConverter,
		tpn.NodesSetup,
		tpn.ShardCoordinator,
		tpn.Storage,
		tpn.BlockChain,
		TestMarshalizer,
		TestHasher,
		TestUint64Converter,
		tpn.DataPool,
		tpn.EconomicsData,
		tpn.EnableEpochs,
	)
}

func (tpn *TestProcessorNode) initTestNodeWithArgs(args ArgTestProcessorNode) {
	tpn.initChainHandler()
	tpn.initHeaderValidator()
	tpn.initRoundHandler()
	tpn.NetworkShardingCollector = mock.NewNetworkShardingCollectorMock()
	if check.IfNil(tpn.EpochNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}
	tpn.initStorage()
	if check.IfNil(args.TrieStore) {
		tpn.initAccountDBsWithPruningStorer()
	} else {
		tpn.initAccountDBs(args.TrieStore)
	}

	economicsConfig := args.EconomicsConfig
	if economicsConfig == nil {
		economicsConfig = createDefaultEconomicsConfig()
	}

	tpn.initEconomicsData(economicsConfig)
	tpn.initRatingsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initValidatorStatistics()
	tpn.initGenesisBlocks(args)
	tpn.initBlockTracker()

	strPk := ""
	if !check.IfNil(args.HardforkPk) {
		buff, err := args.HardforkPk.ToByteArray()
		log.LogIfError(err)

		strPk = hex.EncodeToString(buff)
	}
	tpn.initInterceptors(strPk)

	gasMap := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	if args.GasScheduleMap != nil {
		gasMap = args.GasScheduleMap
	}
	vmConfig := getDefaultVMConfig()
	if args.VMConfig != nil {
		vmConfig = args.VMConfig
	}
	tpn.initInnerProcessors(gasMap, vmConfig)

	if check.IfNil(args.TrieStore) {
		argsNewScQueryService := smartContract.ArgsNewSCQueryService{
			VmContainer:              tpn.VMContainer,
			EconomicsFee:             tpn.EconomicsData,
			BlockChainHook:           tpn.BlockchainHook,
			BlockChain:               tpn.BlockChain,
			ArwenChangeLocker:        tpn.ArwenChangeLocker,
			Bootstrapper:             tpn.Bootstrapper,
			AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
		}
		tpn.SCQueryService, _ = smartContract.NewSCQueryService(argsNewScQueryService)
	} else {
		tpn.createFullSCQueryService(gasMap, vmConfig)
	}

	if args.WithSync {
		tpn.initBlockProcessorWithSync()
	} else {
		scm := stateCheckpointModulus
		if args.StateCheckpointModulus != nil {
			scm = args.StateCheckpointModulus.Value
		}
		tpn.initBlockProcessor(scm)
	}

	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		TestHasher,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.PeerSigHandler,
		tpn.DataPool.Headers(),
		tpn.InterceptorsContainer,
		&testscommon.AlarmSchedulerStub{},
	)

	if args.WithSync {
		tpn.initBootstrapper()
	}
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()

	if args.GenesisFile != "" {
		tpn.createHeartbeatWithHardforkTrigger()
	}
}

func (tpn *TestProcessorNode) createFullSCQueryService(gasMap map[string]map[string]uint64, vmConfig *config.VirtualMachineConfig) {
	var vmFactory process.VirtualMachinesContainerFactory

	gasSchedule := mock.NewGasScheduleNotifierMock(gasMap)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasSchedule,
		MapDNSAddresses:           make(map[string]struct{}),
		Marshalizer:               TestMarshalizer,
		Accounts:                  tpn.AccntState,
		ShardCoordinator:          tpn.ShardCoordinator,
		EpochNotifier:             tpn.EpochNotifier,
		EnableEpochsHandler:       tpn.EnableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncFactory, _ := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)

	smartContractsCache := testscommon.NewCacherMock()

	argsHook := hooks.ArgBlockChainHook{
		Accounts:              tpn.AccntState,
		PubkeyConv:            TestAddressPubkeyConverter,
		StorageService:        tpn.Storage,
		BlockChain:            tpn.BlockChain,
		ShardCoordinator:      tpn.ShardCoordinator,
		Marshalizer:           TestMarshalizer,
		Uint64Converter:       TestUint64Converter,
		BuiltInFunctions:      builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:     builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler: builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:              tpn.DataPool,
		CompiledSCPool:        smartContractsCache,
		EpochNotifier:         tpn.EpochNotifier,
		EnableEpochsHandler:   tpn.EnableEpochsHandler,
		NilCompiledSCStore:    true,
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.EnableEpochs = config.EnableEpochs{}
		sigVerifier, _ := disabled.NewMessageSignVerifier(&mock.KeyGenMock{})

		blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
		argsNewVmFactory := metaProcess.ArgsNewVMContainerFactory{
			BlockChainHook:      blockChainHookImpl,
			PubkeyConv:          argsHook.PubkeyConv,
			Economics:           tpn.EconomicsData,
			MessageSignVerifier: sigVerifier,
			GasSchedule:         gasSchedule,
			NodesConfigProvider: tpn.NodesSetup,
			Hasher:              TestHasher,
			Marshalizer:         TestMarshalizer,
			SystemSCConfig: &config.SystemSmartContractsConfig{
				ESDTSystemSCConfig: config.ESDTSystemSCConfig{
					BaseIssuingCost: "1000",
					OwnerAddress:    "aaaaaa",
				},
				GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
					V1: config.GovernanceSystemSCConfigV1{
						ProposalCost:     "500",
						NumNodes:         100,
						MinQuorum:        50,
						MinPassThreshold: 50,
						MinVetoThreshold: 50,
					},
					Active: config.GovernanceSystemSCConfigActive{
						ProposalCost:     "500",
						MinQuorum:        "50",
						MinPassThreshold: "50",
						MinVetoThreshold: "50",
					},
					FirstWhitelistedAddress: DelegationManagerConfigChangeAddress,
				},
				StakingSystemSCConfig: config.StakingSystemSCConfig{
					GenesisNodePrice:                     "1000",
					UnJailValue:                          "10",
					MinStepValue:                         "10",
					MinStakeValue:                        "1",
					UnBondPeriod:                         1,
					UnBondPeriodInEpochs:                 1,
					NumRoundsWithoutBleed:                1,
					MaximumPercentageToBleed:             1,
					BleedPercentagePerRound:              1,
					MaxNumberOfNodesForStake:             100,
					ActivateBLSPubKeyMessageVerification: false,
					MinUnstakeTokensValue:                "1",
				},
				DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
					MinCreationDeposit:  "100",
					MinStakeAmount:      "100",
					ConfigChangeAddress: DelegationManagerConfigChangeAddress,
				},
				DelegationSystemSCConfig: config.DelegationSystemSCConfig{
					MinServiceFee: 0,
					MaxServiceFee: 100000,
				},
			},
			ValidatorAccountsDB: tpn.PeerState,
			ChanceComputer:      tpn.NodesCoordinator,
			ShardCoordinator:    tpn.ShardCoordinator,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
		}
		tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
			EpochField: tpn.EnableEpochs.DelegationSmartContractEnableEpoch,
		})
		vmFactory, _ = metaProcess.NewVMContainerFactory(argsNewVmFactory)
	} else {
		esdtTransferParser, _ := parsers.NewESDTTransferParser(TestMarshalizer)
		blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
		argsNewVMFactory := shard.ArgVMContainerFactory{
			Config:              *vmConfig,
			BlockChainHook:      blockChainHookImpl,
			BuiltInFunctions:    argsHook.BuiltInFunctions,
			BlockGasLimit:       tpn.EconomicsData.MaxGasLimitPerBlock(tpn.ShardCoordinator.SelfId()),
			GasSchedule:         gasSchedule,
			EpochNotifier:       tpn.EpochNotifier,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
			ArwenChangeLocker:   tpn.ArwenChangeLocker,
			ESDTTransferParser:  esdtTransferParser,
		}
		vmFactory, _ = shard.NewVMContainerFactory(argsNewVMFactory)
	}

	vmContainer, _ := vmFactory.Create()

	_ = builtInFuncFactory.SetPayableHandler(vmFactory.BlockChainHookImpl())
	argsNewScQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             tpn.EconomicsData,
		BlockChainHook:           vmFactory.BlockChainHookImpl(),
		BlockChain:               tpn.BlockChain,
		ArwenChangeLocker:        tpn.ArwenChangeLocker,
		Bootstrapper:             tpn.Bootstrapper,
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
	}
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(argsNewScQueryService)
}

// InitializeProcessors will reinitialize processors
func (tpn *TestProcessorNode) InitializeProcessors(gasMap map[string]map[string]uint64) {
	tpn.initValidatorStatistics()
	tpn.initBlockTracker()
	tpn.initInnerProcessors(gasMap, getDefaultVMConfig())
	argsNewScQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              tpn.VMContainer,
		EconomicsFee:             tpn.EconomicsData,
		BlockChainHook:           tpn.BlockchainHook,
		BlockChain:               tpn.BlockChain,
		ArwenChangeLocker:        tpn.ArwenChangeLocker,
		Bootstrapper:             tpn.Bootstrapper,
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
	}
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(argsNewScQueryService)
	tpn.initBlockProcessor(stateCheckpointModulus)
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		TestHasher,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.PeerSigHandler,
		tpn.DataPool.Headers(),
		tpn.InterceptorsContainer,
		&testscommon.AlarmSchedulerStub{},
	)
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()
}

func (tpn *TestProcessorNode) initDataPools() {
	tpn.DataPool = dataRetrieverMock.CreatePoolsHolder(1, tpn.ShardCoordinator.SelfId())
	cacherCfg := storageunit.CacheConfig{Capacity: 10000, Type: storageunit.LRUCache, Shards: 1}
	suCache, _ := storageunit.NewCache(cacherCfg)
	tpn.WhiteListHandler, _ = interceptors.NewWhiteListDataVerifier(suCache)

	cacherVerifiedCfg := storageunit.CacheConfig{Capacity: 5000, Type: storageunit.LRUCache, Shards: 1}
	cacheVerified, _ := storageunit.NewCache(cacherVerifiedCfg)
	tpn.WhiteListerVerifiedTxs, _ = interceptors.NewWhiteListDataVerifier(cacheVerified)
}

func (tpn *TestProcessorNode) initStorage() {
	tpn.Storage = CreateStore(tpn.ShardCoordinator.NumberOfShards())
}

func (tpn *TestProcessorNode) initChainHandler() {
	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.BlockChain = CreateMetaChain()
	} else {
		tpn.BlockChain = CreateShardChain()
	}
}

func (tpn *TestProcessorNode) initEconomicsData(economicsConfig *config.EconomicsConfig) {
	tpn.EnableEpochs.PenalizedTooMuchGasEnableEpoch = 0
	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics:                   economicsConfig,
		EpochNotifier:               tpn.EpochNotifier,
		EnableEpochsHandler:         tpn.EnableEpochsHandler,
		BuiltInFunctionsCostHandler: &mock.BuiltInCostHandlerStub{},
	}
	economicsData, _ := economics.NewEconomicsData(argsNewEconomicsData)
	tpn.EconomicsData = economics.NewTestEconomicsData(economicsData)
}

func createDefaultEconomicsConfig() *config.EconomicsConfig {
	maxGasLimitPerBlock := strconv.FormatUint(MaxGasLimitPerBlock, 10)
	minGasPrice := strconv.FormatUint(MinTxGasPrice, 10)
	minGasLimit := strconv.FormatUint(MinTxGasLimit, 10)

	return &config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "2000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			RewardsConfigByEpoch: []config.EpochRewardSettings{
				{
					LeaderPercentage:                 0.1,
					DeveloperPercentage:              0.1,
					ProtocolSustainabilityAddress:    testProtocolSustainabilityAddress,
					TopUpFactor:                      0.25,
					TopUpGradientPoint:               "300000000000000000000",
					ProtocolSustainabilityPercentage: 0.1,
				},
			},
		},
		FeeSettings: config.FeeSettings{
			GasLimitSettings: []config.GasLimitSetting{
				{
					MaxGasLimitPerBlock:         maxGasLimitPerBlock,
					MaxGasLimitPerMiniBlock:     maxGasLimitPerBlock,
					MaxGasLimitPerMetaBlock:     maxGasLimitPerBlock,
					MaxGasLimitPerMetaMiniBlock: maxGasLimitPerBlock,
					MaxGasLimitPerTx:            maxGasLimitPerBlock,
					MinGasLimit:                 minGasLimit,
				},
			},
			MinGasPrice:      minGasPrice,
			GasPerDataByte:   "1",
			GasPriceModifier: 0.01,
		},
	}
}

func (tpn *TestProcessorNode) initRatingsData() {
	if tpn.RatingsData == nil {
		tpn.RatingsData = CreateRatingsData()
	}
}

// CreateRatingsData creates a mock RatingsData object
func CreateRatingsData() *rating.RatingsData {
	ratingsConfig := config.RatingsConfig{
		ShardChain: config.ShardChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 50,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  1.1,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 50,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  1.1,
			},
		},
		General: config.General{
			StartRating:           500000,
			MaxRating:             1000000,
			MinRating:             1,
			SignedBlocksThreshold: 0.025,
			SelectionChances: []*config.SelectionChance{
				{
					MaxThreshold:  0,
					ChancePercent: 5,
				},
				{
					MaxThreshold:  100000,
					ChancePercent: 0,
				},
				{
					MaxThreshold:  200000,
					ChancePercent: 16,
				},
				{
					MaxThreshold:  300000,
					ChancePercent: 17,
				},
				{
					MaxThreshold:  400000,
					ChancePercent: 18,
				},
				{
					MaxThreshold:  500000,
					ChancePercent: 19,
				},
				{
					MaxThreshold:  600000,
					ChancePercent: 20,
				},
				{
					MaxThreshold:  700000,
					ChancePercent: 21,
				},
				{
					MaxThreshold:  800000,
					ChancePercent: 22,
				},
				{
					MaxThreshold:  900000,
					ChancePercent: 23,
				},
				{
					MaxThreshold:  1000000,
					ChancePercent: 24,
				},
			},
		},
	}

	ratingDataArgs := rating.RatingsDataArg{
		Config:                   ratingsConfig,
		ShardConsensusSize:       63,
		MetaConsensusSize:        400,
		ShardMinNodes:            400,
		MetaMinNodes:             400,
		RoundDurationMiliseconds: 6000,
	}

	ratingsData, _ := rating.NewRatingsData(ratingDataArgs)
	return ratingsData
}

func (tpn *TestProcessorNode) initInterceptors(heartbeatPk string) {
	var err error
	tpn.BlockBlackListHandler = cache.NewTimeCache(TimeSpanForBadHeaders)
	if check.IfNil(tpn.EpochStartNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.TxMarshalizerField = TestTxSignMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.ChainIdCalled = func() string {
		return string(tpn.ChainID)
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return tpn.MinTransactionVersion
	}
	coreComponents.TxVersionCheckField = versioning.NewTxVersionChecker(tpn.MinTransactionVersion)
	coreComponents.EnableEpochsHandlerField = tpn.EnableEpochsHandler
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.EconomicsDataField = tpn.EconomicsData

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.PubKey = nil
	cryptoComponents.BlockSig = tpn.OwnAccount.BlockSingleSigner
	cryptoComponents.TxSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(TestMultiSig)
	cryptoComponents.BlKeyGen = tpn.OwnAccount.KeygenBlockSign
	cryptoComponents.TxKeyGen = tpn.OwnAccount.KeygenTxSign

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime: tpn.RoundHandler.TimeStamp(),
			Settings: &config.EpochStartConfig{
				MinRoundsBetweenEpochs: 1000,
				RoundsPerEpoch:         10000,
			},
			Epoch:              0,
			EpochStartNotifier: tpn.EpochStartNotifier,
			Storage:            tpn.Storage,
			Marshalizer:        TestMarshalizer,
			Hasher:             TestHasher,
			AppStatusHandler:   &statusHandlerMock.AppStatusHandlerStub{},
			DataPool:           tpn.DataPool,
		}
		epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
		tpn.EpochStartTrigger = &metachain.TestTrigger{}
		tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)
		providedHardforkPk := tpn.createHardforkTrigger(heartbeatPk)
		coreComponents.HardforkTriggerPubKeyField = providedHardforkPk

		metaInterceptorContainerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
			CoreComponents:               coreComponents,
			CryptoComponents:             cryptoComponents,
			Accounts:                     tpn.AccntState,
			ShardCoordinator:             tpn.ShardCoordinator,
			NodesCoordinator:             tpn.NodesCoordinator,
			Messenger:                    tpn.Messenger,
			Store:                        tpn.Storage,
			DataPool:                     tpn.DataPool,
			MaxTxNonceDeltaAllowed:       maxTxNonceDeltaAllowed,
			TxFeeHandler:                 tpn.EconomicsData,
			BlockBlackList:               tpn.BlockBlackListHandler,
			HeaderSigVerifier:            tpn.HeaderSigVerifier,
			HeaderIntegrityVerifier:      tpn.HeaderIntegrityVerifier,
			ValidityAttester:             tpn.BlockTracker,
			EpochStartTrigger:            tpn.EpochStartTrigger,
			WhiteListHandler:             tpn.WhiteListHandler,
			WhiteListerVerifiedTxs:       tpn.WhiteListerVerifiedTxs,
			AntifloodHandler:             &mock.NilAntifloodHandler{},
			ArgumentsParser:              smartContract.NewArgumentParser(),
			PreferredPeersHolder:         &p2pmocks.PeersHolderStub{},
			SizeCheckDelta:               sizeCheckDelta,
			RequestHandler:               tpn.RequestHandler,
			PeerSignatureHandler:         &processMock.PeerSignatureHandlerStub{},
			SignaturesHandler:            &processMock.SignaturesHandlerStub{},
			HeartbeatExpiryTimespanInSec: 30,
			PeerShardMapper:              tpn.PeerShardMapper,
			HardforkTrigger:              tpn.HardforkTrigger,
		}
		interceptorContainerFactory, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(metaInterceptorContainerFactoryArgs)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			log.Debug("interceptor container factory Create", "error", err.Error())
		}
	} else {
		argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool:     tpn.DataPool.MiniBlocks(),
			ValidatorsInfoPool: tpn.DataPool.ValidatorsInfo(),
			RequestHandler:     tpn.RequestHandler,
		}
		peerMiniBlockSyncer, _ := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)
		argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          TestMarshalizer,
			Hasher:               TestHasher,
			HeaderValidator:      tpn.HeaderValidator,
			Uint64Converter:      TestUint64Converter,
			DataPool:             tpn.DataPool,
			Storage:              tpn.Storage,
			RequestHandler:       tpn.RequestHandler,
			Epoch:                0,
			Validity:             1,
			Finality:             1,
			EpochStartNotifier:   tpn.EpochStartNotifier,
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
			RoundHandler:         tpn.RoundHandler,
			AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
			EnableEpochsHandler:  tpn.EnableEpochsHandler,
		}
		epochStartTrigger, _ := shardchain.NewEpochStartTrigger(argsShardEpochStart)
		tpn.EpochStartTrigger = &shardchain.TestTrigger{}
		tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)
		providedHardforkPk := tpn.createHardforkTrigger(heartbeatPk)
		coreComponents.HardforkTriggerPubKeyField = providedHardforkPk

		shardIntereptorContainerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
			CoreComponents:               coreComponents,
			CryptoComponents:             cryptoComponents,
			Accounts:                     tpn.AccntState,
			ShardCoordinator:             tpn.ShardCoordinator,
			NodesCoordinator:             tpn.NodesCoordinator,
			Messenger:                    tpn.Messenger,
			Store:                        tpn.Storage,
			DataPool:                     tpn.DataPool,
			MaxTxNonceDeltaAllowed:       maxTxNonceDeltaAllowed,
			TxFeeHandler:                 tpn.EconomicsData,
			BlockBlackList:               tpn.BlockBlackListHandler,
			HeaderSigVerifier:            tpn.HeaderSigVerifier,
			HeaderIntegrityVerifier:      tpn.HeaderIntegrityVerifier,
			ValidityAttester:             tpn.BlockTracker,
			EpochStartTrigger:            tpn.EpochStartTrigger,
			WhiteListHandler:             tpn.WhiteListHandler,
			WhiteListerVerifiedTxs:       tpn.WhiteListerVerifiedTxs,
			AntifloodHandler:             &mock.NilAntifloodHandler{},
			ArgumentsParser:              smartContract.NewArgumentParser(),
			PreferredPeersHolder:         &p2pmocks.PeersHolderStub{},
			SizeCheckDelta:               sizeCheckDelta,
			RequestHandler:               tpn.RequestHandler,
			PeerSignatureHandler:         &processMock.PeerSignatureHandlerStub{},
			SignaturesHandler:            &processMock.SignaturesHandlerStub{},
			HeartbeatExpiryTimespanInSec: 30,
			PeerShardMapper:              tpn.PeerShardMapper,
			HardforkTrigger:              tpn.HardforkTrigger,
		}
		interceptorContainerFactory, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(shardIntereptorContainerFactoryArgs)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (tpn *TestProcessorNode) createHardforkTrigger(heartbeatPk string) []byte {
	pkBytes, _ := tpn.NodeKeys.Pk.ToByteArray()
	argHardforkTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        pkBytes,
		Enabled:                   true,
		EnabledAuthenticated:      true,
		ArgumentParser:            smartContract.NewArgumentParser(),
		EpochProvider:             tpn.EpochStartTrigger,
		ExportFactoryHandler:      &mock.ExportFactoryHandlerStub{},
		CloseAfterExportInMinutes: 5,
		ChanStopNodeProcess:       make(chan endProcess.ArgEndProcess),
		EpochConfirmedNotifier:    tpn.EpochStartNotifier,
		SelfPubKeyBytes:           pkBytes,
		ImportStartHandler:        &mock.ImportStartHandlerStub{},
		RoundHandler:              &mock.RoundHandlerMock{},
	}

	var err error
	if len(heartbeatPk) > 0 {
		argHardforkTrigger.TriggerPubKeyBytes, err = hex.DecodeString(heartbeatPk)
		log.LogIfError(err)
	}
	tpn.HardforkTrigger, err = trigger.NewTrigger(argHardforkTrigger)
	log.LogIfError(err)

	return argHardforkTrigger.TriggerPubKeyBytes
}

func (tpn *TestProcessorNode) initResolvers() {
	dataPacker, _ := partitioning.NewSimpleDataPacker(TestMarshalizer)

	_ = tpn.Messenger.CreateTopic(common.ConsensusTopic+tpn.ShardCoordinator.CommunicationIdentifier(tpn.ShardCoordinator.SelfId()), true)
	payloadValidator, _ := validator.NewPeerAuthenticationPayloadValidator(60)

	resolverContainerFactory := resolverscontainer.FactoryArgs{
		ShardCoordinator:            tpn.ShardCoordinator,
		Messenger:                   tpn.Messenger,
		Store:                       tpn.Storage,
		Marshalizer:                 TestMarshalizer,
		DataPools:                   tpn.DataPool,
		Uint64ByteSliceConverter:    TestUint64Converter,
		DataPacker:                  dataPacker,
		TriesContainer:              tpn.TrieContainer,
		SizeCheckDelta:              100,
		InputAntifloodHandler:       &mock.NilAntifloodHandler{},
		OutputAntifloodHandler:      &mock.NilAntifloodHandler{},
		NumConcurrentResolvingJobs:  10,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		ResolverConfig: config.ResolverConfig{
			NumCrossShardPeers:  2,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3,
		},
		PeersRatingHandler: tpn.PeersRatingHandler,
		PayloadValidator:   payloadValidator,
	}

	var err error
	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		resolversContainerFactory, _ := resolverscontainer.NewMetaResolversContainerFactory(resolverContainerFactory)

		tpn.ResolversContainer, err = resolversContainerFactory.Create()
		log.LogIfError(err)

		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
			tpn.WhiteListHandler,
			100,
			tpn.ShardCoordinator.SelfId(),
			time.Second,
		)
	} else {
		resolversContainerFactory, _ := resolverscontainer.NewShardResolversContainerFactory(resolverContainerFactory)

		tpn.ResolversContainer, err = resolversContainerFactory.Create()
		log.LogIfError(err)

		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewResolverRequestHandler(
			tpn.ResolverFinder,
			tpn.RequestedItemsHandler,
			tpn.WhiteListHandler,
			100,
			tpn.ShardCoordinator.SelfId(),
			time.Second,
		)
	}
}

func (tpn *TestProcessorNode) initInnerProcessors(gasMap map[string]map[string]uint64, vmConfig *config.VirtualMachineConfig) {
	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.initMetaInnerProcessors(gasMap)
		return
	}

	if tpn.ValidatorStatisticsProcessor == nil {
		tpn.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorStub{}
	}

	interimProcFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressPubkeyConverter,
		tpn.Storage,
		tpn.DataPool,
		tpn.EconomicsData,
	)

	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = postprocess.NewTestIntermediateResultsProcessor(
		TestHasher,
		TestMarshalizer,
		tpn.ShardCoordinator,
		TestAddressPubkeyConverter,
		tpn.Storage,
		dataBlock.SmartContractResultBlock,
		tpn.DataPool.CurrentBlockTxs(),
		tpn.EconomicsData,
	)

	tpn.InterimProcContainer.Remove(dataBlock.SmartContractResultBlock)
	_ = tpn.InterimProcContainer.Add(dataBlock.SmartContractResultBlock, tpn.ScrForwarder)

	tpn.RewardsProcessor, _ = rewardTransaction.NewRewardTxProcessor(
		tpn.AccntState,
		TestAddressPubkeyConverter,
		tpn.ShardCoordinator,
	)

	mapDNSAddresses := make(map[string]struct{})
	if !check.IfNil(tpn.SmartContractParser) {
		mapDNSAddresses, _ = tpn.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	}

	gasSchedule := mock.NewGasScheduleNotifierMock(gasMap)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasSchedule,
		MapDNSAddresses:           mapDNSAddresses,
		Marshalizer:               TestMarshalizer,
		Accounts:                  tpn.AccntState,
		ShardCoordinator:          tpn.ShardCoordinator,
		EpochNotifier:             tpn.EpochNotifier,
		EnableEpochsHandler:       tpn.EnableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncFactory, _ := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)

	for name, function := range TestBuiltinFunctions {
		err := builtInFuncFactory.BuiltInFunctionContainer().Add(name, function)
		log.LogIfError(err)
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:              tpn.AccntState,
		PubkeyConv:            TestAddressPubkeyConverter,
		StorageService:        tpn.Storage,
		BlockChain:            tpn.BlockChain,
		ShardCoordinator:      tpn.ShardCoordinator,
		Marshalizer:           TestMarshalizer,
		Uint64Converter:       TestUint64Converter,
		BuiltInFunctions:      builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:     builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler: builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:              tpn.DataPool,
		CompiledSCPool:        tpn.DataPool.SmartContracts(),
		EpochNotifier:         tpn.EpochNotifier,
		EnableEpochsHandler:   tpn.EnableEpochsHandler,
		NilCompiledSCStore:    true,
	}
	esdtTransferParser, _ := parsers.NewESDTTransferParser(TestMarshalizer)
	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)
	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
	tpn.EnableEpochs.FailExecutionOnEveryAPIErrorEnableEpoch = 1
	argsNewVMFactory := shard.ArgVMContainerFactory{
		Config:              *vmConfig,
		BlockGasLimit:       maxGasLimitPerBlock,
		GasSchedule:         gasSchedule,
		BlockChainHook:      blockChainHookImpl,
		BuiltInFunctions:    argsHook.BuiltInFunctions,
		EpochNotifier:       tpn.EpochNotifier,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
		ArwenChangeLocker:   tpn.ArwenChangeLocker,
		ESDTTransferParser:  esdtTransferParser,
	}
	vmFactory, _ := shard.NewVMContainerFactory(argsNewVMFactory)

	var err error
	tpn.VMContainer, err = vmFactory.Create()
	if err != nil {
		panic(err)
	}

	tpn.BlockchainHook, _ = vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	_ = builtInFuncFactory.SetPayableHandler(tpn.BlockchainHook)

	mockVM, _ := mock.NewOneSCExecutorMockVM(tpn.BlockchainHook, TestHasher)
	mockVM.GasForOperation = OpGasValueForMockVm
	_ = tpn.VMContainer.Add(procFactory.InternalTestingVM, mockVM)

	tpn.FeeAccumulator, _ = postprocess.NewFeeAccumulator()
	tpn.ArgsParser = smartContract.NewArgumentParser()

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	tpn.GasHandler, _ = preprocess.NewGasComputation(tpn.EconomicsData, txTypeHandler, tpn.EnableEpochsHandler)
	badBlocksHandler, _ := tpn.InterimProcContainer.Get(dataBlock.InvalidBlock)

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:         tpn.VMContainer,
		ArgsParser:          tpn.ArgsParser,
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		AccountsDB:          tpn.AccntState,
		BlockChainHook:      vmFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		PubkeyConv:          TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		ScrForwarder:        tpn.ScrForwarder,
		TxFeeHandler:        tpn.FeeAccumulator,
		EconomicsFee:        tpn.EconomicsData,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          tpn.GasHandler,
		GasSchedule:         gasSchedule,
		TxLogsProcessor:     tpn.TransactionLogProcessor,
		BadTxForwarder:      badBlocksHandler,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
		VMOutputCacher:      txcache.NewDisabledCache(),
		ArwenChangeLocker:   tpn.ArwenChangeLocker,
	}
	sc, _ := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	tpn.ScProcessor = smartContract.NewTestScProcessor(sc)

	receiptsHandler, _ := tpn.InterimProcContainer.Get(dataBlock.ReceiptBlock)
	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:            tpn.AccntState,
		Hasher:              TestHasher,
		PubkeyConv:          TestAddressPubkeyConverter,
		Marshalizer:         TestMarshalizer,
		SignMarshalizer:     TestTxSignMarshalizer,
		ShardCoordinator:    tpn.ShardCoordinator,
		ScProcessor:         tpn.ScProcessor,
		TxFeeHandler:        tpn.FeeAccumulator,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        tpn.EconomicsData,
		ReceiptForwarder:    receiptsHandler,
		BadTxForwarder:      badBlocksHandler,
		ArgsParser:          tpn.ArgsParser,
		ScrForwarder:        tpn.ScrForwarder,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	tpn.TxProcessor, _ = transaction.NewTxProcessor(argsNewTxProcessor)
	scheduledSCRsStorer, _ := tpn.Storage.GetStorer(dataRetriever.ScheduledSCRsUnit)
	scheduledTxsExecutionHandler, _ := preprocess.NewScheduledTxsExecution(
		tpn.TxProcessor,
		&mock.TransactionCoordinatorMock{},
		scheduledSCRsStorer,
		TestMarshalizer,
		TestHasher,
		tpn.ShardCoordinator,
	)
	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()

	fact, _ := shard.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.DataPool,
		TestAddressPubkeyConverter,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		tpn.ScProcessor,
		tpn.ScProcessor,
		tpn.RewardsProcessor,
		tpn.EconomicsData,
		tpn.GasHandler,
		tpn.BlockTracker,
		TestBlockSizeComputationHandler,
		TestBalanceComputationHandler,
		tpn.EnableEpochsHandler,
		txTypeHandler,
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		MiniBlockPool:                tpn.DataPool.MiniBlocks(),
		RequestHandler:               tpn.RequestHandler,
		PreProcessors:                tpn.PreProcessorsContainer,
		InterProcessors:              tpn.InterimProcContainer,
		GasHandler:                   tpn.GasHandler,
		FeeHandler:                   tpn.FeeAccumulator,
		BlockSizeComputation:         TestBlockSizeComputationHandler,
		BalanceComputation:           TestBalanceComputationHandler,
		EconomicsFee:                 tpn.EconomicsData,
		TxTypeHandler:                txTypeHandler,
		TransactionsLogProcessor:     tpn.TransactionLogProcessor,
		EnableEpochsHandler:          tpn.EnableEpochsHandler,
		ScheduledTxsExecutionHandler: scheduledTxsExecutionHandler,
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   processedMiniBlocksTracker,
	}
	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	scheduledTxsExecutionHandler.SetTransactionCoordinator(tpn.TxCoordinator)
}

func (tpn *TestProcessorNode) initMetaInnerProcessors(gasMap map[string]map[string]uint64) {
	interimProcFactory, _ := metaProcess.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressPubkeyConverter,
		tpn.Storage,
		tpn.DataPool,
		tpn.EconomicsData,
	)

	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = postprocess.NewTestIntermediateResultsProcessor(
		TestHasher,
		TestMarshalizer,
		tpn.ShardCoordinator,
		TestAddressPubkeyConverter,
		tpn.Storage,
		dataBlock.SmartContractResultBlock,
		tpn.DataPool.CurrentBlockTxs(),
		tpn.EconomicsData,
	)

	tpn.InterimProcContainer.Remove(dataBlock.SmartContractResultBlock)
	_ = tpn.InterimProcContainer.Add(dataBlock.SmartContractResultBlock, tpn.ScrForwarder)

	gasSchedule := mock.NewGasScheduleNotifierMock(gasMap)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasSchedule,
		MapDNSAddresses:           make(map[string]struct{}),
		Marshalizer:               TestMarshalizer,
		Accounts:                  tpn.AccntState,
		ShardCoordinator:          tpn.ShardCoordinator,
		EpochNotifier:             tpn.EpochNotifier,
		EnableEpochsHandler:       tpn.EnableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncFactory, _ := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)
	argsHook := hooks.ArgBlockChainHook{
		Accounts:              tpn.AccntState,
		PubkeyConv:            TestAddressPubkeyConverter,
		StorageService:        tpn.Storage,
		BlockChain:            tpn.BlockChain,
		ShardCoordinator:      tpn.ShardCoordinator,
		Marshalizer:           TestMarshalizer,
		Uint64Converter:       TestUint64Converter,
		BuiltInFunctions:      builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:     builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler: builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:              tpn.DataPool,
		CompiledSCPool:        tpn.DataPool.SmartContracts(),
		EpochNotifier:         tpn.EpochNotifier,
		EnableEpochsHandler:   tpn.EnableEpochsHandler,
		NilCompiledSCStore:    true,
	}

	var signVerifier vm.MessageSignVerifier
	if tpn.UseValidVmBlsSigVerifier {
		signVerifier, _ = vmProcess.NewMessageSigVerifier(
			signing.NewKeyGenerator(mcl.NewSuiteBLS12()),
			mclsig.NewBlsSigner(),
		)
	} else {
		signVerifier, _ = disabled.NewMessageSignVerifier(&mock.KeyGenMock{})
	}
	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
	argsVMContainerFactory := metaProcess.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          argsHook.PubkeyConv,
		Economics:           tpn.EconomicsData,
		MessageSignVerifier: signVerifier,
		GasSchedule:         gasSchedule,
		NodesConfigProvider: tpn.NodesSetup,
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: DelegationManagerConfigChangeAddress,
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				UnBondPeriodInEpochs:                 1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             100,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: DelegationManagerConfigChangeAddress,
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100000,
			},
		},
		ValidatorAccountsDB: tpn.PeerState,
		ChanceComputer:      &mock.RaterMock{},
		ShardCoordinator:    tpn.ShardCoordinator,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	vmFactory, _ := metaProcess.NewVMContainerFactory(argsVMContainerFactory)

	tpn.VMContainer, _ = vmFactory.Create()
	tpn.BlockchainHook, _ = vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	tpn.SystemSCFactory = vmFactory.SystemSmartContractContainerFactory()
	tpn.addMockVm(tpn.BlockchainHook)

	tpn.FeeAccumulator, _ = postprocess.NewFeeAccumulator()
	tpn.ArgsParser = smartContract.NewArgumentParser()
	esdtTransferParser, _ := parsers.NewESDTTransferParser(TestMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	tpn.GasHandler, _ = preprocess.NewGasComputation(tpn.EconomicsData, txTypeHandler, tpn.EnableEpochsHandler)
	badBlocksHandler, _ := tpn.InterimProcContainer.Get(dataBlock.InvalidBlock)
	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:         tpn.VMContainer,
		ArgsParser:          tpn.ArgsParser,
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		AccountsDB:          tpn.AccntState,
		BlockChainHook:      vmFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		PubkeyConv:          TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		ScrForwarder:        tpn.ScrForwarder,
		TxFeeHandler:        tpn.FeeAccumulator,
		EconomicsFee:        tpn.EconomicsData,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          tpn.GasHandler,
		GasSchedule:         gasSchedule,
		TxLogsProcessor:     tpn.TransactionLogProcessor,
		BadTxForwarder:      badBlocksHandler,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
		VMOutputCacher:      txcache.NewDisabledCache(),
		ArwenChangeLocker:   tpn.ArwenChangeLocker,
	}
	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	tpn.ScProcessor = smartContract.NewTestScProcessor(scProcessor)
	argsNewMetaTxProc := transaction.ArgsNewMetaTxProcessor{
		Hasher:              TestHasher,
		Marshalizer:         TestMarshalizer,
		Accounts:            tpn.AccntState,
		PubkeyConv:          TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		ScProcessor:         tpn.ScProcessor,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        tpn.EconomicsData,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	tpn.TxProcessor, _ = transaction.NewMetaTxProcessor(argsNewMetaTxProc)
	scheduledSCRsStorer, _ := tpn.Storage.GetStorer(dataRetriever.ScheduledSCRsUnit)
	scheduledTxsExecutionHandler, _ := preprocess.NewScheduledTxsExecution(
		tpn.TxProcessor,
		&mock.TransactionCoordinatorMock{},
		scheduledSCRsStorer,
		TestMarshalizer,
		TestHasher,
		tpn.ShardCoordinator)
	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()

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
		tpn.EconomicsData,
		tpn.GasHandler,
		tpn.BlockTracker,
		TestAddressPubkeyConverter,
		TestBlockSizeComputationHandler,
		TestBalanceComputationHandler,
		tpn.EnableEpochsHandler,
		txTypeHandler,
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		MiniBlockPool:                tpn.DataPool.MiniBlocks(),
		RequestHandler:               tpn.RequestHandler,
		PreProcessors:                tpn.PreProcessorsContainer,
		InterProcessors:              tpn.InterimProcContainer,
		GasHandler:                   tpn.GasHandler,
		FeeHandler:                   tpn.FeeAccumulator,
		BlockSizeComputation:         TestBlockSizeComputationHandler,
		BalanceComputation:           TestBalanceComputationHandler,
		EconomicsFee:                 tpn.EconomicsData,
		TxTypeHandler:                txTypeHandler,
		TransactionsLogProcessor:     tpn.TransactionLogProcessor,
		EnableEpochsHandler:          tpn.EnableEpochsHandler,
		ScheduledTxsExecutionHandler: scheduledTxsExecutionHandler,
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   processedMiniBlocksTracker,
	}
	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	scheduledTxsExecutionHandler.SetTransactionCoordinator(tpn.TxCoordinator)
}

// InitDelegationManager will initialize the delegation manager whenever required
func (tpn *TestProcessorNode) InitDelegationManager() {
	if tpn.ShardCoordinator.SelfId() != core.MetachainShardId {
		return
	}

	systemVM, err := tpn.VMContainer.Get(factory.SystemVirtualMachine)
	log.LogIfError(err)

	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	vmInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.DelegationManagerSCAddress,
			Arguments:  [][]byte{},
			CallValue:  zero,
		},
		ContractCode:         vm.DelegationManagerSCAddress,
		ContractCodeMetadata: codeMetaData.ToBytes(),
	}

	vmOutput, err := systemVM.RunSmartContractCreate(vmInput)
	log.LogIfError(err)
	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Error("error while initializing system SC", "return code", vmOutput.ReturnCode)
	}

	err = tpn.processSCOutputAccounts(vmOutput)
	log.LogIfError(err)

	err = tpn.updateSystemSCContractsCode(vmInput.ContractCodeMetadata, vm.DelegationManagerSCAddress)
	log.LogIfError(err)

	_, err = tpn.AccntState.Commit()
	log.LogIfError(err)
}

func (tpn *TestProcessorNode) updateSystemSCContractsCode(contractMetadata []byte, scAddress []byte) error {
	userAcc, err := tpn.getUserAccount(scAddress)
	if err != nil {
		return err
	}

	userAcc.SetOwnerAddress(scAddress)
	userAcc.SetCodeMetadata(contractMetadata)
	userAcc.SetCode(scAddress)

	return tpn.AccntState.SaveAccount(userAcc)
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (tpn *TestProcessorNode) processSCOutputAccounts(vmOutput *vmcommon.VMOutput) error {
	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		acc, err := tpn.getUserAccount(outAcc.Address)
		if err != nil {
			return err
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			err = acc.SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				return err
			}
		}

		if outAcc.BalanceDelta != nil && outAcc.BalanceDelta.Cmp(zero) != 0 {
			err = acc.AddToBalance(outAcc.BalanceDelta)
			if err != nil {
				return err
			}
		}

		err = tpn.AccntState.SaveAccount(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tpn *TestProcessorNode) getUserAccount(address []byte) (state.UserAccountHandler, error) {
	acnt, err := tpn.AccntState.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	stAcc, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return stAcc, nil
}

func (tpn *TestProcessorNode) addMockVm(blockchainHook vmcommon.BlockchainHook) {
	mockVM, _ := mock.NewOneSCExecutorMockVM(blockchainHook, TestHasher)
	mockVM.GasForOperation = OpGasValueForMockVm

	_ = tpn.VMContainer.Add(factory.InternalTestingVM, mockVM)
}

func (tpn *TestProcessorNode) initBlockProcessor(stateCheckpointModulus uint) {
	var err error

	if tpn.ShardCoordinator.SelfId() != core.MetachainShardId {
		tpn.ForkDetector, _ = processSync.NewShardForkDetector(tpn.RoundHandler, tpn.BlockBlackListHandler, tpn.BlockTracker, tpn.NodesSetup.GetStartTime())
	} else {
		tpn.ForkDetector, _ = processSync.NewMetaForkDetector(tpn.RoundHandler, tpn.BlockBlackListHandler, tpn.BlockTracker, tpn.NodesSetup.GetStartTime())
	}

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = tpn.AccntState
	accountsDb[state.PeerAccountsState] = tpn.PeerState

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.RoundHandlerField = tpn.RoundHandler
	coreComponents.EnableEpochsHandlerField = tpn.EnableEpochsHandler
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.EconomicsDataField = tpn.EconomicsData

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = tpn.Storage
	dataComponents.DataPool = tpn.DataPool
	dataComponents.BlockChain = tpn.BlockChain

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator)
	bootstrapComponents.HdrIntegrityVerifier = tpn.HeaderIntegrityVerifier

	statusComponents := GetDefaultStatusComponents()

	triesConfig := config.Config{
		StateTriesConfig: config.StateTriesConfig{
			SnapshotsEnabled:          true,
			CheckpointRoundsModulus:   stateCheckpointModulus,
			UserStatePruningQueueSize: uint(5),
			PeerStatePruningQueueSize: uint(3),
		},
	}

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:      coreComponents,
		DataComponents:      dataComponents,
		BootstrapComponents: bootstrapComponents,
		StatusComponents:    statusComponents,
		Config:              triesConfig,
		AccountsDB:          accountsDb,
		ForkDetector:        tpn.ForkDetector,
		NodesCoordinator:    tpn.NodesCoordinator,
		FeeHandler:          tpn.FeeAccumulator,
		RequestHandler:      tpn.RequestHandler,
		BlockChainHook:      tpn.BlockchainHook,
		HeaderValidator:     tpn.HeaderValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                 tpn.BlockTracker,
		BlockSizeThrottler:           TestBlockSizeThrottler,
		HistoryRepository:            tpn.HistoryRepository,
		EnableRoundsHandler:          coreComponents.EnableRoundsHandler(),
		GasHandler:                   tpn.GasHandler,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:           &testscommon.ReceiptsRepositoryStub{},
		OutportDataProvider:          &outport.OutportDataProviderStub{},
	}

	if check.IfNil(tpn.EpochStartNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		if check.IfNil(tpn.EpochStartTrigger) {
			argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
				GenesisTime: argumentsBase.CoreComponents.RoundHandler().TimeStamp(),
				Settings: &config.EpochStartConfig{
					MinRoundsBetweenEpochs: 1000,
					RoundsPerEpoch:         10000,
				},
				Epoch:              0,
				EpochStartNotifier: tpn.EpochStartNotifier,
				Storage:            tpn.Storage,
				Marshalizer:        TestMarshalizer,
				Hasher:             TestHasher,
				AppStatusHandler:   &statusHandlerMock.AppStatusHandlerStub{},
				DataPool:           tpn.DataPool,
			}
			epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
			tpn.EpochStartTrigger = &metachain.TestTrigger{}
			tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)
		}

		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.TxCoordinator = tpn.TxCoordinator

		argsStakingToPeer := scToProtocol.ArgStakingToPeer{
			PubkeyConv:          TestValidatorPubkeyConverter,
			Hasher:              TestHasher,
			Marshalizer:         TestMarshalizer,
			PeerState:           tpn.PeerState,
			BaseState:           tpn.AccntState,
			ArgParser:           tpn.ArgsParser,
			CurrTxs:             tpn.DataPool.CurrentBlockTxs(),
			RatingsData:         tpn.RatingsData,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
		}
		scToProtocolInstance, _ := scToProtocol.NewStakingToPeer(argsStakingToPeer)

		argsEpochStartData := metachain.ArgsNewEpochStartData{
			Marshalizer:         TestMarshalizer,
			Hasher:              TestHasher,
			Store:               tpn.Storage,
			DataPool:            tpn.DataPool,
			BlockTracker:        tpn.BlockTracker,
			ShardCoordinator:    tpn.ShardCoordinator,
			EpochStartTrigger:   tpn.EpochStartTrigger,
			RequestHandler:      tpn.RequestHandler,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
		}
		epochStartDataCreator, _ := metachain.NewEpochStartData(argsEpochStartData)

		economicsDataProvider := metachain.NewEpochEconomicsStatistics()
		argsEpochEconomics := metachain.ArgsNewEpochEconomics{
			Marshalizer:           TestMarshalizer,
			Hasher:                TestHasher,
			Store:                 tpn.Storage,
			ShardCoordinator:      tpn.ShardCoordinator,
			RewardsHandler:        tpn.EconomicsData,
			RoundTime:             tpn.RoundHandler,
			GenesisTotalSupply:    tpn.EconomicsData.GenesisTotalSupply(),
			EconomicsDataNotified: economicsDataProvider,
			StakingV2EnableEpoch:  tpn.EnableEpochs.StakingV2EnableEpoch,
		}
		epochEconomics, _ := metachain.NewEndOfEpochEconomicsDataCreator(argsEpochEconomics)

		systemVM, errGet := tpn.VMContainer.Get(factory.SystemVirtualMachine)
		if errGet != nil {
			log.Error("initBlockProcessor tpn.VMContainer.Get", "error", errGet)
		}
		stakingDataProvider, errRsp := metachain.NewStakingDataProvider(systemVM, "1000")
		if errRsp != nil {
			log.Error("initBlockProcessor NewRewardsStakingProvider", "error", errRsp)
		}

		rewardsStorage, _ := tpn.Storage.GetStorer(dataRetriever.RewardTransactionUnit)
		miniBlockStorage, _ := tpn.Storage.GetStorer(dataRetriever.MiniBlockUnit)
		argsEpochRewards := metachain.RewardsCreatorProxyArgs{
			BaseRewardsCreatorArgs: metachain.BaseRewardsCreatorArgs{
				ShardCoordinator:              tpn.ShardCoordinator,
				PubkeyConverter:               TestAddressPubkeyConverter,
				RewardsStorage:                rewardsStorage,
				MiniBlockStorage:              miniBlockStorage,
				Hasher:                        TestHasher,
				Marshalizer:                   TestMarshalizer,
				DataPool:                      tpn.DataPool,
				ProtocolSustainabilityAddress: testProtocolSustainabilityAddress,
				NodesConfigProvider:           tpn.NodesCoordinator,
				UserAccountsDB:                tpn.AccntState,
				EnableEpochsHandler:           tpn.EnableEpochsHandler,
			},
			StakingDataProvider:   stakingDataProvider,
			RewardsHandler:        tpn.EconomicsData,
			EconomicsDataProvider: economicsDataProvider,
		}
		epochStartRewards, _ := metachain.NewRewardsCreatorProxy(argsEpochRewards)

		validatorInfoStorage, _ := tpn.Storage.GetStorer(dataRetriever.UnsignedTransactionUnit)
		argsEpochValidatorInfo := metachain.ArgsNewValidatorInfoCreator{
			ShardCoordinator:     tpn.ShardCoordinator,
			ValidatorInfoStorage: validatorInfoStorage,
			MiniBlockStorage:     miniBlockStorage,
			Hasher:               TestHasher,
			Marshalizer:          TestMarshalizer,
			DataPool:             tpn.DataPool,
			EnableEpochsHandler:  tpn.EnableEpochsHandler,
		}
		epochStartValidatorInfo, _ := metachain.NewValidatorInfoCreator(argsEpochValidatorInfo)
		argsEpochSystemSC := metachain.ArgsNewEpochStartSystemSCProcessing{
			SystemVM:                systemVM,
			UserAccountsDB:          tpn.AccntState,
			PeerAccountsDB:          tpn.PeerState,
			Marshalizer:             TestMarshalizer,
			StartRating:             tpn.RatingsData.StartRating(),
			ValidatorInfoCreator:    tpn.ValidatorStatisticsProcessor,
			EndOfEpochCallerAddress: vm.EndOfEpochAddress,
			StakingSCAddress:        vm.StakingSCAddress,
			ChanceComputer:          tpn.NodesCoordinator,
			EpochNotifier:           tpn.EpochNotifier,
			GenesisNodesConfig:      tpn.NodesSetup,
			StakingDataProvider:     stakingDataProvider,
			NodesConfigProvider:     tpn.NodesCoordinator,
			ShardCoordinator:        tpn.ShardCoordinator,
			ESDTOwnerAddressBytes:   vm.EndOfEpochAddress,
			EnableEpochsHandler:     tpn.EnableEpochsHandler,
		}
		epochStartSystemSCProcessor, _ := metachain.NewSystemSCProcessor(argsEpochSystemSC)
		tpn.EpochStartSystemSCProcessor = epochStartSystemSCProcessor

		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:             argumentsBase,
			SCToProtocol:                 scToProtocolInstance,
			PendingMiniBlocksHandler:     &mock.PendingMiniBlocksHandlerStub{},
			EpochEconomics:               epochEconomics,
			EpochStartDataCreator:        epochStartDataCreator,
			EpochRewardsCreator:          epochStartRewards,
			EpochValidatorInfoCreator:    epochStartValidatorInfo,
			ValidatorStatisticsProcessor: tpn.ValidatorStatisticsProcessor,
			EpochSystemSCProcessor:       epochStartSystemSCProcessor,
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		if check.IfNil(tpn.EpochStartTrigger) {
			argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
				MiniBlocksPool:     tpn.DataPool.MiniBlocks(),
				ValidatorsInfoPool: tpn.DataPool.ValidatorsInfo(),
				RequestHandler:     tpn.RequestHandler,
			}
			peerMiniBlocksSyncer, _ := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)
			argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
				Marshalizer:          TestMarshalizer,
				Hasher:               TestHasher,
				HeaderValidator:      tpn.HeaderValidator,
				Uint64Converter:      TestUint64Converter,
				DataPool:             tpn.DataPool,
				Storage:              tpn.Storage,
				RequestHandler:       tpn.RequestHandler,
				Epoch:                0,
				Validity:             1,
				Finality:             1,
				EpochStartNotifier:   tpn.EpochStartNotifier,
				PeerMiniBlocksSyncer: peerMiniBlocksSyncer,
				RoundHandler:         tpn.RoundHandler,
				AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
				EnableEpochsHandler:  tpn.EnableEpochsHandler,
			}
			epochStartTrigger, _ := shardchain.NewEpochStartTrigger(argsShardEpochStart)
			tpn.EpochStartTrigger = &shardchain.TestTrigger{}
			tpn.EpochStartTrigger.SetTrigger(epochStartTrigger)
		}

		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.BlockChainHook = tpn.BlockchainHook
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		argumentsBase.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{}

		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		panic(fmt.Sprintf("error creating blockprocessor: %s", err.Error()))
	}
}

func (tpn *TestProcessorNode) setGenesisBlock() {
	genesisBlock := tpn.GenesisBlocks[tpn.ShardCoordinator.SelfId()]
	_ = tpn.BlockChain.SetGenesisHeader(genesisBlock)
	hash, _ := core.CalculateHash(TestMarshalizer, TestHasher, genesisBlock)
	tpn.BlockChain.SetGenesisHeaderHash(hash)
	log.Info("set genesis",
		"shard ID", tpn.ShardCoordinator.SelfId(),
		"hash", hex.EncodeToString(hash),
	)
}

func (tpn *TestProcessorNode) initNode() {
	var err error

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.VmMarshalizerField = TestVmMarshalizer
	coreComponents.TxMarshalizerField = TestTxSignMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.AddressPubKeyConverterField = TestAddressPubkeyConverter
	coreComponents.ValidatorPubKeyConverterField = TestValidatorPubkeyConverter
	coreComponents.ChainIdCalled = func() string {
		return string(tpn.ChainID)
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return tpn.MinTransactionVersion
	}
	coreComponents.TxVersionCheckField = versioning.NewTxVersionChecker(tpn.MinTransactionVersion)
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.EconomicsDataField = tpn.EconomicsData
	coreComponents.APIEconomicsHandler = tpn.EconomicsData
	coreComponents.SyncTimerField = &mock.SyncTimerMock{}
	coreComponents.EnableEpochsHandlerField = tpn.EnableEpochsHandler
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.ArwenChangeLockerInternal = tpn.ArwenChangeLocker
	hardforkPubKeyBytes, _ := coreComponents.ValidatorPubKeyConverterField.Decode(hardforkPubKey)
	coreComponents.HardforkTriggerPubKeyField = hardforkPubKeyBytes

	dataComponents := GetDefaultDataComponents()
	dataComponents.BlockChain = tpn.BlockChain
	dataComponents.DataPool = tpn.DataPool
	dataComponents.Store = tpn.Storage

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator)

	processComponents := GetDefaultProcessComponents()
	processComponents.BlockProcess = tpn.BlockProcessor
	processComponents.ResFinder = tpn.ResolverFinder
	processComponents.HeaderIntegrVerif = tpn.HeaderIntegrityVerifier
	processComponents.HeaderSigVerif = tpn.HeaderSigVerifier
	processComponents.BlackListHdl = tpn.BlockBlackListHandler
	processComponents.NodesCoord = tpn.NodesCoordinator
	processComponents.ShardCoord = tpn.ShardCoordinator
	processComponents.IntContainer = tpn.InterceptorsContainer
	processComponents.HistoryRepositoryInternal = tpn.HistoryRepository
	processComponents.WhiteListHandlerInternal = tpn.WhiteListHandler
	processComponents.WhiteListerVerifiedTxsInternal = tpn.WhiteListerVerifiedTxs
	processComponents.TxsSenderHandlerField = createTxsSender(tpn.ShardCoordinator, tpn.Messenger)
	processComponents.HardforkTriggerField = tpn.HardforkTrigger

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.PrivKey = tpn.NodeKeys.Sk
	cryptoComponents.PubKey = tpn.NodeKeys.Pk
	cryptoComponents.TxSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.BlockSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(tpn.MultiSigner)
	cryptoComponents.BlKeyGen = tpn.OwnAccount.KeygenTxSign
	cryptoComponents.TxKeyGen = TestKeyGenForAccounts

	stateComponents := GetDefaultStateComponents()
	stateComponents.Accounts = tpn.AccntState
	stateComponents.AccountsAPI = tpn.AccntState

	finalProvider, _ := blockInfoProviders.NewFinalBlockInfo(dataComponents.BlockChain)
	finalAccountsApi, _ := state.NewAccountsDBApi(tpn.AccntState, finalProvider)

	currentProvider, _ := blockInfoProviders.NewCurrentBlockInfo(dataComponents.BlockChain)
	currentAccountsApi, _ := state.NewAccountsDBApi(tpn.AccntState, currentProvider)

	historicalAccountsApi, _ := state.NewAccountsDBApiWithHistory(tpn.AccntState)

	argsAccountsRepo := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      finalAccountsApi,
		CurrentStateAccountsWrapper:    currentAccountsApi,
		HistoricalStateAccountsWrapper: historicalAccountsApi,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(argsAccountsRepo)

	networkComponents := GetDefaultNetworkComponents()
	networkComponents.Messenger = tpn.Messenger

	tpn.Node, err = node.NewNode(
		node.WithAddressSignatureSize(64),
		node.WithValidatorSignatureSize(48),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithStateComponents(stateComponents),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{}),
	)
	log.LogIfError(err)

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		tpn.Node,
		tpn.InterceptorsContainer,
		tpn.ResolverFinder,
		config.InterceptorResolverDebugConfig{
			Enabled:                    true,
			CacheSize:                  1000,
			EnablePrint:                true,
			IntervalAutoPrintInSeconds: 1,
			NumRequestsThreshold:       1,
			NumResolveFailureThreshold: 1,
			DebugLineExpiration:        1000,
		},
	)
	log.LogIfError(err)
}

// SendTransaction can send a transaction (it does the dispatching)
func (tpn *TestProcessorNode) SendTransaction(tx *dataTransaction.Transaction) (string, error) {
	tx, txHash, err := tpn.Node.CreateTransaction(
		tx.Nonce,
		tx.Value.String(),
		TestAddressPubkeyConverter.Encode(tx.RcvAddr),
		nil,
		TestAddressPubkeyConverter.Encode(tx.SndAddr),
		nil,
		tx.GasPrice,
		tx.GasLimit,
		tx.Data,
		hex.EncodeToString(tx.Signature),
		string(tx.ChainID),
		tx.Version,
		tx.Options,
	)
	if err != nil {
		return "", err
	}

	err = tpn.Node.ValidateTransaction(tx)
	if err != nil {
		return "", err
	}

	_, err = tpn.Node.SendBulkTransactions([]*dataTransaction.Transaction{tx})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(txHash), err
}

func (tpn *TestProcessorNode) addHandlersForCounters() {
	hdrHandlers := func(header data.HeaderHandler, key []byte) {
		atomic.AddInt32(&tpn.CounterHdrRecv, 1)
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.DataPool.Headers().RegisterHandler(hdrHandlers)
	} else {
		txHandler := func(key []byte, value interface{}) {
			tx, _ := tpn.DataPool.Transactions().SearchFirstData(key)
			tpn.ReceivedTransactions.Store(string(key), tx)
			atomic.AddInt32(&tpn.CounterTxRecv, 1)
		}
		mbHandlers := func(key []byte, value interface{}) {
			atomic.AddInt32(&tpn.CounterMbRecv, 1)
		}

		tpn.DataPool.UnsignedTransactions().RegisterOnAdded(txHandler)
		tpn.DataPool.Transactions().RegisterOnAdded(txHandler)
		tpn.DataPool.RewardTransactions().RegisterOnAdded(txHandler)
		tpn.DataPool.Headers().RegisterHandler(hdrHandlers)
		tpn.DataPool.MiniBlocks().RegisterHandler(mbHandlers, core.UniqueIdentifier())
	}
}

// StartSync calls Bootstrapper.StartSync. Errors if bootstrapper is not set
func (tpn *TestProcessorNode) StartSync() error {
	if tpn.Bootstrapper == nil || fmt.Sprintf("%T", tpn.Bootstrapper) == "*mock.testBootstrapperMock" {
		return errors.New("no bootstrapper available")
	}

	tpn.Bootstrapper.StartSyncingBlocks()

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

	blockHeader, err := tpn.BlockProcessor.CreateNewHeader(round, nonce)
	if err != nil {
		return nil, nil, nil
	}

	err = blockHeader.SetShardID(tpn.ShardCoordinator.SelfId())
	if err != nil {
		log.Warn("blockHeader.SetShardID", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetPubKeysBitmap([]byte{1})
	if err != nil {
		log.Warn("blockHeader.SetPubKeysBitmap", "error", err.Error())
		return nil, nil, nil
	}

	currHdr := tpn.BlockChain.GetCurrentBlockHeader()
	currHdrHash := tpn.BlockChain.GetCurrentBlockHeaderHash()
	if check.IfNil(currHdr) {
		currHdr = tpn.BlockChain.GetGenesisHeader()
		currHdrHash = tpn.BlockChain.GetGenesisHeaderHash()
	}

	err = blockHeader.SetPrevHash(currHdrHash)
	if err != nil {
		log.Warn("blockHeader.SetPrevHash", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	if err != nil {
		log.Warn("blockHeader.SetPrevRandSeed", "error", err.Error())
		return nil, nil, nil
	}
	sig := []byte("aggregated signature")
	err = blockHeader.SetSignature(sig)
	if err != nil {
		log.Warn("blockHeader.SetSignature", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetRandSeed(sig)
	if err != nil {
		log.Warn("blockHeader.SetRandSeed", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetLeaderSignature([]byte("leader sign"))
	if err != nil {
		log.Warn("blockHeader.SetLeaderSignature", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetChainID(tpn.ChainID)
	if err != nil {
		log.Warn("blockHeader.SetChainID", "error", err.Error())
		return nil, nil, nil
	}

	err = blockHeader.SetSoftwareVersion(SoftwareVersion)
	if err != nil {
		log.Warn("blockHeader.SetSoftwareVersion", "error", err.Error())
		return nil, nil, nil
	}

	genesisRound := tpn.BlockChain.GetGenesisHeader().GetRound()
	err = blockHeader.SetTimeStamp((round - genesisRound) * uint64(tpn.RoundHandler.TimeDuration().Seconds()))
	if err != nil {
		log.Warn("blockHeader.SetTimeStamp", "error", err.Error())
		return nil, nil, nil
	}

	blockHeader, blockBody, err := tpn.BlockProcessor.CreateBlock(blockHeader, haveTime)
	if err != nil {
		log.Warn("createBlockBody", "error", err.Error())
		return nil, nil, nil
	}

	shardBlockBody, ok := blockBody.(*dataBlock.Body)
	txHashes := make([][]byte, 0)
	if !ok {
		return blockBody, blockHeader, txHashes
	}

	for _, mb := range shardBlockBody.MiniBlocks {
		if mb.Type == dataBlock.PeerBlock {
			continue
		}
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

	time.Sleep(tpn.WaitTime)

	miniBlocks, transactions, _ := tpn.BlockProcessor.MarshalizedDataToBroadcast(header, body)
	_ = tpn.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = tpn.BroadcastMessenger.BroadcastTransactions(transactions)
}

// WhiteListBody will whitelist all miniblocks from the given body for all the given nodes
func (tpn *TestProcessorNode) WhiteListBody(nodes []*TestProcessorNode, bodyHandler data.BodyHandler) {
	body, ok := bodyHandler.(*dataBlock.Body)
	if !ok {
		return
	}

	mbHashes := make([][]byte, 0)
	txHashes := make([][]byte, 0)
	for _, miniBlock := range body.MiniBlocks {
		mbHash, err := core.CalculateHash(TestMarshalizer, TestHasher, miniBlock)
		if err != nil {
			continue
		}

		mbHashes = append(mbHashes, mbHash)
		txHashes = append(txHashes, miniBlock.TxHashes...)
	}

	if len(mbHashes) > 0 {
		for _, n := range nodes {
			n.WhiteListHandler.Add(mbHashes)
			n.WhiteListHandler.Add(txHashes)
		}
	}
}

// CommitBlock commits the block and body
func (tpn *TestProcessorNode) CommitBlock(body data.BodyHandler, header data.HeaderHandler) {
	_ = tpn.BlockProcessor.CommitBlock(header, body)
}

// GetShardHeader returns the first *dataBlock.Header stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetShardHeader(nonce uint64) (data.HeaderHandler, error) {
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.Headers() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	headerObjects, _, err := tpn.DataPool.Headers().GetHeadersByNonceAndShardId(nonce, tpn.ShardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("no headers found for nonce %d and shard id %d %s", nonce, tpn.ShardCoordinator.SelfId(), err.Error()))
	}

	headerObject := headerObjects[len(headerObjects)-1]

	header, ok := headerObject.(*dataBlock.Header)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.Header stored in headers found for nonce and shard id %d %d", nonce, tpn.ShardCoordinator.SelfId()))
	}

	return header, nil
}

// GetBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetBlockBody(header data.HeaderHandler) (*dataBlock.Body, error) {
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := &dataBlock.Body{}
	for _, miniBlockHeader := range header.GetMiniBlockHeaderHandlers() {
		miniBlockHash := miniBlockHeader.GetHash()

		mbObject, ok := tpn.DataPool.MiniBlocks().Get(miniBlockHash)
		if !ok {
			return nil, errors.New(fmt.Sprintf("no miniblock found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		mb, ok := mbObject.(*dataBlock.MiniBlock)
		if !ok {
			return nil, errors.New(fmt.Sprintf("not a *dataBlock.MiniBlock stored in miniblocks found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		body.MiniBlocks = append(body.MiniBlocks, mb)
	}

	return body, nil
}

// GetMetaBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetMetaBlockBody(header *dataBlock.MetaBlock) (*dataBlock.Body, error) {
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := &dataBlock.Body{}
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

		body.MiniBlocks = append(body.MiniBlocks, mb)
	}

	return body, nil
}

// GetMetaHeader returns the first *dataBlock.MetaBlock stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetMetaHeader(nonce uint64) (*dataBlock.MetaBlock, error) {
	invalidCachers := tpn.DataPool == nil || tpn.DataPool.Headers() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	headerObjects, _, err := tpn.DataPool.Headers().GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("no headers found for nonce and shard id %d %d %s", nonce, core.MetachainShardId, err.Error()))
	}

	headerObject := headerObjects[len(headerObjects)-1]

	header, ok := headerObject.(*dataBlock.MetaBlock)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.MetaBlock stored in headers found for nonce and shard id %d %d", nonce, core.MetachainShardId))
	}

	return header, nil
}

// SyncNode tries to process and commit a block already stored in data pool with provided nonce
func (tpn *TestProcessorNode) SyncNode(nonce uint64) error {
	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
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
		header,
		body,
		func() time.Duration {
			return time.Second * 5
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(header, body)
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
		header,
		body,
		func() time.Duration {
			return time.Second * 2
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(header, body)
	if err != nil {
		return err
	}

	return nil
}

// SetAccountNonce sets the account nonce with journal
func (tpn *TestProcessorNode) SetAccountNonce(nonce uint64) error {
	nodeAccount, _ := tpn.AccntState.LoadAccount(tpn.OwnAccount.Address)
	nodeAccount.(state.UserAccountHandler).IncreaseNonce(nonce)

	err := tpn.AccntState.SaveAccount(nodeAccount)
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

func (tpn *TestProcessorNode) initRoundHandler() {
	tpn.RoundHandler = &mock.RoundHandlerMock{TimeDurationField: 5 * time.Second}
}

func (tpn *TestProcessorNode) initRequestedItemsHandler() {
	tpn.RequestedItemsHandler = cache.NewTimeCache(roundDuration)
}

func (tpn *TestProcessorNode) initBlockTracker() {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:           TestHasher,
		HeaderValidator:  tpn.HeaderValidator,
		Marshalizer:      TestMarshalizer,
		RequestHandler:   tpn.RequestHandler,
		RoundHandler:     tpn.RoundHandler,
		ShardCoordinator: tpn.ShardCoordinator,
		Store:            tpn.Storage,
		StartHeaders:     tpn.GenesisBlocks,
		PoolsHolder:      tpn.DataPool,
		WhitelistHandler: tpn.WhiteListHandler,
		FeeHandler:       tpn.EconomicsData,
	}

	if tpn.ShardCoordinator.SelfId() != core.MetachainShardId {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		tpn.BlockTracker, _ = track.NewShardBlockTrack(arguments)
	} else {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
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

func (tpn *TestProcessorNode) createHeartbeatWithHardforkTrigger() {
	cacher := testscommon.NewCacherMock()
	psh, err := peerSignatureHandler.NewPeerSignatureHandler(
		cacher,
		tpn.OwnAccount.BlockSingleSigner,
		tpn.OwnAccount.KeygenBlockSign,
	)
	log.LogIfError(err)

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.PrivKey = tpn.NodeKeys.Sk
	cryptoComponents.PubKey = tpn.NodeKeys.Pk
	cryptoComponents.TxSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.BlockSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(tpn.MultiSigner)
	cryptoComponents.BlKeyGen = tpn.OwnAccount.KeygenTxSign
	cryptoComponents.TxKeyGen = TestKeyGenForAccounts
	cryptoComponents.PeerSignHandler = psh

	networkComponents := GetDefaultNetworkComponents()
	networkComponents.Messenger = tpn.Messenger
	networkComponents.InputAntiFlood = &mock.NilAntifloodHandler{}

	processComponents := GetDefaultProcessComponents()
	processComponents.BlockProcess = tpn.BlockProcessor
	processComponents.ResFinder = tpn.ResolverFinder
	processComponents.HeaderIntegrVerif = tpn.HeaderIntegrityVerifier
	processComponents.HeaderSigVerif = tpn.HeaderSigVerifier
	processComponents.BlackListHdl = tpn.BlockBlackListHandler
	processComponents.NodesCoord = tpn.NodesCoordinator
	processComponents.ShardCoord = tpn.ShardCoordinator
	processComponents.IntContainer = tpn.InterceptorsContainer
	processComponents.ValidatorStatistics = &mock.ValidatorStatisticsProcessorStub{
		GetValidatorInfoForRootHashCalled: func(_ []byte) (map[uint32][]*state.ValidatorInfo, error) {
			return map[uint32][]*state.ValidatorInfo{
				0: {{PublicKey: []byte("pk0")}},
			}, nil
		},
	}
	processComponents.ValidatorProvider = &mock.ValidatorsProviderStub{}
	processComponents.EpochTrigger = tpn.EpochStartTrigger
	processComponents.EpochNotifier = tpn.EpochStartNotifier
	processComponents.WhiteListerVerifiedTxsInternal = tpn.WhiteListerVerifiedTxs
	processComponents.WhiteListHandlerInternal = tpn.WhiteListHandler
	processComponents.HistoryRepositoryInternal = tpn.HistoryRepository
	processComponents.TxsSenderHandlerField = createTxsSender(tpn.ShardCoordinator, tpn.Messenger)

	processComponents.HardforkTriggerField = tpn.HardforkTrigger

	err = tpn.Node.ApplyOptions(
		node.WithCryptoComponents(cryptoComponents),
		node.WithProcessComponents(processComponents),
	)
	log.LogIfError(err)

	// TODO: remove it with heartbeat v1 cleanup
	// =============== Heartbeat ============== //
	redundancyHandler := &mock.RedundancyHandlerStub{}

	hbConfig := config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 4,
		MaxTimeToWaitBetweenBroadcastsInSec: 6,
		DurationToConsiderUnresponsiveInSec: 60,
		HeartbeatRefreshIntervalInSec:       5,
		HideInactiveValidatorIntervalInSec:  600,
	}

	hbFactoryArgs := heartbeatComp.HeartbeatComponentsFactoryArgs{
		Config: config.Config{
			Heartbeat: hbConfig,
		},
		Prefs:             config.Preferences{},
		RedundancyHandler: redundancyHandler,
		CoreComponents:    tpn.Node.GetCoreComponents(),
		DataComponents:    tpn.Node.GetDataComponents(),
		NetworkComponents: tpn.Node.GetNetworkComponents(),
		CryptoComponents:  tpn.Node.GetCryptoComponents(),
		ProcessComponents: tpn.Node.GetProcessComponents(),
	}

	heartbeatFactory, err := heartbeatComp.NewHeartbeatComponentsFactory(hbFactoryArgs)
	log.LogIfError(err)

	managedHeartbeatComponents, err := heartbeatComp.NewManagedHeartbeatComponents(heartbeatFactory)
	log.LogIfError(err)

	err = managedHeartbeatComponents.Create()
	log.LogIfError(err)

	err = tpn.Node.ApplyOptions(
		node.WithHeartbeatComponents(managedHeartbeatComponents),
	)
	log.LogIfError(err)

	// ============== HeartbeatV2 ============= //
	hbv2Config := config.HeartbeatV2Config{
		PeerAuthenticationTimeBetweenSendsInSec:          5,
		PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
		PeerAuthenticationThresholdBetweenSends:          0.1,
		HeartbeatTimeBetweenSendsInSec:                   2,
		HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
		HeartbeatThresholdBetweenSends:                   0.1,
		HeartbeatExpiryTimespanInSec:                     300,
		MinPeersThreshold:                                0.8,
		DelayBetweenRequestsInSec:                        10,
		MaxTimeoutInSec:                                  60,
		DelayBetweenConnectionNotificationsInSec:         5,
		MaxMissingKeysInRequest:                          100,
		MaxDurationPeerUnresponsiveInSec:                 10,
		HideInactiveValidatorIntervalInSec:               60,
		HardforkTimeBetweenSendsInSec:                    2,
		TimeBetweenConnectionsMetricsUpdateInSec:         10,
		HeartbeatPool: config.CacheConfig{
			Type:     "LRU",
			Capacity: 1000,
			Shards:   1,
		},
	}

	hbv2FactoryArgs := heartbeatComp.ArgHeartbeatV2ComponentsFactory{
		Config: config.Config{
			HeartbeatV2: hbv2Config,
			Hardfork: config.HardforkConfig{
				PublicKeyToListenFrom: hardforkPubKey,
			},
		},
		BoostrapComponents: tpn.Node.GetBootstrapComponents(),
		CoreComponents:     tpn.Node.GetCoreComponents(),
		DataComponents:     tpn.Node.GetDataComponents(),
		NetworkComponents:  tpn.Node.GetNetworkComponents(),
		CryptoComponents:   tpn.Node.GetCryptoComponents(),
		ProcessComponents:  tpn.Node.GetProcessComponents(),
	}

	heartbeatV2Factory, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(hbv2FactoryArgs)
	log.LogIfError(err)

	managedHeartbeatV2Components, err := heartbeatComp.NewManagedHeartbeatV2Components(heartbeatV2Factory)
	log.LogIfError(err)

	err = managedHeartbeatV2Components.Create()
	log.LogIfError(err)

	err = tpn.Node.ApplyOptions(
		node.WithHeartbeatV2Components(managedHeartbeatV2Components),
	)
	log.LogIfError(err)
}

// CreateEnableEpochsConfig creates enable epochs definitions to be used in tests
func CreateEnableEpochsConfig() config.EnableEpochs {
	return config.EnableEpochs{
		SCDeployEnableEpoch:                               UnreachableEpoch,
		BuiltInFunctionsEnableEpoch:                       0,
		RelayedTransactionsEnableEpoch:                    UnreachableEpoch,
		PenalizedTooMuchGasEnableEpoch:                    UnreachableEpoch,
		SwitchJailWaitingEnableEpoch:                      UnreachableEpoch,
		SwitchHysteresisForMinNodesEnableEpoch:            UnreachableEpoch,
		BelowSignedThresholdEnableEpoch:                   UnreachableEpoch,
		TransactionSignedWithTxHashEnableEpoch:            UnreachableEpoch,
		MetaProtectionEnableEpoch:                         UnreachableEpoch,
		AheadOfTimeGasUsageEnableEpoch:                    UnreachableEpoch,
		GasPriceModifierEnableEpoch:                       UnreachableEpoch,
		RepairCallbackEnableEpoch:                         UnreachableEpoch,
		BlockGasAndFeesReCheckEnableEpoch:                 UnreachableEpoch,
		StakingV2EnableEpoch:                              UnreachableEpoch,
		StakeEnableEpoch:                                  0,
		DoubleKeyProtectionEnableEpoch:                    0,
		ESDTEnableEpoch:                                   UnreachableEpoch,
		GovernanceEnableEpoch:                             UnreachableEpoch,
		DelegationManagerEnableEpoch:                      UnreachableEpoch,
		DelegationSmartContractEnableEpoch:                UnreachableEpoch,
		CorrectLastUnjailedEnableEpoch:                    UnreachableEpoch,
		BalanceWaitingListsEnableEpoch:                    UnreachableEpoch,
		ReturnDataToLastTransferEnableEpoch:               UnreachableEpoch,
		SenderInOutTransferEnableEpoch:                    UnreachableEpoch,
		RelayedTransactionsV2EnableEpoch:                  UnreachableEpoch,
		UnbondTokensV2EnableEpoch:                         UnreachableEpoch,
		SaveJailedAlwaysEnableEpoch:                       UnreachableEpoch,
		ValidatorToDelegationEnableEpoch:                  UnreachableEpoch,
		ReDelegateBelowMinCheckEnableEpoch:                UnreachableEpoch,
		WaitingListFixEnableEpoch:                         UnreachableEpoch,
		IncrementSCRNonceInMultiTransferEnableEpoch:       UnreachableEpoch,
		ESDTMultiTransferEnableEpoch:                      UnreachableEpoch,
		GlobalMintBurnDisableEpoch:                        UnreachableEpoch,
		ESDTTransferRoleEnableEpoch:                       UnreachableEpoch,
		BuiltInFunctionOnMetaEnableEpoch:                  UnreachableEpoch,
		ComputeRewardCheckpointEnableEpoch:                UnreachableEpoch,
		SCRSizeInvariantCheckEnableEpoch:                  UnreachableEpoch,
		BackwardCompSaveKeyValueEnableEpoch:               UnreachableEpoch,
		ESDTNFTCreateOnMultiShardEnableEpoch:              UnreachableEpoch,
		MetaESDTSetEnableEpoch:                            UnreachableEpoch,
		AddTokensToDelegationEnableEpoch:                  UnreachableEpoch,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       UnreachableEpoch,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       UnreachableEpoch,
		CorrectFirstQueuedEpoch:                           UnreachableEpoch,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           UnreachableEpoch,
		FixOOGReturnCodeEnableEpoch:                       UnreachableEpoch,
		RemoveNonUpdatedStorageEnableEpoch:                UnreachableEpoch,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       UnreachableEpoch,
		OptimizeNFTStoreEnableEpoch:                       UnreachableEpoch,
		CreateNFTThroughExecByCallerEnableEpoch:           UnreachableEpoch,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: UnreachableEpoch,
		FrontRunningProtectionEnableEpoch:                 UnreachableEpoch,
		IsPayableBySCEnableEpoch:                          UnreachableEpoch,
		CleanUpInformativeSCRsEnableEpoch:                 UnreachableEpoch,
		StorageAPICostOptimizationEnableEpoch:             UnreachableEpoch,
		TransformToMultiShardCreateEnableEpoch:            UnreachableEpoch,
		ESDTRegisterAndSetAllRolesEnableEpoch:             UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:                    UnreachableEpoch,
		FailExecutionOnEveryAPIErrorEnableEpoch:           UnreachableEpoch,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        UnreachableEpoch,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        UnreachableEpoch,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     UnreachableEpoch,
		HeartbeatDisableEpoch:                             UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:              UnreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:                UnreachableEpoch,
	}
}

// GetDefaultCoreComponents -
func GetDefaultCoreComponents() *mock.CoreComponentsStub {
	enableEpochsCfg := CreateEnableEpochsConfig()
	genericEpochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsCfg, genericEpochNotifier)

	return &mock.CoreComponentsStub{
		InternalMarshalizerField:      TestMarshalizer,
		TxMarshalizerField:            TestTxSignMarshalizer,
		VmMarshalizerField:            TestVmMarshalizer,
		HasherField:                   TestHasher,
		TxSignHasherField:             TestTxSignHasher,
		Uint64ByteSliceConverterField: TestUint64Converter,
		AddressPubKeyConverterField:   TestAddressPubkeyConverter,
		ValidatorPubKeyConverterField: TestValidatorPubkeyConverter,
		PathHandlerField:              &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return string(ChainID)
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		StatusHandlerField:           &statusHandlerMock.AppStatusHandlerStub{},
		WatchdogField:                &testscommon.WatchdogMock{},
		AlarmSchedulerField:          &testscommon.AlarmSchedulerStub{},
		SyncTimerField:               &testscommon.SyncTimerStub{},
		RoundHandlerField:            &testscommon.RoundHandlerMock{},
		EconomicsDataField:           &economicsmocks.EconomicsHandlerStub{},
		RatingsDataField:             &testscommon.RatingsInfoMock{},
		RaterField:                   &testscommon.RaterMock{},
		GenesisNodesSetupField:       &testscommon.NodesSetupStub{},
		GenesisTimeField:             time.Time{},
		EpochNotifierField:           genericEpochNotifier,
		EnableRoundsHandlerField:     &testscommon.EnableRoundsHandlerStub{},
		TxVersionCheckField:          versioning.NewTxVersionChecker(MinTransactionVersion),
		ProcessStatusHandlerInternal: &testscommon.ProcessStatusHandlerStub{},
		EnableEpochsHandlerField:     enableEpochsHandler,
	}
}

// GetDefaultProcessComponents -
func GetDefaultProcessComponents() *mock.ProcessComponentsStub {
	return &mock.ProcessComponentsStub{
		NodesCoord: &shardingMocks.NodesCoordinatorMock{},
		ShardCoord: &testscommon.ShardsCoordinatorMock{
			NoShards:     1,
			CurrentShard: 0,
		},
		IntContainer:             &testscommon.InterceptorsContainerStub{},
		ResFinder:                &mock.ResolversFinderStub{},
		RoundHandlerField:        &testscommon.RoundHandlerMock{},
		EpochTrigger:             &testscommon.EpochStartTriggerStub{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		ForkDetect:               &mock.ForkDetectorStub{},
		BlockProcess:             &mock.BlockProcessorMock{},
		BlackListHdl:             &testscommon.TimeCacheStub{},
		BootSore:                 &mock.BoostrapStorerMock{},
		HeaderSigVerif:           &mock.HeaderSigVerifierStub{},
		HeaderIntegrVerif:        &mock.HeaderIntegrityVerifierStub{},
		ValidatorStatistics:      &mock.ValidatorStatisticsProcessorStub{},
		ValidatorProvider:        &mock.ValidatorsProviderStub{},
		BlockTrack:               &mock.BlockTrackerStub{},
		PendingMiniBlocksHdl:     &mock.PendingMiniBlocksHandlerStub{},
		ReqHandler:               &testscommon.RequestHandlerStub{},
		TxLogsProcess:            &mock.TxLogProcessorMock{},
		HeaderConstructValidator: &mock.HeaderValidatorStub{},
		PeerMapper:               &p2pmocks.NetworkShardingCollectorStub{},
		FallbackHdrValidator:     &testscommon.FallBackHeaderValidatorStub{},
		NodeRedundancyHandlerInternal: &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
			ObserverPrivateKeyCalled: func() crypto.PrivateKey {
				return &mock.PrivateKeyMock{}
			},
		},
		CurrentEpochProviderInternal: &testscommon.CurrentEpochProviderStub{},
		HistoryRepositoryInternal:    &dblookupextMock.HistoryRepositoryStub{},
		HardforkTriggerField:         &testscommon.HardforkTriggerStub{},
	}
}

// GetDefaultDataComponents -
func GetDefaultDataComponents() *mock.DataComponentsStub {
	return &mock.DataComponentsStub{
		BlockChain: &testscommon.ChainHandlerMock{},
		Store:      &storageStubs.ChainStorerStub{},
		DataPool:   &dataRetrieverMock.PoolsHolderMock{},
		MbProvider: &mock.MiniBlocksProviderStub{},
	}
}

// GetDefaultCryptoComponents -
func GetDefaultCryptoComponents() *mock.CryptoComponentsStub {
	return &mock.CryptoComponentsStub{
		PubKey:            &mock.PublicKeyMock{},
		PrivKey:           &mock.PrivateKeyMock{},
		PubKeyString:      "pubKey",
		PrivKeyBytes:      []byte("privKey"),
		PubKeyBytes:       []byte("pubKey"),
		BlockSig:          &mock.SignerMock{},
		TxSig:             &mock.SignerMock{},
		MultiSigContainer: cryptoMocks.NewMultiSignerContainerMock(TestMultiSig),
		PeerSignHandler:   &mock.PeerSignatureHandler{},
		BlKeyGen:          &mock.KeyGenMock{},
		TxKeyGen:          &mock.KeyGenMock{},
		MsgSigVerifier:    &testscommon.MessageSignVerifierMock{},
	}
}

// GetDefaultStateComponents -
func GetDefaultStateComponents() *testscommon.StateComponentsMock {
	return &testscommon.StateComponentsMock{
		PeersAcc:     &stateMock.AccountsStub{},
		Accounts:     &stateMock.AccountsStub{},
		AccountsRepo: &stateMock.AccountsRepositoryStub{},
		Tries:        &trieMock.TriesHolderStub{},
		StorageManagers: map[string]common.StorageManager{
			"0":                         &testscommon.StorageManagerStub{},
			trieFactory.UserAccountTrie: &testscommon.StorageManagerStub{},
			trieFactory.PeerAccountTrie: &testscommon.StorageManagerStub{},
		},
	}
}

// GetDefaultNetworkComponents -
func GetDefaultNetworkComponents() *mock.NetworkComponentsStub {
	return &mock.NetworkComponentsStub{
		Messenger:               &p2pmocks.MessengerStub{},
		InputAntiFlood:          &mock.P2PAntifloodHandlerStub{},
		OutputAntiFlood:         &mock.P2PAntifloodHandlerStub{},
		PeerBlackList:           &mock.PeerBlackListCacherStub{},
		PeersRatingHandlerField: &p2pmocks.PeersRatingHandlerStub{},
	}
}

// GetDefaultStatusComponents -
func GetDefaultStatusComponents() *mock.StatusComponentsStub {
	return &mock.StatusComponentsStub{
		Outport:              mock.NewNilOutport(),
		SoftwareVersionCheck: &mock.SoftwareVersionCheckerMock{},
		AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
	}
}

// getDefaultBootstrapComponents -
func getDefaultBootstrapComponents(shardCoordinator sharding.Coordinator) *mainFactoryMocks.BootstrapComponentsStub {
	var versionedHeaderFactory nodeFactory.VersionedHeaderFactory

	headerVersionHandler := &testscommon.HeaderVersionHandlerStub{}
	versionedHeaderFactory, _ = hdrFactory.NewShardHeaderFactory(headerVersionHandler)
	if shardCoordinator.SelfId() == core.MetachainShardId {
		versionedHeaderFactory, _ = hdrFactory.NewMetaHeaderFactory(headerVersionHandler)
	}

	return &mainFactoryMocks.BootstrapComponentsStub{
		Bootstrapper: &bootstrapMocks.EpochStartBootstrapperStub{
			TrieHolder:      &trieMock.TriesHolderStub{},
			StorageManagers: map[string]common.StorageManager{"0": &testscommon.StorageManagerStub{}},
			BootstrapCalled: nil,
		},
		BootstrapParams:      &bootstrapMocks.BootstrapParamsHandlerMock{},
		NodeRole:             "",
		ShCoordinator:        shardCoordinator,
		HdrVersionHandler:    headerVersionHandler,
		VersionedHdrFactory:  versionedHeaderFactory,
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tpn *TestProcessorNode) IsInterfaceNil() bool {
	return tpn == nil
}

// GetTokenIdentifier returns the token identifier from the metachain for the given ticker
func GetTokenIdentifier(nodes []*TestProcessorNode, ticker []byte) []byte {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		acc, _ := n.AccntState.LoadAccount(vm.ESDTSCAddress)
		userAcc, _ := acc.(state.UserAccountHandler)

		rootHash, _ := userAcc.DataTrie().RootHash()
		chLeaves := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
		_ = userAcc.DataTrie().GetAllLeavesOnChannel(chLeaves, context.Background(), rootHash, keyBuilder.NewKeyBuilder())
		for leaf := range chLeaves {
			if !bytes.HasPrefix(leaf.Key(), ticker) {
				continue
			}

			return leaf.Key()
		}
	}

	return nil
}

func createTxsSender(shardCoordinator storage.ShardCoordinator, messenger txsSender.NetworkMessenger) process.TxsSenderHandler {
	txAccumulatorConfig := config.TxAccumulatorConfig{
		MaxAllowedTimeInMilliseconds:   10,
		MaxDeviationTimeInMilliseconds: 1,
	}
	dataPacker, err := partitioning.NewSimpleDataPacker(TestMarshalizer)
	log.LogIfError(err)

	argsTxsSender := txsSender.ArgsTxsSenderWithAccumulator{
		Marshaller:        TestMarshalizer,
		ShardCoordinator:  shardCoordinator,
		NetworkMessenger:  messenger,
		AccumulatorConfig: txAccumulatorConfig,
		DataPacker:        dataPacker,
	}
	txsSenderHandler, err := txsSender.NewTxsSenderWithAccumulator(argsTxsSender)
	log.LogIfError(err)

	return txsSenderHandler
}

func getDefaultVMConfig() *config.VirtualMachineConfig {
	return &config.VirtualMachineConfig{
		ArwenVersions: []config.ArwenVersionByEpoch{
			{StartEpoch: 0, Version: "*"},
		},
	}
}

func getDefaultNodesSetup(maxShards, numNodes uint32, address []byte, pksBytes map[uint32][]byte) sharding.GenesisNodesSetupHandler {
	return &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
			for i := uint32(0); i < maxShards; i++ {
				oneMap[i] = append(oneMap[i], mock.NewNodeInfo(address, pksBytes[i], i, InitialRating))
			}
			oneMap[core.MetachainShardId] = append(oneMap[core.MetachainShardId],
				mock.NewNodeInfo(address, pksBytes[core.MetachainShardId], core.MetachainShardId, InitialRating))
			return oneMap, nil
		},
		InitialNodesInfoForShardCalled: func(shardId uint32) (handlers []nodesCoordinator.GenesisNodeInfoHandler, handlers2 []nodesCoordinator.GenesisNodeInfoHandler, err error) {
			list := make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
			list = append(list, mock.NewNodeInfo(address, pksBytes[shardId], shardId, InitialRating))
			return list, nil, nil
		},
		MinNumberOfNodesCalled: func() uint32 {
			return numNodes
		},
	}
}

func getDefaultNodesCoordinator(maxShards uint32, pksBytes map[uint32][]byte) nodesCoordinator.NodesCoordinator {
	return &shardingMocks.NodesCoordinatorStub{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pksBytes[shardId], 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
		GetAllValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			keys := make(map[uint32][][]byte)
			for shardID := uint32(0); shardID < maxShards; shardID++ {
				keys[shardID] = append(keys[shardID], pksBytes[shardID])
			}

			shardID := core.MetachainShardId
			keys[shardID] = append(keys[shardID], pksBytes[shardID])

			return keys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validatorInstance, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validatorInstance, 0, nil
		},
	}
}

// GetDefaultEnableEpochsConfig returns a default EnableEpochs config
func GetDefaultEnableEpochsConfig() *config.EnableEpochs {
	return &config.EnableEpochs{
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        UnreachableEpoch,
		FailExecutionOnEveryAPIErrorEnableEpoch:     UnreachableEpoch,
	}
}
