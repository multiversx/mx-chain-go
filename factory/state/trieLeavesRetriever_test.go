package state

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	factoryMock "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestTrieRetrieverFactoryEnabledAndDisabled(t *testing.T) {
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
	lr, err := scf.createTrieLeavesRetriever(tr.GetStorageManager())
	require.NoError(t, err)
	_, _, err = lr.GetLeaves(10, [][]byte{root}, parser, context.Background())
	require.ErrorIs(t, err, common.ErrTrieLeavesRetrieverDisabled)
	scf.config.TrieLeavesRetrieverConfig = config.TrieLeavesRetrieverConfig{Enabled: true, MaxSizeInBytes: 100000}
	lr, err = scf.createTrieLeavesRetriever(tr.GetStorageManager())
	require.NoError(t, err)
	next := [][]byte{root}
	all := make(map[string]string)
	for page := 0; len(next) > 0; page++ {
		require.Less(t, page, 5)
		var leaves map[string]string
		leaves, next, err = lr.GetLeaves(10, next, parser, context.Background())
		require.NoError(t, err)
		for k, v := range leaves {
			require.NotContains(t, all, k)
			all[k] = v
		}
	}
	require.Len(t, all, 25)
}
