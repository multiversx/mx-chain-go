package connectionMonitor

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNilConnectionMonitor_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	ncm := &NilConnectionMonitor{}

	ncm.OpenedStream(nil, nil)
	ncm.ListenClose(nil, nil)
	ncm.Listen(nil, nil)
	ncm.Disconnected(nil, nil)
	ncm.ClosedStream(nil, nil)
	ncm.Connected(nil, nil)
}

func TestNilConnectionMonitor_SetSharderShouldErr(t *testing.T) {
	t.Parallel()

	ncm := &NilConnectionMonitor{}
	err := ncm.SetSharder(nil)

	assert.True(t, errors.Is(err, p2p.ErrIncompatibleMethodCalled))
}
