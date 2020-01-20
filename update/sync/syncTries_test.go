package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewSyncState_NilTrieSyncersShouldErr(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncTriesHandler{
		TrieSyncers: nil,
		ActiveTries: &mock.TriesHolderMock{},
	}

	triesSyncHandler, err := NewSyncTriesHandler(args)
	require.Nil(t, triesSyncHandler)
	require.Equal(t, update.ErrNilTrieSyncers, err)
}

func TestNewSyncState_NilActiveTriesShouldErr(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncTriesHandler{
		TrieSyncers: &mock.TrieSyncersStub{},
		ActiveTries: nil,
	}

	triesSyncHandler, err := NewSyncTriesHandler(args)
	require.Nil(t, triesSyncHandler)
	require.Equal(t, update.ErrNilActiveTries, err)
}

func TestNewSyncState(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncTriesHandler{
		TrieSyncers: &mock.TrieSyncersStub{
			GetCalled: func(key string) (syncer update.TrieSyncer, err error) {
				return &mock.TrieSyncersStub{}, nil
			},
		},
		ActiveTries: &mock.TriesHolderMock{
			GetCalled: func(bytes []byte) data.Trie {
				return &mock.TrieStub{
					RecreateCalled: func(root []byte) (trie data.Trie, err error) {
						return &mock.TrieStub{
							CommitCalled: func() error {
								return nil
							},
						}, nil
					},
				}
			},
		},
	}

	triesSyncHandler, err := NewSyncTriesHandler(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash")},
			},
		},
	}

	err = triesSyncHandler.SyncTriesFrom(metaBlock, time.Second)
	require.Nil(t, err)

	tries, err := triesSyncHandler.GetTries()
	assert.NotNil(t, tries)
	assert.Nil(t, err)
}
