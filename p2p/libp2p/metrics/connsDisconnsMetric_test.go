package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnsDisconnsMetric_EmptyFunctionsDoNotPanicWhenCalled(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "test should not have failed")
		}
	}()

	cdm := NewConnsDisconnsMetric()

	cdm.ClosedStream(nil, nil)
	cdm.Listen(nil, nil)
	cdm.ListenClose(nil, nil)
	cdm.OpenedStream(nil, nil)
}

func TestConnsDisconnsMetric_ResetNumConnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cdm := NewConnsDisconnsMetric()

	cdm.Connected(nil, nil)
	cdm.Connected(nil, nil)

	existing := cdm.ResetNumConnections()
	assert.Equal(t, uint32(2), existing)

	existing = cdm.ResetNumConnections()
	assert.Equal(t, uint32(0), existing)
}

func TestConnsDisconnsMetric_ResetNumDisconnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cdm := NewConnsDisconnsMetric()

	cdm.Disconnected(nil, nil)
	cdm.Disconnected(nil, nil)

	existing := cdm.ResetNumDisconnections()
	assert.Equal(t, uint32(2), existing)

	existing = cdm.ResetNumDisconnections()
	assert.Equal(t, uint32(0), existing)
}
