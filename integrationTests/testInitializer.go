package integrationTests

import (
	"context"
	"crypto/ecdsa"
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

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/epochStart/genesis"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/p2p/memp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	procFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	txProc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var stepDelay = time.Second
var p2pBootstrapStepDelay = 5 * time.Second

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

// CreateMessengerWithKadDht creates a new libp2p messenger with kad-dht peer discovery
func CreateMessengerWithKadDht(ctx context.Context, initialAddr string) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessengerOnFreePort(
		ctx,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewKadDhtPeerDiscoverer(stepDelay, "test", []string{initialAddr}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

// CreateTestShardDataPool creates a test data pool for shard nodes
func CreateTestShardDataPool(txPool dataRetriever.ShardedDataCacherNotifier) dataRetriever.PoolsHolder {
	if txPool == nil {
		txPool, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})
	}

	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})
	rewardsTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 300, Type: storageUnit.LRUCache, Shards: 1})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache, Shards: 1}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	currTxs, _ := dataPool.NewCurrentBlockPool()

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		rewardsTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		currTxs,
	)

	return dPool
}

// CreateTestMetaDataPool creates a test data pool for meta nodes
func CreateTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	shardHeaders, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	shardHeadersNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	shardHeadersNonces, _ := dataPool.NewNonceSyncMapCacher(shardHeadersNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})
	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})

	currTxs, _ := dataPool.NewCurrentBlockPool()

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		txBlockBody,
		shardHeaders,
		shardHeadersNonces,
		txPool,
		uTxPool,
		currTxs,
	)

	return dPool
}

// CreateMemUnit returns an in-memory storer implementation (the vast majority of tests do not require effective
// disk I/O)
func CreateMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.NewlruDB(100000)
	unit, _ := storageUnit.NewStorageUnit(cache, persist)

	return unit
}

// CreateShardStore creates a storage service for shard nodes
func CreateShardStore(numOfShards uint32) dataRetriever.StorageService {
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

	for i := uint32(0); i < numOfShards; i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, CreateMemUnit())
	}

	return store
}

// CreateMetaStore creates a storage service for meta nodes
func CreateMetaStore(coordinator sharding.Coordinator) dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.TransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.StatusMetricsUnit, CreateMemUnit())

	for i := uint32(0); i < coordinator.NumberOfShards(); i++ {
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(i), CreateMemUnit())
	}

	return store
}

// CreateAccountsDB creates an account state with a valid trie implementation but with a memory storage
func CreateAccountsDB(accountType factory.Type) (*state.AccountsDB, data.Trie, storage.Storer) {
	hasher := sha256.Sha256{}
	store := CreateMemUnit()

	tr, _ := trie.NewTrie(store, TestMarshalizer, hasher)
	accountFactory, _ := factory.NewAccountFactoryCreator(accountType)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, TestMarshalizer, accountFactory)

	return adb, tr, store
}

// CreateShardChain creates a blockchain implementation used by the shard nodes
func CreateShardChain() *blockchain.BlockChain {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}
	genesisHeaderM, _ := TestMarshalizer.Marshal(blockChain.GenesisHeader)

	blockChain.SetGenesisHeaderHash(TestHasher.Compute(string(genesisHeaderM)))

	return blockChain
}

// CreateMetaChain creates a blockchain implementation used by the meta nodes
func CreateMetaChain() data.ChainHandler {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	metaChain, _ := blockchain.NewMetaChain(
		badBlockCache,
	)
	metaChain.GenesisBlock = &dataBlock.MetaBlock{}

	return metaChain
}

// CreateSimpleGenesisBlocks creates empty genesis blocks for all known shards, including metachain
func CreateSimpleGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = CreateSimpleGenesisBlock(shardId)
	}

	genesisBlocks[sharding.MetachainShardId] = CreateSimpleGenesisMetaBlock()

	return genesisBlocks
}

// CreateSimpleGenesisBlock creates a new mock shard genesis block
func CreateSimpleGenesisBlock(shardId uint32) *dataBlock.Header {
	rootHash := []byte("root hash")

	return &dataBlock.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       shardId,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
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
		PeerInfo:               nil,
		Signature:              nil,
		PubKeysBitmap:          nil,
		PrevHash:               rootHash,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: rootHash,
		TxCount:                0,
		MiniBlockHeaders:       nil,
	}
}

// CreateGenesisBlocks creates empty genesis blocks for all known shards, including metachain
func CreateGenesisBlocks(
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
	nodesSetup *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	metaDataPool dataRetriever.MetaPoolsHolder,
	economics *economics.EconomicsData,
) map[uint32]data.HeaderHandler {

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = CreateSimpleGenesisBlock(shardId)
	}

	genesisBlocks[sharding.MetachainShardId] = CreateGenesisMetaBlock(
		accounts,
		addrConv,
		nodesSetup,
		shardCoordinator,
		store,
		blkc,
		marshalizer,
		hasher,
		uint64Converter,
		metaDataPool,
		economics,
	)

	return genesisBlocks
}

// CreateGenesisMetaBlock creates a new mock meta genesis block
func CreateGenesisMetaBlock(
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
	nodesSetup *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	metaDataPool dataRetriever.MetaPoolsHolder,
	economics *economics.EconomicsData,
) data.HeaderHandler {
	argsMetaGenesis := genesis.ArgsMetaGenesisBlockCreator{
		GenesisTime:              0,
		Accounts:                 accounts,
		AddrConv:                 addrConv,
		NodesSetup:               nodesSetup,
		ShardCoordinator:         shardCoordinator,
		Store:                    store,
		Blkc:                     blkc,
		Marshalizer:              marshalizer,
		Hasher:                   hasher,
		Uint64ByteSliceConverter: uint64Converter,
		MetaDatapool:             metaDataPool,
		Economics:                economics,
		ValidatorStatsRootHash:   []byte("validator stats root hash"),
	}

	if shardCoordinator.SelfId() != sharding.MetachainShardId {
		newShardCoordinator, _ := sharding.NewMultiShardCoordinator(
			shardCoordinator.NumberOfShards(),
			sharding.MetachainShardId,
		)

		newStore := CreateMetaStore(newShardCoordinator)
		newMetaDataPool := CreateTestMetaDataPool()

		cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
		newBlkc, _ := blockchain.NewMetaChain(cache)
		newAccounts, _, _ := CreateAccountsDB(factory.UserAccount)

		argsMetaGenesis.ShardCoordinator = newShardCoordinator
		argsMetaGenesis.Accounts = newAccounts
		argsMetaGenesis.Store = newStore
		argsMetaGenesis.Blkc = newBlkc
		argsMetaGenesis.MetaDatapool = newMetaDataPool
	}

	metaHdr, _ := genesis.CreateMetaGenesisBlock(argsMetaGenesis)
	fmt.Printf("meta genesis root hash %s \n", hex.EncodeToString(metaHdr.GetRootHash()))

	return metaHdr
}

// CreateAddressFromAddrBytes creates an address container object from address bytes provided
func CreateAddressFromAddrBytes(addressBytes []byte) state.AddressContainer {
	addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(addressBytes)
	return addr
}

// CreateRandomAddress creates a random byte array with fixed size
func CreateRandomAddress() state.AddressContainer {
	addr, _ := TestAddressConverter.CreateAddressFromHex(CreateRandomHexString(64))
	return addr
}

// MintAddress will create an account (if it does not exists), update the balance with required value,
// save the account and commit the trie.
func MintAddress(accnts state.AccountsAdapter, addressBytes []byte, value *big.Int) {
	accnt, _ := accnts.GetAccountWithJournal(CreateAddressFromAddrBytes(addressBytes))
	_ = accnt.(*state.Account).SetBalanceWithJournal(value)
	_, _ = accnts.Commit()
}

// CreateAccount creates a new account and returns the address
func CreateAccount(accnts state.AccountsAdapter, nonce uint64, balance *big.Int) state.AddressContainer {
	address, _ := TestAddressConverter.CreateAddressFromHex(CreateRandomHexString(64))
	account, _ := accnts.GetAccountWithJournal(address)
	_ = account.(*state.Account).SetNonceWithJournal(nonce)
	_ = account.(*state.Account).SetBalanceWithJournal(balance)

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
				fmt.Sprintf("%d", len(n.Messenger.ConnectedPeers())),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)

	return table
}

// PrintShardAccount outputs on console a shard account data contained
func PrintShardAccount(accnt *state.Account, tag string) {
	str := fmt.Sprintf("%s Address: %s\n", tag, base64.StdEncoding.EncodeToString(accnt.AddressContainer().Bytes()))
	str += fmt.Sprintf("  Nonce: %d\n", accnt.Nonce)
	str += fmt.Sprintf("  Balance: %d\n", accnt.Balance.Uint64())
	str += fmt.Sprintf("  Code hash: %s\n", base64.StdEncoding.EncodeToString(accnt.CodeHash))
	str += fmt.Sprintf("  Root hash: %s\n", base64.StdEncoding.EncodeToString(accnt.RootHash))

	fmt.Println(str)
}

// CreateRandomHexString returns a string encoded in hex with the given size
func CreateRandomHexString(chars int) string {
	if chars < 1 {
		return ""
	}

	buff := make([]byte, chars/2)
	_, _ = rand.Reader.Read(buff)

	return hex.EncodeToString(buff)
}

// GenerateAddressJournalAccountAccountsDB returns an account, the accounts address, and the accounts database
func GenerateAddressJournalAccountAccountsDB() (state.AddressContainer, state.AccountHandler, *state.AccountsDB) {
	adr := CreateRandomAddress()
	adb, _, _ := CreateAccountsDB(factory.UserAccount)
	account, _ := state.NewAccount(adr, adb)

	return adr, account, adb
}

// AdbEmulateBalanceTxSafeExecution emulates a tx execution by altering the accounts
// balance and nonce, and printing any encountered error
func AdbEmulateBalanceTxSafeExecution(acntSrc, acntDest *state.Account, accounts state.AccountsAdapter, value *big.Int) {

	snapshot := accounts.JournalLen()
	err := AdbEmulateBalanceTxExecution(acntSrc, acntDest, value)

	if err != nil {
		fmt.Printf("Error executing tx (value: %v), reverting...\n", value)
		err = accounts.RevertToSnapshot(snapshot)

		if err != nil {
			panic(err)
		}
	}
}

// AdbEmulateBalanceTxExecution emulates a tx execution by altering the accounts
// balance and nonce, and printing any encountered error
func AdbEmulateBalanceTxExecution(acntSrc, acntDest *state.Account, value *big.Int) error {

	srcVal := acntSrc.Balance
	destVal := acntDest.Balance

	if srcVal.Cmp(value) < 0 {
		return errors.New("not enough funds")
	}

	err := acntSrc.SetBalanceWithJournal(srcVal.Sub(srcVal, value))
	if err != nil {
		return err
	}

	err = acntDest.SetBalanceWithJournal(destVal.Add(destVal, value))
	if err != nil {
		return err
	}

	err = acntSrc.SetNonceWithJournal(acntSrc.Nonce + 1)
	if err != nil {
		return err
	}

	return nil
}

// CreateSimpleTxProcessor returns a transaction processor
func CreateSimpleTxProcessor(accnts state.AccountsAdapter) process.TransactionProcessor {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	txProcessor, _ := txProc.NewTxProcessor(
		accnts,
		TestHasher,
		TestAddressConverter,
		TestMarshalizer,
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.UnsignedTxHandlerMock{},
		&mock.TxTypeHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return tx.GetGasLimit()
			},
			CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
				return nil
			},
			ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
				fee := big.NewInt(0).SetUint64(tx.GetGasLimit())
				fee.Mul(fee, big.NewInt(0).SetUint64(tx.GetGasPrice()))

				return fee
			},
		},
	)

	return txProcessor
}

// CreateNewDefaultTrie returns a new trie with test hasher and marsahalizer
func CreateNewDefaultTrie() data.Trie {
	tr, _ := trie.NewTrie(CreateMemUnit(), TestMarshalizer, TestHasher)
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
		if n.ShardCoordinator.SelfId() == sharding.MetachainShardId {
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
		MintAddress(targetNode.AccntState, n.OwnAccount.Address.Bytes(), value)
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

			MintAddress(n.AccntState, player.Address.Bytes(), value)
			player.Balance = big.NewInt(0).Set(value)
		}
	}
}

// IncrementAndPrintRound increments the given variable, and prints the message for the beginning of the round
func IncrementAndPrintRound(round uint64) uint64 {
	round++
	fmt.Printf("#################################### ROUND %d BEGINS ####################################\n\n", round)

	return round
}

// ProposeBlock proposes a block for every shard
func ProposeBlock(nodes []*TestProcessorNode, idxProposers []int, round uint64, nonce uint64) {
	fmt.Println("All shards propose blocks...")

	for idx, n := range nodes {
		if !IsIntInSlice(idx, idxProposers) {
			continue
		}

		body, header, _ := n.ProposeBlock(round, nonce)
		n.BroadcastBlock(body, header)
		n.CommitBlock(body, header)
	}

	fmt.Println("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelay)
	fmt.Println(MakeDisplayTable(nodes))
}

// SyncBlock synchronizes the proposed block in all the other shard nodes
func SyncBlock(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxProposers []int,
	round uint64,
) {

	fmt.Println("All other shard nodes sync the proposed block...")
	for idx, n := range nodes {
		if IsIntInSlice(idx, idxProposers) {
			continue
		}

		err := n.SyncNode(round)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
	}

	time.Sleep(stepDelay)
	fmt.Println(MakeDisplayTable(nodes))
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

		fmt.Printf("Testing roothash for node index %d, shard ID %d...\n", i, n.ShardCoordinator.SelfId())
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
				fmt.Printf("unsigned tx with nonce %d is missing\n", i)
			}
		}
		assert.Fail(t, fmt.Sprintf("should have been %d, got %d", noOfTxs, len(txHashes)))

		return
	}

	bitmap := make([]bool, noOfTxs+int(startingNonce))
	//set for each nonce from found tx a true flag in bitmap
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

	//for the first startingNonce values, the bitmap should be false
	//for the rest, true
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

	buff, _ := hex.DecodeString(txHandler.GetData())
	return binary.BigEndian.Uint64(buff)
}

// CreateNodes creates multiple nodes in different shards
func CreateNodes(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	serviceID string,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNode(uint32(numOfShards), shardId, shardId, serviceID)

			nodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(uint32(numOfShards), sharding.MetachainShardId, 0, serviceID)
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
	}

	return nodes
}

// CreateNodesWithMemP2P creates multiple nodes in different shards
func CreateNodesWithMemP2P(
	numOfShards int,
	nodesPerShard int,
	numMetaChainNodes int,
	network *memp2p.Network,
) []*TestProcessorNode {
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := NewTestProcessorNodeWithMemP2P(uint32(numOfShards), shardId, shardId, network)

			nodes[idx] = n
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNodeWithMemP2P(uint32(numOfShards), sharding.MetachainShardId, 0, network)
		idx = i + numOfShards*nodesPerShard
		nodes[idx] = metaNode
	}

	return nodes
}

// DisplayAndStartNodes prints each nodes shard ID, sk and pk, and then starts the node
func DisplayAndStartNodes(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		skBuff, _ := n.OwnAccount.SkTxSign.ToByteArray()
		pkBuff, _ := n.OwnAccount.PkTxSign.ToByteArray()

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			n.ShardCoordinator.SelfId(),
			hex.EncodeToString(skBuff),
			hex.EncodeToString(pkBuff),
		)
		_ = n.Node.Start()
		_ = n.Node.P2PBootstrap()
	}

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(p2pBootstrapStepDelay)
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
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := make([]uint64, len(senders))
		for _, shardReceiversPublicKeys := range receiversPublicKeysMap {
			receiverPubKey := shardReceiversPublicKeys[i]
			tx := generateTransferTx(incrementalNonce[i], senderKey, receiverPubKey, valToTransfer, gasPrice, gasLimit)
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
		if shardId == sharding.MetachainShardId {
			continue
		}

		sendersPrivateKeys[shardId], _ = CreateSendersAndReceiversInShard(
			nodes[0],
			1,
		)

		fmt.Println("Minting sender addresses...")
		CreateMintingForSenders(
			nodes,
			shardId,
			sendersPrivateKeys[shardId],
			mintValue,
		)
	}

	return sendersPrivateKeys
}

func CreateAndSendTransaction(
	node *TestProcessorNode,
	txValue *big.Int,
	rcvAddress []byte,
	txData string,
) {
	tx := &transaction.Transaction{
		Nonce:    node.OwnAccount.Nonce,
		Value:    txValue,
		SndAddr:  node.OwnAccount.Address.Bytes(),
		RcvAddr:  rcvAddress,
		Data:     txData,
		GasPrice: MinTxGasPrice,
		GasLimit: MinTxGasLimit*100 + uint64(len(txData)),
	}

	txBuff, _ := TestMarshalizer.Marshal(tx)
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

func generateTransferTx(
	nonce uint64,
	senderPrivateKey crypto.PrivateKey,
	receiverPublicKey crypto.PublicKey,
	valToTransfer *big.Int,
	gasPrice uint64,
	gasLimit uint64,
) *transaction.Transaction {

	receiverPubKeyBytes, _ := receiverPublicKey.ToByteArray()
	tx := transaction.Transaction{
		Nonce:    nonce,
		Value:    valToTransfer,
		RcvAddr:  receiverPubKeyBytes,
		SndAddr:  skToPk(senderPrivateKey),
		Data:     "",
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}
	txBuff, _ := TestMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}
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
		Value:    args.value,
		RcvAddr:  args.rcvAddr,
		SndAddr:  args.sndAddr,
		GasPrice: args.gasPrice,
		GasLimit: args.gasLimit,
		Data:     args.data,
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
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
	addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
	account, _ := n.AccntState.GetExistingAccount(addr)
	assert.Equal(t, expectedBalance, account.(*state.Account).Balance)
}

// TestPrivateKeyHasBalance checks if the private key has the expected balance
func TestPrivateKeyHasBalance(t *testing.T, n *TestProcessorNode, sk crypto.PrivateKey, expectedBalance *big.Int) {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
	account, _ := n.AccntState.GetExistingAccount(addr)
	assert.Equal(t, expectedBalance, account.(*state.Account).Balance)
}

// GetMiniBlocksHashesFromShardIds returns miniblock hashes from body
func GetMiniBlocksHashesFromShardIds(body dataBlock.Body, shardIds ...uint32) [][]byte {
	hashes := make([][]byte, 0)

	for _, miniblock := range body {
		for _, shardId := range shardIds {
			if miniblock.ReceiverShardID == shardId {
				buff, _ := TestMarshalizer.Marshal(miniblock)
				hashes = append(hashes, TestHasher.Compute(string(buff)))
			}
		}
	}

	return hashes
}

// GenerateSkAndPkInShard generates and returns a private and a public key that reside in a given shard.
// It also returns the key generator
func GenerateSkAndPkInShard(
	coordinator sharding.Coordinator,
	shardId uint32,
) (crypto.PrivateKey, crypto.PublicKey, crypto.KeyGenerator) {
	suite := kyber.NewBlakeSHA256Ed25519()
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	if shardId == sharding.MetachainShardId {
		// for metachain generate in shard 0
		shardId = 0
	}

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if coordinator.ComputeId(addr) == shardId {
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
		if shardId == sharding.MetachainShardId {
			continue
		}

		nodeInShard := nodes[shardId][0]

		fmt.Println("Generating transactions...")
		GenerateAndDisseminateTxs(
			nodeInShard,
			sendersPrivKeysMap[shardId],
			receiversPubKeysMap,
			valueToTransfer,
			gasPricePerTx,
			gasLimitPerTx,
		)
	}

	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)
}

// CreateMintingForSenders creates account with balances for every node in a given shard
func CreateMintingForSenders(
	nodes []*TestProcessorNode,
	senderShard uint32,
	sendersPrivateKeys []crypto.PrivateKey,
	value *big.Int,
) {

	for _, n := range nodes {
		//only sender shard nodes will be minted
		if n.ShardCoordinator.SelfId() != senderShard {
			continue
		}

		for _, sk := range sendersPrivateKeys {
			pkBuff, _ := sk.GeneratePublic().ToByteArray()
			adr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
			account, _ := n.AccntState.GetAccountWithJournal(adr)
			_ = account.(*state.Account).SetBalanceWithJournal(value)
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

	fmt.Println("Proposing block without commit...")

	body, header, txHashes := node.ProposeBlock(round, nonce)
	node.BroadcastBlock(body, header)
	isEmptyBlock := len(txHashes) == 0

	fmt.Println("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelay)

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
	addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(node.OwnAccount.PkTxSignBytes)
	_, _ = node.AccntState.GetAccountWithJournal(addr)
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
	neededTxs := make([][]byte, 0)

	for i := 0; i < len(generatedTxHashes); i++ {
		_, ok := n.ShardDataPool.Transactions().SearchFirstData(generatedTxHashes[i])
		if !ok {
			neededTxs = append(neededTxs, generatedTxHashes[i])
		}
	}

	return neededTxs
}

func requestMissingTransactions(n *TestProcessorNode, shardResolver uint32, neededTxs [][]byte) {
	txResolver, _ := n.ResolverFinder.CrossShardResolver(procFactory.TransactionTopic, shardResolver)

	for i := 0; i < len(neededTxs); i++ {
		_ = txResolver.RequestDataFromHash(neededTxs[i])
	}
}

// CreateRequesterDataPool creates a datapool with a mock txPool
func CreateRequesterDataPool(t *testing.T, recvTxs map[int]map[string]struct{}, mutRecvTxs *sync.Mutex, nodeIndex int) dataRetriever.PoolsHolder {

	//not allowed to request data from the same shard
	return CreateTestShardDataPool(&mock.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			assert.Fail(t, "same-shard requesters should not be queried")
			return nil, false
		},
		ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
			assert.Fail(t, "same-shard requesters should not be queried")
			return nil
		},
		AddDataCalled: func(key []byte, data interface{}, cacheId string) {
			mutRecvTxs.Lock()
			defer mutRecvTxs.Unlock()

			txMap := recvTxs[nodeIndex]
			if txMap == nil {
				txMap = make(map[string]struct{})
				recvTxs[nodeIndex] = txMap
			}

			txMap[string(key)] = struct{}{}
		},
		RegisterHandlerCalled: func(i func(key []byte)) {
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
	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})

	for i := 0; i < maxTxs; i++ {
		tx, txHash := generateValidTx(t, shardCoordinator, senderShardID, recvShardId)
		cacherIdentifier := process.ShardCacherIdentifier(1, 0)
		txPool.AddData(txHash, tx, cacherIdentifier)
		txHashes[i] = txHash
		txsSndAddr = append(txsSndAddr, tx.SndAddr)
	}

	return CreateTestShardDataPool(txPool), txHashes, txsSndAddr
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

	accnts, _, _ := CreateAccountsDB(factory.UserAccount)
	addrSender, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkSenderBuff)
	_, _ = accnts.GetAccountWithJournal(addrSender)
	_, _ = accnts.Commit()

	mockNode, _ := node.NewNode(
		node.WithMarshalizer(TestMarshalizer, 100),
		node.WithHasher(TestHasher),
		node.WithAddressConverter(TestAddressConverter),
		node.WithKeyGen(signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())),
		node.WithTxSingleSigner(&singlesig.SchnorrSigner{}),
		node.WithTxSignPrivKey(skSender),
		node.WithTxSignPubKey(pkSender),
		node.WithAccountsAdapter(accnts),
	)

	tx, err := mockNode.GenerateTransaction(
		hex.EncodeToString(pkSenderBuff),
		hex.EncodeToString(pkRecvBuff),
		big.NewInt(1),
		"",
	)
	assert.Nil(t, err)

	txBuff, _ := TestMarshalizer.Marshal(tx)
	txHash := TestHasher.Compute(string(txBuff))

	return tx, txHash
}

// GetNumTxsWithDst returns the total number of transactions that have a certain destination shard
func GetNumTxsWithDst(dstShardId uint32, dataPool dataRetriever.PoolsHolder, nrShards uint32) int {
	txPool := dataPool.Transactions()
	if txPool == nil {
		return 0
	}

	sumTxs := 0

	for i := uint32(0); i < nrShards; i++ {
		strCache := process.ShardCacherIdentifier(i, dstShardId)
		txStore := txPool.ShardDataStore(strCache)
		if txStore == nil {
			continue
		}
		sumTxs += txStore.Len()
	}

	return sumTxs
}

// ProposeAndSyncBlocks proposes and syncs blocks until all transaction pools are empty
func ProposeAndSyncBlocks(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxProposers []int,
	round uint64,
	nonce uint64,
) (uint64, uint64) {

	// if there are many transactions, they might not fit into the block body in only one round
	for {
		numTxsInPool := 0
		round, nonce = ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

		for _, idProposer := range idxProposers {
			proposerNode := nodes[idProposer]
			numTxsInPool = GetNumTxsWithDst(
				proposerNode.ShardCoordinator.SelfId(),
				proposerNode.ShardDataPool,
				proposerNode.ShardCoordinator.NumberOfShards(),
			)

			if numTxsInPool > 0 {
				break
			}
		}

		if numTxsInPool == 0 {
			break
		}
	}

	if nodes[0].ShardCoordinator.NumberOfShards() == 1 {
		return round, nonce
	}

	// cross shard smart contract call is first processed at sender shard, notarized by metachain, processed at
	// shard with smart contract, smart contract result is notarized by metachain, then finally processed at the
	// sender shard
	numberToPropagateToEveryShard := 5
	for i := 0; i < numberToPropagateToEveryShard; i++ {
		round, nonce = ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	return round, nonce
}

// ProposeAndSyncOneBlock proposes a block, syncs the block and then increments the round
func ProposeAndSyncOneBlock(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxProposers []int,
	round uint64,
	nonce uint64,
) (uint64, uint64) {
	ProposeBlock(nodes, idxProposers, round, nonce)
	SyncBlock(t, nodes, idxProposers, round)
	round = IncrementAndPrintRound(round)
	nonce++

	return round, nonce
}

// WaitForBootstrapAndShowConnected will delay a given duration in order to wait for bootstraping  and print the
// number of peers that each node is connected to
func WaitForBootstrapAndShowConnected(peers []p2p.Messenger, durationBootstrapingTime time.Duration) {
	fmt.Printf("Waiting %v for peer discovery...\n", durationBootstrapingTime)
	time.Sleep(durationBootstrapingTime)

	fmt.Println("Connected peers:")
	for _, peer := range peers {
		fmt.Printf("Peer %s is connected to %d peers\n", peer.ID().Pretty(), len(peer.ConnectedPeers()))
	}
}

// PubKeysMapFromKeysMap returns a map of public keys per shard from the key pairs per shard map.
func PubKeysMapFromKeysMap(keyPairMap map[uint32][]*TestKeyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string, 0)

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

// GenValidatorsFromPubKeys generates a map of validators per shard out of public keys map
func GenValidatorsFromPubKeys(pubKeysMap map[uint32][]string, nbShards uint32) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(nbShards, shardId)
		for i := 0; i < len(shardNodesPks); i++ {
			_, pk, _ := GenerateSkAndPkInShard(shardCoordinator, shardId)
			address, err := pk.ToByteArray()
			if err != nil {
				return nil
			}
			v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]), address)
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

// CreateCryptoParams generates the crypto parameters (key pairs, key generator and suite) for multiple nodes
func CreateCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards uint32) *CryptoParams {
	suite := kyber.NewSuitePairingBn256()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)

	keysMap := make(map[uint32][]*TestKeyPair)
	keyPairs := make([]*TestKeyPair, nodesPerShard)
	for shardId := uint32(0); shardId < nbShards; shardId++ {
		for n := 0; n < nodesPerShard; n++ {
			kp := &TestKeyPair{}
			kp.Sk, kp.Pk = keyGen.GeneratePair()
			keyPairs[n] = kp
		}
		keysMap[shardId] = keyPairs
	}

	keyPairs = make([]*TestKeyPair, nbMetaNodes)
	for n := 0; n < nbMetaNodes; n++ {
		kp := &TestKeyPair{}
		kp.Sk, kp.Pk = keyGen.GeneratePair()
		keyPairs[n] = kp
	}
	keysMap[sharding.MetachainShardId] = keyPairs

	params := &CryptoParams{
		Keys:         keysMap,
		KeyGen:       keyGen,
		SingleSigner: singleSigner,
	}

	return params
}

// CloseProcessorNodes closes the used TestProcessorNodes and advertiser
func CloseProcessorNodes(nodes []*TestProcessorNode, advertiser p2p.Messenger) {
	_ = advertiser.Close()
	for _, n := range nodes {
		_ = n.Messenger.Close()
	}
}

// StartP2pBootstrapOnProcessorNodes will start the p2p discovery on processor nodes and wait a predefined time
func StartP2pBootstrapOnProcessorNodes(nodes []*TestProcessorNode) {
	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(p2pBootstrapStepDelay)
}
