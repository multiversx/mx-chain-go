package integrationTests

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/genesis"
	genesisProcess "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	procFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	txProc "github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	testStorage "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	testcommonStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	logger "github.com/multiversx/mx-chain-logger-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// StepDelay is used so that transactions can disseminate properly
var StepDelay = time.Millisecond * 180

// SyncDelay is used so that nodes have enough time to sync
var SyncDelay = time.Second / 5

// P2pBootstrapDelay is used so that nodes have enough time to bootstrap
var P2pBootstrapDelay = 5 * time.Second

// InitialRating is used to initiate a node's info
var InitialRating = uint32(50)

// AdditionalGasLimit is the value that can be added on a transaction in the GasLimit
var AdditionalGasLimit = uint64(999000)

// GasSchedulePath --
const GasSchedulePath = "../../../../cmd/node/config/gasSchedules/gasScheduleV4.toml"

var log = logger.GetOrCreate("integrationtests")

// shuffler constants
const (
	shuffleBetweenShards    = false
	adaptivity              = false
	hysteresis              = float32(0.2)
	maxTrieLevelInMemory    = uint(5)
	delegationManagementKey = "delegationManagement"
	delegationContractsList = "delegationContracts"
)

// Type defines account types to save in accounts trie
type Type uint8

const (
	// UserAccount identifies an account holding balance, storage updates, code
	UserAccount Type = 0
	// ValidatorAccount identifies an account holding stake, crypto public keys, assigned shard, rating
	ValidatorAccount Type = 1
)

const defaultChancesSelection = 1

// GetConnectableAddress returns a non circuit, non windows default connectable address for provided messenger
func GetConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}
		return addr
	}
	return ""
}

func createP2PConfig(initialPeerList []string) p2pConfig.P2PConfig {
	return p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: p2pConfig.P2PTCPTransport{
					ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
				},
			},
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled:                          true,
			Type:                             "optimized",
			RefreshIntervalInSec:             2,
			ProtocolID:                       "/erd/kad/1.0.0",
			InitialPeerList:                  initialPeerList,
			BucketSize:                       100,
			RoutingTableRefreshIntervalInSec: 100,
		},
		Sharding: p2pConfig.ShardingConfig{
			Type: p2p.NilListSharder,
		},
	}
}

// CreateMessengerWithKadDht creates a new libp2p messenger with kad-dht peer discovery
func CreateMessengerWithKadDht(initialAddr string) p2p.Messenger {
	initialAddresses := make([]string, 0)
	if len(initialAddr) > 0 {
		initialAddresses = append(initialAddresses, initialAddr)
	}
	arg := p2pFactory.ArgsNetworkMessenger{
		Marshaller:            TestMarshalizer,
		P2pConfig:             createP2PConfig(initialAddresses),
		SyncTimer:             &p2pFactory.LocalSyncTimer{},
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:    &p2pmocks.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         mock.NewPrivateKeyMock(),
		P2pSingleSigner:       &mock.SignerMock{},
		P2pKeyGenerator:       &mock.KeyGenMock{},
		Logger:                logger.GetOrCreate("tests/p2p"),
	}

	libP2PMes, err := p2pFactory.NewNetworkMessenger(arg)
	log.LogIfError(err)

	return libP2PMes
}

// CreateMessengerFromConfig creates a new libp2p messenger with provided configuration
func CreateMessengerFromConfig(p2pConfig p2pConfig.P2PConfig) p2p.Messenger {
	arg := p2pFactory.ArgsNetworkMessenger{
		Marshaller:            TestMarshalizer,
		P2pConfig:             p2pConfig,
		SyncTimer:             &p2pFactory.LocalSyncTimer{},
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:    &p2pmocks.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         mock.NewPrivateKeyMock(),
		P2pSingleSigner:       &mock.SignerMock{},
		P2pKeyGenerator:       &mock.KeyGenMock{},
		Logger:                logger.GetOrCreate("tests/p2p"),
	}

	libP2PMes, err := p2pFactory.NewNetworkMessenger(arg)
	log.LogIfError(err)

	return libP2PMes
}

// CreateMessengerFromConfigWithPeersRatingHandler creates a new libp2p messenger with provided configuration
func CreateMessengerFromConfigWithPeersRatingHandler(p2pConfig p2pConfig.P2PConfig, peersRatingHandler p2p.PeersRatingHandler, p2pKey crypto.PrivateKey) p2p.Messenger {
	arg := p2pFactory.ArgsNetworkMessenger{
		Marshaller:            TestMarshalizer,
		P2pConfig:             p2pConfig,
		SyncTimer:             &p2pFactory.LocalSyncTimer{},
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:    peersRatingHandler,
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		P2pPrivateKey:         p2pKey,
		P2pSingleSigner:       &mock.SignerMock{},
		P2pKeyGenerator:       &mock.KeyGenMock{},
		Logger:                logger.GetOrCreate("tests/p2p"),
	}

	libP2PMes, err := p2pFactory.NewNetworkMessenger(arg)
	log.LogIfError(err)

	return libP2PMes
}

// CreateP2PConfigWithNoDiscovery creates a new libp2p messenger with no peer discovery
func CreateP2PConfigWithNoDiscovery() p2pConfig.P2PConfig {
	return p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: p2pConfig.P2PTCPTransport{
					ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
				},
			},
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
		Sharding: p2pConfig.ShardingConfig{
			Type: p2p.NilListSharder,
		},
	}
}

// CreateMessengerWithNoDiscovery creates a new libp2p messenger with no peer discovery
func CreateMessengerWithNoDiscovery() p2p.Messenger {
	p2pCfg := CreateP2PConfigWithNoDiscovery()

	return CreateMessengerFromConfig(p2pCfg)
}

// CreateMessengerWithNoDiscoveryAndPeersRatingHandler creates a new libp2p messenger with no peer discovery
func CreateMessengerWithNoDiscoveryAndPeersRatingHandler(peersRatingHanlder p2p.PeersRatingHandler, p2pKey crypto.PrivateKey) p2p.Messenger {
	p2pCfg := p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: p2pConfig.P2PTCPTransport{
					ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
				},
			},
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled: false,
		},
		Sharding: p2pConfig.ShardingConfig{
			Type: p2p.NilListSharder,
		},
	}

	return CreateMessengerFromConfigWithPeersRatingHandler(p2pCfg, peersRatingHanlder, p2pKey)
}

// CreateFixedNetworkOf8Peers assembles a network as following:
//
//	                     0------------------- 1
//	                     |                    |
//	2 ------------------ 3 ------------------ 4
//	|                    |                    |
//	5                    6                    7
func CreateFixedNetworkOf8Peers() ([]p2p.Messenger, error) {
	peers := createMessengersWithNoDiscovery(8)

	connections := map[int][]int{
		0: {1, 3},
		1: {4},
		2: {5, 3},
		3: {4, 6},
		4: {7},
	}

	err := createConnections(peers, connections)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

// CreateFixedNetworkOf14Peers assembles a network as following:
//
//	               0
//	               |
//	               1
//	               |
//	+--+--+--+--+--2--+--+--+--+--+
//	|  |  |  |  |  |  |  |  |  |  |
//	3  4  5  6  7  8  9  10 11 12 13
func CreateFixedNetworkOf14Peers() ([]p2p.Messenger, error) {
	peers := createMessengersWithNoDiscovery(14)

	connections := map[int][]int{
		0: {1},
		1: {2},
		2: {3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
	}

	err := createConnections(peers, connections)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func createMessengersWithNoDiscovery(numPeers int) []p2p.Messenger {
	peers := make([]p2p.Messenger, numPeers)

	for i := 0; i < numPeers; i++ {
		peers[i] = CreateMessengerWithNoDiscovery()
	}

	return peers
}

func createConnections(peers []p2p.Messenger, connections map[int][]int) error {
	for pid, connectTo := range connections {
		err := connectPeerToOthers(peers, pid, connectTo)
		if err != nil {
			return err
		}
	}

	return nil
}

func connectPeerToOthers(peers []p2p.Messenger, idx int, connectToIdxes []int) error {
	for _, connectToIdx := range connectToIdxes {
		err := peers[idx].ConnectToPeer(peers[connectToIdx].Addresses()[0])
		if err != nil {
			return fmt.Errorf("%w connecting %s to %s", err, peers[idx].ID(), peers[connectToIdx].ID())
		}
	}

	return nil
}

// ClosePeers calls Messenger.Close on the provided peers
func ClosePeers(peers []p2p.Messenger) {
	for _, p := range peers {
		_ = p.Close()
	}
}

// CreateMemUnit returns an in-memory storer implementation (the vast majority of tests do not require effective
// disk I/O)
func CreateMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	persist, _ := database.NewlruDB(10000000)
	unit, _ := storageunit.NewStorageUnit(cache, persist)

	return unit
}

// CreateStore creates a storage service for shard nodes
func CreateStore(numOfShards uint32) dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.StatusMetricsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, CreateMemUnit())

	for i := uint32(0); i < numOfShards; i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, CreateMemUnit())
	}

	return store
}

// CreateTrieStorageManagerWithPruningStorer creates the trie storage manager for the tests
func CreateTrieStorageManagerWithPruningStorer(coordinator sharding.Coordinator, notifier pruning.EpochStartNotifier) common.StorageManager {
	mainStorer, _, err := testStorage.CreateTestingTriePruningStorer(coordinator, notifier)
	if err != nil {
		fmt.Println("err creating main storer" + err.Error())
	}
	checkpointsStorer, _, err := testStorage.CreateTestingTriePruningStorer(coordinator, notifier)
	if err != nil {
		fmt.Println("err creating checkpoints storer" + err.Error())
	}

	args := testcommonStorage.GetStorageManagerArgs()
	args.MainStorer = mainStorer
	args.CheckpointsStorer = checkpointsStorer
	args.Marshalizer = TestMarshalizer
	args.Hasher = TestHasher
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(10000000, uint64(TestHasher.Size()))

	trieStorageManager, _ := trie.NewTrieStorageManager(args)

	return trieStorageManager
}

// CreateTrieStorageManager creates the trie storage manager for the tests
func CreateTrieStorageManager(store storage.Storer) (common.StorageManager, storage.Storer) {
	args := testcommonStorage.GetStorageManagerArgs()
	args.MainStorer = store
	args.Marshalizer = TestMarshalizer
	args.Hasher = TestHasher
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(10000000, uint64(TestHasher.Size()))

	trieStorageManager, _ := trie.NewTrieStorageManager(args)

	return trieStorageManager, store
}

// CreateAccountsDB creates an account state with a valid trie implementation but with a memory storage
func CreateAccountsDB(
	accountType Type,
	trieStorageManager common.StorageManager,
) (*state.AccountsDB, common.Trie) {
	return CreateAccountsDBWithEnableEpochsHandler(accountType, trieStorageManager, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
}

// CreateAccountsDBWithEnableEpochsHandler creates a new AccountsDb with the given enableEpochsHandler
func CreateAccountsDBWithEnableEpochsHandler(
	accountType Type,
	trieStorageManager common.StorageManager,
	enableEpochsHandler common.EnableEpochsHandler,
) (*state.AccountsDB, common.Trie) {
	tr, _ := trie.NewTrie(trieStorageManager, TestMarshalizer, TestHasher, enableEpochsHandler, maxTrieLevelInMemory)

	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	accountFactory, _ := getAccountFactory(accountType, enableEpochsHandler)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)
	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                sha256.NewSha256(),
		Marshaller:            TestMarshalizer,
		AccountFactory:        accountFactory,
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
		AppStatusHandler:      &statusHandler.AppStatusHandlerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
	}
	adb, _ := state.NewAccountsDB(args)

	return adb, tr
}

func getAccountFactory(accountType Type, enableEpochsHandler common.EnableEpochsHandler) (state.AccountFactory, error) {
	switch accountType {
	case UserAccount:
		argsAccCreator := factory.ArgsAccountCreator{
			Hasher:              TestHasher,
			Marshaller:          TestMarshalizer,
			EnableEpochsHandler: enableEpochsHandler,
		}
		return factory.NewAccountCreator(argsAccCreator)
	case ValidatorAccount:
		return factory.NewPeerAccountCreator(), nil
	default:
		return nil, fmt.Errorf("invalid account type provided")
	}
}

// CreateShardChain creates a blockchain implementation used by the shard nodes
func CreateShardChain() data.ChainHandler {
	blockChain, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blockChain.SetGenesisHeader(&dataBlock.Header{})
	genesisHeaderM, _ := TestMarshalizer.Marshal(blockChain.GetGenesisHeader())

	blockChain.SetGenesisHeaderHash(TestHasher.Compute(string(genesisHeaderM)))

	return blockChain
}

// CreateMetaChain creates a blockchain implementation used by the meta nodes
func CreateMetaChain() data.ChainHandler {
	metaChain, _ := blockchain.NewMetaChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = metaChain.SetGenesisHeader(&dataBlock.MetaBlock{})
	genesisHeaderHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, metaChain.GetGenesisHeader())
	metaChain.SetGenesisHeaderHash(genesisHeaderHash)

	return metaChain
}

// CreateSimpleGenesisBlocks creates empty genesis blocks for all known shards, including metachain
func CreateSimpleGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = CreateSimpleGenesisBlock(shardId)
	}

	genesisBlocks[core.MetachainShardId] = CreateSimpleGenesisMetaBlock()

	return genesisBlocks
}

// CreateSimpleGenesisBlock creates a new mock shard genesis block
func CreateSimpleGenesisBlock(shardId uint32) *dataBlock.Header {
	rootHash := []byte("root hash")

	return &dataBlock.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardID:         shardId,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

// CreateSimpleGenesisMetaBlock creates a new mock meta genesis block
func CreateSimpleGenesisMetaBlock() *dataBlock.MetaBlock {
	rootHash := []byte("root hash")

	return &dataBlock.MetaBlock{
		Nonce:                  0,
		Epoch:                  0,
		Round:                  0,
		TimeStamp:              0,
		ShardInfo:              nil,
		Signature:              rootHash,
		PubKeysBitmap:          rootHash,
		PrevHash:               rootHash,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: rootHash,
		TxCount:                0,
		MiniBlockHeaders:       nil,
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
	}
}

// CreateGenesisBlocks creates empty genesis blocks for all known shards, including metachain
func CreateGenesisBlocks(
	accounts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	trieStorageManagers map[string]common.StorageManager,
	pubkeyConv core.PubkeyConverter,
	nodesSetup sharding.GenesisNodesSetupHandler,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	dataPool dataRetriever.PoolsHolder,
	economics process.EconomicsDataHandler,
	enableEpochsConfig config.EnableEpochs,
) map[uint32]data.HeaderHandler {

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = CreateSimpleGenesisBlock(shardId)
	}

	genesisBlocks[core.MetachainShardId] = CreateGenesisMetaBlock(
		accounts,
		validatorAccounts,
		trieStorageManagers,
		pubkeyConv,
		nodesSetup,
		shardCoordinator,
		store,
		blkc,
		marshalizer,
		hasher,
		uint64Converter,
		dataPool,
		economics,
		enableEpochsConfig,
	)

	return genesisBlocks
}

// CreateFullGenesisBlocks does the full genesis process, deploys smart contract at genesis
func CreateFullGenesisBlocks(
	accounts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	trieStorageManagers map[string]common.StorageManager,
	nodesSetup sharding.GenesisNodesSetupHandler,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	dataPool dataRetriever.PoolsHolder,
	economics process.EconomicsDataHandler,
	accountsParser genesis.AccountsParser,
	smartContractParser genesis.InitialSmartContractParser,
	enableEpochsConfig config.EnableEpochs,
) map[uint32]data.HeaderHandler {
	gasSchedule := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasSchedule, 1)

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.TxMarshalizerField = TestTxSignMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.AddressPubKeyConverterField = TestAddressPubkeyConverter
	coreComponents.ChainIdCalled = func() string {
		return "undefined"
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return 1
	}

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = store
	dataComponents.DataPool = dataPool
	dataComponents.BlockChain = blkc

	roundsConfig := GetDefaultRoundsConfig()
	runTypeComponents := mainFactoryMocks.NewRunTypeComponentsStub()
	runTypeComponents.BlockChainHookHandlerFactory, _ = hooks.NewBlockChainHookFactory()
	runTypeComponents.TransactionCoordinatorFactory, _ = coordinator.NewShardTransactionCoordinatorFactory()
	runTypeComponents.SCResultsPreProcessorFactory, _ = preprocess.NewSmartContractResultPreProcessorFactory()

	argsGenesis := genesisProcess.ArgsGenesisBlockCreator{
		Core:              coreComponents,
		Data:              dataComponents,
		GenesisTime:       0,
		StartEpochNum:     0,
		Accounts:          accounts,
		InitialNodesSetup: nodesSetup,
		Economics:         economics,
		ShardCoordinator:  shardCoordinator,
		ValidatorAccounts: validatorAccounts,
		GasSchedule:       mock.NewGasScheduleNotifierMock(gasSchedule),
		TxLogsProcessor:   &mock.TxLogsProcessorStub{},
		VirtualMachineConfig: config.VirtualMachineConfig{
			WasmVMVersions: []config.WasmVMVersionByEpoch{
				{StartEpoch: 0, Version: "*"},
			},
		},
		TrieStorageManagers: trieStorageManagers,
		SystemSCConfig: config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				OwnerAddress: DelegationManagerConfigChangeAddress,
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost: "500",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
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
				MaxServiceFee: 100,
			},
		},
		AccountsParser:      accountsParser,
		SmartContractParser: smartContractParser,
		BlockSignKeyGen:     &mock.KeyGenMock{},
		EpochConfig: &config.EpochConfig{
			EnableEpochs: enableEpochsConfig,
		},
		RoundConfig: &roundsConfig,
		RunType:     runTypeComponents,
	}

	genesisProcessor, _ := genesisProcess.NewGenesisBlockCreator(argsGenesis)
	genesisBlocks, _ := genesisProcessor.CreateGenesisBlocks()

	return genesisBlocks
}

// CreateGenesisMetaBlock creates a new mock meta genesis block
func CreateGenesisMetaBlock(
	accounts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	trieStorageManagers map[string]common.StorageManager,
	pubkeyConv core.PubkeyConverter,
	nodesSetup sharding.GenesisNodesSetupHandler,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	dataPool dataRetriever.PoolsHolder,
	economics process.EconomicsDataHandler,
	enableEpochsConfig config.EnableEpochs,
) data.MetaHeaderHandler {
	gasSchedule := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasSchedule, 1)

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = marshalizer
	coreComponents.HasherField = hasher
	coreComponents.Uint64ByteSliceConverterField = uint64Converter
	coreComponents.AddressPubKeyConverterField = pubkeyConv

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = store
	dataComponents.DataPool = dataPool
	dataComponents.BlockChain = blkc

	argsMetaGenesis := genesisProcess.ArgsGenesisBlockCreator{
		Core:                coreComponents,
		Data:                dataComponents,
		GenesisTime:         0,
		Accounts:            accounts,
		TrieStorageManagers: trieStorageManagers,
		InitialNodesSetup:   nodesSetup,
		ShardCoordinator:    shardCoordinator,
		Economics:           economics,
		ValidatorAccounts:   validatorAccounts,
		GasSchedule:         mock.NewGasScheduleNotifierMock(gasSchedule),
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
		VirtualMachineConfig: config.VirtualMachineConfig{
			WasmVMVersions: []config.WasmVMVersionByEpoch{
				{StartEpoch: 0, Version: "*"},
			},
		},
		HardForkConfig: config.HardforkConfig{},
		SystemSCConfig: config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost: "500",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: DelegationManagerConfigChangeAddress,
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
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
				MaxServiceFee: 100,
			},
		},
		BlockSignKeyGen:  &mock.KeyGenMock{},
		GenesisNodePrice: big.NewInt(1000),
		EpochConfig: &config.EpochConfig{
			EnableEpochs: enableEpochsConfig,
		},
		RunType: &mainFactoryMocks.RunTypeComponentsStub{},
	}

	if shardCoordinator.SelfId() != core.MetachainShardId {
		newShardCoordinator, _ := sharding.NewMultiShardCoordinator(
			shardCoordinator.NumberOfShards(),
			core.MetachainShardId,
		)

		newDataPool := dataRetrieverMock.CreatePoolsHolder(1, shardCoordinator.SelfId())

		newBlkc, _ := blockchain.NewMetaChain(&statusHandlerMock.AppStatusHandlerStub{})
		trieStorage, _ := CreateTrieStorageManager(CreateMemUnit())
		newAccounts, _ := CreateAccountsDBWithEnableEpochsHandler(UserAccount, trieStorage, coreComponents.EnableEpochsHandler())

		argsMetaGenesis.ShardCoordinator = newShardCoordinator
		argsMetaGenesis.Accounts = newAccounts

		_ = argsMetaGenesis.Data.SetBlockchain(newBlkc)
		dataComponents.DataPool = newDataPool
	}

	nodesHandler, err := mock.NewNodesHandlerMock(nodesSetup)
	log.LogIfError(err)

	metaHdr, _, _, err := genesisProcess.CreateMetaGenesisBlock(argsMetaGenesis, nil, nodesHandler, nil)
	log.LogIfError(err)

	log.Info("meta genesis root hash", "hash", hex.EncodeToString(metaHdr.GetRootHash()))
	log.Info("meta genesis validatorStatistics",
		"shardID", shardCoordinator.SelfId(),
		"hash", hex.EncodeToString(metaHdr.GetValidatorStatsRootHash()),
	)

	return metaHdr
}

// CreateRandomAddress creates a random byte array with fixed size
func CreateRandomAddress() []byte {
	return CreateRandomBytes(32)
}

// MintAddress will create an account (if it does not exist), update the balance with required value,
// save the account and commit the trie.
func MintAddress(accnts state.AccountsAdapter, addressBytes []byte, value *big.Int) {
	accnt, _ := accnts.LoadAccount(addressBytes)
	_ = accnt.(state.UserAccountHandler).AddToBalance(value)
	_ = accnts.SaveAccount(accnt)
	_, _ = accnts.Commit()
}

// CreateAccount creates a new account and returns the address
func CreateAccount(accnts state.AccountsAdapter, nonce uint64, balance *big.Int) []byte {
	address := CreateRandomBytes(32)
	account, _ := accnts.LoadAccount(address)
	account.(state.UserAccountHandler).IncreaseNonce(nonce)
	_ = account.(state.UserAccountHandler).AddToBalance(balance)
	_ = accnts.SaveAccount(account)

	return address
}

// MakeDisplayTable will output a string containing counters for received transactions, headers, miniblocks and
// meta headers for all provided test nodes
func MakeDisplayTable(nodes []*TestProcessorNode) string {
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))

	for idx, n := range nodes {
		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(n.OwnAccount.PkTxSignBytes),
				fmt.Sprintf("%d", n.ShardCoordinator.SelfId()),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterTxRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterMbRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterHdrRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterMetaRcv)),
				fmt.Sprintf("%d", len(n.MainMessenger.ConnectedPeers())),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

// PrintShardAccount outputs on console a shard account data contained
func PrintShardAccount(accnt state.UserAccountHandler, tag string) {
	str := fmt.Sprintf("%s Address: %s\n", tag, base64.StdEncoding.EncodeToString(accnt.AddressBytes()))
	str += fmt.Sprintf("  Nonce: %d\n", accnt.GetNonce())
	str += fmt.Sprintf("  Balance: %d\n", accnt.GetBalance().Uint64())
	str += fmt.Sprintf("  Code hash: %s\n", base64.StdEncoding.EncodeToString(accnt.GetCodeHash()))
	str += fmt.Sprintf("  Root hash: %s\n", base64.StdEncoding.EncodeToString(accnt.GetRootHash()))

	log.Info(str)
}

// CreateRandomBytes returns a random byte slice with the given size
func CreateRandomBytes(chars int) []byte {
	buff := make([]byte, chars)
	_, _ = rand.Reader.Read(buff)

	return buff
}

// GenerateAddressJournalAccountAccountsDB returns an account, the accounts address, and the accounts database
func GenerateAddressJournalAccountAccountsDB() ([]byte, state.UserAccountHandler, *state.AccountsDB) {
	adr := CreateRandomAddress()
	trieStorage, _ := CreateTrieStorageManager(CreateMemUnit())
	adb, _ := CreateAccountsDB(UserAccount, trieStorage)

	dtlp, _ := parsers.NewDataTrieLeafParser(adr, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	dtt, _ := trackableDataTrie.NewTrackableDataTrie(adr, &testscommon.HasherStub{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	account, _ := accounts.NewUserAccount(adr, dtt, dtlp)

	return adr, account, adb
}

// AdbEmulateBalanceTxSafeExecution emulates a tx execution by altering the accounts
// balance and nonce, and printing any encountered error
func AdbEmulateBalanceTxSafeExecution(acntSrc, acntDest state.UserAccountHandler, accounts state.AccountsAdapter, value *big.Int) {

	snapshot := accounts.JournalLen()
	err := AdbEmulateBalanceTxExecution(accounts, acntSrc, acntDest, value)

	if err != nil {
		log.Error("Error executing tx (value: %v), reverting...", value)
		err = accounts.RevertToSnapshot(snapshot)

		if err != nil {
			panic(err)
		}
	}
}

// AdbEmulateBalanceTxExecution emulates a tx execution by altering the accounts
// balance and nonce, and printing any encountered error
func AdbEmulateBalanceTxExecution(accounts state.AccountsAdapter, acntSrc, acntDest state.UserAccountHandler, value *big.Int) error {

	srcVal := acntSrc.GetBalance()
	if srcVal.Cmp(value) < 0 {
		return errors.New("not enough funds")
	}

	err := acntSrc.SubFromBalance(value)
	if err != nil {
		return err
	}

	err = acntDest.AddToBalance(value)
	if err != nil {
		return err
	}

	acntSrc.IncreaseNonce(1)

	err = accounts.SaveAccount(acntSrc)
	if err != nil {
		return err
	}

	err = accounts.SaveAccount(acntDest)
	if err != nil {
		return err
	}

	return nil
}

// CreateSimpleTxProcessor returns a transaction processor
func CreateSimpleTxProcessor(accnts state.AccountsAdapter) process.TransactionProcessor {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	argsNewTxProcessor := txProc.ArgsNewTxProcessor{
		Accounts:         accnts,
		Hasher:           TestHasher,
		PubkeyConv:       TestAddressPubkeyConverter,
		Marshalizer:      TestMarshalizer,
		SignMarshalizer:  TestTxSignMarshalizer,
		ShardCoordinator: shardCoordinator,
		ScProcessor:      &testscommon.SCProcessorMock{},
		TxFeeHandler:     &testscommon.UnsignedTxHandlerStub{},
		TxTypeHandler:    &testscommon.TxTypeHandlerMock{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
			CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
				return nil
			},
			ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				fee := big.NewInt(0).SetUint64(tx.GetGasLimit())
				fee.Mul(fee, big.NewInt(0).SetUint64(tx.GetGasPrice()))

				return fee
			},
		},
		ReceiptForwarder:    &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:      &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:          smartContract.NewArgumentParser(),
		ScrForwarder:        &mock.IntermediateTransactionHandlerMock{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxVersionChecker:    &testscommon.TxVersionCheckerStub{},
		GuardianChecker:     &guardianMocks.GuardedAccountHandlerStub{},
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
	}
	txProcessor, _ := txProc.NewTxProcessor(argsNewTxProcessor)

	return txProcessor
}

// CreateNewDefaultTrie returns a new trie with test hasher and marsahalizer
func CreateNewDefaultTrie() common.Trie {
	args := testcommonStorage.GetStorageManagerArgs()
	args.Marshalizer = TestMarshalizer
	args.Hasher = TestHasher
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(10000000, uint64(TestHasher.Size()))

	trieStorage, _ := trie.NewTrieStorageManager(args)

	tr, _ := trie.NewTrie(trieStorage, TestMarshalizer, TestHasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, maxTrieLevelInMemory)
	return tr
}

// GenerateRandomSlice returns a random byte slice with the given size
func GenerateRandomSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Reader.Read(buff)

	return buff
}

// MintAllNodes will take each shard node (n) and will mint all nodes that have their pk managed by the iterating node n
func MintAllNodes(nodes []*TestProcessorNode, value *big.Int) {
	for idx, n := range nodes {
		if n.ShardCoordinator.SelfId() == core.MetachainShardId {
			continue
		}

		mintAddressesFromSameShard(nodes, idx, value)
	}
}

func mintAddressesFromSameShard(nodes []*TestProcessorNode, targetNodeIdx int, value *big.Int) {
	targetNode := nodes[targetNodeIdx]

	for _, n := range nodes {
		shardId := targetNode.ShardCoordinator.ComputeId(n.OwnAccount.Address)
		if shardId != targetNode.ShardCoordinator.SelfId() {
			continue
		}

		n.OwnAccount.Balance = big.NewInt(0).Set(value)
		MintAddress(targetNode.AccntState, n.OwnAccount.Address, value)
	}
}

// MintAllPlayers mints addresses for all players
func MintAllPlayers(nodes []*TestProcessorNode, players []*TestWalletAccount, value *big.Int) {
	shardCoordinator := nodes[0].ShardCoordinator

	for _, player := range players {
		pShardId := shardCoordinator.ComputeId(player.Address)

		for _, n := range nodes {
			if pShardId != n.ShardCoordinator.SelfId() {
				continue
			}

			MintAddress(n.AccntState, player.Address, value)
			player.Balance = big.NewInt(0).Set(value)
		}
	}
}

// IncrementAndPrintRound increments the given variable, and prints the message for the beginning of the round
func IncrementAndPrintRound(round uint64) uint64 {
	round++
	log.Info(fmt.Sprintf("#################################### ROUND %d BEGINS ####################################", round))

	return round
}

// ProposeBlock proposes a block for every shard
func ProposeBlock(nodes []*TestProcessorNode, idxProposers []int, round uint64, nonce uint64) {
	log.Info("All shards propose blocks...")

	stepDelayAdjustment := StepDelay * time.Duration(1+len(nodes)/3)

	for idx, n := range nodes {
		if !IsIntInSlice(idx, idxProposers) {
			continue
		}

		body, header, _ := n.ProposeBlock(round, nonce)
		n.WhiteListBody(nodes, body)
		pk := n.NodeKeys.MainKey.Pk
		n.BroadcastBlock(body, header, pk)
		n.CommitBlock(body, header)
	}

	log.Info("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelayAdjustment)
	log.Info("Proposed block\n" + MakeDisplayTable(nodes))
}

// SyncBlock synchronizes the proposed block in all the other shard nodes
func SyncBlock(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxProposers []int,
	round uint64,
) {

	log.Info("All other shard nodes sync the proposed block...")
	for idx, n := range nodes {
		if IsIntInSlice(idx, idxProposers) {
			continue
		}

		err := n.SyncNode(round)
		if err != nil {
			log.Warn(fmt.Sprintf("SyncNode on round %v could not be synced. Error: %s", round, err.Error()))
			assert.Fail(t, err.Error())
			continue
		}
	}

	time.Sleep(StepDelay)
	log.Info("Synchronized block\n" + MakeDisplayTable(nodes))
}

// IsIntInSlice returns true if idx is found on any position in the provided slice
func IsIntInSlice(idx int, slice []int) bool {
	for _, value := range slice {
		if value == idx {
			return true
		}
	}

	return false
}

// Uint32InSlice checks if a uint32 value is in a slice
func Uint32InSlice(searched uint32, list []uint32) bool {
	for _, val := range list {
		if val == searched {
			return true
		}
	}
	return false
}

// CheckRootHashes checks the root hash of the proposer in every shard
func CheckRootHashes(t *testing.T, nodes []*TestProcessorNode, idxProposers []int) {
	for _, idx := range idxProposers {
		checkRootHashInShard(t, nodes, idx)
	}
}

func checkRootHashInShard(t *testing.T, nodes []*TestProcessorNode, idxProposer int) {
	proposerNode := nodes[idxProposer]
	proposerRootHash, _ := proposerNode.AccntState.RootHash()

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]

		if n.ShardCoordinator.SelfId() != proposerNode.ShardCoordinator.SelfId() {
			continue
		}

		log.Info(fmt.Sprintf("Testing roothash for node index %d, shard ID %d...", i, n.ShardCoordinator.SelfId()))
		nodeRootHash, _ := n.AccntState.RootHash()
		assert.Equal(t, proposerRootHash, nodeRootHash)
	}
}

// CheckTxPresentAndRightNonce verifies that the nonce was updated correctly after the exec of bulk txs
func CheckTxPresentAndRightNonce(
	t *testing.T,
	startingNonce uint64,
	noOfTxs int,
	txHashes [][]byte,
	txs []data.TransactionHandler,
	cache dataRetriever.ShardedDataCacherNotifier,
	shardCoordinator sharding.Coordinator,
) {

	if noOfTxs != len(txHashes) {
		for i := startingNonce; i < startingNonce+uint64(noOfTxs); i++ {
			found := false

			for _, txHandler := range txs {
				nonce := extractUint64ValueFromTxHandler(txHandler)
				if nonce == i {
					found = true
					break
				}
			}

			if !found {
				log.Info(fmt.Sprintf("unsigned tx with nonce %d is missing", i))
			}
		}
		assert.Fail(t, fmt.Sprintf("should have been %d, got %d", noOfTxs, len(txHashes)))

		return
	}

	bitmap := make([]bool, noOfTxs+int(startingNonce))
	// set for each nonce from found tx a true flag in bitmap
	for i := 0; i < noOfTxs; i++ {
		selfId := shardCoordinator.SelfId()
		shardDataStore := cache.ShardDataStore(process.ShardCacherIdentifier(selfId, selfId))
		val, _ := shardDataStore.Get(txHashes[i])
		if val == nil {
			continue
		}

		nonce := extractUint64ValueFromTxHandler(val.(data.TransactionHandler))
		bitmap[nonce] = true
	}

	// for the first startingNonce values, the bitmap should be false
	// for the rest, true
	for i := 0; i < noOfTxs+int(startingNonce); i++ {
		if i < int(startingNonce) {
			assert.False(t, bitmap[i])
			continue
		}

		assert.True(t, bitmap[i])
	}
}

func extractUint64ValueFromTxHandler(txHandler data.TransactionHandler) uint64 {
	tx, ok := txHandler.(*transaction.Transaction)
	if ok {
		return tx.Nonce
	}

	buff := txHandler.GetData()
	return binary.BigEndian.Uint64(buff)
}

// CreateHeaderIntegrityVerifier outputs a valid header integrity verifier handler
func CreateHeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	hvh := &testscommon.HeaderVersionHandlerStub{}

	headerVersioning, _ := headerCheck.NewHeaderIntegrityVerifier(
		ChainID,
		hvh,
	)

	return headerVersioning
}

// CreateNodes creates multiple nodes in different shards
func CreateNodes(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
) []*TestProcessorNode {
	return createNodesWithEpochsConfig(numOfShards, nodesPerShard, numMetaChainNodes, GetDefaultEnableEpochsConfig())
}

// CreateNodesWithEnableEpochsConfig creates multiple nodes in different shards but with custom enable epochs config
func CreateNodesWithEnableEpochsConfig(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	enableEpochsConfig *config.EnableEpochs,
) []*TestProcessorNode {
	return createNodesWithEpochsConfig(numOfShards, nodesPerShard, numMetaChainNodes, enableEpochsConfig)
}

func createNodesWithEpochsConfig(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	enableEpochsConfig *config.EnableEpochs,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)
	connectableNodes := make([]Connectable, len(nodes))

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(numOfShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				EpochsConfig:         enableEpochsConfig,
			})
			nodes[idx] = n
			connectableNodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            uint32(numOfShards),
			NodeShardId:          core.MetachainShardId,
			TxSignPrivKeyShardId: 0,
			EpochsConfig:         enableEpochsConfig,
		})
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
		connectableNodes[idx] = metaNode
	}

	ConnectNodes(connectableNodes)

	return nodes
}

// CreateNodesWithEnableEpochs creates multiple nodes with custom epoch config
func CreateNodesWithEnableEpochs(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	epochConfig config.EnableEpochs,
) []*TestProcessorNode {
	return CreateNodesWithEnableEpochsAndVmConfig(numOfShards, nodesPerShard, numMetaChainNodes, epochConfig, nil)
}

// CreateNodesWithEnableEpochsAndVmConfig creates multiple nodes with custom epoch and vm config
func CreateNodesWithEnableEpochsAndVmConfig(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	epochConfig config.EnableEpochs,
	vmConfig *config.VirtualMachineConfig,
) []*TestProcessorNode {
	return CreateNodesWithEnableEpochsAndVmConfigWithRoundsConfig(
		numOfShards,
		nodesPerShard,
		numMetaChainNodes,
		epochConfig,
		GetDefaultRoundsConfig(),
		vmConfig,
	)
}

// CreateNodesWithEnableEpochsAndVmConfigWithRoundsConfig creates multiple nodes with custom epoch and vm config
func CreateNodesWithEnableEpochsAndVmConfigWithRoundsConfig(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	epochConfig config.EnableEpochs,
	roundsConfig config.RoundConfig,
	vmConfig *config.VirtualMachineConfig,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)
	connectableNodes := make([]Connectable, len(nodes))

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(numOfShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				EpochsConfig:         &epochConfig,
				RoundsConfig:         &roundsConfig,
				VMConfig:             vmConfig,
			})
			nodes[idx] = n
			connectableNodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            uint32(numOfShards),
			NodeShardId:          core.MetachainShardId,
			TxSignPrivKeyShardId: 0,
			EpochsConfig:         &epochConfig,
			RoundsConfig:         &roundsConfig,
			VMConfig:             vmConfig,
		})
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
		connectableNodes[idx] = metaNode
	}

	ConnectNodes(connectableNodes)

	return nodes
}

// ConnectNodes will try to connect all provided connectable instances in a full mesh fashion
func ConnectNodes(nodes []Connectable) {
	encounteredErrors := make([]error, 0)

	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			src := nodes[i]
			dst := nodes[j]
			err := src.ConnectOnMain(dst)
			if err != nil {
				encounteredErrors = append(encounteredErrors,
					fmt.Errorf("%w while %s was connecting to %s", err, src.GetMainConnectableAddress(), dst.GetMainConnectableAddress()))
			}
		}
	}

	printEncounteredErrors(encounteredErrors)
}

func printEncounteredErrors(encounteredErrors []error) {
	if len(encounteredErrors) == 0 {
		return
	}

	printArguments := make([]interface{}, 0, len(encounteredErrors)*2)
	for i, err := range encounteredErrors {
		if err == nil {
			continue
		}

		printArguments = append(printArguments, fmt.Sprintf("err%d", i))
		printArguments = append(printArguments, err.Error())
	}

	log.Warn("errors encountered while connecting hosts", printArguments...)
}

// CreateNodesWithBLSSigVerifier creates multiple nodes in different shards
func CreateNodesWithBLSSigVerifier(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)
	connectableNodes := make([]Connectable, len(nodes))

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(numOfShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				WithBLSSigVerifier:   true,
			})
			nodes[idx] = n
			connectableNodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            uint32(numOfShards),
			NodeShardId:          core.MetachainShardId,
			TxSignPrivKeyShardId: 0,
			WithBLSSigVerifier:   true,
		})
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
		connectableNodes[idx] = metaNode
	}

	ConnectNodes(connectableNodes)

	return nodes
}

// CreateNodesWithFullGenesis creates multiple nodes in different shards
func CreateNodesWithFullGenesis(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	genesisFile string,
) ([]*TestProcessorNode, *TestProcessorNode) {
	enableEpochsConfig := GetDefaultEnableEpochsConfig()
	enableEpochsConfig.StakingV2EnableEpoch = UnreachableEpoch
	return CreateNodesWithFullGenesisCustomEnableEpochs(numOfShards, nodesPerShard, numMetaChainNodes, genesisFile, enableEpochsConfig)
}

// CreateNodesWithFullGenesisCustomEnableEpochs creates multiple nodes in different shards
func CreateNodesWithFullGenesisCustomEnableEpochs(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	genesisFile string,
	enableEpochsConfig *config.EnableEpochs,
) ([]*TestProcessorNode, *TestProcessorNode) {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)
	connectableNodes := make([]Connectable, len(nodes))

	economicsConfig := createDefaultEconomicsConfig()
	economicsConfig.GlobalSettings.YearSettings = append(
		economicsConfig.GlobalSettings.YearSettings,
		&config.YearSetting{
			Year:             1,
			MaximumInflation: 0.01,
		},
	)

	hardforkStarter := NewTestProcessorNode(ArgTestProcessorNode{
		MaxShards:            uint32(numOfShards),
		NodeShardId:          0,
		TxSignPrivKeyShardId: 0,
		GenesisFile:          genesisFile,
		EpochsConfig:         enableEpochsConfig,
		EconomicsConfig:      economicsConfig,
	})

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			nodes[idx] = NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(numOfShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				GenesisFile:          genesisFile,
				HardforkPk:           hardforkStarter.NodeKeys.MainKey.Pk,
				EpochsConfig:         enableEpochsConfig,
				EconomicsConfig:      economicsConfig,
			})
			connectableNodes[idx] = nodes[idx]
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            uint32(numOfShards),
			NodeShardId:          core.MetachainShardId,
			TxSignPrivKeyShardId: 0,
			GenesisFile:          genesisFile,
			HardforkPk:           hardforkStarter.NodeKeys.MainKey.Pk,
			EpochsConfig:         enableEpochsConfig,
			EconomicsConfig:      economicsConfig,
		})
		connectableNodes[idx] = nodes[idx]
	}

	connectableNodes = append(connectableNodes, hardforkStarter)
	ConnectNodes(connectableNodes)

	return nodes, hardforkStarter
}

// CreateNodesWithCustomStateCheckpointModulus creates multiple nodes in different shards with custom stateCheckpointModulus
func CreateNodesWithCustomStateCheckpointModulus(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	stateCheckpointModulus uint,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)
	connectableNodes := make([]Connectable, len(nodes))

	enableEpochsConfig := GetDefaultEnableEpochsConfig()
	enableEpochsConfig.StakingV2EnableEpoch = UnreachableEpoch

	scm := &IntWrapper{
		Value: stateCheckpointModulus,
	}

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:              uint32(numOfShards),
				NodeShardId:            shardId,
				TxSignPrivKeyShardId:   shardId,
				StateCheckpointModulus: scm,
				EpochsConfig:           enableEpochsConfig,
			})

			nodes[idx] = n
			connectableNodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:              uint32(numOfShards),
			NodeShardId:            core.MetachainShardId,
			TxSignPrivKeyShardId:   0,
			StateCheckpointModulus: scm,
			EpochsConfig:           enableEpochsConfig,
		})
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
		connectableNodes[idx] = metaNode
	}

	ConnectNodes(connectableNodes)

	return nodes
}

// DisplayAndStartNodes prints each nodes shard ID, sk and pk, and then starts the node
func DisplayAndStartNodes(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		skTxBuff, _ := n.OwnAccount.SkTxSign.ToByteArray()
		pkTxBuff, _ := n.OwnAccount.PkTxSign.ToByteArray()
		pkNode := n.NodesCoordinator.GetOwnPublicKey()

		encodedPkNode := TestValidatorPubkeyConverter.SilentEncode(pkNode, log)

		log.Info(fmt.Sprintf("Shard ID: %v, pkNode: %s",
			n.ShardCoordinator.SelfId(),
			encodedPkNode))

		encodedPkTxBuff := TestAddressPubkeyConverter.SilentEncode(pkTxBuff, log)

		log.Info(fmt.Sprintf("skTx: %s, pkTx: %s",
			hex.EncodeToString(skTxBuff),
			encodedPkTxBuff))
	}

	log.Info("Delaying for node bootstrap and topic announcement...")
	time.Sleep(P2pBootstrapDelay)
}

// SetEconomicsParameters will set maxGasLimitPerBlock, minGasPrice and minGasLimits to provided nodes
func SetEconomicsParameters(nodes []*TestProcessorNode, maxGasLimitPerBlock uint64, minGasPrice uint64, minGasLimit uint64) {
	for _, n := range nodes {
		n.EconomicsData.SetMaxGasLimitPerBlock(maxGasLimitPerBlock)
		n.EconomicsData.SetMinGasPrice(minGasPrice)
		n.EconomicsData.SetMinGasLimit(minGasLimit)
	}
}

// GenerateAndDisseminateTxs generates and sends multiple txs
func GenerateAndDisseminateTxs(
	n *TestProcessorNode,
	senders []crypto.PrivateKey,
	receiversPublicKeysMap map[uint32][]crypto.PublicKey,
	valToTransfer *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	chainID []byte,
	version uint32,
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := make([]uint64, len(senders))
		for _, shardReceiversPublicKeys := range receiversPublicKeysMap {
			receiverPubKey := shardReceiversPublicKeys[i]
			tx := GenerateTransferTx(incrementalNonce[i], senderKey, receiverPubKey, valToTransfer, gasPrice, gasLimit, chainID, version)
			_, _ = n.SendTransaction(tx)
			incrementalNonce[i]++
		}
	}
}

// CreateSendersWithInitialBalances creates a map of 1 sender per shard with an initial balance
func CreateSendersWithInitialBalances(
	nodesMap map[uint32][]*TestProcessorNode,
	mintValue *big.Int,
) map[uint32][]crypto.PrivateKey {

	sendersPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	for shardId, nodes := range nodesMap {
		if shardId == core.MetachainShardId {
			continue
		}

		sendersPrivateKeys[shardId], _ = CreateSendersAndReceiversInShard(
			nodes[0],
			1,
		)

		log.Info("Minting sender addresses...")
		CreateMintingForSenders(
			nodes,
			shardId,
			sendersPrivateKeys[shardId],
			mintValue,
		)
	}

	return sendersPrivateKeys
}

// CreateAndSendTransaction will generate a transaction with provided parameters, sign it with the provided
// node's tx sign private key and send it on the transaction topic using the correct node that can send the transaction
func CreateAndSendTransaction(
	node *TestProcessorNode,
	nodes []*TestProcessorNode,
	txValue *big.Int,
	rcvAddress []byte,
	txData string,
	additionalGasLimit uint64,
) {
	CreateAndSendTransactionWithSenderAccount(
		node,
		nodes,
		txValue,
		node.OwnAccount,
		rcvAddress,
		txData,
		additionalGasLimit)
}

// CreateAndSendTransactionWithSenderAccount will generate a transaction with provided parameters, sign it with the provided
// node's tx sign private key and send it on the transaction topic using the correct node that can send the transaction
func CreateAndSendTransactionWithSenderAccount(
	node *TestProcessorNode,
	nodes []*TestProcessorNode,
	txValue *big.Int,
	senderAccount *TestWalletAccount,
	rcvAddress []byte,
	txData string,
	additionalGasLimit uint64,
) {
	tx := &transaction.Transaction{
		Nonce:    senderAccount.Nonce,
		Value:    new(big.Int).Set(txValue),
		SndAddr:  senderAccount.Address,
		RcvAddr:  rcvAddress,
		Data:     []byte(txData),
		GasPrice: MinTxGasPrice,
		GasLimit: MinTxGasLimit + uint64(len(txData)) + additionalGasLimit,
		ChainID:  ChainID,
		Version:  MinTransactionVersion,
	}

	txBuff, _ := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer, TestTxSignHasher)
	tx.Signature, _ = senderAccount.SingleSigner.Sign(senderAccount.SkTxSign, txBuff)
	senderShardID := node.ShardCoordinator.ComputeId(senderAccount.Address)

	wasSent := false
	for _, senderNode := range nodes {
		if senderNode.ShardCoordinator.SelfId() != senderShardID {
			continue
		}

		_, err := senderNode.SendTransaction(tx)
		if err != nil {
			log.Error("could not send transaction", "address", senderAccount.Address, "error", err)
		} else {
			wasSent = true
		}
		break
	}

	if !wasSent {
		log.Error("no suitable node found to send the provided transaction", "address", senderAccount.Address)
	}
	senderAccount.Nonce++
}

// CreateAndSendTransactionWithGasLimit generates and send a transaction with provided gas limit/gas price
func CreateAndSendTransactionWithGasLimit(
	node *TestProcessorNode,
	txValue *big.Int,
	gasLimit uint64,
	rcvAddress []byte,
	txData []byte,
	chainID []byte,
	version uint32,
) {
	tx := &transaction.Transaction{
		Nonce:    node.OwnAccount.Nonce,
		Value:    txValue,
		SndAddr:  node.OwnAccount.Address,
		RcvAddr:  rcvAddress,
		Data:     txData,
		GasPrice: MinTxGasPrice,
		GasLimit: gasLimit,
		ChainID:  chainID,
		Version:  version,
	}

	txBuff, _ := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer, TestTxSignHasher)
	tx.Signature, _ = node.OwnAccount.SingleSigner.Sign(node.OwnAccount.SkTxSign, txBuff)

	_, _ = node.SendTransaction(tx)
	node.OwnAccount.Nonce++
}

type txArgs struct {
	nonce    uint64
	value    *big.Int
	rcvAddr  []byte
	sndAddr  []byte
	data     string
	gasPrice uint64
	gasLimit uint64
}

// GenerateTransferTx will generate a move balance transaction
func GenerateTransferTx(
	nonce uint64,
	senderPrivateKey crypto.PrivateKey,
	receiverPublicKey crypto.PublicKey,
	valToTransfer *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	chainID []byte,
	version uint32,
) *transaction.Transaction {

	receiverPubKeyBytes, _ := receiverPublicKey.ToByteArray()
	tx := transaction.Transaction{
		Nonce:    nonce,
		Value:    new(big.Int).Set(valToTransfer),
		RcvAddr:  receiverPubKeyBytes,
		SndAddr:  skToPk(senderPrivateKey),
		Data:     []byte(""),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		ChainID:  chainID,
		Version:  version,
	}
	txBuff, _ := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer, TestTxSignHasher)
	signer := TestSingleSigner
	tx.Signature, _ = signer.Sign(senderPrivateKey, txBuff)

	return &tx
}

func generateTx(
	skSign crypto.PrivateKey,
	signer crypto.SingleSigner,
	args *txArgs,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    args.nonce,
		Value:    new(big.Int).Set(args.value),
		RcvAddr:  args.rcvAddr,
		SndAddr:  args.sndAddr,
		GasPrice: args.gasPrice,
		GasLimit: args.gasLimit,
		Data:     []byte(args.data),
		ChainID:  ChainID,
		Version:  MinTransactionVersion,
	}
	txBuff, _ := tx.GetDataForSigning(TestAddressPubkeyConverter, TestTxSignMarshalizer, TestTxSignHasher)
	tx.Signature, _ = signer.Sign(skSign, txBuff)

	return tx
}

func skToPk(sk crypto.PrivateKey) []byte {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	return pkBuff
}

// TestPublicKeyHasBalance checks if the account corresponding to the given public key has the expected balance
func TestPublicKeyHasBalance(t *testing.T, n *TestProcessorNode, pk crypto.PublicKey, expectedBalance *big.Int) {
	pkBuff, _ := pk.ToByteArray()
	account, _ := n.AccntState.GetExistingAccount(pkBuff)
	assert.Equal(t, expectedBalance, account.(state.UserAccountHandler).GetBalance())
}

// TestPrivateKeyHasBalance checks if the private key has the expected balance
func TestPrivateKeyHasBalance(t *testing.T, n *TestProcessorNode, sk crypto.PrivateKey, expectedBalance *big.Int) {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	account, _ := n.AccntState.GetExistingAccount(pkBuff)
	assert.Equal(t, expectedBalance, account.(state.UserAccountHandler).GetBalance())
}

// GetMiniBlocksHashesFromShardIds returns miniblock hashes from body
func GetMiniBlocksHashesFromShardIds(body *dataBlock.Body, shardIds ...uint32) [][]byte {
	var hashes [][]byte

	for _, miniblock := range body.MiniBlocks {
		for _, shardId := range shardIds {
			if miniblock.ReceiverShardID == shardId {
				buff, _ := TestMarshalizer.Marshal(miniblock)
				hashes = append(hashes, TestHasher.Compute(string(buff)))
			}
		}
	}

	return hashes
}

// GenerateIntraShardTransactions generates intra shard transactions
func GenerateIntraShardTransactions(
	nodesMap map[uint32][]*TestProcessorNode,
	nbTxsPerShard uint32,
	mintValue *big.Int,
	valToTransfer *big.Int,
	gasPrice uint64,
	gasLimit uint64,
) {
	sendersPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)

	for shardId, nodes := range nodesMap {
		if shardId == core.MetachainShardId {
			continue
		}

		sendersPrivateKeys[shardId], receiversPublicKeys[shardId] = CreateSendersAndReceiversInShard(
			nodes[0],
			nbTxsPerShard,
		)

		log.Info("Minting sender addresses...")
		CreateMintingForSenders(
			nodes,
			shardId,
			sendersPrivateKeys[shardId],
			mintValue,
		)
	}

	CreateAndSendTransactions(
		nodesMap,
		sendersPrivateKeys,
		receiversPublicKeys,
		gasPrice,
		gasLimit,
		valToTransfer,
	)
}

// GenerateSkAndPkInShard generates and returns a private and a public key that reside in a given shard.
// It also returns the key generator
func GenerateSkAndPkInShard(
	coordinator sharding.Coordinator,
	shardId uint32,
) (crypto.PrivateKey, crypto.PublicKey, crypto.KeyGenerator) {
	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	if shardId == core.MetachainShardId {
		// for metachain generate in shard 0
		shardId = 0
	}

	for {
		pkBytes, _ := pk.ToByteArray()
		if coordinator.ComputeId(pkBytes) == shardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	return sk, pk, keyGen
}

// CreateSendersAndReceiversInShard creates given number of sender private key and receiver public key pairs,
// with account in same shard as given node
func CreateSendersAndReceiversInShard(
	nodeInShard *TestProcessorNode,
	nbSenderReceiverPairs uint32,
) ([]crypto.PrivateKey, []crypto.PublicKey) {
	shardId := nodeInShard.ShardCoordinator.SelfId()
	receiversPublicKeys := make([]crypto.PublicKey, nbSenderReceiverPairs)
	sendersPrivateKeys := make([]crypto.PrivateKey, nbSenderReceiverPairs)

	for i := uint32(0); i < nbSenderReceiverPairs; i++ {
		sendersPrivateKeys[i], _, _ = GenerateSkAndPkInShard(nodeInShard.ShardCoordinator, shardId)
		_, receiversPublicKeys[i], _ = GenerateSkAndPkInShard(nodeInShard.ShardCoordinator, shardId)
	}

	return sendersPrivateKeys, receiversPublicKeys
}

// CreateAndSendTransactions creates and sends transactions between given senders and receivers.
func CreateAndSendTransactions(
	nodes map[uint32][]*TestProcessorNode,
	sendersPrivKeysMap map[uint32][]crypto.PrivateKey,
	receiversPubKeysMap map[uint32][]crypto.PublicKey,
	gasPricePerTx uint64,
	gasLimitPerTx uint64,
	valueToTransfer *big.Int,
) {
	for shardId := range nodes {
		if shardId == core.MetachainShardId {
			continue
		}

		nodeInShard := nodes[shardId][0]

		log.Info("Generating transactions...")
		GenerateAndDisseminateTxs(
			nodeInShard,
			sendersPrivKeysMap[shardId],
			receiversPubKeysMap,
			valueToTransfer,
			gasPricePerTx,
			gasLimitPerTx,
			ChainID,
			MinTransactionVersion,
		)
	}

	log.Info("Delaying for disseminating transactions...")
	time.Sleep(time.Second)
}

// CreateMintingForSenders creates account with balances for every node in a given shard
func CreateMintingForSenders(
	nodes []*TestProcessorNode,
	senderShard uint32,
	sendersPrivateKeys []crypto.PrivateKey,
	value *big.Int,
) {

	for _, n := range nodes {
		// only sender shard nodes will be minted
		if n.ShardCoordinator.SelfId() != senderShard {
			continue
		}

		for _, sk := range sendersPrivateKeys {
			pkBuff, _ := sk.GeneratePublic().ToByteArray()
			account, _ := n.AccntState.LoadAccount(pkBuff)
			_ = account.(state.UserAccountHandler).AddToBalance(value)
			_ = n.AccntState.SaveAccount(account)
		}

		_, _ = n.AccntState.Commit()
	}
}

// CreateMintingFromAddresses creates account with balances for given address
func CreateMintingFromAddresses(
	nodes []*TestProcessorNode,
	addresses [][]byte,
	value *big.Int,
) {
	for _, n := range nodes {
		for _, address := range addresses {
			MintAddress(n.AccntState, address, value)
		}
	}
}

// ProposeBlockSignalsEmptyBlock proposes and broadcasts a block
func ProposeBlockSignalsEmptyBlock(
	node *TestProcessorNode,
	round uint64,
	nonce uint64,
) (data.HeaderHandler, data.BodyHandler, bool) {

	log.Info("Proposing block without commit...")

	body, header, txHashes := node.ProposeBlock(round, nonce)
	pk := node.NodeKeys.MainKey.Pk
	node.BroadcastBlock(body, header, pk)
	isEmptyBlock := len(txHashes) == 0

	log.Info("Delaying for disseminating headers and miniblocks...")
	time.Sleep(StepDelay)

	return header, body, isEmptyBlock
}

// CreateAccountForNodes creates accounts for each node and commits the accounts state
func CreateAccountForNodes(nodes []*TestProcessorNode) {
	for i := 0; i < len(nodes); i++ {
		CreateAccountForNode(nodes[i])
	}
}

// CreateAccountForNode creates an account for the given node
func CreateAccountForNode(node *TestProcessorNode) {
	acc, _ := node.AccntState.LoadAccount(node.OwnAccount.PkTxSignBytes)
	_ = node.AccntState.SaveAccount(acc)
	_, _ = node.AccntState.Commit()
}

// ComputeAndRequestMissingTransactions computes missing transactions for each node, and requests them
func ComputeAndRequestMissingTransactions(
	nodes []*TestProcessorNode,
	generatedTxHashes [][]byte,
	shardResolver uint32,
	shardRequesters ...uint32,
) {
	for _, n := range nodes {
		if !Uint32InSlice(n.ShardCoordinator.SelfId(), shardRequesters) {
			continue
		}

		neededTxs := getMissingTxsForNode(n, generatedTxHashes)
		requestMissingTransactions(n, shardResolver, neededTxs)
	}
}

func getMissingTxsForNode(n *TestProcessorNode, generatedTxHashes [][]byte) [][]byte {
	var neededTxs [][]byte

	for i := 0; i < len(generatedTxHashes); i++ {
		_, ok := n.DataPool.Transactions().SearchFirstData(generatedTxHashes[i])
		if !ok {
			neededTxs = append(neededTxs, generatedTxHashes[i])
		}
	}

	return neededTxs
}

func requestMissingTransactions(n *TestProcessorNode, shardResolver uint32, neededTxs [][]byte) {
	txResolver, _ := n.RequestersFinder.CrossShardRequester(procFactory.TransactionTopic, shardResolver)

	for i := 0; i < len(neededTxs); i++ {
		_ = txResolver.RequestDataFromHash(neededTxs[i], 0)
	}
}

// CreateRequesterDataPool creates a datapool with a mock txPool
func CreateRequesterDataPool(recvTxs map[int]map[string]struct{}, mutRecvTxs *sync.Mutex, nodeIndex int, _ uint32) dataRetriever.PoolsHolder {
	// not allowed requesting data from the same shard
	return dataRetrieverMock.CreatePoolsHolderWithTxPool(&testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			return nil
		},
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			mutRecvTxs.Lock()
			defer mutRecvTxs.Unlock()

			txMap := recvTxs[nodeIndex]
			if txMap == nil {
				txMap = make(map[string]struct{})
				recvTxs[nodeIndex] = txMap
			}

			txMap[string(key)] = struct{}{}
		},
	})
}

// CreateResolversDataPool creates a datapool containing a given number of transactions
func CreateResolversDataPool(
	t *testing.T,
	maxTxs int,
	senderShardID uint32,
	recvShardId uint32,
	shardCoordinator sharding.Coordinator,
) (dataRetriever.PoolsHolder, [][]byte, [][]byte) {

	txHashes := make([][]byte, maxTxs)
	txsSndAddr := make([][]byte, 0)
	poolsHolder := dataRetrieverMock.CreatePoolsHolder(1, shardCoordinator.SelfId())
	txPool := poolsHolder.Transactions()

	for i := 0; i < maxTxs; i++ {
		tx, txHash := generateValidTx(t, shardCoordinator, senderShardID, recvShardId)
		cacherIdentifier := process.ShardCacherIdentifier(1, 0)
		txPool.AddData(txHash, tx, tx.Size(), cacherIdentifier)
		txHashes[i] = txHash
		txsSndAddr = append(txsSndAddr, tx.SndAddr)
	}

	return poolsHolder, txHashes, txsSndAddr
}

func generateValidTx(
	t *testing.T,
	shardCoordinator sharding.Coordinator,
	senderShardId uint32,
	receiverShardId uint32,
) (*transaction.Transaction, []byte) {

	skSender, pkSender, _ := GenerateSkAndPkInShard(shardCoordinator, senderShardId)
	pkSenderBuff, _ := pkSender.ToByteArray()

	_, pkRecv, _ := GenerateSkAndPkInShard(shardCoordinator, receiverShardId)
	pkRecvBuff, _ := pkRecv.ToByteArray()

	trieStorage, _ := CreateTrieStorageManager(CreateMemUnit())
	accnts, _ := CreateAccountsDB(UserAccount, trieStorage)
	acc, _ := accnts.LoadAccount(pkSenderBuff)
	_ = accnts.SaveAccount(acc)
	_, _ = accnts.Commit()

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.TxMarshalizerField = TestTxSignMarshalizer
	coreComponents.VmMarshalizerField = TestMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.AddressPubKeyConverterField = TestAddressPubkeyConverter
	coreComponents.ValidatorPubKeyConverterField = TestValidatorPubkeyConverter

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.TxSig = TestSingleSigner
	cryptoComponents.TxKeyGen = signing.NewKeyGenerator(ed25519.NewEd25519())
	cryptoComponents.BlKeyGen = signing.NewKeyGenerator(ed25519.NewEd25519())

	stateComponents := GetDefaultStateComponents()
	stateComponents.Accounts = accnts
	stateComponents.AccountsAPI = accnts

	mockNode, _ := node.NewNode(
		node.WithAddressSignatureSize(64),
		node.WithValidatorSignatureSize(48),
		node.WithCoreComponents(coreComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithStateComponents(stateComponents),
	)

	encodedPkSenderBuff, err := TestAddressPubkeyConverter.Encode(pkSenderBuff)
	assert.Nil(t, err)

	encodedPkRecvBuff, err := TestAddressPubkeyConverter.Encode(pkRecvBuff)
	assert.Nil(t, err)

	tx, err := mockNode.GenerateTransaction(
		encodedPkSenderBuff,
		encodedPkRecvBuff,
		big.NewInt(1),
		"",
		skSender,
		ChainID,
		MinTransactionVersion,
	)
	assert.Nil(t, err)

	txBuff, _ := TestMarshalizer.Marshal(tx)
	txHash := TestHasher.Compute(string(txBuff))

	return tx, txHash
}

// ProposeAndSyncOneBlock proposes a block, syncs the block and then increments the round
func ProposeAndSyncOneBlock(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxProposers []int,
	round uint64,
	nonce uint64,
) (uint64, uint64) {

	UpdateRound(nodes, round)
	ProposeBlock(nodes, idxProposers, round, nonce)
	SyncBlock(t, nodes, idxProposers, round)
	round = IncrementAndPrintRound(round)
	nonce++

	return round, nonce
}

// PubKeysMapFromTxKeysMap returns a map of public keys per shard from the key pairs per shard map.
func PubKeysMapFromTxKeysMap(keyPairMap map[uint32][]*TestKeyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string)

	for shardId, pairList := range keyPairMap {
		shardKeys := make([]string, len(pairList))
		for i, pair := range pairList {
			b, _ := pair.Pk.ToByteArray()
			shardKeys[i] = string(b)
		}
		keysMap[shardId] = shardKeys
	}

	return keysMap
}

// PubKeysMapFromNodesKeysMap returns a map of public keys per shard from the key pairs per shard map.
func PubKeysMapFromNodesKeysMap(keyPairMap map[uint32][]*TestNodeKeys) map[uint32][]string {
	keysMap := make(map[uint32][]string)

	for shardId, keys := range keyPairMap {
		addAllKeysOnShard(keysMap, shardId, keys)
	}

	return keysMap
}

func addAllKeysOnShard(m map[uint32][]string, shardID uint32, keys []*TestNodeKeys) {
	for _, keyOfTheNode := range keys {
		addKeysToMap(m, shardID, keyOfTheNode)
	}
}

func addKeysToMap(m map[uint32][]string, shardID uint32, keysOfTheNode *TestNodeKeys) {
	if len(keysOfTheNode.HandledKeys) == 0 {
		b, _ := keysOfTheNode.MainKey.Pk.ToByteArray()
		m[shardID] = append(m[shardID], string(b))
		return
	}

	for _, handledKey := range keysOfTheNode.HandledKeys {
		b, _ := handledKey.Pk.ToByteArray()
		m[shardID] = append(m[shardID], string(b))
	}
}

// GenValidatorsFromPubKeys generates a map of validators per shard out of public keys map
func GenValidatorsFromPubKeys(pubKeysMap map[uint32][]string, _ uint32) map[uint32][]nodesCoordinator.GenesisNodeInfoHandler {
	validatorsMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)

	for shardId, shardNodesPks := range pubKeysMap {
		var shardValidators []nodesCoordinator.GenesisNodeInfoHandler
		for i := 0; i < len(shardNodesPks); i++ {
			v := mock.NewNodeInfo([]byte(shardNodesPks[i][:32]), []byte(shardNodesPks[i]), shardId, InitialRating)
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

// GenValidatorsFromPubKeysAndTxPubKeys generates a map of validators per shard out of public keys map
func GenValidatorsFromPubKeysAndTxPubKeys(
	blsPubKeysMap map[uint32][]string,
	txPubKeysMap map[uint32][]string,
) map[uint32][]nodesCoordinator.GenesisNodeInfoHandler {
	validatorsMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)

	for shardId, shardNodesPks := range blsPubKeysMap {
		var shardValidators []nodesCoordinator.GenesisNodeInfoHandler
		for i := 0; i < len(shardNodesPks); i++ {
			v := mock.NewNodeInfo([]byte(txPubKeysMap[shardId][i]), []byte(shardNodesPks[i]), shardId, InitialRating)
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

// CreateCryptoParams generates the crypto parameters (key pairs, key generator and suite) for multiple nodes
func CreateCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards uint32, numKeysOnEachNode int) *CryptoParams {
	txSuite := ed25519.NewEd25519()
	txKeyGen := signing.NewKeyGenerator(txSuite)
	suite := mcl.NewSuiteBLS12()
	singleSigner := TestSingleSigner
	keyGen := signing.NewKeyGenerator(suite)
	p2pSuite := secp256k1.NewSecp256k1()
	p2pKeyGen := signing.NewKeyGenerator(p2pSuite)

	nodesKeysMap := make(map[uint32][]*TestNodeKeys)
	txKeysMap := make(map[uint32][]*TestKeyPair)
	for shardId := uint32(0); shardId < nbShards; shardId++ {
		for n := 0; n < nodesPerShard; n++ {
			createAndAddKeys(keyGen, txKeyGen, shardId, nodesKeysMap, txKeysMap, numKeysOnEachNode)
		}
	}

	for n := 0; n < nbMetaNodes; n++ {
		createAndAddKeys(keyGen, txKeyGen, core.MetachainShardId, nodesKeysMap, txKeysMap, numKeysOnEachNode)
	}

	params := &CryptoParams{
		NodesKeys:    nodesKeysMap,
		KeyGen:       keyGen,
		P2PKeyGen:    p2pKeyGen,
		SingleSigner: singleSigner,
		TxKeyGen:     txKeyGen,
		TxKeys:       txKeysMap,
	}

	return params
}

func createAndAddKeys(
	keyGen crypto.KeyGenerator,
	txKeyGen crypto.KeyGenerator,
	shardId uint32,
	nodeKeysMap map[uint32][]*TestNodeKeys,
	txKeysMap map[uint32][]*TestKeyPair,
	numKeysOnEachNode int,
) {
	kp := &TestKeyPair{}
	kp.Sk, kp.Pk = keyGen.GeneratePair()

	txKp := &TestKeyPair{}
	txKp.Sk, txKp.Pk = txKeyGen.GeneratePair()

	nodeKey := &TestNodeKeys{
		MainKey: kp,
	}

	txKeysMap[shardId] = append(txKeysMap[shardId], txKp)
	nodeKeysMap[shardId] = append(nodeKeysMap[shardId], nodeKey)
	if numKeysOnEachNode == 1 {
		return
	}

	for i := 0; i < numKeysOnEachNode; i++ {
		validatorKp := &TestKeyPair{}
		validatorKp.Sk, validatorKp.Pk = keyGen.GeneratePair()

		nodeKey.HandledKeys = append(nodeKey.HandledKeys, validatorKp)
	}
}

// CloseProcessorNodes closes the used TestProcessorNodes and advertiser
func CloseProcessorNodes(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		n.Close()
	}
}

// BootstrapDelay will delay the execution to allow the p2p bootstrap
func BootstrapDelay() {
	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(P2pBootstrapDelay)
}

// SetupSyncNodesOneShardAndMeta creates nodes with sync capabilities divided into one shard and a metachain
func SetupSyncNodesOneShardAndMeta(
	numNodesPerShard int,
	numNodesMeta int,
) ([]*TestProcessorNode, []int) {

	maxShardsLocal := uint32(1)
	shardId := uint32(0)

	var nodes []*TestProcessorNode
	var connectableNodes []Connectable
	for i := 0; i < numNodesPerShard; i++ {
		shardNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            maxShardsLocal,
			NodeShardId:          shardId,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
		})
		nodes = append(nodes, shardNode)
		connectableNodes = append(connectableNodes, shardNode)
	}
	idxProposerShard0 := 0

	for i := 0; i < numNodesMeta; i++ {
		metaNode := NewTestProcessorNode(ArgTestProcessorNode{
			MaxShards:            maxShardsLocal,
			NodeShardId:          core.MetachainShardId,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
		})
		nodes = append(nodes, metaNode)
		connectableNodes = append(connectableNodes, metaNode)
	}
	idxProposerMeta := len(nodes) - 1

	idxProposers := []int{idxProposerShard0, idxProposerMeta}

	ConnectNodes(connectableNodes)

	return nodes, idxProposers
}

// StartSyncingBlocks starts the syncing process of all the nodes
func StartSyncingBlocks(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		_ = n.StartSync()
	}

	log.Info("Delaying for nodes to start syncing blocks...")
	time.Sleep(StepDelay)
}

// ForkChoiceOneBlock rollbacks a block from the given shard
func ForkChoiceOneBlock(nodes []*TestProcessorNode, shardId uint32) {
	for idx, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}
		err := n.Bootstrapper.RollBack(false)
		if err != nil {
			log.Error(err.Error())
		}

		newNonce := n.BlockChain.GetCurrentBlockHeader().GetNonce()
		log.Info(fmt.Sprintf("Node's id %d is at block height %d", idx, newNonce))
	}
}

// ResetHighestProbableNonce resets the highest probable nonce
func ResetHighestProbableNonce(nodes []*TestProcessorNode, shardId uint32, targetNonce uint64) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}
		if n.BlockChain.GetCurrentBlockHeader().GetNonce() != targetNonce {
			continue
		}

		n.Bootstrapper.SetProbableHighestNonce(targetNonce)
	}
}

// EmptyDataPools clears all the data pools
func EmptyDataPools(nodes []*TestProcessorNode, shardId uint32) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}

		emptyNodeDataPool(n)
	}
}

func emptyNodeDataPool(node *TestProcessorNode) {
	if node.DataPool != nil {
		emptyDataPool(node.DataPool)
	}
}

func emptyDataPool(sdp dataRetriever.PoolsHolder) {
	sdp.Headers().Clear()
	sdp.UnsignedTransactions().Clear()
	sdp.Transactions().Clear()
	sdp.MiniBlocks().Clear()
	sdp.PeerChangesBlocks().Clear()
}

// UpdateRound updates the round for every node
func UpdateRound(nodes []*TestProcessorNode, round uint64) {
	for _, n := range nodes {
		n.RoundHandler.IndexField = int64(round)
	}

	// this delay is needed in order for the round to be properly updated in the nodes
	time.Sleep(10 * time.Millisecond)
}

// ProposeBlocks proposes blocks for a given number of rounds
func ProposeBlocks(
	nodes []*TestProcessorNode,
	round *uint64,
	idxProposers []int,
	nonces []*uint64,
	numOfRounds int,
) {

	for i := 0; i < numOfRounds; i++ {
		crtRound := atomic.LoadUint64(round)
		proposeBlocks(nodes, idxProposers, nonces, crtRound)

		time.Sleep(SyncDelay)

		crtRound = IncrementAndPrintRound(crtRound)
		atomic.StoreUint64(round, crtRound)
		UpdateRound(nodes, crtRound)
		IncrementNonces(nonces)
	}
	time.Sleep(SyncDelay)
}

// IncrementNonces increments all the nonces
func IncrementNonces(nonces []*uint64) {
	for i := 0; i < len(nonces); i++ {
		atomic.AddUint64(nonces[i], 1)
	}
}

func proposeBlocks(
	nodes []*TestProcessorNode,
	idxProposers []int,
	nonces []*uint64,
	crtRound uint64,
) {
	for idx, proposer := range idxProposers {
		crtNonce := atomic.LoadUint64(nonces[idx])
		ProposeBlock(nodes, []int{proposer}, crtRound, crtNonce)
	}
}

// WaitOperationToBeDone -
func WaitOperationToBeDone(t *testing.T, nodes []*TestProcessorNode, nrOfRounds int, nonce uint64, round uint64, idxProposers []int) (uint64, uint64) {
	for i := 0; i < nrOfRounds; i++ {
		round, nonce = ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	return nonce, round
}

// AddSelfNotarizedHeaderByMetachain -
func AddSelfNotarizedHeaderByMetachain(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == core.MetachainShardId {
			continue
		}

		header := n.BlockChain.GetCurrentBlockHeader()
		if check.IfNil(header) {
			continue
		}

		n.BlockTracker.AddSelfNotarizedHeader(core.MetachainShardId, header, nil)
	}
}

// WhiteListTxs -
func WhiteListTxs(nodes []*TestProcessorNode, txs []*transaction.Transaction) {
	txHashes := make([][]byte, 0)
	for _, tx := range txs {
		txHash, err := core.CalculateHash(TestMarshalizer, TestHasher, tx)
		if err != nil {
			return
		}

		txHashes = append(txHashes, txHash)
	}

	for _, n := range nodes {
		for index, txHash := range txHashes {
			senderShardID := n.ShardCoordinator.ComputeId(txs[index].SndAddr)
			receiverShardID := n.ShardCoordinator.ComputeId(txs[index].RcvAddr)
			if senderShardID == n.ShardCoordinator.SelfId() ||
				receiverShardID == n.ShardCoordinator.SelfId() {
				n.WhiteListHandler.Add([][]byte{txHash})
			}
		}
	}
}

// SaveDelegationManagerConfig will save a mock configuration for the delegation manager SC
func SaveDelegationManagerConfig(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		acc, _ := n.AccntState.LoadAccount(vm.DelegationManagerSCAddress)
		userAcc, _ := acc.(state.UserAccountHandler)

		managementData := &systemSmartContracts.DelegationManagement{
			MinDeposit:          big.NewInt(100),
			LastAddress:         vm.FirstDelegationSCAddress,
			MinDelegationAmount: big.NewInt(1),
		}
		marshaledData, _ := TestMarshalizer.Marshal(managementData)
		_ = userAcc.SaveKeyValue([]byte(delegationManagementKey), marshaledData)
		_ = n.AccntState.SaveAccount(userAcc)
		_, _ = n.AccntState.Commit()
	}
}

// SaveDelegationContractsList will save a mock configuration for the delegation contracts list
func SaveDelegationContractsList(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		acc, _ := n.AccntState.LoadAccount(vm.DelegationManagerSCAddress)
		userAcc, _ := acc.(state.UserAccountHandler)

		managementData := &systemSmartContracts.DelegationContractList{
			Addresses: [][]byte{[]byte("addr")},
		}
		marshaledData, _ := TestMarshalizer.Marshal(managementData)
		_ = userAcc.SaveKeyValue([]byte(delegationContractsList), marshaledData)
		_ = n.AccntState.SaveAccount(userAcc)
		_, _ = n.AccntState.Commit()
	}
}

// PrepareRelayedTxDataV1 repares the data for a relayed transaction V1
func PrepareRelayedTxDataV1(innerTx *transaction.Transaction) []byte {
	userTxBytes, _ := TestMarshalizer.Marshal(innerTx)
	return []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxBytes))
}

// PrepareRelayedTxDataV2 prepares the data for a relayed transaction V2
func PrepareRelayedTxDataV2(innerTx *transaction.Transaction) []byte {
	dataBuilder := txDataBuilder.NewBuilder()
	txData := dataBuilder.
		Func(core.RelayedTransactionV2).
		Bytes(innerTx.RcvAddr).
		Int64(int64(innerTx.Nonce)).
		Bytes(innerTx.Data).
		Bytes(innerTx.Signature)

	return txData.ToBytes()
}
