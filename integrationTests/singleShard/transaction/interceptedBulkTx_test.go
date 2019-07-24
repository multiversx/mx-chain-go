package transaction

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptBulkTransactionsWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	dPool := createTestDataPool()

	startingNonce := uint64(6)

	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	accntAdapter := createAccountsDB()

	shardCoordinator := &sharding.OneShardCoordinator{}
	nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		hasher,
		0,
		1,
		make(map[uint32][]sharding.Validator),
	)
	n, _, sk, _ := createNetNode(dPool, accntAdapter, shardCoordinator, nodesCoordinator)

	_ = n.Start()
	defer func() {
		_ = n.Stop()
	}()

	_ = n.P2PBootstrap()

	time.Sleep(time.Second)

	//set the account's nonce to startingNonce
	nodePubKeyBytes, _ := sk.GeneratePublic().ToByteArray()
	nodeAddress, _ := addrConverter.CreateAddressFromPublicKeyBytes(nodePubKeyBytes)
	nodeAccount, _ := accntAdapter.GetAccountWithJournal(nodeAddress)
	_ = nodeAccount.(*state.Account).SetNonceWithJournal(startingNonce)
	_, _ = accntAdapter.Commit()

	noOfTx := 8000

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(noOfTx)

	chanDone := make(chan bool)

	go func() {
		wg.Wait()

		chanDone <- true
	}()

	mut := sync.Mutex{}
	txHashes := make([][]byte, 0)
	transactions := make([]*transaction.Transaction, 0)

	//wire up handler
	dPool.Transactions().RegisterHandler(func(key []byte) {
		mut.Lock()
		defer mut.Unlock()

		txHashes = append(txHashes, key)

		dataStore := dPool.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
		)
		val, _ := dataStore.Get(key)

		if val == nil {
			assert.Fail(t, fmt.Sprintf("key %s not in store?", base64.StdEncoding.EncodeToString(key)))
			return
		}

		transactions = append(transactions, val.(*transaction.Transaction))
		wg.Done()
	})

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), big.NewInt(1), uint64(noOfTx))

	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 60):
		assert.Fail(t, "timeout")
		return
	}

	if noOfTx != len(txHashes) {

		for i := startingNonce; i < startingNonce+uint64(noOfTx); i++ {
			found := false

			for _, tx := range transactions {
				if tx.Nonce == i {
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("tx with nonce %d is missing\n", i)
			}

		}

		assert.Fail(t, fmt.Sprintf("should have been %d, got %d", noOfTx, len(txHashes)))

		return
	}

	bitmap := make([]bool, noOfTx+int(startingNonce))
	//set for each nonce from found tx a true flag in bitmap
	for i := 0; i < noOfTx; i++ {
		val, _ := dPool.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
		).Get(txHashes[i])

		if val == nil {
			continue
		}

		tx := val.(*transaction.Transaction)

		bitmap[tx.Nonce] = true
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
