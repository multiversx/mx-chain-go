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

func createMockHistoryRepoArgs() HistoryRepositoryArguments {
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

	args := createMockHistoryRepoArgs()
	args.HistoryStorer = nil

	repo, err := NewHistoryRepository(args)
	assert.Nil(t, repo)
	assert.Equal(t, core.ErrNilStore, err)
}

func TestNewHistoryRepository_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
	args.Hasher = nil

	repo, err := NewHistoryRepository(args)
	assert.Nil(t, repo)
	assert.Equal(t, core.ErrNilHasher, err)
}

func TestNewHistoryRepository_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
	args.Marshalizer = nil

	repo, err := NewHistoryRepository(args)
	assert.Nil(t, repo)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestNewHistoryRepository_NilHashEpochStorerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
	args.HashEpochStorer = nil

	repo, err := NewHistoryRepository(args)
	assert.Nil(t, repo)
	assert.Equal(t, core.ErrNilStore, err)
}

func TestNewHistoryRepository(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()

	repo, err := NewHistoryRepository(args)
	assert.NotNil(t, repo)
	assert.NoError(t, err)
	assert.True(t, repo.IsEnabled())
	assert.False(t, repo.IsInterfaceNil())
}

func TestHistoryRepository_PutTransactionsData(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	countCalledHashEpoch := 0
	args := createMockHistoryRepoArgs()
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

	repo, _ := NewHistoryRepository(args)

	headerHash := []byte("headerHash")
	txsData := &HistoryTransactionsData{
		BlockHeaderHash: headerHash,
		BlockHeader: &block.Header{
			Epoch: 0,
		},
		BlockBody: &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes:        [][]byte{txHash},
					SenderShardID:   0,
					ReceiverShardID: 1,
				},
			},
		},
	}

	err := repo.PutTransactionsData(txsData)
	assert.Nil(t, err)
	assert.Equal(t, 3, countCalledHashEpoch)
}

func TestHistoryRepository_GetTransaction(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	args := createMockHistoryRepoArgs()
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

	repo, _ := NewHistoryRepository(args)

	historyTx, err := repo.GetTransactionsGroupMetadata([]byte("txHash"))
	assert.Nil(t, err)
	assert.Equal(t, round, historyTx.Round)
	assert.Equal(t, epoch, historyTx.Epoch)
}

func TestHistoryRepository_GetEpochForHash(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	args := createMockHistoryRepoArgs()
	args.HashEpochStorer = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			hashEpochData := EpochByHash{
				Epoch: epoch,
			}

			hashEpochBytes, _ := json.Marshal(hashEpochData)
			return hashEpochBytes, nil
		},
	}
	repo, _ := NewHistoryRepository(args)

	resEpoch, err := repo.GetEpochByHash([]byte("txHash"))
	assert.NoError(t, err)
	assert.Equal(t, epoch, resEpoch)
}
