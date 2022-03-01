package process_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/process"
	"github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
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
		AppStatusHandler:                   &statusHandlerMock.AppStatusHandlerStub{},
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
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	time.Sleep(sleepDuration)
	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(60)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(61)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(70)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	time.Sleep(20 * time.Millisecond)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(151)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	time.Sleep(sleepDuration)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)

	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(120)
	mon2.RefreshHeartbeatMessageInfo()
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})

	timer.SetSeconds(160)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)
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
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)
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
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(201)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)
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
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.False(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)
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
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(tenSeconds)
	mon1.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twentySeconds)
	mon1.RefreshHeartbeatMessageInfo()
	heartBeats = mon1.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredSeconds)
	mon2 := createMonitor(storer, genesisTime, unresponsiveDuration, timer)
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(hundredFiftySeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)

	timer.SetSeconds(twoHundredSeconds)
	mon2.AddHeartbeatMessageToMap(&data.Heartbeat{Pubkey: []byte(pkValidator)})
	mon2.RefreshHeartbeatMessageInfo()
	heartBeats = mon2.GetHeartbeats()
	assert.True(t, heartBeats[0].IsActive)
}
