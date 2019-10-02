package heartbeat_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

var pkValidator = "pk"

func createMonitor(
	genesisTime time.Time,
	maxDurationPeerUnresponsive time.Duration,
	timeHandler func() time.Time,
) *heartbeat.Monitor {

	heartbeatStorerStub := &mock.HeartbeatStorerStub{
		UpdateGenesisTimeCalled: func(genesisTime time.Time) error {
			return nil
		},
		LoadHbmiDTOCalled: func(pubKey string) (*heartbeat.HeartbeatDTO, error) {
			return nil, errors.New("not found")
		},
		LoadKeysCalled: func() ([][]byte, error) {
			return nil, nil
		},
		SavePubkeyDataCalled: func(pubkey []byte, heartbeat *heartbeat.HeartbeatDTO) error {
			return nil
		},
	}

	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerFake{},
		maxDurationPeerUnresponsive,
		map[uint32][]string{0: {pkValidator}},
		genesisTime,
		&mock.MessageHandlerStub{},
		heartbeatStorerStub,
		timeHandler,
	)

	return mon
}

// v: |.................................
// o: |___________|.........|___________
func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
	t.Parallel()

	genesisTime := time.Unix(0, 0)
	unresponsiveDuration := time.Second * 3
	observerDownDuration := 1

	_ = createMonitor(genesisTime, unresponsiveDuration, func() time.Time {
		return time.Unix(0, 0)
	})

	mon2 := createMonitor(genesisTime, unresponsiveDuration, func() time.Time {
		return time.Unix(int64(observerDownDuration), 0)
	})

	heartBeats := mon2.GetHeartbeats()
	assert.Equal(t, 1, len(heartBeats))
	assert.Equal(t, 0, heartBeats[0].TotalUpTime)
	assert.Equal(t, observerDownDuration, heartBeats[0].TotalDownTime)
}
