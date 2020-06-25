package timecache

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerTimeCache_NilTimeCacheShouldErr(t *testing.T) {
	t.Parallel()

	ptc, err := NewPeerTimeCache(nil)

	assert.Equal(t, storage.ErrNilTimeCache, err)
	assert.True(t, check.IfNil(ptc))
}

func TestNewPeerTimeCache_ShouldWork(t *testing.T) {
	t.Parallel()

	ptc, err := NewPeerTimeCache(&mock.TimeCacheStub{})

	assert.Nil(t, err)
	assert.False(t, check.IfNil(ptc))
}

func TestPeerTimeCache_Methods(t *testing.T) {
	t.Parallel()

	pid := core.PeerID("test peer id")
	unexpectedErr := errors.New("unexpected error")
	updateWasCalled := false
	hasWasCalled := false
	sweepWasCalled := false
	ptc, _ := NewPeerTimeCache(&mock.TimeCacheStub{
		UpsertCalled: func(key string, span time.Duration) error {
			if key != string(pid) {
				return unexpectedErr
			}

			updateWasCalled = true
			return nil
		},
		HasCalled: func(key string) bool {
			if key != string(pid) {
				return false
			}

			hasWasCalled = true
			return true
		},
		SweepCalled: func() {
			sweepWasCalled = true
		},
	})

	assert.Nil(t, ptc.Upsert(pid, time.Second))
	assert.True(t, ptc.Has(pid))
	ptc.Sweep()

	assert.True(t, updateWasCalled)
	assert.True(t, hasWasCalled)
	assert.True(t, sweepWasCalled)
}
