package heartbeat_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

var pkValidator = "pk"

func createMonitor(
	storer heartbeat.HeartbeatStorageHandler,
	genesisTime time.Time,
	maxDurationPeerUnresponsive time.Duration,
	timeHandler func() time.Time,
) *heartbeat.Monitor {
	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerFake{},
		maxDurationPeerUnresponsive,
		map[uint32][]string{0: {pkValidator}},
		genesisTime,
		&mock.MessageHandlerStub{},
		storer,
		timeHandler,
	)

	return mon
}

// v: |.................................
// o: |___________|.........|___________
func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})
	genesisTime := time.Unix(0, 0)
	unresponsiveDuration := time.Second * 3
	observerDownDuration := 1

	_ = createMonitor(storer, genesisTime, unresponsiveDuration, func() time.Time {
		return time.Unix(0, 0)
	})

	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, func() time.Time {
		return time.Unix(int64(observerDownDuration), 0)
	})

	heartBeats := mon2.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats))
	assert.Equal(t, 0, heartBeats[0].TotalUpTime)
	assert.Equal(t, observerDownDuration, heartBeats[0].TotalDownTime)
}
