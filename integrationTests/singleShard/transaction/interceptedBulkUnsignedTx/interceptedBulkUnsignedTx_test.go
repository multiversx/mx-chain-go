package interceptedBulkUnsignedTx

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptBulkUnsignedTransactionsWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startingNonce := uint64(6)

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	n := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            numOfShards,
		NodeShardId:          shardID,
		TxSignPrivKeyShardId: txSignPrivKeyShardId,
	})

	defer func() {
		n.Close()
	}()

	time.Sleep(integrationTests.P2pBootstrapDelay)

	//set the account's nonce to startingNonce
	_ = n.SetAccountNonce(startingNonce)
	noOfUnsignedTx := 8000

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(noOfUnsignedTx)

	chanDone := make(chan bool)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	mut := sync.Mutex{}
	unsignedtxHashes := make([][]byte, 0)
	unsignedTransactions := make([]data.TransactionHandler, 0)

	//wire up handler
	n.DataPool.UnsignedTransactions().RegisterOnAdded(func(key []byte, value interface{}) {
		mut.Lock()
		defer mut.Unlock()

		unsignedtxHashes = append(unsignedtxHashes, key)

		dataStore := n.DataPool.UnsignedTransactions().ShardDataStore(
			process.ShardCacherIdentifier(n.ShardCoordinator.SelfId(), n.ShardCoordinator.SelfId()),
		)
		val, _ := dataStore.Get(key)
		if val == nil {
			assert.Fail(t, fmt.Sprintf("key %s not in store?", base64.StdEncoding.EncodeToString(key)))
			return
		}

		unsignedTransactions = append(unsignedTransactions, val.(*smartContractResult.SmartContractResult))
		wg.Done()
	})

	err := generateAndSendBulkSmartContractResults(
		startingNonce,
		noOfUnsignedTx,
		integrationTests.TestMarshalizer,
		n.ShardCoordinator,
		n.MainMessenger,
	)

	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 60):
		assert.Fail(t, "timeout")
		return
	}

	integrationTests.CheckTxPresentAndRightNonce(
		t,
		startingNonce,
		noOfUnsignedTx,
		unsignedtxHashes,
		unsignedTransactions,
		n.DataPool.UnsignedTransactions(),
		n.ShardCoordinator,
	)
}

func generateAndSendBulkSmartContractResults(
	startingNonce uint64,
	noOfUnsignedTx int,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	messenger p2p.Messenger,
) error {

	dataPacker, err := partitioning.NewSimpleDataPacker(marshalizer)
	if err != nil {
		return err
	}

	sender := make([]byte, 32)
	_, _ = rand.Reader.Read(sender)

	dest := make([]byte, 32)
	_, _ = rand.Reader.Read(dest)

	unsigedTxs := make([][]byte, 0)
	for nonce := startingNonce; nonce < startingNonce+uint64(noOfUnsignedTx); nonce++ {
		uTx := &smartContractResult.SmartContractResult{
			Nonce:      nonce,
			PrevTxHash: []byte("tx hash"),
			SndAddr:    sender,
			RcvAddr:    dest,
			Value:      big.NewInt(0),
		}
		buff := make([]byte, 8)
		binary.BigEndian.PutUint64(buff, nonce)
		uTx.Data = buff

		uTxBytes, _ := marshalizer.Marshal(uTx)
		unsigedTxs = append(unsigedTxs, uTxBytes)
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.UnsignedTransactionTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())

	packets, err := dataPacker.PackDataInChunks(unsigedTxs, common.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	for _, buff := range packets {
		messenger.BroadcastOnChannel(
			identifier,
			identifier,
			buff,
		)
	}

	return nil
}
