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

const unresponsiveDuration = time.Second * 50
const tenSeconds = 10
const twentySeconds = 20
const hundredSeconds = 100
const hundredFiftySeconds = 150
const twoHundredSeconds = 200

// v: |.................................
// o: |___________|.........|___________
//    0           20        100   150   200
func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})
	timer := &mock2.MockTimer{}

	genesisTime := timer.Now()
	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	timer.SetSeconds(tenSeconds)
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 10)

	timer.SetSeconds(twentySeconds)
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 20)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	timer.SetSeconds(hundredFiftySeconds)
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 70)

	timer.SetSeconds(twoHundredSeconds)
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 120)
}

// v: |_________________________________
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorOnline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 70, 0)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 120, 0)
}

// v: |_____................
// o: |_____________________
//    0    10    60  61   70
func TestMonitor_ObserverGapValidatorActiveUnitlMaxPeriodEnds(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(60)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 60, 0)

	timer.SetSeconds(61)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 60, 1)

	timer.SetSeconds(70)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 60, 10)

}

// v: |_____.............................
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline1(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 20, 0)

	timer.SetSeconds(hundredFiftySeconds)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 20, 50)

	timer.SetSeconds(twoHundredSeconds)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 20, 100)
}

// v: |____________________________......
// o: |___________|.........|___________
//    0    10    20        100   120   200
func TestMonitor_ObserverGapValidatorPartlyOnline2(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(120)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 40, 0)

	timer.SetSeconds(twoHundredSeconds)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 40, 80)
}

// v: |____________________________......
// o: |___________|.........|___________
//    0    10    20        100   160   200
func TestMonitor_ObserverGapValidatorPartlyOnline3(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})

	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(160)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 60)

	timer.SetSeconds(twoHundredSeconds)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 60, 60)
}

// v: |.....____________________________
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline4(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 60, 10)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 110, 10)
}

// v: |.....______________________.......
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline5(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 60, 10)

	timer.SetSeconds(twoHundredSeconds)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 110, 10)
}

// v: |............................_____
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline6(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 10)

	timer.SetSeconds(twentySeconds)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 20)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 20)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 70)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 50, 70)
}

// v: |_____......................._____
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline7(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})

	timer := &mock2.MockTimer{}
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	//mon1.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	//mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 20, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 50)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&heartbeat.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 70, 50)
}

func verifyHeartBeat(
	t *testing.T,
	heartBeat heartbeat.PubKeyHeartbeat,
	isActive bool,
	upTimeSeconds int,
	downTimeSeconds int,
) {
	assert.Equal(t, isActive, heartBeat.IsActive)
	assert.Equal(t, upTimeSeconds, heartBeat.TotalUpTime)
	assert.Equal(t, downTimeSeconds, heartBeat.TotalDownTime)
}
