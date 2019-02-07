package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTransactionWithMemMessenger(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPool := createTestDataPool()

	n, _, sk, _ := createMemNode(1, dPool, adbCreateAccountsDB())

	n.Start()
	defer func() { _ = n.Stop() }()

	defer p2p.ReInitializeGloballyRegisteredPeers()

	pkBuff, _ := sk.GeneratePublic().ToByteArray()

	//Step 1. Generate a transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: pkBuff,
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
