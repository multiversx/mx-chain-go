package sync

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgsNewPendingTransactionsSyncer {
	return ArgsNewPendingTransactionsSyncer{
		DataPools:      mock.NewPoolsHolderMock(),
		Storages:       &mock.ChainStorerMock{},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestNewPendingTransactionsSyncer(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = &mock.ChainStorerMock{GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
		return &mock.StorerStub{}
	}}
	pendingTxsSyncer, err := NewPendingTransactionsSyncer(args)
	require.Nil(t, err)
	require.NotNil(t, pendingTxsSyncer)
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
	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, time.Second)
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

	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, time.Second)
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
		pendingTxsSyncer.txPools[block.TxBlock].AddData(txHash, tx, "0")

		pendingTxsSyncer.receivedTransaction(txHash)
	}()

	err = pendingTxsSyncer.SyncPendingTransactionsFor(miniBlocks, 1, time.Second)
	require.Nil(t, err)
}
