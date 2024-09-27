package stateChanges

import (
	"math/big"
	"testing"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func TestNewDataAnalysisCollector(t *testing.T) {
	t.Parallel()

	t.Run("nil storer", func(t *testing.T) {
		t.Parallel()

		dsc, err := NewDataAnalysisStateChangesCollector(nil)
		require.Nil(t, dsc)
		require.Equal(t, storage.ErrNilPersister, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
		require.Nil(t, err)
		require.False(t, dsc.IsInterfaceNil())
	})
}

func TestDataAnalysisStateChangesCollector_AddStateChange(t *testing.T) {
	t.Parallel()

	dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
	require.Nil(t, err)

	require.Equal(t, 0, len(dsc.stateChanges))

	dsc.AddStateChange(&data.StateChange{
		Type: "write",
	})
	dsc.AddStateChange(&data.StateChange{
		Type: "read",
	})
	dsc.AddStateChange(&data.StateChange{
		Type: "write",
	})

	require.Equal(t, 3, len(dsc.stateChanges))
}

func TestDataAnalysisStateChangesCollector_AddSaveAccountStateChange(t *testing.T) {
	t.Parallel()

	t.Run("nil old account should return early", func(t *testing.T) {
		t.Parallel()

		dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
		require.Nil(t, err)

		dsc.AddSaveAccountStateChange(
			nil,
			&state.UserAccountStub{},
			&data.StateChange{
				Type:        "saveAccount",
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		dsc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := dsc.GetStateChanges()
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

		dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
		require.Nil(t, err)

		dsc.AddSaveAccountStateChange(
			&state.UserAccountStub{},
			nil,
			&data.StateChange{
				Type:        "saveAccount",
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		dsc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := dsc.GetStateChanges()
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

		dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
		require.Nil(t, err)

		dsc.AddSaveAccountStateChange(
			&state.UserAccountStub{
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
			&state.UserAccountStub{
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
				Type:        "saveAccount",
				Index:       2,
				TxHash:      []byte("txHash1"),
				MainTrieKey: []byte("key1"),
			},
		)

		dsc.AddTxHashToCollectedStateChanges([]byte("txHash1"), &transaction.Transaction{})

		stateChangesForTx := dsc.GetStateChanges()
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

	dsc, err := NewDataAnalysisStateChangesCollector(&mock.PersisterStub{})
	require.Nil(t, err)

	numStateChanges := 10
	for i := 0; i < numStateChanges; i++ {
		dsc.AddStateChange(getDefaultStateChange())
	}
	require.Equal(t, numStateChanges, len(dsc.stateChanges))

	dsc.Reset()
	require.Equal(t, 0, len(dsc.GetStateChanges()))
}
