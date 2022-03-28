package sender

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockBaseArgs() argBaseSender {
	return argBaseSender{
		messenger:                 &p2pmocks.MessengerStub{},
		marshaller:                &testscommon.MarshalizerMock{},
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

	var d1, d2, d3 time.Duration
	for i := 0; i < 100; i++ {
		d1 = bs.computeRandomDuration()
		d2 = bs.computeRandomDuration()
		d3 = bs.computeRandomDuration()
		if d1 != d2 && d2 != d3 && d1 != d3 {
			break
		}
	}

	assert.False(t, d1 == d2)
	assert.False(t, d2 == d3)
}
