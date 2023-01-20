package sender

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockBaseArgs() argBaseSender {
	return argBaseSender{
		messenger:                 &p2pmocks.MessengerStub{},
		marshaller:                &marshallerMock.MarshalizerMock{},
		topic:                     "topic",
		timeBetweenSends:          time.Second,
		timeBetweenSendsWhenError: time.Second,
		thresholdBetweenSends:     0.1,
		privKey:                   &cryptoMocks.PrivateKeyStub{},
		redundancyHandler:         &mock.RedundancyHandlerStub{},
	}
}

func TestBaseSender_computeRandomDuration(t *testing.T) {
	t.Parallel()

	bs := createBaseSender(createMockBaseArgs())
	assert.NotNil(t, bs)

	var d1, d2, d3 time.Duration
	for i := 0; i < 100; i++ {
		d1 = bs.computeRandomDuration(bs.timeBetweenSends)
		d2 = bs.computeRandomDuration(bs.timeBetweenSends)
		d3 = bs.computeRandomDuration(bs.timeBetweenSends)
		if d1 != d2 && d2 != d3 && d1 != d3 {
			break
		}
	}

	assert.False(t, d1 == d2)
	assert.False(t, d2 == d3)
}
