package integrationTests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

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
		discovery.NewKadDhtPeerDiscoverer(time.Second, "test", []string{initialAddr}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

// CreateTestShardDataPool creates a test data pool for shard nodes
func CreateTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
	)

	return dPool
}

// CreateTestMetaDataPool creates a test data pool for meta nodes
func CreateTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache}
	miniblockHashes, _ := shardedData.NewShardedData(cacherCfg)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	shardHeaders, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	shardHeadersNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	shardHeadersNonces, _ := dataPool.NewNonceSyncMapCacher(shardHeadersNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		miniblockHashes,
		shardHeaders,
		shardHeadersNonces,
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
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, CreateMemUnit())

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
	for i := uint32(0); i < coordinator.NumberOfShards(); i++ {
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(i), CreateMemUnit())
	}

	return store
}

// CreateAccountsDB creates an account state with a valid trie implementation but with a memory storage
func CreateAccountsDB(shardCoordinator sharding.Coordinator,
) (*state.AccountsDB, data.Trie, storage.Storer) {

	accountFactory, _ := factory.NewAccountFactoryCreator(shardCoordinator)
	if shardCoordinator == nil {
		accountFactory = factory.NewAccountCreator()
	}

	store := CreateMemUnit()
	tr, _ := trie.NewTrie(store, TestMarshalizer, TestHasher)
	adb, _ := state.NewAccountsDB(tr, TestHasher, TestMarshalizer, accountFactory)

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

// CreateGenesisBlocks creates empty genesis blocks for all known shards, including metachain
func CreateGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = CreateGenesisBlock(shardId)
	}

	genesisBlocks[sharding.MetachainShardId] = CreateGenesisMetaBlock()

	return genesisBlocks
}

// CreateGenesisBlock creates a new mock shard genesis block
func CreateGenesisBlock(shardId uint32) *dataBlock.Header {
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

// CreateGenesisMetaBlock creates a new mock meta genesis block
func CreateGenesisMetaBlock() *dataBlock.MetaBlock {
	rootHash := []byte("root hash")

	return &dataBlock.MetaBlock{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

// CreateIeleVMAndBlockchainHook creates a new instance of a iele VM
func CreateIeleVMAndBlockchainHook(accnts state.AccountsAdapter) (vmcommon.VMExecutionHandler, *hooks.VMAccountsDB) {
	blockChainHook, _ := hooks.NewVMAccountsDB(accnts, TestAddressConverter)
	cryptoHook := hooks.NewVMCryptoHook()
	vm := endpoint.NewElrondIeleVM(blockChainHook, cryptoHook, endpoint.ElrondTestnet)

	return vm, blockChainHook
}

// CreateAddressFromAddrBytes creates an address container object from address bytes provided
func CreateAddressFromAddrBytes(addressBytes []byte) state.AddressContainer {
	addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(addressBytes)
	return addr
}

// CreateAddressFromHex creates an address container object from hex
func CreateAddressFromHex() state.AddressContainer {
	addr, _ := TestAddressConverter.CreateAddressFromHex(CreateRandomHexAddress(64))
	return addr
}

// MintAddress will create an account (if it does not exists), updated the balance with required value,
// saves the account and commit the trie.
func MintAddress(accnts state.AccountsAdapter, addressBytes []byte, value *big.Int) {
	accnt, _ := accnts.GetAccountWithJournal(CreateAddressFromAddrBytes(addressBytes))
	_ = accnt.(*state.Account).SetBalanceWithJournal(value)
	_, _ = accnts.Commit()
}

// CreateAccount creates a new account and returns the address
func CreateAccount(accnts state.AccountsAdapter, nonce uint64, balance *big.Int) state.AddressContainer {
	address, _ := TestAddressConverter.CreateAddressFromHex(CreateRandomHexAddress(64))
	account, _ := accnts.GetAccountWithJournal(address)
	_ = account.(*state.Account).SetNonceWithJournal(nonce)
	_ = account.(*state.Account).SetBalanceWithJournal(balance)

	return address
}

// MakeDisplayTable will output a string containing counters for received transactions, headers, miniblocks and
// meta headers for all provided test nodes
func MakeDisplayTable(nodes []*TestProcessorNode) string {
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers"}
	dataLines := make([]*display.LineData, len(nodes))

	for idx, n := range nodes {
		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(n.PkTxSignBytes),
				fmt.Sprintf("%d", n.ShardCoordinator.SelfId()),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterTxRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterMbRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterHdrRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.CounterMetaRcv)),
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

// CreateRandomAddress creates a random byte array with fixed size
func CreateRandomAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())
	_, _ = rand.Reader.Read(buff)

	return state.NewAddress(buff)
}

// CreateRandomHexAddress returns a string encoded in hex with the given size
func CreateRandomHexAddress(chars int) string {
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
	adb, _, _ := CreateAccountsDB(nil)
	account, _ := state.NewAccount(adr, adb)

	return adr, account, adb
}

// AdbEmulateBalanceTxSafeExecution emulates a tx execution by altering the accounts
// balance and nonce, and printing any encountered error
func AdbEmulateBalanceTxSafeExecution(acntSrc, acntDest *state.Account, accounts state.AccountsAdapter, value *big.Int) {

	snapshot := accounts.JournalLen()
	err := AdbEmulateBalanceTxExecution(acntSrc, acntDest, value)

	if err != nil {
		fmt.Printf("!!!! Error executing tx (value: %v), reverting...\n", value)
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
	txProcessor, _ := transaction.NewTxProcessor(accnts, TestHasher, TestAddressConverter, TestMarshalizer, shardCoordinator, &mock.SCProcessorMock{})

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
		shardId := targetNode.ShardCoordinator.ComputeId(n.TxSignAddress)
		if shardId != targetNode.ShardCoordinator.SelfId() {
			continue
		}

		MintAddress(targetNode.AccntState, n.PkTxSignBytes, value)
	}
}
