package pruning_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	storageCore "github.com/multiversx/mx-chain-core-go/storage"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewFullHistoryTriePruningStorer(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 10,
	}
	fhps, err := pruning.NewFullHistoryTriePruningStorer(fhArgs)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(fhps))
}

func TestFullHistoryTriePruningStorer_CallsMethodsFromUndelyingFHPS(t *testing.T) {
	t.Parallel()

	t.Run("GetFromEpoch called", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		fhArgs := pruning.FullHistoryStorerArgs{
			StorerArgs:               args,
			NumOfOldActivePersisters: 10,
		}
		fhps, _ := pruning.NewFullHistoryTriePruningStorer(fhArgs)

		getFromEpochCalled := false
		sweo := &storage.StorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, nil
			},
		}
		fhps.SetStorerWithEpochOperations(sweo)
		_, _ = fhps.GetFromEpoch([]byte("key"), 0)

		assert.True(t, getFromEpochCalled)
	})

	t.Run("GetBulkFromEpoch called", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		fhArgs := pruning.FullHistoryStorerArgs{
			StorerArgs:               args,
			NumOfOldActivePersisters: 10,
		}
		fhps, _ := pruning.NewFullHistoryTriePruningStorer(fhArgs)

		getBulkFromEpochCalled := false
		sweo := &storage.StorerStub{
			GetBulkFromEpochCalled: func(_ [][]byte, _ uint32) ([]storageCore.KeyValuePair, error) {
				getBulkFromEpochCalled = true
				return nil, nil
			},
		}
		fhps.SetStorerWithEpochOperations(sweo)
		_, _ = fhps.GetBulkFromEpoch([][]byte{[]byte("key")}, 0)

		assert.True(t, getBulkFromEpochCalled)
	})

	t.Run("PutInEpoch called", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		fhArgs := pruning.FullHistoryStorerArgs{
			StorerArgs:               args,
			NumOfOldActivePersisters: 10,
		}
		fhps, _ := pruning.NewFullHistoryTriePruningStorer(fhArgs)

		putInEpochCalled := false
		sweo := &storage.StorerStub{
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				putInEpochCalled = true
				return nil
			},
		}
		fhps.SetStorerWithEpochOperations(sweo)
		_ = fhps.PutInEpoch([]byte("key"), []byte("data"), 0)

		assert.True(t, putInEpochCalled)
	})

	t.Run("Close called", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		fhArgs := pruning.FullHistoryStorerArgs{
			StorerArgs:               args,
			NumOfOldActivePersisters: 10,
		}
		fhps, _ := pruning.NewFullHistoryTriePruningStorer(fhArgs)

		closeCalled := false
		sweo := &storage.StorerStub{
			CloseCalled: func() error {
				closeCalled = true
				return nil
			},
		}
		fhps.SetStorerWithEpochOperations(sweo)
		_ = fhps.Close()

		assert.True(t, closeCalled)
	})
}
