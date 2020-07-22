package fullHistory

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func createMockHistoryProcArgs() HistoryRepositoryArguments {
	return HistoryRepositoryArguments{
		Marshalizer:     &mock.MarshalizerMock{},
		Hasher:          &mock.HasherMock{},
		HistoryStorer:   &mock.StorerStub{},
		HashEpochStorer: &mock.StorerStub{},
		SelfShardID:     0,
	}
}

func TestNewHistoryRepository_NilHistoryStorerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryProcArgs()
	args.HistoryStorer = nil

	proc, err := NewHistoryRepository(args)
	assert.Nil(t, proc)
	assert.Equal(t, core.ErrNilStore, err)
}

func TestNewHistoryRepository_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryProcArgs()
	args.Hasher = nil

	proc, err := NewHistoryRepository(args)
	assert.Nil(t, proc)
	assert.Equal(t, core.ErrNilHasher, err)
}

func TestNewHistoryRepository_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryProcArgs()
	args.Marshalizer = nil

	proc, err := NewHistoryRepository(args)
	assert.Nil(t, proc)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestNewHistoryRepository_NilHashEpochStorerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryProcArgs()
	args.HashEpochStorer = nil

	proc, err := NewHistoryRepository(args)
	assert.Nil(t, proc)
	assert.Equal(t, core.ErrNilStore, err)
}

func TestNewHistoryRepository(t *testing.T) {
	t.Parallel()

	args := createMockHistoryProcArgs()

	proc, err := NewHistoryRepository(args)
	assert.NotNil(t, proc)
	assert.NoError(t, err)
	assert.True(t, proc.IsEnabled())
	assert.False(t, proc.IsInterfaceNil())
}

func TestHistoryRepository_PutTransactionsData(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	countCalledHashEpoch := 0
	args := createMockHistoryProcArgs()
	args.HashEpochStorer = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			countCalledHashEpoch++
			return nil
		},
	}
	args.HistoryStorer = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			assert.True(t, bytes.Equal(txHash, key))
			return nil
		},
	}

	proc, _ := NewHistoryRepository(args)

	headerHash := []byte("headerHash")
	txsData := &HistoryTransactionsData{
		HeaderHash: headerHash,
		HeaderHandler: &block.Header{
			Epoch: 0,
		},
		BodyHandler: &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes:        [][]byte{txHash},
					SenderShardID:   0,
					ReceiverShardID: 1,
				},
			},
		},
	}

	err := proc.PutTransactionsData(txsData)
	assert.Nil(t, err)
	assert.Equal(t, 3, countCalledHashEpoch)
}

func TestHistoryRepository_GetTransaction(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	args := createMockHistoryProcArgs()
	args.HashEpochStorer = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			hashEpochData := EpochByHash{
				Epoch: epoch,
			}

			hashEpochBytes, _ := json.Marshal(hashEpochData)
			return hashEpochBytes, nil
		},
	}

	round := uint64(1000)
	args.HistoryStorer = &mock.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			if epoch == epoch {
				historyTx := &TransactionsGroupMetadata{
					Round: round,
				}
				historyTxBytes, _ := json.Marshal(historyTx)
				return historyTxBytes, nil
			}
			return nil, nil
		},
	}

	proc, _ := NewHistoryRepository(args)

	historyTx, err := proc.GetTransaction([]byte("txHash"))
	assert.Nil(t, err)
	assert.Equal(t, round, historyTx.Round)
	assert.Equal(t, epoch, historyTx.Epoch)
}

func TestHistoryRepository_GetEpochForHash(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	args := createMockHistoryProcArgs()
	args.HashEpochStorer = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			hashEpochData := EpochByHash{
				Epoch: epoch,
			}

			hashEpochBytes, _ := json.Marshal(hashEpochData)
			return hashEpochBytes, nil
		},
	}
	proc, _ := NewHistoryRepository(args)

	resEpoch, err := proc.GetEpochForHash([]byte("txHash"))
	assert.NoError(t, err)
	assert.Equal(t, epoch, resEpoch)
}
