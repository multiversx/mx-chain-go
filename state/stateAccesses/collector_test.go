package stateAccesses

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/mock"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"

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

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		stateAccessesCollector, err := NewCollector(nil)
		require.True(t, stateAccessesCollector.IsInterfaceNil())
		require.Equal(t, state.ErrNilMarshalizer, err)
	})
	t.Run("should work with default options", func(t *testing.T) {
		t.Parallel()

		stateAccessesCollector, err := NewCollector(&mock.MarshalizerMock{})
		require.False(t, stateAccessesCollector.IsInterfaceNil())
		require.Nil(t, err)
	})
}

func TestStateAccessesCollector_AddStateAccess(t *testing.T) {
	t.Parallel()

	t.Run("default collector", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{})
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		assert.Equal(t, 0, len(c.stateAccesses))
	})

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}

		c.AddStateAccess(getReadStateAccess())
		assert.Equal(t, numStateChanges, len(c.stateAccesses))
	})

	t.Run("collect only read", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectRead())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getReadStateAccess())
		}

		c.AddStateAccess(getWriteStateAccess())
		assert.Equal(t, numStateChanges, len(c.stateAccesses))
	})

	t.Run("collect both read and write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			if i%2 == 0 {
				c.AddStateAccess(getReadStateAccess())
			} else {
				c.AddStateAccess(getWriteStateAccess())
			}
		}
		assert.Equal(t, numStateChanges, len(c.stateAccesses))
	})
}

func TestStateAccessesCollector_GetStateChanges(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
	assert.Equal(t, 0, len(c.stateAccesses))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateAccess(&data.StateAccess{
			Type:        data.Write,
			MainTrieKey: []byte(strconv.Itoa(i)),
		})
	}
	assert.Equal(t, numStateChanges, len(c.stateAccesses))
	stateChangesForTxs := getStateAccessesForTxs(c.stateAccesses, false)
	assert.Equal(t, 0, len(stateChangesForTxs))
	c.AddTxHashToCollectedStateChanges([]byte("txHash"))
	assert.Equal(t, numStateChanges, len(c.stateAccesses))
	assert.Equal(t, 1, len(c.GetStateChanges()))
	stateAccesses, ok := c.GetStateChanges()["txHash"]
	require.True(t, ok)
	assert.Equal(t, numStateChanges, len(stateAccesses.StateAccess))

	for i := 0; i < len(stateAccesses.StateAccess); i++ {
		assert.Equal(t, []byte(strconv.Itoa(i)), stateAccesses.StateAccess[i].MainTrieKey)
	}
}

func TestStateAccessesCollector_AddTxHashToCollectedStateChanges(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
	assert.Equal(t, 0, len(c.stateAccesses))
	assert.Equal(t, 0, len(c.GetStateChanges()))

	c.AddTxHashToCollectedStateChanges([]byte("txHash0"))

	stateChange := &data.StateAccess{
		Type:            data.Write,
		MainTrieKey:     []byte("mainTrieKey"),
		MainTrieVal:     []byte("mainTrieVal"),
		DataTrieChanges: []*data.DataTrieChange{{Key: []byte("dataTrieKey"), Val: []byte("dataTrieVal")}},
	}
	c.AddStateAccess(stateChange)

	assert.Equal(t, 1, len(c.stateAccesses))
	assert.Equal(t, 0, len(c.GetStateChanges()))
	c.AddTxHashToCollectedStateChanges([]byte("txHash"))
	assert.Equal(t, 1, len(c.stateAccesses))
	assert.Equal(t, 1, len(c.GetStateChanges()))

	stateChangesForTx := c.GetStateChanges()
	stateAccesses, ok := stateChangesForTx["txHash"]
	require.True(t, ok)
	assert.Equal(t, 1, len(stateChangesForTx))
	assert.Equal(t, 1, len(stateAccesses.StateAccess))

	sc := stateAccesses.StateAccess[0]
	assert.Equal(t, []byte("mainTrieKey"), sc.MainTrieKey)
	assert.Equal(t, []byte("mainTrieVal"), sc.MainTrieVal)
	assert.Equal(t, 1, len(sc.DataTrieChanges))
}

func TestStateAccessesCollector_RevertToIndex_FailIfWrongIndex(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
	numStateChanges := len(c.stateAccesses)

	err := c.RevertToIndex(-1)
	require.True(t, errors.Is(err, state.ErrStateChangesIndexOutOfBounds))

	err = c.RevertToIndex(numStateChanges + 1)
	require.Nil(t, err)
}

func TestStateAccessesCollector_RevertToIndex(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateAccess(getWriteStateAccess())
		err := c.SetIndexToLastStateChange(i)
		require.Nil(t, err)
	}
	c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

	for i := numStateChanges; i < numStateChanges*2; i++ {
		c.AddStateAccess(getWriteStateAccess())
		c.AddTxHashToCollectedStateChanges([]byte("txHash" + fmt.Sprintf("%d", i)))
	}
	err := c.SetIndexToLastStateChange(numStateChanges)
	require.Nil(t, err)

	assert.Equal(t, numStateChanges*2, len(c.stateAccesses))

	err = c.RevertToIndex(numStateChanges)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges*2, len(c.stateAccesses))

	err = c.RevertToIndex(numStateChanges - 1)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges, len(c.stateAccesses))

	err = c.RevertToIndex(numStateChanges / 2)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges/2+1, len(c.stateAccesses))

	err = c.RevertToIndex(1)
	require.Nil(t, err)
	assert.Equal(t, 2, len(c.stateAccesses))

	err = c.RevertToIndex(0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(c.stateAccesses))
}

func TestStateAccessesCollector_SetIndexToLastStateChange(t *testing.T) {
	t.Parallel()

	t.Run("should fail if invalid index", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())

		err := c.SetIndexToLastStateChange(-1)
		require.True(t, errors.Is(err, state.ErrStateChangesIndexOutOfBounds))

		numStateChanges := len(c.stateAccesses)
		err = c.SetIndexToLastStateChange(numStateChanges + 1)
		require.Nil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getWriteStateAccess())
			err := c.SetIndexToLastStateChange(i)
			require.Nil(t, err)
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		for i := numStateChanges; i < numStateChanges*2; i++ {
			c.AddStateAccess(getWriteStateAccess())
			c.AddTxHashToCollectedStateChanges([]byte("txHash" + fmt.Sprintf("%d", i)))
		}
		err := c.SetIndexToLastStateChange(numStateChanges)
		require.Nil(t, err)

		assert.Equal(t, numStateChanges*2, len(c.stateAccesses))
	})
}

func TestStateAccessesCollector_Reset(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
	assert.Equal(t, 0, len(c.stateAccesses))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateAccess(getWriteStateAccess())
	}
	c.AddTxHashToCollectedStateChanges([]byte("txHash"))
	for i := numStateChanges; i < numStateChanges*2; i++ {
		c.AddStateAccess(getWriteStateAccess())
	}
	assert.Equal(t, numStateChanges*2, len(c.stateAccesses))

	assert.Equal(t, 1, len(c.GetStateChanges()))

	c.Reset()
	assert.Equal(t, 0, len(c.stateAccesses))

	assert.Equal(t, 0, len(c.GetStateChanges()))
}

func TestStateAccessesCollector_Publish(t *testing.T) {
	t.Parallel()

	t.Run("collect only write", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
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

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 1)
		require.Len(t, stateChangesForTx["hash0"].StateAccess, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateAccesses{
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

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectRead())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
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

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 1)
		require.Len(t, stateChangesForTx["hash1"].StateAccess, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateAccesses{
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

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectRead(), WithCollectWrite())
		assert.Equal(t, 0, len(c.stateAccesses))

		numStateChanges := 20
		for i := 0; i < numStateChanges; i++ {
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

		stateChangesForTx, err := c.Publish()
		require.NoError(t, err)

		require.Len(t, stateChangesForTx, 2)
		require.Len(t, stateChangesForTx["hash0"].StateAccess, 10)
		require.Len(t, stateChangesForTx["hash1"].StateAccess, 10)

		require.Equal(t, stateChangesForTx, map[string]*data.StateAccesses{
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

func TestNewDataAnalysisCollector(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithStorer(&mock.PersisterStub{}))
		require.False(t, c.IsInterfaceNil())
		require.NotNil(t, c.storer)
	})
}

func TestDataAnalysisStateAccessesCollector_AddSaveAccountStateChange(t *testing.T) {
	t.Parallel()

	t.Run("nil old account should return early", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateAccess(
			nil,
			&mockState.UserAccountStub{},
			&data.StateAccess{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))
		stateAccesses, ok := stateChangesForTx["txHash1"]
		assert.True(t, ok)

		sc := stateAccesses.StateAccess[0]
		require.Nil(t, sc.AccountChanges)
	})

	t.Run("nil new account should return early", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateAccess(
			&mockState.UserAccountStub{},
			nil,
			&data.StateAccess{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))
		stateAccesses, ok := stateChangesForTx["txHash1"]
		assert.True(t, ok)

		sc := stateAccesses.StateAccess[0]
		require.Nil(t, sc.AccountChanges)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

		c.AddSaveAccountStateAccess(
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
			&data.StateAccess{
				Type:        data.Write,
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		stateChangesForTx := c.GetStateChanges()
		require.Equal(t, 1, len(stateChangesForTx))

		stateAccesses, ok := stateChangesForTx["txHash1"]
		assert.True(t, ok)

		sc := stateAccesses.StateAccess[0]
		require.True(t, sc.AccountChanges.Nonce)
		require.True(t, sc.AccountChanges.Balance)
		require.True(t, sc.AccountChanges.CodeHash)
		require.True(t, sc.AccountChanges.RootHash)
		require.True(t, sc.AccountChanges.DeveloperReward)
		require.True(t, sc.AccountChanges.OwnerAddress)
		require.True(t, sc.AccountChanges.UserName)
		require.True(t, sc.AccountChanges.CodeMetadata)
	})
}

func TestDataAnalysisStateAccessesCollector_Reset(t *testing.T) {
	t.Parallel()

	c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite(), WithStorer(&mock.PersisterStub{}))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		c.AddStateAccess(getWriteStateAccess())
	}
	require.Equal(t, numStateChanges, len(c.stateAccesses))

	c.Reset()
	require.Equal(t, 0, len(c.GetStateChanges()))
}

func TestDataAnalysisStateAccessesCollector_Store(t *testing.T) {
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

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite(), WithStorer(storer))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		err := c.Store()
		require.Nil(t, err)

		require.True(t, putCalled)
	})

	t.Run("without storer, should return nil directly", func(t *testing.T) {
		t.Parallel()

		c, _ := NewCollector(&mock.MarshalizerMock{}, WithCollectWrite())

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			c.AddStateAccess(getWriteStateAccess())
		}
		c.AddTxHashToCollectedStateChanges([]byte("txHash1"))

		err := c.Store()
		require.Nil(t, err)
	})
}
