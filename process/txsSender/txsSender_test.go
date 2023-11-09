package txsSender

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	scrData "github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	epochStartMock "github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTxsSenderWithAccumulator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args          func() ArgsTxsSenderWithAccumulator
		expectedError error
	}{
		{
			args: func() ArgsTxsSenderWithAccumulator {
				args := generateMockArgsTxsSender()
				args.Marshaller = nil
				return args
			},
			expectedError: process.ErrNilMarshalizer,
		},
		{
			args: func() ArgsTxsSenderWithAccumulator {
				args := generateMockArgsTxsSender()
				args.ShardCoordinator = nil
				return args
			},
			expectedError: process.ErrNilShardCoordinator,
		},
		{
			args: func() ArgsTxsSenderWithAccumulator {
				args := generateMockArgsTxsSender()
				args.NetworkMessenger = nil
				return args
			},
			expectedError: process.ErrNilMessenger,
		},
		{
			args: func() ArgsTxsSenderWithAccumulator {
				args := generateMockArgsTxsSender()
				args.AccumulatorConfig = config.TxAccumulatorConfig{
					MaxAllowedTimeInMilliseconds:   0,
					MaxDeviationTimeInMilliseconds: 0,
				}
				return args
			},
			expectedError: core.ErrInvalidValue,
		},
		{
			args: func() ArgsTxsSenderWithAccumulator {
				args := generateMockArgsTxsSender()
				args.DataPacker = nil
				return args
			},
			expectedError: dataRetriever.ErrNilDataPacker,
		},
		{
			args: func() ArgsTxsSenderWithAccumulator {
				return generateMockArgsTxsSender()
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		instance, err := NewTxsSenderWithAccumulator(test.args())
		if test.expectedError == nil {
			require.NoError(t, err)
			require.False(t, instance.IsInterfaceNil())
			require.NoError(t, instance.Close())
		} else {
			require.Error(t, err)
			require.True(t, strings.Contains(err.Error(), test.expectedError.Error()))
			require.True(t, instance.IsInterfaceNil())
		}
	}
}

func TestTxsSender_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	mutRecoveredTransactions := &sync.RWMutex{}
	recoveredTransactions := make(map[uint32][]*transaction.Transaction)
	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
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
		BroadcastOnChannelCalled: func(pipe string, topic string, buff []byte) {

			b := &batch.Batch{}
			err := marshaller.Unmarshal(b, buff)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			for _, txBuff := range b.Data {
				tx := transaction.Transaction{}
				errMarshal := marshaller.Unmarshal(&tx, txBuff)
				require.Nil(t, errMarshal)

				mutRecoveredTransactions.Lock()
				sId := shardCoordinator.ComputeId(tx.SndAddr)
				recoveredTransactions[sId] = append(recoveredTransactions[sId], &tx)
				mutRecoveredTransactions.Unlock()

				wg.Done()
			}
		},
	}
	dataPacker, _ := partitioning.NewSimpleDataPacker(&marshallerMock.MarshalizerMock{})
	args := ArgsTxsSenderWithAccumulator{
		Marshaller:       &marshallerMock.MarshalizerMock{},
		ShardCoordinator: shardCoordinator,
		NetworkMessenger: mes,
		DataPacker:       dataPacker,
		AccumulatorConfig: config.TxAccumulatorConfig{
			MaxAllowedTimeInMilliseconds:   250,
			MaxDeviationTimeInMilliseconds: 25,
		},
	}
	sender, _ := NewTxsSenderWithAccumulator(args)

	numTxs, err := sender.SendBulkTransactions(txsToSend)
	assert.Equal(t, len(txsToSend), int(numTxs))
	assert.Nil(t, err)

	// we need to wait a little bit as the txsSender.printTxSentCounter should iterate and avoid different code coverage computation
	time.Sleep(time.Second + time.Millisecond*500)
	var timeoutWait = time.Second
	select {
	case <-chDone:
	case <-time.After(timeoutWait):
		assert.Fail(t, "timeout while waiting the broadcast of the generated transactions")
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

func TestTxsSender_sendFromTxAccumulatorSendOneTxOneSCRExpectOnlyTxToBeSent(t *testing.T) {
	t.Parallel()

	senderAddrTx := []byte("senderAddrTx")
	senderAddrSCR := []byte("senderAddrSCR")
	tx := &transaction.Transaction{SndAddr: senderAddrTx}
	scr := &scrData.SmartContractResult{SndAddr: senderAddrSCR}

	ctMarshallCalled := atomic.Counter{}
	ctComputeIdCalled := atomic.Counter{}
	ctCommunicationIdCalled := atomic.Counter{}
	ctPackDataCalled := atomic.Counter{}
	ctBroadCastCalled := atomic.Counter{}

	shardIDSenderAddrTx := uint32(1234)
	communicationIdentifier := "idTx"
	shardCoordinator := &epochStartMock.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			ctComputeIdCalled.Increment()
			require.Equal(t, senderAddrTx, address)
			return shardIDSenderAddrTx
		},
		CommunicationIdentifierCalled: func(destShardID uint32) string {
			ctCommunicationIdCalled.Increment()
			require.Equal(t, shardIDSenderAddrTx, destShardID)
			return communicationIdentifier
		},
	}

	txMarshalled := []byte("txMarshalled")
	marshaller := &marshallerMock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			ctMarshallCalled.Increment()
			require.Equal(t, tx, obj)
			return txMarshalled, nil

		},
	}

	txChunk := []byte("txChunk")
	dataPacker := &dataRetrieverMock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			ctPackDataCalled.Increment()
			require.Equal(t, [][]byte{txMarshalled}, data)
			return [][]byte{txChunk}, nil
		},
	}
	messenger := &p2pmocks.MessengerStub{
		BroadcastOnChannelCalled: func(channel string, topic string, buff []byte) {
			ctBroadCastCalled.Increment()
			require.Equal(t, SendTransactionsPipe, channel)
			require.Equal(t, factory.TransactionTopic+communicationIdentifier, topic)
			require.Equal(t, txChunk, buff)
		},
	}

	args := generateMockArgsTxsSender()
	args.ShardCoordinator = shardCoordinator
	args.Marshaller = marshaller
	args.NetworkMessenger = messenger
	args.DataPacker = dataPacker
	txsHandler, _ := NewTxsSenderWithAccumulator(args)
	defer func() {
		err := txsHandler.Close()
		require.Nil(t, err)
	}()

	txsHandler.txAccumulator.AddData(tx)
	txsHandler.txAccumulator.AddData(scr)

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(1), ctMarshallCalled.Get())
	require.Equal(t, int64(1), ctComputeIdCalled.Get())
	require.Equal(t, int64(1), ctCommunicationIdCalled.Get())
	require.Equal(t, int64(1), ctPackDataCalled.Get())
	require.Equal(t, int64(1), ctBroadCastCalled.Get())
}

func TestTxsSender_sendBulkTransactionsSendTwoTxsFailToMarshallOneExpectOnlyOneTxSend(t *testing.T) {
	t.Parallel()

	senderAddrTx1 := []byte("senderAddrTx1")
	senderAddrTx2 := []byte("senderAddrTx2")
	tx1 := &transaction.Transaction{SndAddr: senderAddrTx1}
	tx2 := &transaction.Transaction{SndAddr: senderAddrTx2}
	txs := []*transaction.Transaction{tx1, tx2}

	ctMarshallCalled := atomic.Counter{}
	ctComputeIdCalled := atomic.Counter{}
	ctCommunicationIdCalled := atomic.Counter{}
	ctPackDataCalled := atomic.Counter{}
	ctBroadCastCalled := atomic.Counter{}

	communicationIdentifierTx1 := "idTx1"
	shardIdTx1 := uint32(1234)
	shardCoordinator := &epochStartMock.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			ctComputeIdCalled.Increment()
			require.Equal(t, senderAddrTx1, address)
			return shardIdTx1
		},
		CommunicationIdentifierCalled: func(destShardID uint32) string {
			ctCommunicationIdCalled.Increment()
			require.Equal(t, shardIdTx1, destShardID)
			return communicationIdentifierTx1
		},
	}

	tx1Marshalled := []byte("tx1Marshalled")
	marshaller := &marshallerMock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			switch ctMarshallCalled.Get() {
			case 0:
				ctMarshallCalled.Increment()
				require.Equal(t, tx1, obj)
				return tx1Marshalled, nil

			case 1:
				ctMarshallCalled.Increment()
				require.Equal(t, tx2, obj)
				return nil, errors.New("error marshal tx2")
			default:
				require.Fail(t, "this marshaller should not be called for more than 2 txs")
				return nil, nil
			}
		},
	}

	tx1Chunk := []byte("tx1Chunk")
	dataPacker := &dataRetrieverMock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			ctPackDataCalled.Increment()
			require.Equal(t, [][]byte{tx1Marshalled}, data)
			return [][]byte{tx1Chunk}, nil
		},
	}
	messenger := &p2pmocks.MessengerStub{
		BroadcastOnChannelCalled: func(channel string, topic string, buff []byte) {
			ctBroadCastCalled.Increment()
			require.Equal(t, SendTransactionsPipe, channel)
			require.Equal(t, factory.TransactionTopic+communicationIdentifierTx1, topic)
			require.Equal(t, tx1Chunk, buff)
		},
	}

	args := generateMockArgsTxsSender()
	args.DataPacker = dataPacker
	args.NetworkMessenger = messenger
	args.Marshaller = marshaller
	args.ShardCoordinator = shardCoordinator

	txsHandler, _ := NewTxsSenderWithAccumulator(args)
	defer func() {
		err := txsHandler.Close()
		require.Nil(t, err)
	}()

	txsHandler.sendBulkTransactions(txs)
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(2), ctMarshallCalled.Get())
	require.Equal(t, int64(1), ctComputeIdCalled.Get())
	require.Equal(t, int64(1), ctCommunicationIdCalled.Get())
	require.Equal(t, int64(1), ctPackDataCalled.Get())
	require.Equal(t, int64(1), ctBroadCastCalled.Get())
}

func TestTxsSender_sendBulkTransactionsFromShardCannotPackDataExpectError(t *testing.T) {
	t.Parallel()

	txs := [][]byte{[]byte("tx")}
	errPackData := errors.New("error packing data in chunks")
	dataPacker := &dataRetrieverMock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			require.Equal(t, txs, data)
			return nil, errPackData
		},
	}

	args := generateMockArgsTxsSender()
	args.DataPacker = dataPacker
	txsHandler, _ := NewTxsSenderWithAccumulator(args)
	defer func() {
		err := txsHandler.Close()
		require.Nil(t, err)
	}()

	err := txsHandler.sendBulkTransactionsFromShard(txs, 0)
	require.Equal(t, errPackData, err)
}

func TestTxsSender_SendBulkTransactionsNoTxToProcessExpectError(t *testing.T) {
	t.Parallel()

	args := generateMockArgsTxsSender()
	txsHandler, _ := NewTxsSenderWithAccumulator(args)

	numOfProcessedTxs, err := txsHandler.SendBulkTransactions([]*transaction.Transaction{})
	assert.Equal(t, uint64(0), numOfProcessedTxs)
	assert.Equal(t, process.ErrNoTxToProcess, err)
}

func generateMockArgsTxsSender() ArgsTxsSenderWithAccumulator {
	marshaller := marshallerMock.MarshalizerMock{}
	dataPacker, _ := partitioning.NewSimpleDataPacker(marshaller)
	accumulatorConfig := config.TxAccumulatorConfig{
		MaxAllowedTimeInMilliseconds:   10,
		MaxDeviationTimeInMilliseconds: 1,
	}
	return ArgsTxsSenderWithAccumulator{
		Marshaller:        marshaller,
		ShardCoordinator:  &testscommon.ShardsCoordinatorMock{},
		NetworkMessenger:  &p2pmocks.MessengerStub{},
		DataPacker:        dataPacker,
		AccumulatorConfig: accumulatorConfig,
	}
}
