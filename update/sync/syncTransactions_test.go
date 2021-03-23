package sync

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgsNewPendingTransactionsSyncer {
	return ArgsNewPendingTransactionsSyncer{
		DataPools: testscommon.NewPoolsHolderMock(),
		Storages: &mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return &mock.StorerStub{}
			},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestNewPendingTransactionsSyncer(t *testing.T) {
	t.Parallel()

	args := createMockArgs()

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, err)
	require.NotNil(t, pendingTxsSyncer)
	require.False(t, pendingTxsSyncer.IsInterfaceNil())
}

func TestNewPendingTransactionsSyncer_NilStorages(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = nil

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestNewPendingTransactionsSyncer_NilDataPools(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.DataPools = nil

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewPendingTransactionsSyncer_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Marshalizer = nil

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewPendingTransactionsSyncer_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.RequestHandler = nil

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, process.ErrNilRequestHandler, err)
}

func TestSyncPendingTransactionsFor(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					tx := &dataTransaction.Transaction{
						Nonce: 1, Value: big.NewInt(10), SndAddr: []byte("snd"), RcvAddr: []byte("rcv"),
					}
					return json.Marshal(tx)
				},
			}
		},
	}

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{[]byte("txHash")}}
	miniBlocks["key"] = mb
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Nil(t, err)
}

func TestSyncPendingTransactionsFor_MissingTxFromPool(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					dummy := 10
					return json.Marshal(dummy)
				},
			}
		},
	}

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{[]byte("txHash")}}
	miniBlocks["key"] = mb

	// we need a value larger than the request interval as to also test what happens after the normal request interval has expired
	timeout := time.Second + time.Millisecond*500
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Equal(t, process.ErrTimeIsOut, err)
}

func TestSyncPendingTransactionsFor_ReceiveMissingTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	args := createMockArgs()
	args.Storages = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					dummy := 10
					return json.Marshal(dummy)
				},
			}
		},
	}

	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	miniBlocks["key"] = mb

	go func() {
		time.Sleep(500 * time.Millisecond)

		tx := &dataTransaction.Transaction{
			Nonce: 1, Value: big.NewInt(10), SndAddr: []byte("snd"), RcvAddr: []byte("rcv"),
		}
		pendingTxsSyncer.txPools[block.TxBlock].AddData(txHash, tx, tx.Size(), "0")

		pendingTxsSyncer.receivedTransaction(txHash, tx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Nil(t, err)
}
