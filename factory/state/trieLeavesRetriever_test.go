package state

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	factoryMock "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestCreateTrieLeavesRetriever(t *testing.T) {
	tr := stateMock.GetNewTrie()
	t.Cleanup(func() { _ = tr.GetStorageManager().Close() })
	stateMock.AddDataToTrie(tr, 25)
	root, err := tr.RootHash()
	require.NoError(t, err)
	parser := &trieMock.TrieLeafParserStub{ParseLeafCalled: func(k, v []byte, _ core.TrieNodeVersion) (core.KeyValueHolder, error) {
		return keyValStorage.NewKeyValStorage(k, v), nil
	}}
	scf := &stateComponentsFactory{core: &factoryMock.CoreComponentsHolderStub{
		InternalMarshalizerCalled: func() marshal.Marshalizer { return &marshallerMock.MarshalizerMock{} },
		HasherCalled:              func() hashing.Hasher { return &hashingMocks.HasherMock{} },
	}}

	t.Run("always enabled, default budget", func(t *testing.T) {
		lr, err := scf.createTrieLeavesRetriever()
		require.NoError(t, err)
		next := [][]byte{root}
		all := make(map[string]string)
		visit := func(leaf core.KeyValueHolder) (bool, uint64, error) {
			require.NotContains(t, all, string(leaf.Key()))
			all[string(leaf.Key())] = string(leaf.Value())
			return true, uint64(len(leaf.Key()) + len(leaf.Value())), nil
		}
		for page := 0; len(next) > 0; page++ {
			require.Less(t, page, 5)
			next, err = lr.IterateLeaves(10, 0, next, tr.GetStorageManager(), parser, context.Background(), visit)
			require.NoError(t, err)
		}
		require.Len(t, all, 25)
	})
	t.Run("page byte budget comes from the trie operations config", func(t *testing.T) {
		scf.config.TrieOperations = config.TrieOperationsConfig{MaxBytes: 1}
		lr, err := scf.createTrieLeavesRetriever()
		require.NoError(t, err)
		numLeaves := 0
		visit := func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numLeaves++
			return true, uint64(len(leaf.Key()) + len(leaf.Value())), nil
		}
		next, err := lr.IterateLeaves(0, 0, [][]byte{root}, tr.GetStorageManager(), parser, context.Background(), visit)
		require.NoError(t, err)
		require.NotZero(t, numLeaves)
		require.Less(t, numLeaves, 25)
		require.NotEmpty(t, next)
	})
}
