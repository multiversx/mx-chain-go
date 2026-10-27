package stateAccesses

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewStateAccessesStorer(t *testing.T) {
	t.Parallel()

	t.Run("nil underlying storer", func(t *testing.T) {
		t.Parallel()

		storer, err := NewStateAccessesStorer(nil, &marshallerMock.MarshalizerMock{})
		assert.True(t, check.IfNil(storer))
		assert.Equal(t, state.ErrNilStateAccessesStorer, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		storer, err := NewStateAccessesStorer(&storage.StorerStub{}, nil)
		assert.True(t, check.IfNil(storer))
		assert.Equal(t, state.ErrNilMarshalizer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		storer, err := NewStateAccessesStorer(&storage.StorerStub{}, &marshallerMock.MarshalizerMock{})
		assert.NotNil(t, storer)
		assert.Nil(t, err)
		assert.False(t, storer.IsInterfaceNil())
	})
}

func TestStateAccessesStorer_Store(t *testing.T) {
	t.Parallel()

	marshallCalled := 0
	putCalled := 0
	marshalledBytes := []byte("stateChange")
	marshaller := &marshallerMock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			marshallCalled++
			return marshalledBytes, nil
		},
	}
	db := &storage.StorerStub{
		PutCalled: func(key []byte, value []byte) error {
			putCalled++
			assert.Equal(t, marshalledBytes, value)
			return nil
		},
	}
	storer, _ := NewStateAccessesStorer(db, marshaller)
	stateAccessesForTxs := map[string]*data.StateAccesses{
		"tx1": {StateAccess: []*data.StateAccess{}},
		"tx2": {StateAccess: []*data.StateAccess{}},
	}

	err := storer.Store(stateAccessesForTxs)
	assert.Nil(t, err)
	assert.Equal(t, 2, marshallCalled)
	assert.Equal(t, 2, putCalled)
}
