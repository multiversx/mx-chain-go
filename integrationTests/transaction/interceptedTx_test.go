package transaction

import (
	"context"
	"encoding/hex"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTransaction(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	keyGen := schnorr.NewKeyGenerator()
	sk, pk := keyGen.GeneratePair()
	buffPk, _ := pk.ToByteArray()

	dPool := createTestDataPool()

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	//TODO change when injecting a messenger is possible
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(createTestBlockChain()),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	n.Start()
	defer n.Stop()

	_ = n.BindInterceptorsResolvers()

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
	_, err := n.SendTransaction(tx.Nonce, hex.EncodeToString(tx.SndAddr), hex.EncodeToString(tx.RcvAddr),
		tx.Value, string(tx.Data), tx.Signature)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
