package factory_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getArgs() factory.TrieFactoryArgs {
	return factory.TrieFactoryArgs{
		Marshalizer:              &testscommon.MarshalizerMock{},
		Hasher:                   &hashingMocks.HasherMock{},
		PathManager:              &testscommon.PathManagerStub{},
		TrieStorageManagerConfig: config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
	}
}

func getCreateArgs() factory.TrieCreateArgs {
	return factory.TrieCreateArgs{
		MainStorer:         testscommon.CreateMemUnit(),
		CheckpointsStorer:  testscommon.CreateMemUnit(),
		PruningEnabled:     false,
		CheckpointsEnabled: false,
		MaxTrieLevelInMem:  5,
		IdleProvider:       &testscommon.ProcessStatusHandlerStub{},
	}
}

func TestNewTrieFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Marshalizer = nil
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Hasher = nil
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.PathManager = nil
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilPathManager, err)
}

func TestNewTrieFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()

	tf, err := factory.NewTrieFactory(args)
	require.Nil(t, err)
	require.False(t, check.IfNil(tf))
}

func TestTrieFactory_CreateWithoutPruningShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	_, tr, err := tf.Create(getCreateArgs())
	require.NotNil(t, tr)
	require.Nil(t, err)
}

func TestTrieCreator_CreateWithPruningShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	_, tr, err := tf.Create(createArgs)
	require.Nil(t, err)
	require.NotNil(t, tr)
}

func TestTrieCreator_CreateWithoutCheckpointShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	createArgs.CheckpointsEnabled = true
	_, tr, err := tf.Create(createArgs)
	require.NotNil(t, tr)
	require.Nil(t, err)
}

func TestTrieCreator_CreateWithNilMainStorerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	createArgs.MainStorer = nil
	_, tr, err := tf.Create(createArgs)
	require.Nil(t, tr)
	require.True(t, strings.Contains(err.Error(), trie.ErrNilStorer.Error()))
}

func TestTrieCreator_CreateWithNilCheckpointsStorerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	createArgs.CheckpointsStorer = nil
	_, tr, err := tf.Create(createArgs)
	require.Nil(t, tr)
	require.True(t, strings.Contains(err.Error(), trie.ErrNilStorer.Error()))
}

func TestTrieCreator_CreateTriesComponentsForShardIdMissingStorer(t *testing.T) {
	t.Parallel()

	t.Run("missing UserAccountsUnit", testWithMissingStorer(dataRetriever.UserAccountsUnit))
	t.Run("missing UserAccountsCheckpointsUnit", testWithMissingStorer(dataRetriever.UserAccountsCheckpointsUnit))
	t.Run("missing PeerAccountsUnit", testWithMissingStorer(dataRetriever.PeerAccountsUnit))
	t.Run("missing PeerAccountsCheckpointsUnit", testWithMissingStorer(dataRetriever.PeerAccountsCheckpointsUnit))
}

func testWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		holder, storageManager, err := factory.CreateTriesComponentsForShardId(
			testscommon.GetGeneralConfig(),
			&mock.CoreComponentsStub{
				InternalMarshalizerField: &testscommon.MarshalizerMock{},
				HasherField:              &hashingMocks.HasherMock{},
				PathHandlerField:         &testscommon.PathManagerStub{},
			},
			&storageStubs.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					if unitType == missingUnit {
						return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
					}
					return &storageStubs.StorerStub{}, nil
				},
			})
		require.True(t, check.IfNil(holder))
		require.Nil(t, storageManager)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
	}
}
