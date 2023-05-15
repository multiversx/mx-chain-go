package syncer_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	t.Run("invalid base args (nil hasher) should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Hasher = nil

		syncer, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("invalid timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Timeout = 0

		s, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, s)
		assert.True(t, errors.Is(err, common.ErrInvalidTimeout))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		v, err := syncer.NewValidatorAccountsSyncer(args)
		require.Nil(t, err)
		require.NotNil(t, v)
	})
}

func TestValidatorAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
	}

	key := []byte("rootHash")
	serializedLeafNode := getSerializedTrieNode(key, args.Marshalizer, args.Hasher)
	itn, err := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
	require.Nil(t, err)

	args.TrieStorageManager = &testscommon.StorageManagerStub{
		GetCalled: func(b []byte) ([]byte, error) {
			return serializedLeafNode, nil
		},
	}

	cacher := testscommon.NewCacherMock()
	cacher.Put(key, itn, 0)
	args.Cacher = cacher

	v, err := syncer.NewValidatorAccountsSyncer(args)
	require.Nil(t, err)

	err = v.SyncAccounts(key)
	require.Nil(t, err)
}

func TestValidatorAccountsSyncer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var vas *syncer.ValidatorAccountsSyncer
	assert.True(t, vas.IsInterfaceNil())

	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
	}
	vas, err := syncer.NewValidatorAccountsSyncer(args)
	require.Nil(t, err)
	assert.False(t, vas.IsInterfaceNil())
}
