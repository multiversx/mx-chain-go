package node

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestLoadOptionsForIteration(t *testing.T) {
	t.Parallel()

	callerOptions := api.AccountQueryOptions{BlockNonce: core.OptionalUint64{Value: 7, HasValue: true}}
	epoch := uint32(3)

	t.Run("no checkpoint or no block recorded uses the caller options", func(t *testing.T) {
		t.Parallel()

		options, err := loadOptionsForIteration(trieIterationRequest{options: callerOptions})
		require.Nil(t, err)
		require.Equal(t, callerOptions, options)

		options, err = loadOptionsForIteration(trieIterationRequest{options: callerOptions, checkpoint: &checkpoint.Checkpoint{IteratorState: [][]byte{[]byte("s")}}})
		require.Nil(t, err)
		require.Equal(t, callerOptions, options)
	})
	t.Run("too many entries", func(t *testing.T) {
		t.Parallel()

		_, err := loadOptionsForIteration(trieIterationRequest{checkpoint: &checkpoint.Checkpoint{IteratorState: make([][]byte, maxIteratorStateEntries+1)}})
		require.ErrorIs(t, err, common.ErrInvalidIteratorState)
	})
	t.Run("block hash without root hash is not a recorded block", func(t *testing.T) {
		t.Parallel()

		options, err := loadOptionsForIteration(trieIterationRequest{options: callerOptions, checkpoint: &checkpoint.Checkpoint{BlockNonce: 1, BlockHash: "aabb"}})
		require.Nil(t, err)
		require.Equal(t, callerOptions, options)
	})
	t.Run("recorded block is loaded by its root hash", func(t *testing.T) {
		t.Parallel()

		cp := &checkpoint.Checkpoint{BlockNonce: 7, BlockHash: "aabb", BlockRootHash: "ccdd", Epoch: &epoch}
		options, err := loadOptionsForIteration(trieIterationRequest{options: callerOptions, checkpoint: cp})
		require.Nil(t, err)
		require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte{0xcc, 0xdd}, HintEpoch: core.OptionalUint32{Value: 3, HasValue: true}}, options)

		options, err = loadOptionsForIteration(trieIterationRequest{checkpoint: &checkpoint.Checkpoint{BlockRootHash: "ccdd", Epoch: &epoch}})
		require.Nil(t, err)
		require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte{0xcc, 0xdd}, HintEpoch: core.OptionalUint32{Value: 3, HasValue: true}}, options)

		callerHint := api.AccountQueryOptions{BlockRootHash: []byte{0xcc, 0xdd}, HintEpoch: core.OptionalUint32{Value: 9, HasValue: true}}
		options, err = loadOptionsForIteration(trieIterationRequest{options: callerHint, checkpoint: &checkpoint.Checkpoint{BlockRootHash: "ccdd"}})
		require.Nil(t, err)
		require.Equal(t, callerHint, options)

		_, err = loadOptionsForIteration(trieIterationRequest{checkpoint: &checkpoint.Checkpoint{BlockRootHash: "zz"}})
		require.ErrorIs(t, err, common.ErrInvalidIteratorState)
	})
	t.Run("caller coordinates must match the checkpoint", func(t *testing.T) {
		t.Parallel()

		cp := &checkpoint.Checkpoint{BlockNonce: 7, BlockHash: "aabb", BlockRootHash: "ccdd", Epoch: &epoch}
		for _, options := range []api.AccountQueryOptions{
			{BlockNonce: core.OptionalUint64{Value: 7, HasValue: true}},
			{BlockHash: []byte{0xaa, 0xbb}},
			{BlockRootHash: []byte{0xcc, 0xdd}, HintEpoch: core.OptionalUint32{Value: 3, HasValue: true}},
			{OnFinalBlock: true}, // the block selectors only apply to the first page
			{OnStartOfEpoch: core.OptionalUint32{Value: 5, HasValue: true}},
			{},
		} {
			_, err := loadOptionsForIteration(trieIterationRequest{options: options, checkpoint: cp})
			require.Nil(t, err)
		}
		for _, options := range []api.AccountQueryOptions{
			{BlockNonce: core.OptionalUint64{Value: 8, HasValue: true}},
			{BlockHash: []byte{0xaa}},
			{BlockRootHash: []byte{0xcc}},
			{BlockRootHash: []byte{0xcc, 0xdd}, HintEpoch: core.OptionalUint32{Value: 4, HasValue: true}},
		} {
			_, err := loadOptionsForIteration(trieIterationRequest{options: options, checkpoint: cp})
			require.ErrorIs(t, err, common.ErrInvalidIteratorState)
		}
	})
}

func TestNewCheckpoint(t *testing.T) {
	t.Parallel()

	blockInfo := api.BlockInfo{Nonce: 5, Hash: "aa", RootHash: "bb"}
	state := [][]byte{[]byte("s")}
	epoch := uint32(2)
	require.Equal(t, &checkpoint.Checkpoint{BlockNonce: 5, BlockHash: "aa", BlockRootHash: "bb", Epoch: &epoch, IteratorState: state}, newCheckpoint(blockInfo, 2, state))
}

func TestNumLeavesForPage(t *testing.T) {
	t.Parallel()

	n, _ := NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxLeaves: 10}))
	require.Equal(t, 10, n.numLeavesForPage(0))
	require.Equal(t, 10, n.numLeavesForPage(11))
	require.Equal(t, 3, n.numLeavesForPage(3))
}

func TestAcquireTrieOperationsSlot(t *testing.T) {
	t.Parallel()

	n, _ := NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxConcurrentRequests: 1}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := n.acquireTrieOperationsSlot(ctx)
	require.ErrorIs(t, err, ErrTrieOperationsTimeout)

	release, err := n.acquireTrieOperationsSlot(context.Background())
	require.Nil(t, err)
	_, err = n.acquireTrieOperationsSlot(context.Background())
	require.ErrorIs(t, err, common.ErrTrieOperationsBusy)
	release()
	release, err = n.acquireTrieOperationsSlot(context.Background())
	require.Nil(t, err)
	release()
	require.Empty(t, n.trieOperationsSlots)
}

func TestKeyValueCollector(t *testing.T) {
	t.Parallel()

	pairs := make(map[string]string)
	visit := keyValueCollector(pairs)

	accepted, size, err := visit(keyValStorage.NewKeyValStorage([]byte{0xaa}, []byte{0xbb, 0xcc}))
	require.Nil(t, err)
	require.True(t, accepted)
	require.Equal(t, uint64(6), size)

	accepted, size, err = visit(keyValStorage.NewKeyValStorage([]byte{0xaa}, []byte{0xdd}))
	require.Nil(t, err)
	require.False(t, accepted)
	require.Equal(t, uint64(4), size)
	require.Equal(t, map[string]string{"aa": "dd"}, pairs)
}
