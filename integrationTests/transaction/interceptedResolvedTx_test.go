package transaction

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTransaction(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	keyGen := schnorr.NewKeyGenerator()
	sk, pk := keyGen.GeneratePair()
	buffPk, _ := pk.ToByteArray()

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	blkcRequestor := createTestBlockChain()
	blkcResolver := createTestBlockChain()

	nRequestor, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
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
		node.WithDataPool(dPoolResolver),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(blkcResolver),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	nRequestor.Start()
	nResolver.Start()
	defer nRequestor.Stop()
	defer nResolver.Stop()

	nRequestor.P2PBootstrap()
	nResolver.P2PBootstrap()

	time.Sleep(time.Second)

	_ = nRequestor.BindInterceptorsResolvers()
	_ = nResolver.BindInterceptorsResolvers()

	//we wait 1 second to be sure that both nodes advertised their topics
	//TODO change when injecting a messenger is possible
	time.Sleep(time.Second)

	//Step 1. Generate a signed transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   *big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: buffPk,
		Data:    []byte("tx notarized data"),
	}

	txBuff, _ := marshalizer.Marshal(&tx)
	tx.Signature, _ = sk.Sign(txBuff)

	signedTxBuff, _ := marshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(signedTxBuff))

	chanDone := make(chan bool)

	txHash := hasher.Compute(string(signedTxBuff))

	//step 2. wire up a received handler for requestor
	dPoolRequestor.Transactions().RegisterHandler(func(key []byte) {
		txStored, _ := dPoolRequestor.Transactions().ShardDataStore(0).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	dPoolResolver.Transactions().AddData(txHash, &tx, 0)

	//Step 4. request tx
	txResolver := nRequestor.GetResolvers()[0].(*transaction2.TxResolver)
	err := txResolver.RequestTransactionFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
