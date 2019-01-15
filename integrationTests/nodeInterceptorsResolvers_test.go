package integrationTests_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTransaction(t *testing.T) {
	hasher := mock.HasherFake{}
	marshalizer := &mock.MarshalizerFake{}

	keyGen := schnorr.NewKeyGenerator()
	sk, pk := keyGen.GeneratePair()
	buffPk, _ := pk.ToByteArray()

	dPool := createTestDataPool()

	addrConverter := mock.NewAddressConverterFake(32, "0x")

	//TODO change when injecting a messenger is possible
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(createTestBlockChain()),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	n.Start()
	defer n.Stop()

	err := n.BindInterceptorsResolvers()
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	//Step 1. Generate a transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   *big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: buffPk,
		Data:    []byte("tx notarized data"),
	}

	//Step 2. Sign transaction
	txBuff, _ := marshalizer.Marshal(&tx)
	tx.Signature, _ = sk.Sign(txBuff)

	signedTxBuff, _ := marshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(signedTxBuff))

	chanDone := make(chan bool)

	//step 3. wire up a received handler
	dPool.Transactions().RegisterHandler(func(key []byte) {
		txStored, _ := dPool.Transactions().ShardDataStore(0).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)

	})

	//Step 4. Send Tx
	_, err = n.SendTransaction(tx.Nonce, hex.EncodeToString(tx.SndAddr), hex.EncodeToString(tx.RcvAddr),
		tx.Value, string(tx.Data), string(tx.Signature))
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}

func createTestDataPool() data.TransientDataHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100, Type: storage.LRUCache})
	hdrPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100, Type: storage.LRUCache})

	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	stateBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	dPool, _ := dataPool.NewDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		stateBlockBody,
	)

	return dPool
}

func createTestBlockChain() *blockchain.BlockChain {

	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}

	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)

	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit())

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func TestNode_GenerateSendInterceptHeaderByNonce(t *testing.T) {
	hasher := mock.HasherFake{}
	marshalizer := &mock.MarshalizerFake{}

	keyGen := schnorr.NewKeyGenerator()
	//sk, pk := keyGen.GeneratePair()
	//buffPk, _ := pk.ToByteArray()

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	addrConverter := mock.NewAddressConverterFake(32, "0x")

	blkcRequestor := createTestBlockChain()
	blkcResolver := createTestBlockChain()

	nRequestor, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
		node.WithDataPool(dPoolRequestor),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(blkcRequestor),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	nResolver, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(&mock.AccountsAdapterStub{}),
		node.WithPrivateKey(&mock.PrivateKeyStub{}),
		node.WithDataPool(dPoolResolver),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(blkcResolver),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	fmt.Println("Requestor:")
	nRequestor.Start()
	fmt.Println("Resolver:")
	nResolver.Start()
	nRequestor.P2PBootstrap()
	nResolver.P2PBootstrap()
	defer nRequestor.Stop()
	defer nResolver.Stop()

	time.Sleep(time.Second)

	err := nRequestor.BindInterceptorsResolvers()
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	err = nResolver.BindInterceptorsResolvers()
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	//we wait 1 second to be sure that both nodes advertised their topics
	//TODO change when injecting a messenger is possible
	time.Sleep(time.Second)

	//Step 1. Generate a header
	hdr := block.Header{
		Nonce:         0,
		PubKeysBitmap: []byte{255, 0},
		Commitment:    []byte("commitment"),
		Signature:     []byte("signature"),
		BlockBodyHash: []byte("block body hash"),
		PrevHash:      []byte("prev hash"),
		TimeStamp:     uint64(time.Now().Unix()),
		Round:         1,
		Epoch:         2,
		ShardId:       0,
		BlockBodyType: block.TxBlock,
	}

	hdrBuff, _ := marshalizer.Marshal(&hdr)
	hdrHash := hasher.Compute(string(hdrBuff))

	//Step 2. resolver has the header
	dPoolResolver.Headers().AddData(hdrHash, &hdr, 0)
	dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash)

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequestor.Headers().ShardDataStore(0).Get(key)

		if reflect.DeepEqual(hdrStored, &hdr) && hdr.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, hdrStored, &hdr)

	})

	//Step 4. request header
	hdrResolver := nRequestor.GetResolvers()[1].(*block2.HeaderResolver)
	hdrResolver.RequestHeaderFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
