package heartbeat_test

//
//import (
//	"testing"
//	"time"
//
//	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
//	"github.com/ElrondNetwork/elrond-go/node/mock"
//	"github.com/ElrondNetwork/elrond-go/storage"
//	"github.com/stretchr/testify/assert"
//)
//
//var pkValidator = "pk"
//
//func createMonitor(
//	genesisTime time.Time,
//	maxDurationPeerUnresponsive time.Duration,
//) *heartbeat.Monitor {
//
//	mon, _ := heartbeat.NewMonitor(
//		&mock.MarshalizerFake{},
//		maxDurationPeerUnresponsive,
//		map[uint32][]string{0: {pkValidator}},
//		genesisTime,
//		&mock.MessageHandlerStub{},
//	)
//
//	return mon
//}
//
//// v: |.................................
//// o: |___________|.........|___________
//func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
//	t.Parallel()
//
//	db := mock.NewStorerMock()
//	genesisTime := time.Now()
//	unresponsiveDuration := time.Second * 3
//	observerDownDuration := time.Second
//
//	_ = createMonitor(db, genesisTime, unresponsiveDuration)
//	time.Sleep(observerDownDuration)
//	mon2 := createMonitor(db, genesisTime, unresponsiveDuration)
//
//	heartBeats := mon2.GetHeartbeats()
//	assert.Equal(t, 1, len(heartBeats))
//	assert.Equal(t, 0, heartBeats[0].TotalUpTime)
//	//assert.True(t, heartBeats[0].TotalUpTime)
//}
