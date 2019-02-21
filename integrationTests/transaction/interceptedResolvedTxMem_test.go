package transaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTransactionWithMemMessenger(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	nRequestor, _, sk1, pf := createMemNode(1, dPoolRequestor, adbCreateAccountsDB())
	nResolver, _, _, _ := createMemNode(2, dPoolResolver, adbCreateAccountsDB())

	nRequestor.Start()
	nResolver.Start()
	defer func() {
		_ = nRequestor.Stop()
		_ = nResolver.Stop()
	}()

	defer p2p.ReInitializeGloballyRegisteredPeers()

	time.Sleep(time.Second)

	buffPk1, _ := sk1.GeneratePublic().ToByteArray()

	//Step 1. Generate a signed transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: buffPk1,
		Data:    []byte("tx notarized data"),
	}

	txBuff, _ := marshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}

	tx.Signature, _ = signer.Sign(sk1, txBuff)

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
	res, _ := pf.ResolverContainer().Get(string(factory.TransactionTopic))
	txResolver := res.(*transaction2.TxResolver)
	err := txResolver.RequestTransactionFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
