package transaction

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptBulkTransactionsWithMemMessenger(t *testing.T) {
	dPool := createTestDataPool()

	startingNonce := uint64(6)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")
	accntAdapter := adbCreateAccountsDB()

	//TODO change when injecting a messenger is possible
	n, _, sk, _ := createMemNode(1, dPool, accntAdapter)

	n.Start()
	defer func() { _ = n.Stop() }()

	defer p2p.ReInitializeGloballyRegisteredPeers()

	//set the account's nonce to startingNonce
	nodePubKeyBytes, _ := sk.GeneratePublic().ToByteArray()
	nodeAddress, _ := addrConverter.CreateAddressFromPublicKeyBytes(nodePubKeyBytes)
	nodeAccount, _ := accntAdapter.GetJournalizedAccount(nodeAddress)
	nodeAccount.SetNonceWithJournal(startingNonce)
	accntAdapter.Commit()

	noOfTx := 50

	wg := sync.WaitGroup{}
	wg.Add(noOfTx)

	chanDone := make(chan bool)

	go func() {
		wg.Wait()

		chanDone <- true
	}()

	mut := sync.Mutex{}
	txHashes := make([][]byte, 0)

	//wire up handler
	dPool.Transactions().RegisterHandler(func(key []byte) {
		mut.Lock()
		defer mut.Unlock()

		txHashes = append(txHashes, key)
		wg.Done()
	})

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), uint64(noOfTx))
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
		return
	}

	assert.Equal(t, noOfTx, len(txHashes))

	bitmap := make([]bool, noOfTx+int(startingNonce))
	//set for each nonce from found tx a true flag in bitmap
	for i := 0; i < noOfTx; i++ {
		tx, _ := dPool.Transactions().ShardDataStore(0).Get(txHashes[i])

		assert.NotNil(t, tx)
		bitmap[tx.(*transaction.Transaction).Nonce] = true
	}

	//for the first startingNonce values, the bitmap should be false
	//for the rest, true
	for i := 0; i < noOfTx+int(startingNonce); i++ {
		if i < int(startingNonce) {
			assert.False(t, bitmap[i])
			continue
		}

		assert.True(t, bitmap[i])
	}
}
