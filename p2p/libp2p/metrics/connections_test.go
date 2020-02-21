package metrics_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConnections_EmptyFunctionsDoNotPanicWhenCalled(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "test should not have failed")
		}
	}()

	cdm := metrics.NewConnections()

	cdm.ClosedStream(nil, nil)
	cdm.Listen(nil, nil)
	cdm.ListenClose(nil, nil)
	cdm.OpenedStream(nil, nil)
}

func TestConnections_ResetNumConnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cdm := metrics.NewConnections()

	cdm.Connected(nil, nil)
	cdm.Connected(nil, nil)

	existing := cdm.ResetNumConnections()
	assert.Equal(t, uint32(2), existing)

	existing = cdm.ResetNumConnections()
	assert.Equal(t, uint32(0), existing)
}

func TestConnsDisconnsMetric_ResetNumDisconnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cdm := metrics.NewConnections()

	cdm.Disconnected(nil, nil)
	cdm.Disconnected(nil, nil)

	existing := cdm.ResetNumDisconnections()
	assert.Equal(t, uint32(2), existing)

	existing = cdm.ResetNumDisconnections()
	assert.Equal(t, uint32(0), existing)
}
