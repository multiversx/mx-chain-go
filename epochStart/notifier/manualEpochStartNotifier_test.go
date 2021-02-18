package notifier

import (
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
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

	mesn.RegisterHandler(&genericMocks.ActionHandlerStub{})
	assert.Equal(t, 1, len(mesn.Handlers()))

	mesn.RegisterHandler(&genericMocks.ActionHandlerStub{})
	assert.Equal(t, 2, len(mesn.Handlers()))
}

func TestManualEpochStartNotifier_NewEpochWorks(t *testing.T) {
	t.Parallel()

	newEpoch := uint32(6483)
	calledEpoch := uint32(0)
	mesn := NewManualEpochStartNotifier()
	mesn.RegisterHandler(&genericMocks.ActionHandlerStub{
		EpochStartActionCalled: func(hdr data.HeaderHandler) {
			atomic.StoreUint32(&calledEpoch, hdr.GetEpoch())
		},
	})

	mesn.NewEpoch(newEpoch)

	assert.Equal(t, newEpoch, atomic.LoadUint32(&calledEpoch))
}

func TestManualEpochStartNotifier_NewEpochLowerValueShouldNotCallUpdate(t *testing.T) {
	t.Parallel()

	newEpoch := uint32(6483)
	calledEpoch := uint32(0)
	mesn := NewManualEpochStartNotifier()
	mesn.RegisterHandler(&genericMocks.ActionHandlerStub{
		EpochStartActionCalled: func(hdr data.HeaderHandler) {
			atomic.StoreUint32(&calledEpoch, hdr.GetEpoch())
		},
	})
	assert.Equal(t, uint32(0), mesn.CurrentEpoch())

	mesn.NewEpoch(newEpoch)
	assert.Equal(t, newEpoch, mesn.CurrentEpoch())

	mesn.NewEpoch(newEpoch - 1)
	assert.Equal(t, newEpoch, mesn.CurrentEpoch())

	assert.Equal(t, newEpoch, atomic.LoadUint32(&calledEpoch))
}
