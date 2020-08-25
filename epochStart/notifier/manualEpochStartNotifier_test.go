package notifier

import (
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewManualEpochStartNotifier(t *testing.T) {
	t.Parallel()

	mesn := NewManualEpochStartNotifier()

	assert.False(t, check.IfNil(mesn))
}

func TestManualEpochStartNotifier_RegisterHandler(t *testing.T) {
	t.Parallel()

	mesn := NewManualEpochStartNotifier()
	assert.Equal(t, 0, len(mesn.Handlers()))

	mesn.RegisterHandler(&mock.ActionHandlerStub{})
	assert.Equal(t, 1, len(mesn.Handlers()))

	mesn.RegisterHandler(&mock.ActionHandlerStub{})
	assert.Equal(t, 2, len(mesn.Handlers()))
}

func TestManualEpochStartNotifier_NewEpochWorks(t *testing.T) {
	t.Parallel()

	newEpoch := uint32(6483)
	epoch := uint32(0)
	mesn := NewManualEpochStartNotifier()
	mesn.RegisterHandler(&mock.ActionHandlerStub{
		EpochStartActionCalled: func(hdr data.HeaderHandler) {
			atomic.StoreUint32(&epoch, hdr.GetEpoch())
		},
	})

	mesn.NewEpoch(newEpoch)

	assert.Equal(t, newEpoch, atomic.LoadUint32(&epoch))
}
