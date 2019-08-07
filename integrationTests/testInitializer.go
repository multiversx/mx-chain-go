package integrationTests

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	txProc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var stepDelay = time.Second

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
func CreateAccountsDB(shardCoordinator sharding.Coordinator) (*state.AccountsDB, data.Trie, storage.Storer) {

	var accountFactory state.AccountFactory
	if shardCoordinator == nil {
		accountFactory = factory.NewAccountCreator()
	} else {
		accountFactory, _ = factory.NewAccountFactoryCreator(shardCoordinator)
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
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers"}
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
	txProcessor, _ := txProc.NewTxProcessor(accnts, TestHasher, TestAddressConverter, TestMarshalizer, shardCoordinator, &mock.SCProcessorMock{})

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

		MintAddress(targetNode.AccntState, n.OwnAccount.PkTxSignBytes, value)
	}
}

func MintAllPlayers(nodes []*TestProcessorNode, players []*TestWalletAccount, value *big.Int) {
	shardCoordinator := nodes[0].ShardCoordinator

	for _, player := range players {
		pShardId := shardCoordinator.ComputeId(player.Address)

		for _, node := range nodes {
			if pShardId != node.ShardCoordinator.SelfId() {
				continue
			}

			MintAddress(node.AccntState, player.Address.Bytes(), value)
			player.Balance = big.NewInt(0).Set(value)
		}
	}
}

// IncrementAndPrintRound increments the given variable, and prints the message for teh beginning of the round
func IncrementAndPrintRound(round uint64) uint64 {
	round++
	fmt.Printf("#################################### ROUND %d BEGINS ####################################\n\n", round)

	return round
}

// ProposeBlock proposes a block with SC txs for every shard
func ProposeBlock(nodes []*TestProcessorNode, idxProposers []int, round uint64) {
	fmt.Println("All shards propose blocks...")
	for idx, n := range nodes {
		if !isIntInSlice(idx, idxProposers) {
			continue
		}

		body, header, _ := n.ProposeBlock(round)
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
		if isIntInSlice(idx, idxProposers) {
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

func isIntInSlice(idx int, slice []int) bool {
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
		node := nodes[i]

		if node.ShardCoordinator.SelfId() != proposerNode.ShardCoordinator.SelfId() {
			continue
		}

		fmt.Printf("Testing roothash for node index %d, shard ID %d...\n", i, node.ShardCoordinator.SelfId())
		nodeRootHash, _ := node.AccntState.RootHash()
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
	serviceID string,
) []*TestProcessorNode {
	//first node generated will have is pk belonging to firstSkShardId
	numMetaChainNodes := 1
	nodes := make([]*TestProcessorNode, numOfShards*nodesPerShard+numMetaChainNodes)

	idx := 0
	for shardId := 0; shardId < numOfShards; shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			node := NewTestProcessorNode(uint32(numOfShards), uint32(shardId), uint32(shardId), serviceID)

			nodes[idx] = node
			idx++
		}
	}

	for i := 0; i < numMetaChainNodes; i++ {
		metaNode := NewTestProcessorNode(uint32(numOfShards), sharding.MetachainShardId, 0, serviceID)
		idx := i + numOfShards*nodesPerShard
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
}

// GenerateAndDisseminateTxs generates and sends multiple txs
func GenerateAndDisseminateTxs(
	n *TestProcessorNode,
	senders []crypto.PrivateKey,
	receiversPrivateKeys map[uint32][]crypto.PrivateKey,
	valToTransfer *big.Int,
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := 0
		for _, recvPrivateKeys := range receiversPrivateKeys {
			receiverKey := recvPrivateKeys[i]
			tx := generateTx(
				senderKey,
				n.OwnAccount.SingleSigner,
				&txArgs{
					nonce:   incrementalNonce,
					value:   valToTransfer,
					rcvAddr: skToPk(receiverKey),
					sndAddr: skToPk(senderKey),
				},
			)
			_, _ = n.SendTransaction(tx)
			incrementalNonce++
		}
	}
}

type txArgs struct {
	nonce    int
	value    *big.Int
	rcvAddr  []byte
	sndAddr  []byte
	data     string
	gasPrice int
	gasLimit int
}

func generateTx(
	skSign crypto.PrivateKey,
	signer crypto.SingleSigner,
	args *txArgs,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    uint64(args.nonce),
		Value:    args.value,
		RcvAddr:  args.rcvAddr,
		SndAddr:  args.sndAddr,
		GasPrice: uint64(args.gasPrice),
		GasLimit: uint64(args.gasLimit),
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

// CreateBlockBodyAndHeader creates and returns the block body and header
func CreateBlockBodyAndHeader(t *testing.T, proposer *TestProcessorNode, round uint64, coordinator sharding.Coordinator) (data.BodyHandler, data.HeaderHandler) {
	haveTime := func() bool { return true }
	var err error
	var blockHeader data.HeaderHandler
	var blockBody data.BodyHandler

	if coordinator.SelfId() == sharding.MetachainShardId {
		blockHeader, err = proposer.BlockProcessor.CreateBlockHeader(nil, round, haveTime)
		assert.Nil(t, err)
	} else {
		blockBody, err = proposer.BlockProcessor.CreateBlockBody(round, haveTime)
		assert.Nil(t, err)

		blockHeader, err = proposer.BlockProcessor.CreateBlockHeader(blockBody, round, haveTime)
		assert.Nil(t, err)
	}

	blockHeader.SetRound(round)
	blockHeader.SetNonce(round)
	blockHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := TestMultiSig.AggregateSigs(nil)
	blockHeader.SetSignature(sig)
	currHdr := proposer.BlockChain.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = proposer.BlockChain.GetGenesisHeader()
	}
	buff, _ := TestMarshalizer.Marshal(currHdr)
	blockHeader.SetPrevHash(TestHasher.Compute(string(buff)))
	blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	blockHeader.SetRandSeed(sig)

	return blockBody, blockHeader
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

// EqualSlices checks if two slices are equal
func EqualSlices(slice1 [][]byte, slice2 [][]byte) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	//check slice1 has all elements in slice2
	for _, buff1 := range slice1 {
		found := false
		for _, buff2 := range slice2 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	//check slice2 has all elements in slice1
	for _, buff2 := range slice2 {
		found := false
		for _, buff1 := range slice1 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GeneratePrivateKeyInShardId generates a private key that is in a given shard
func GeneratePrivateKeyInShardId(
	coordinator sharding.Coordinator,
	shardId uint32,
) crypto.PrivateKey {

	suite := kyber.NewBlakeSHA256Ed25519()
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		buff, _ := pk.ToByteArray()
		addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(buff)

		if coordinator.ComputeId(addr) == shardId {
			return sk
		}

		sk, pk = keyGen.GeneratePair()
	}
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
