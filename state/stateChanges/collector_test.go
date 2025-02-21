package stateChanges

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/mock"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"

	"github.com/stretchr/testify/require"
)

func getWriteStateChange() *data.StateChange {
	return &data.StateChange{
		Type: data.Write,
	}
}

func getReadStateChange() *data.StateChange {
	return &data.StateChange{
		Type: data.Read,
	}
}

func TestNewStateChangesCollector(t *testing.T) {
	t.Parallel()

	stateChangesCollector := NewCollector()
	require.False(t, stateChangesCollector.IsInterfaceNil())
}

func TestStateChangesCollector_AddStateChange(t *testing.T) {
	t.Parallel()

	t.Run("default collector", func(t *testing.T) {
		t.Parallel()

		c := NewCollector()
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getWriteStateChange())
		}
		assert.Equal(t, 0, len(c.stateChanges))
	})

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getWriteStateChange())
		}

		c.AddStateChange(getReadStateChange())
		assert.Equal(t, numStateChanges, len(c.stateChanges))
	})

	t.Run("collect only read", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectRead())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getReadStateChange())
		}

		c.AddStateChange(getWriteStateChange())
		assert.Equal(t, numStateChanges, len(c.stateChanges))
	})

	t.Run("collect both read and write", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			if i%2 == 0 {
				c.AddStateChange(getReadStateChange())
			} else {
				c.AddStateChange(getWriteStateChange())
			}
		}
		assert.Equal(t, numStateChanges, len(c.stateChanges))
	})
}

func TestStateChangesCollector_GetStateChanges(t *testing.T) {
	t.Parallel()

	t.Run("getStateChanges with tx hash", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(&data.StateChange{
				Type:        data.Write,
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(c.stateChanges))
		stateChangesForTxs, _ := c.getStateChangesForTxs()
		assert.Equal(t, 0, len(stateChangesForTxs))
		c.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
		assert.Equal(t, numStateChanges, len(c.stateChanges))
		assert.Equal(t, 1, len(c.GetStateChanges()))
		assert.Equal(t, []byte("txHash"), c.GetStateChanges()[0].TxHash)
		assert.Equal(t, numStateChanges, len(c.GetStateChanges()[0].StateChanges))

		stateChangesForTx := c.GetStateChanges()
		assert.Equal(t, 1, len(stateChangesForTx))
		assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
		for i := 0; i < len(stateChangesForTx[0].StateChanges); i++ {
			sc, ok := stateChangesForTx[0].StateChanges[i].(*data.StateChange)
			require.True(t, ok)

			assert.Equal(t, []byte(strconv.Itoa(i)), sc.MainTrieKey)
		}
	})

	t.Run("getStateChanges without tx hash", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite(), WithStorer(&mock.PersisterStub{}))
		assert.Equal(t, 0, len(c.stateChanges))
		assert.Equal(t, 0, len(c.GetStateChanges()))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(&data.StateChange{
				Type:        data.Write,
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(c.stateChanges))
		assert.Equal(t, 0, len(c.GetStateChanges()))

		stateChangesForTx := c.GetStateChanges()
		assert.Equal(t, 0, len(stateChangesForTx))
	})
}

func TestStateChangesCollector_AddTxHashToCollectedStateChanges(t *testing.T) {
	t.Parallel()

	c := NewCollector(WithCollectWrite())
	assert.Equal(t, 0, len(c.stateChanges))
	assert.Equal(t, 0, len(c.GetStateChanges()))

	c.AddTxHashToCollectedStateChanges([]byte("txHash0"), &transaction.Transaction{})

	stateChange := &data.StateChange{
		Type:            data.Write,
		MainTrieKey:     []byte("mainTrieKey"),
		MainTrieVal:     []byte("mainTrieVal"),
		DataTrieChanges: []*data.DataTrieChange{{Key: []byte("dataTrieKey"), Val: []byte("dataTrieVal")}},
	}
	c.AddStateChange(stateChange)

	assert.Equal(t, 1, len(c.stateChanges))
	assert.Equal(t, 0, len(c.GetStateChanges()))
	c.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	assert.Equal(t, 1, len(c.stateChanges))
	assert.Equal(t, 1, len(c.GetStateChanges()))

	stateChangesForTx := c.GetStateChanges()
	assert.Equal(t, 1, len(stateChangesForTx))
	assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
	assert.Equal(t, 1, len(stateChangesForTx[0].StateChanges))

	sc, ok := stateChangesForTx[0].StateChanges[0].(*data.StateChange)
	require.True(t, ok)

	assert.Equal(t, []byte("mainTrieKey"), sc.MainTrieKey)
	assert.Equal(t, []byte("mainTrieVal"), sc.MainTrieVal)
	assert.Equal(t, 1, len(sc.DataTrieChanges))
}

func TestStateChangesCollector_RevertToIndex_FailIfWrongIndex(t *testing.T) {
	t.Parallel()

	c := NewCollector(WithCollectWrite())
	numStateChanges := len(c.stateChanges)

	err := c.RevertToIndex(-1)
	require.True(t, errors.Is(err, state.ErrStateChangesIndexOutOfBounds))

	err = c.RevertToIndex(numStateChanges + 1)
	require.Nil(t, err)
}

func TestStateChangesCollector_RevertToIndex(t *testing.T) {
	t.Parallel()

	c := NewCollector(WithCollectWrite())

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateChange(getWriteStateChange())
		err := c.SetIndexToLastStateChange(i)
		require.Nil(t, err)
	}
	c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

	for i := numStateChanges; i < numStateChanges*2; i++ {
		c.AddStateChange(getWriteStateChange())
		c.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
	}
	err := c.SetIndexToLastStateChange(numStateChanges)
	require.Nil(t, err)

	assert.Equal(t, numStateChanges*2, len(c.stateChanges))

	err = c.RevertToIndex(numStateChanges)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges*2-1, len(c.stateChanges))

	err = c.RevertToIndex(numStateChanges - 1)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges-1, len(c.stateChanges))

	err = c.RevertToIndex(numStateChanges / 2)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges/2, len(c.stateChanges))

	err = c.RevertToIndex(1)
	require.Nil(t, err)
	assert.Equal(t, 1, len(c.stateChanges))

	err = c.RevertToIndex(0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(c.stateChanges))
}

func TestStateChangesCollector_SetIndexToLastStateChange(t *testing.T) {
	t.Parallel()

	t.Run("should fail if invalid index", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())

		err := c.SetIndexToLastStateChange(-1)
		require.True(t, errors.Is(err, state.ErrStateChangesIndexOutOfBounds))

		numStateChanges := len(c.stateChanges)
		err = c.SetIndexToLastStateChange(numStateChanges + 1)
		require.Nil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getWriteStateChange())
			err := c.SetIndexToLastStateChange(i)
			require.Nil(t, err)
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		for i := numStateChanges; i < numStateChanges*2; i++ {
			c.AddStateChange(getWriteStateChange())
			c.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
		}
		err := c.SetIndexToLastStateChange(numStateChanges)
		require.Nil(t, err)

		assert.Equal(t, numStateChanges*2, len(c.stateChanges))
	})
}

func TestStateChangesCollector_Reset(t *testing.T) {
	t.Parallel()

	c := NewCollector(WithCollectWrite())
	assert.Equal(t, 0, len(c.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateChange(getWriteStateChange())
	}
	c.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	for i := numStateChanges; i < numStateChanges*2; i++ {
		c.AddStateChange(getWriteStateChange())
	}
	assert.Equal(t, numStateChanges*2, len(c.stateChanges))

	assert.Equal(t, 1, len(c.GetStateChanges()))

	c.Reset()
	assert.Equal(t, 0, len(c.stateChanges))

	assert.Equal(t, 0, len(c.GetStateChanges()))
}

func TestStateChangesCollector_Publish(t *testing.T) {
	t.Parallel()

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
			if i%2 == 0 {
				c.AddStateChange(&data.StateChange{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateChange(&data.StateChange{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 1)
		require.Len(t, stateChangesForTx["hash0"].StateChanges, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateChanges{
			"hash0": {
				StateChanges: []*data.StateChange{
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

		c := NewCollector(WithCollectRead())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
			if i%2 == 0 {
				c.AddStateChange(&data.StateChange{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateChange(&data.StateChange{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 1)
		require.Len(t, stateChangesForTx["hash1"].StateChanges, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateChanges{
			"hash1": {
				StateChanges: []*data.StateChange{
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

		c := NewCollector(WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateChanges))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
			if i%2 == 0 {
				c.AddStateChange(&data.StateChange{
					Type: data.Write,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			} else {
				c.AddStateChange(&data.StateChange{
					Type: data.Read,
					// distribute evenly based on parity of the index
					TxHash: []byte(fmt.Sprintf("hash%d", i%2)),
				})
			}
		}

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 2)
		require.Len(t, stateChangesForTx["hash0"].StateChanges, 10)
		require.Len(t, stateChangesForTx["hash1"].StateChanges, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateChanges{
			"hash0": {
				StateChanges: []*data.StateChange{
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
				StateChanges: []*data.StateChange{
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

func TestNewDataAnalysisCollector(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithStorer(&mock.PersisterStub{}))
		require.False(t, c.IsInterfaceNil())
		require.NotNil(t, c.storer)
	})
}

func TestDataAnalysisStateChangesCollector_AddSaveAccountStateChange(t *testing.T) {
	t.Parallel()

	t.Run("nil old account should return early", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateChange(
			nil,
			&mockState.UserAccountStub{},
			&data.StateChange{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))

		sc := stateChangesForTx[0].StateChanges[0]
		dasc, ok := sc.(*dataAnalysisStateChangeDTO)
		require.True(t, ok)

		require.False(t, dasc.Nonce)
		require.False(t, dasc.Balance)
		require.False(t, dasc.CodeHash)
		require.False(t, dasc.RootHash)
		require.False(t, dasc.DeveloperReward)
		require.False(t, dasc.OwnerAddress)
		require.False(t, dasc.UserName)
		require.False(t, dasc.CodeMetadata)
	})

	t.Run("nil new account should return early", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateChange(
			&mockState.UserAccountStub{},
			nil,
			&data.StateChange{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))

		sc := stateChangesForTx[0].StateChanges[0]
		dasc, ok := sc.(*dataAnalysisStateChangeDTO)
		require.True(t, ok)

		require.False(t, dasc.Nonce)
		require.False(t, dasc.Balance)
		require.False(t, dasc.CodeHash)
		require.False(t, dasc.RootHash)
		require.False(t, dasc.DeveloperReward)
		require.False(t, dasc.OwnerAddress)
		require.False(t, dasc.UserName)
		require.False(t, dasc.CodeMetadata)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateChange(
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
			&data.StateChange{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))

		sc := stateChangesForTx[0].StateChanges[0]
		dasc, ok := sc.(*dataAnalysisStateChangeDTO)
		require.True(t, ok)

		require.True(t, dasc.Nonce)
		require.True(t, dasc.Balance)
		require.True(t, dasc.CodeHash)
		require.True(t, dasc.RootHash)
		require.True(t, dasc.DeveloperReward)
		require.True(t, dasc.OwnerAddress)
		require.True(t, dasc.UserName)
		require.True(t, dasc.CodeMetadata)
	})
}

func TestDataAnalysisStateChangesCollector_Reset(t *testing.T) {
	t.Parallel()

	c := NewCollector(WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateChange(getWriteStateChange())
	}
	require.Equal(t, numStateChanges, len(c.stateChanges))

	c.Reset()
	require.Equal(t, 0, len(c.GetStateChanges()))
}

func TestDataAnalysisStateChangesCollector_Store(t *testing.T) {
	t.Parallel()

	t.Run("with storer", func(t *testing.T) {
		t.Parallel()

		putCalled := false
		storer := &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				putCalled = true
				return nil
			},
		}

		c := NewCollector(WithCollectWrite(), WithStorer(storer))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getWriteStateChange())
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		err := c.Store()
		require.Nil(t, err)

		require.True(t, putCalled)
	})

	t.Run("without storer, should return nil directly", func(t *testing.T) {
		t.Parallel()

		c := NewCollector(WithCollectWrite())

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateChange(getWriteStateChange())
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		err := c.Store()
		require.Nil(t, err)
	})
}
