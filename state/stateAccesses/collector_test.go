package stateAccesses

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/storage/mock"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getWriteStateAccess() *data.StateAccess {
	return &data.StateAccess{
		Type: data.Write,
	}
}

func getReadStateAccess() *data.StateAccess {
	return &data.StateAccess{
		Type: data.Read,
	}
}

func TestNewStateAccessesCollector(t *testing.T) {
	t.Parallel()

	t.Run("nil storer", func(t *testing.T) {
		t.Parallel()

		stateAccessesCollector, err := NewCollector(nil)
		require.True(t, stateAccessesCollector.IsInterfaceNil())
		require.Equal(t, state.ErrNilStateAccessesStorer, err)
	})
	t.Run("should work with default options", func(t *testing.T) {
		t.Parallel()

		stateAccessesCollector, err := NewCollector(disabled.NewDisabledStateAccessesStorer())
		require.False(t, stateAccessesCollector.IsInterfaceNil())
		require.Nil(t, err)
	})
}

func TestStateAccessesCollector_AddStateAccess(t *testing.T) {
	t.Parallel()

	t.Run("default collector", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		assert.Equal(t, 0, len(c.stateAccesses))
	})

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}

		c.AddStateAccess(getReadStateAccess())
		assert.Equal(t, numStateAccesses, len(c.stateAccesses))
	})

	t.Run("collect only read", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectRead())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getReadStateAccess())
		}

		c.AddStateAccess(getWriteStateAccess())
		assert.Equal(t, numStateAccesses, len(c.stateAccesses))
	})

	t.Run("collect both read and write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			if i%2 == 0 {
				c.AddStateAccess(getReadStateAccess())
			} else {
				c.AddStateAccess(getWriteStateAccess())
			}
		}
		assert.Equal(t, numStateAccesses, len(c.stateAccesses))
	})
}

func TestStateAccessesCollector_AddTxHashToCollectedStateAccesses(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())
	assert.Equal(t, 0, len(c.stateAccesses))
	assert.Equal(t, 0, len(c.GetCollectedAccesses()))
	c.AddTxHashToCollectedStateAccesses([]byte("txHash0"))
	assert.Equal(t, 0, len(c.stateAccesses))
	assert.Equal(t, 0, len(c.GetCollectedAccesses()))

	stateAccess := &data.StateAccess{
		Type:            data.Write,
		MainTrieKey:     []byte("mainTrieKey"),
		MainTrieVal:     []byte("mainTrieVal"),
		DataTrieChanges: []*data.DataTrieChange{{Key: []byte("dataTrieKey"), Val: []byte("dataTrieVal")}},
	}
	c.AddStateAccess(stateAccess)
	c.AddTxHashToCollectedStateAccesses([]byte("txHash"))

	assert.Equal(t, 1, len(c.stateAccesses))
	assert.Equal(t, 1, len(c.GetCollectedAccesses()))

	stateAccessesForTx := c.GetCollectedAccesses()
	stateAccesses, ok := stateAccessesForTx["txHash"]
	require.True(t, ok)
	assert.Equal(t, 1, len(stateAccessesForTx))
	assert.Equal(t, 1, len(stateAccesses.StateAccess))

	sc := stateAccesses.StateAccess[0]
	assert.Equal(t, []byte("mainTrieKey"), sc.MainTrieKey)
	assert.Equal(t, []byte("mainTrieVal"), sc.MainTrieVal)
	assert.Equal(t, 1, len(sc.DataTrieChanges))
}

func TestStateAccessesCollector_RevertToIndex(t *testing.T) {
	t.Parallel()

	t.Run("fail if wrong index", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())
		numStateAccesses := len(c.stateAccesses)

		err := c.RevertToIndex(-1)
		require.True(t, errors.Is(err, state.ErrStateAccessesIndexOutOfBounds))

		err = c.RevertToIndex(numStateAccesses + 1)
		require.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
			err := c.SetIndexToLatestStateAccesses(i)
			require.Nil(t, err)
		}
		c.AddTxHashToCollectedStateAccesses([]byte("txHash1"))

		for i := numStateAccesses; i < numStateAccesses*2; i++ {
			c.AddStateAccess(getWriteStateAccess())
			c.AddTxHashToCollectedStateAccesses([]byte("txHash" + fmt.Sprintf("%d", i)))
		}
		err := c.SetIndexToLatestStateAccesses(numStateAccesses)
		require.Nil(t, err)

		assert.Equal(t, numStateAccesses*2, len(c.stateAccesses))

		err = c.RevertToIndex(numStateAccesses)
		require.Nil(t, err)
		assert.Equal(t, numStateAccesses*2, len(c.stateAccesses))

		err = c.RevertToIndex(numStateAccesses - 1)
		require.Nil(t, err)
		assert.Equal(t, numStateAccesses, len(c.stateAccesses))

		err = c.RevertToIndex(numStateAccesses / 2)
		require.Nil(t, err)
		assert.Equal(t, numStateAccesses/2+1, len(c.stateAccesses))

		err = c.RevertToIndex(1)
		require.Nil(t, err)
		assert.Equal(t, 2, len(c.stateAccesses))

		err = c.RevertToIndex(0)
		require.Nil(t, err)
		assert.Equal(t, 0, len(c.stateAccesses))
	})
}

func TestStateAccessesCollector_SetIndexToLatestStateAccesses(t *testing.T) {
	t.Parallel()

	t.Run("should fail if invalid index", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())

		err := c.SetIndexToLatestStateAccesses(-1)
		require.True(t, errors.Is(err, state.ErrStateAccessesIndexOutOfBounds))

		numStateAccesses := len(c.stateAccesses)
		err = c.SetIndexToLatestStateAccesses(numStateAccesses + 1)
		require.Nil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
			err := c.SetIndexToLatestStateAccesses(i)
			require.Nil(t, err)
		}
		c.AddTxHashToCollectedStateAccesses([]byte("txHash1"))

		for i := numStateAccesses; i < numStateAccesses*2; i++ {
			c.AddStateAccess(getWriteStateAccess())
			c.AddTxHashToCollectedStateAccesses([]byte("txHash" + fmt.Sprintf("%d", i)))
		}
		err := c.SetIndexToLatestStateAccesses(numStateAccesses)
		require.Nil(t, err)

		assert.Equal(t, numStateAccesses*2, len(c.stateAccesses))
	})
}

func TestStateAccessesCollector_Reset(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())
	assert.Equal(t, 0, len(c.stateAccesses))

	numStateAccesses := 10
	for i := 0; i < numStateAccesses; i++ {
		c.AddStateAccess(getWriteStateAccess())
	}
	c.AddTxHashToCollectedStateAccesses([]byte("txHash"))
	assert.Equal(t, numStateAccesses, len(c.stateAccesses))
	assert.Equal(t, 1, len(c.GetCollectedAccesses()))
	assert.Equal(t, 1, len(c.stateAccessesForTxs))

	c.Reset()
	assert.Equal(t, 0, len(c.stateAccesses))
	assert.Equal(t, 0, len(c.stateAccessesForTxs))
	assert.Equal(t, 0, len(c.GetCollectedAccesses()))
}

func TestStateAccessesCollector_GetCollectedAccesses(t *testing.T) {
	t.Parallel()

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 20
		for i := 0; i < numStateAccesses; i++ {
			if i%2 == 0 {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateAccessesForTx := c.GetCollectedAccesses()

		require.Len(t, stateAccessesForTx, 1)
		require.Len(t, stateAccessesForTx["hash0"].StateAccess, 10)

		require.Equal(t, stateAccessesForTx, map[string]*data.StateAccesses{
			"hash0": {
				StateAccess: []*data.StateAccess{
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
				},
			},
		})
	})

	t.Run("collect only read", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectRead())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 20
		for i := 0; i < numStateAccesses; i++ {
			if i%2 == 0 {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateAccessesForTx := c.GetCollectedAccesses()

		require.Len(t, stateAccessesForTx, 1)
		require.Len(t, stateAccessesForTx["hash1"].StateAccess, 10)

		require.Equal(t, stateAccessesForTx, map[string]*data.StateAccesses{
			"hash1": {
				StateAccess: []*data.StateAccess{
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
				},
			},
		})
	})

	t.Run("collect both read and write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateAccesses := 20
		for i := 0; i < numStateAccesses; i++ {
			if i%2 == 0 {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateAccess(&data.StateAccess{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateAccessesForTx := c.GetCollectedAccesses()

		require.Len(t, stateAccessesForTx, 2)
		require.Len(t, stateAccessesForTx["hash0"].StateAccess, 10)
		require.Len(t, stateAccessesForTx["hash1"].StateAccess, 10)

		require.Equal(t, stateAccessesForTx, map[string]*data.StateAccesses{
			"hash0": {
				StateAccess: []*data.StateAccess{
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
					{Type: data.Write, TxHash: []byte("hash0")},
				},
			},
			"hash1": {
				StateAccess: []*data.StateAccess{
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
					{Type: data.Read, TxHash: []byte("hash1")},
				},
			},
		})
	})
}

func TestCollector_GetAccountChanges(t *testing.T) {
	t.Parallel()

	t.Run("nil old account should return early", func(t *testing.T) {
		t.Parallel()

		storer, _ := NewStateAccessesStorer(&storage.StorerStub{}, &mock.MarshalizerMock{})
		c, _ := NewCollector(storer, WithCollectWrite(), WithAccountChanges())

		accountChanges := c.GetAccountChanges(
			nil,
			&mockState.UserAccountStub{},
		)
		assert.Nil(t, accountChanges)
	})

	t.Run("nil new account should return early", func(t *testing.T) {
		t.Parallel()

		storer, _ := NewStateAccessesStorer(&storage.StorerStub{}, &mock.MarshalizerMock{})
		c, _ := NewCollector(storer, WithCollectWrite(), WithAccountChanges())

		accountChanges := c.GetAccountChanges(
			&mockState.UserAccountStub{},
			nil,
		)
		assert.Nil(t, accountChanges)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		storer, _ := NewStateAccessesStorer(&storage.StorerStub{}, &mock.MarshalizerMock{})
		c, _ := NewCollector(storer, WithCollectWrite(), WithAccountChanges())

		accountChanges := c.GetAccountChanges(
			&mockState.UserAccountStub{
				Nonce:            0,
				Balance:          big.NewInt(0),
				DeveloperRewards: big.NewInt(0),
				UserName:         []byte{0},
				Owner:            []byte{0},
				Address:          []byte{0},
				CodeMetadata:     []byte{0},
				CodeHash:         []byte{0},
				GetRootHashCalled: func() []byte {
					return []byte{0}
				},
			},
			&mockState.UserAccountStub{
				Nonce:            1,
				Balance:          big.NewInt(1),
				DeveloperRewards: big.NewInt(1),
				UserName:         []byte{1},
				Owner:            []byte{1},
				Address:          []byte{1},
				CodeMetadata:     []byte{1},
				CodeHash:         []byte{1},
				GetRootHashCalled: func() []byte {
					return []byte{1}
				},
			},
		)

		require.True(t, accountChanges.Nonce)
		require.True(t, accountChanges.Balance)
		require.True(t, accountChanges.CodeHash)
		require.True(t, accountChanges.RootHash)
		require.True(t, accountChanges.DeveloperReward)
		require.True(t, accountChanges.OwnerAddress)
		require.True(t, accountChanges.UserName)
		require.True(t, accountChanges.CodeMetadata)
	})
}

func TestCollector_Store(t *testing.T) {
	t.Parallel()

	t.Run("with storer", func(t *testing.T) {
		t.Parallel()

		putCalled := false
		db := &storage.StorerStub{
			PutCalled: func(key, val []byte) error {
				putCalled = true
				return nil
			},
		}

		storer, _ := NewStateAccessesStorer(db, &mock.MarshalizerMock{})
		c, _ := NewCollector(storer, WithCollectWrite())

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		c.AddTxHashToCollectedStateAccesses([]byte("txHash1"))

		err := c.Store()
		require.Nil(t, err)

		require.True(t, putCalled)
	})

	t.Run("without storer, should return nil directly", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(disabled.NewDisabledStateAccessesStorer(), WithCollectWrite())

		numStateAccesses := 10
		for i := 0; i < numStateAccesses; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		c.AddTxHashToCollectedStateAccesses([]byte("txHash1"))

		err := c.Store()
		require.Nil(t, err)
	})
}
