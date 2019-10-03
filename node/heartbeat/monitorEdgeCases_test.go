package heartbeat_test

import (
	mock2 "github.com/ElrondNetwork/elrond-go/node/heartbeat/mock"
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
	timer heartbeat.Timer,
) *heartbeat.Monitor {
	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerFake{},
		maxDurationPeerUnresponsive,
		map[uint32][]string{0: {pkValidator}},
		genesisTime,
		&mock.MessageHandlerStub{},
		storer,
		timer,
	)

	return mon
}

// v: |.................................
// o: |___________|.........|___________
func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})
	timer := &mock2.MockTimer{}
	unresponsiveDuration := time.Second * 3
	observerDownDuration := 3

	genesisTime := timer.Now()
	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	timer.IncrementSeconds(1)

	heartBeats1 := mon1.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats1))
	assert.Equal(t, 0, heartBeats1[0].TotalUpTime)
	assert.Equal(t, 1, heartBeats1[0].TotalDownTime)

	_ = createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	timer.IncrementSeconds(1)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	timer.IncrementSeconds(1)

	heartBeats2 := mon2.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats2))
	assert.Equal(t, 0, heartBeats2[0].TotalUpTime)
	assert.Equal(t, observerDownDuration, heartBeats2[0].TotalDownTime)
}

// v: |_________________________________
// o: |___________|.........|___________
func TestMonitor_ObserverGapValidatorOnline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})
	genesisTime := time.Unix(0, 0)
	unresponsiveDuration := time.Second * 3

	timer := &mock2.MockTimer{}

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats))
	assert.Equal(t, true, heartBeats[0].IsActive)
	assert.Equal(t, 0, heartBeats[0].TotalDownTime)
	assert.Equal(t, 0, heartBeats[0].TotalUpTime)

	timer.IncrementSeconds(1)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	heartBeats = mon2.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats))
	assert.Equal(t, true, heartBeats[0].IsActive)
	assert.Equal(t, 1, heartBeats[0].TotalDownTime)
	assert.Equal(t, 0, heartBeats[0].TotalUpTime)

	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats))
	assert.Equal(t, true, heartBeats[0].IsActive)
	assert.Equal(t, 0, heartBeats[0].TotalDownTime)
	assert.Equal(t, 1, heartBeats[0].TotalUpTime)
}
