package pruning_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestRecoveryReadRetainedEpochs(t *testing.T) {
	args := getDefaultArgsSerialDB()
	args.EpochsData.StartingEpoch = 8
	dir := t.TempDir()
	args.PathManager = &testscommon.PathManagerStub{PathForEpochCalled: func(shard string, epoch uint32, identifier string) string {
		return filepath.Join(dir, fmt.Sprint(epoch), shard, identifier)
	}}
	store, err := pruning.NewTriePruningStorer(args)
	require.NoError(t, err)
	key := []byte("node")
	require.NoError(t, store.PutInEpoch(key, []byte("retained"), 7))
	require.NoError(t, store.PutInEpoch(key, []byte("newer"), 8))
	value, err := store.GetForRecovery(key, 7)
	require.NoError(t, err)
	require.Equal(t, []byte("retained"), value)
	require.NoError(t, store.PutInEpoch([]byte("new-only"), []byte("value"), 8))
	_, err = store.GetForRecovery([]byte("new-only"), 7)
	require.ErrorIs(t, err, storage.ErrKeyNotFound)
	_, err = store.GetForRecovery(key, 6)
	require.Error(t, err)
	require.NoError(t, store.PutForRecovery([]byte("repair"), []byte("repaired"), 7))
	require.NoError(t, store.Close())

	store, err = pruning.NewTriePruningStorer(args)
	require.NoError(t, err)
	defer store.Close()
	value, err = store.GetForRecovery([]byte("repair"), 7)
	require.NoError(t, err)
	require.Equal(t, []byte("repaired"), value)
	value, err = store.GetForRecovery([]byte("repair"), 8)
	require.NoError(t, err)
	require.Equal(t, []byte("repaired"), value)
}
