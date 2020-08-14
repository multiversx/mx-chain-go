package fullHistory

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockHistoryRepoArgs() HistoryRepositoryArguments {
	return HistoryRepositoryArguments{
		SelfShardID:                 0,
		MiniblocksMetadataStorer:    &mock.StorerStub{},
		MiniblockHashByTxHashStorer: &mock.StorerStub{},
		EpochByHashStorer:           &mock.StorerStub{},
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &mock.HasherMock{},
	}
}

func TestNewHistoryRepository(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
	args.MiniblocksMetadataStorer = nil
	repo, err := NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.MiniblockHashByTxHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.EpochByHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.Hasher = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilHasher, err)

	args = createMockHistoryRepoArgs()
	args.Marshalizer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilMarshalizer, err)

	args = createMockHistoryRepoArgs()
	repo, err = NewHistoryRepository(args)
	require.Nil(t, err)
	require.NotNil(t, repo)
}

func TestHistoryRepository_PutTransactionsData(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	countCalledHashEpoch := 0
	args := createMockHistoryRepoArgs()
	args.EpochByHashStorer = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			countCalledHashEpoch++
			return nil
		},
	}
	args.MiniblocksMetadataStorer = &mock.StorerStub{
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
	args.EpochByHashStorer = &mock.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			hashEpochData := EpochByHash{
				Epoch: epoch,
			}

			hashEpochBytes, _ := json.Marshal(hashEpochData)
			return hashEpochBytes, nil
		},
	}

	round := uint64(1000)
	args.MiniblocksMetadataStorer = &mock.StorerStub{
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
	args.EpochByHashStorer = &mock.StorerStub{
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
