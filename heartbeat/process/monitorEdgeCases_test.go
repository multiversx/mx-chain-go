package process_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/process"
	"github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	"github.com/stretchr/testify/assert"
)

var pkValidator = "pk"

func createMonitor(
	storer heartbeat.HeartbeatStorageHandler,
	genesisTime time.Time,
	maxDurationPeerUnresponsive time.Duration,
	timer heartbeat.Timer,
) *process.Monitor {

	arg := process.ArgHeartbeatMonitor{
		Marshalizer:                        &mock.MarshalizerMock{},
		MaxDurationPeerUnresponsive:        maxDurationPeerUnresponsive,
		PubKeysMap:                         map[uint32][]string{0: {pkValidator}},
		GenesisTime:                        genesisTime,
		MessageHandler:                     &mock.MessageHandlerStub{},
		Storer:                             storer,
		PeerTypeProvider:                   &mock.PeerTypeProviderStub{},
		Timer:                              timer,
		AntifloodHandler:                   createMockP2PAntifloodHandler(),
		HardforkTrigger:                    &mock.HardforkTriggerStub{},
		ValidatorPubkeyConverter:           mock.NewPubkeyConverterMock(32),
		HeartbeatRefreshIntervalInSec:      1,
		HideInactiveValidatorIntervalInSec: 600,
		AppStatusHandler:                   &mock.AppStatusHandlerStub{},
	}
	mon, _ := process.NewMonitor(arg)

	return mon
}

const unresponsiveDuration = time.Second * 50
const sleepDuration = time.Millisecond * 10
const tenSeconds = 10
const twentySeconds = 20
const hundredSeconds = 100
const hundredFiftySeconds = 150
const twoHundredSeconds = 200

// v: stands for validator and o: stands for observer
// dotted line means the node (v/o) is offline
// line means the node is online
// values below are timestamps in seconds

// v: |............______________________
// o: |___________|.........|___________
//    0           20        100   150   200

// v: |.................................
// o: |___________|.........|___________
//    0           20        100   150   200
func TestMonitor_ObserverGapValidatorOffline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})
	timer := mock.NewTimerMock()

	genesisTime := timer.Now()
	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	time.Sleep(sleepDuration)
	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 20)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	time.Sleep(sleepDuration)
	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 150)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 200)
}

// v: |_________________________________
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorOnline(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 100, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 150, 0)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 200, 0)
}

// v: |_____................
// o: |_____________________
//    0    10    60  61   70
func TestMonitor_ObserverGapValidatorActiveUnitlMaxPeriodEnds(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(60)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 60, 0)

	timer.SetSeconds(61)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 60, 1)

	timer.SetSeconds(70)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 60, 10)
}

// v: |_____.............................
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline1(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	time.Sleep(20 * time.Millisecond)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 100, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 150, 0)

	timer.SetSeconds(151)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 150, 1)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 150, 50)
}

// v: |____________________________......
// o: |___________|.........|___________
//    0    10    20        100   120   200
func TestMonitor_ObserverGapValidatorPartlyOnline2(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	time.Sleep(sleepDuration)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 100, 0)

	timer.SetSeconds(120)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 120, 0)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 170, 30)
}

// v: |____________________________......
// o: |___________|.........|___________
//    0    10    20        100   160   200
func TestMonitor_ObserverGapValidatorPartlyOnline3(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 100, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	timer.SetSeconds(160)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 160, 0)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 200, 0)
}

// v: |.....____________________________
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline4(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	time.Sleep(20 * time.Millisecond)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 90, 10)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 140, 10)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 190, 10)
}

// v: |.....______________________.......
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline5(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 10)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 90, 10)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 140, 10)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 190, 10)

	timer.SetSeconds(201)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 190, 11)
}

// v: |............................_____
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline6(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	time.Sleep(20 * time.Millisecond)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 10)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 20)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], false, 0, 100)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 150)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 50, 150)
}

// v: |_____......................._____
// o: |___________|.........|___________
//    0    10    20        100   150   200
func TestMonitor_ObserverGapValidatorPartlyOnline7(t *testing.T) {
	t.Parallel()

	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerMock{})

	timer := mock.NewTimerMock()
	genesisTime := timer.Now()

	mon1 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon1.RefreshHeartbeatMessageInfo()
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats := mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 0, 0)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 10, 0)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 20, 0)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 100, 0)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 150, 0)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	verifyHeartBeat(t, heartBeats[0], true, 200, 0)
}

func verifyHeartBeat(
	t *testing.T,
	heartBeat data.PubKeyHeartbeat,
	isActive bool,
	upTimeSeconds int,
	downTimeSeconds int,
) {
	assert.Equal(t, isActive, heartBeat.IsActive)
	assert.Equal(t, int64(upTimeSeconds), heartBeat.TotalUpTime)
	assert.Equal(t, int64(downTimeSeconds), heartBeat.TotalDownTime)
}
