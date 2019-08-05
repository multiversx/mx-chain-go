package transaction

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptBulkUnsignedTransactionsWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	dPool := createTestDataPool()
	startingNonce := uint64(6)
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	accntAdapter := createAccountsDB()
	shardCoordinator := &sharding.OneShardCoordinator{}
	nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		testHasher,
		0,
		1,
		make(map[uint32][]sharding.Validator),
	)
	n, mes, sk, _ := createNetNode(dPool, accntAdapter, shardCoordinator, nodesCoordinator)

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
	dPool.UnsignedTransactions().RegisterHandler(func(key []byte) {
		mut.Lock()
		defer mut.Unlock()

		unsignedtxHashes = append(unsignedtxHashes, key)

		dataStore := dPool.UnsignedTransactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
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
		testMarshalizer,
		shardCoordinator,
		mes,
	)

	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 60):
		assert.Fail(t, "timeout")
		return
	}

	checkResults(t, startingNonce, noOfUnsignedTx, unsignedtxHashes, unsignedTransactions, dPool.UnsignedTransactions(), shardCoordinator)
}

func generateAndSendBulkSmartContractResults(
	startingNonce uint64,
	noOfUnsignedTx int,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	messenger p2p.Messenger,
) error {

	dataPacker, err := partitioning.NewSizeDataPacker(marshalizer)
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
			Nonce:   nonce,
			TxHash:  []byte("tx hash"),
			SndAddr: sender,
			RcvAddr: dest,
			Value:   big.NewInt(0),
		}
		buff := make([]byte, 8)
		binary.BigEndian.PutUint64(buff, nonce)
		uTx.Data = hex.EncodeToString(buff)

		uTxBytes, _ := marshalizer.Marshal(uTx)
		unsigedTxs = append(unsigedTxs, uTxBytes)
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.UnsignedTransactionTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())

	packets, err := dataPacker.PackDataInChunks(unsigedTxs, core.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	for _, buff := range packets {
		go func(bufferToSend []byte) {
			messenger.BroadcastOnChannelBlocking(
				identifier,
				identifier,
				bufferToSend,
			)
		}(buff)
	}

	return nil
}
