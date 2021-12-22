package txsSender

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxsSender_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerFake{}

	mutRecoveredTransactions := &sync.RWMutex{}
	recoveredTransactions := make(map[uint32][]*transaction.Transaction)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		items := strings.Split(string(address), "Shard")
		sId, _ := strconv.ParseUint(items[1], 2, 32)
		return uint32(sId)
	}

	var txsToSend []*transaction.Transaction
	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     10,
		Value:     big.NewInt(15),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  5,
		GasLimit:  11,
		Data:      []byte(""),
		Signature: []byte("sig0"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     11,
		Value:     big.NewInt(25),
		RcvAddr:   []byte("receiverShard1"),
		SndAddr:   []byte("senderShard0"),
		GasPrice:  6,
		GasLimit:  12,
		Data:      []byte(""),
		Signature: []byte("sig1"),
	})

	txsToSend = append(txsToSend, &transaction.Transaction{
		Nonce:     12,
		Value:     big.NewInt(35),
		RcvAddr:   []byte("receiverShard0"),
		SndAddr:   []byte("senderShard1"),
		GasPrice:  7,
		GasLimit:  13,
		Data:      []byte(""),
		Signature: []byte("sig2"),
	})

	wg := sync.WaitGroup{}
	wg.Add(len(txsToSend))

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	mes := &p2pmocks.MessengerStub{
		BroadcastOnChannelBlockingCalled: func(pipe string, topic string, buff []byte) error {

			b := &batch.Batch{}
			err := marshalizer.Unmarshal(b, buff)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			for _, txBuff := range b.Data {
				tx := transaction.Transaction{}
				errMarshal := marshalizer.Unmarshal(&tx, txBuff)
				require.Nil(t, errMarshal)

				mutRecoveredTransactions.Lock()
				sId := shardCoordinator.ComputeId(tx.SndAddr)
				recoveredTransactions[sId] = append(recoveredTransactions[sId], &tx)
				mutRecoveredTransactions.Unlock()

				wg.Done()
			}
			return nil
		},
	}

	sender, _ := NewTxsSenderWithAccumulator(ArgsTxsSenderWithAccumulator{
		Marshaller:       &testscommon.MarshalizerMock{},
		ShardCoordinator: shardCoordinator,
		NetworkMessenger: mes,
		AccumulatorConfig: config.TxAccumulatorConfig{
			MaxAllowedTimeInMilliseconds:   250,
			MaxDeviationTimeInMilliseconds: 25,
		},
	})

	transactions := make([]data.TransactionHandler, 0)
	for _, tx := range txsToSend {
		transactions = append(transactions, tx)
	}
	numTxs, err := sender.SendBulkTransactions(transactions)
	assert.Equal(t, len(txsToSend), int(numTxs))
	assert.Nil(t, err)

	// we need to wait a little bit as the node.printTxSentCounter should iterate and avoid different code coverage computation
	time.Sleep(time.Second + time.Millisecond*500)
	var timeoutWait = time.Second
	select {
	case <-chDone:
	case <-time.After(timeoutWait):
		assert.Fail(t, "timout while waiting the broadcast of the generated transactions")
		return
	}

	mutRecoveredTransactions.RLock()
	// check if all txs were recovered and are assigned to correct shards
	recTxsSize := 0
	for sId, txsSlice := range recoveredTransactions {
		for _, tx := range txsSlice {
			if !strings.Contains(string(tx.SndAddr), fmt.Sprint(sId)) {
				assert.Fail(t, "txs were not distributed correctly to shards")
			}
			recTxsSize++
		}
	}

	assert.Equal(t, len(txsToSend), recTxsSize)
	mutRecoveredTransactions.RUnlock()
}

func TestTxsSender_SendBulkTransactions2(t *testing.T) {
	sender, _ := NewTxsSenderWithAccumulator(ArgsTxsSenderWithAccumulator{
		Marshaller:       &testscommon.MarshalizerMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		NetworkMessenger: &p2pmocks.MessengerStub{},
		AccumulatorConfig: config.TxAccumulatorConfig{
			MaxAllowedTimeInMilliseconds:   250,
			MaxDeviationTimeInMilliseconds: 25,
		},
	})

	txs := make([]data.TransactionHandler, 0)

	numOfTxsProcessed, err := sender.SendBulkTransactions(txs)
	assert.Equal(t, uint64(0), numOfTxsProcessed)
	assert.Equal(t, process.ErrNoTxToProcess, err)
}
