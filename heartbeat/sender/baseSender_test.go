package sender

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/stretchr/testify/assert"
)

func createMockBaseArgs() argBaseSender {
	return argBaseSender{
		messenger:                 &mock.MessengerStub{},
		marshaller:                &mock.MarshallerMock{},
		topic:                     "topic",
		timeBetweenSends:          time.Second,
		timeBetweenSendsWhenError: time.Second,
		thresholdBetweenSends:     0.1,
	}
}

func TestBaseSender_computeRandomDuration(t *testing.T) {
	t.Parallel()

	bs := createBaseSender(createMockBaseArgs())
	assert.NotNil(t, bs)

	d1 := bs.computeRandomDuration()
	d2 := bs.computeRandomDuration()
	d3 := bs.computeRandomDuration()
	assert.False(t, d1 == d2)
	assert.False(t, d2 == d3)
}
