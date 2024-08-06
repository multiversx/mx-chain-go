package state

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestNewStateChangesCollector(t *testing.T) {
	t.Parallel()

	stateChangesCollector := NewStateChangesCollector()
	require.False(t, stateChangesCollector.IsInterfaceNil())
}

func TestStateChangesCollector_AddStateChange(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{})
	}
	assert.Equal(t, numStateChanges, len(scc.stateChanges))
}

func TestStateChangesCollector_GetStateChanges(t *testing.T) {
	t.Parallel()

	t.Run("getStateChanges with tx hash", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()
		assert.Equal(t, 0, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(StateChangeDTO{
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))
		scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 1, len(scc.GetStateChanges()))
		assert.Equal(t, []byte("txHash"), scc.GetStateChanges()[0].TxHash)
		assert.Equal(t, numStateChanges, len(scc.GetStateChanges()[0].StateChanges))

		stateChangesForTx := scc.GetStateChanges()
		assert.Equal(t, 1, len(stateChangesForTx))
		assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
		for i := 0; i < len(stateChangesForTx[0].StateChanges); i++ {
			assert.Equal(t, []byte(strconv.Itoa(i)), stateChangesForTx[0].StateChanges[i].MainTrieKey)
		}
	})

	t.Run("getStateChanges without tx hash", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()
		assert.Equal(t, 0, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(StateChangeDTO{
				MainTrieKey: []byte(strconv.Itoa(i)),
			})
		}
		assert.Equal(t, numStateChanges, len(scc.stateChanges))
		assert.Equal(t, 0, len(scc.GetStateChanges()))

		stateChangesForTx := scc.GetStateChanges()
		assert.Equal(t, 0, len(stateChangesForTx))
	})
}

func TestStateChangesCollector_AddTxHashToCollectedStateChanges(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))
	assert.Equal(t, 0, len(scc.GetStateChanges()))

	scc.AddTxHashToCollectedStateChanges([]byte("txHash0"), &transaction.Transaction{})

	stateChange := StateChangeDTO{
		MainTrieKey:     []byte("mainTrieKey"),
		MainTrieVal:     []byte("mainTrieVal"),
		DataTrieChanges: []DataTrieChange{{Key: []byte("dataTrieKey"), Val: []byte("dataTrieVal")}},
	}
	scc.AddStateChange(stateChange)

	assert.Equal(t, 1, len(scc.stateChanges))
	assert.Equal(t, 0, len(scc.GetStateChanges()))
	scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	assert.Equal(t, 1, len(scc.stateChanges))
	assert.Equal(t, 1, len(scc.GetStateChanges()))

	stateChangesForTx := scc.GetStateChanges()
	assert.Equal(t, 1, len(stateChangesForTx))
	assert.Equal(t, []byte("txHash"), stateChangesForTx[0].TxHash)
	assert.Equal(t, 1, len(stateChangesForTx[0].StateChanges))
	assert.Equal(t, []byte("mainTrieKey"), stateChangesForTx[0].StateChanges[0].MainTrieKey)
	assert.Equal(t, []byte("mainTrieVal"), stateChangesForTx[0].StateChanges[0].MainTrieVal)
	assert.Equal(t, 1, len(stateChangesForTx[0].StateChanges[0].DataTrieChanges))
}

func TestStateChangesCollector_RevertToIndex_FailIfWrongIndex(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	numStateChanges := len(scc.stateChanges)

	err := scc.RevertToIndex(-1)
	require.Equal(t, ErrStateChangesIndexOutOfBounds, err)

	err = scc.RevertToIndex(numStateChanges + 1)
	require.Equal(t, ErrStateChangesIndexOutOfBounds, err)
}

func TestStateChangesCollector_RevertToIndex(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{})
		scc.SetIndexToLastStateChange(i)
	}
	scc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

	for i := numStateChanges; i < numStateChanges*2; i++ {
		scc.AddStateChange(StateChangeDTO{})
		scc.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
	}
	scc.SetIndexToLastStateChange(numStateChanges)

	assert.Equal(t, numStateChanges*2, len(scc.stateChanges))

	err := scc.RevertToIndex(numStateChanges)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges*2-1, len(scc.stateChanges))

	err = scc.RevertToIndex(numStateChanges - 1)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges-1, len(scc.stateChanges))

	err = scc.RevertToIndex(numStateChanges / 2)
	require.Nil(t, err)
	assert.Equal(t, numStateChanges/2, len(scc.stateChanges))

	err = scc.RevertToIndex(1)
	require.Nil(t, err)
	assert.Equal(t, 1, len(scc.stateChanges))

	err = scc.RevertToIndex(0)
	require.Nil(t, err)
	assert.Equal(t, 0, len(scc.stateChanges))
}

func TestStateChangesCollector_SetIndexToLastStateChange(t *testing.T) {
	t.Parallel()

	t.Run("should fail if valid index", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()

		err := scc.SetIndexToLastStateChange(-1)
		require.Equal(t, ErrStateChangesIndexOutOfBounds, err)

		numStateChanges := len(scc.stateChanges)
		err = scc.SetIndexToLastStateChange(numStateChanges + 1)
		require.Equal(t, ErrStateChangesIndexOutOfBounds, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scc := NewStateChangesCollector()

		numStateChanges := 10
		for i := 0; i < numStateChanges; i++ {
			scc.AddStateChange(StateChangeDTO{})
			err := scc.SetIndexToLastStateChange(i)
			require.Nil(t, err)
		}
		scc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		for i := numStateChanges; i < numStateChanges*2; i++ {
			scc.AddStateChange(StateChangeDTO{})
			scc.AddTxHashToCollectedStateChanges([]byte("txHash"+fmt.Sprintf("%d", i)), &transaction.Transaction{})
		}
		err := scc.SetIndexToLastStateChange(numStateChanges)
		require.Nil(t, err)

		assert.Equal(t, numStateChanges*2, len(scc.stateChanges))
	})
}

func TestStateChangesCollector_Reset(t *testing.T) {
	t.Parallel()

	scc := NewStateChangesCollector()
	assert.Equal(t, 0, len(scc.stateChanges))

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		scc.AddStateChange(StateChangeDTO{})
	}
	scc.AddTxHashToCollectedStateChanges([]byte("txHash"), &transaction.Transaction{})
	for i := numStateChanges; i < numStateChanges*2; i++ {
		scc.AddStateChange(StateChangeDTO{})
	}
	assert.Equal(t, numStateChanges*2, len(scc.stateChanges))

	assert.Equal(t, 1, len(scc.GetStateChanges()))

	scc.Reset()
	assert.Equal(t, 0, len(scc.stateChanges))

	assert.Equal(t, 0, len(scc.GetStateChanges()))
}
