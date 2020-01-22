package heartbeat

import (
	"fmt"
	"testing"
	"time"
	"unsafe"
)

func (m *Monitor) GetMessages() map[string]*heartbeatMessageInfo {
	return m.heartbeatMessages
}

func Test_test(t *testing.T) {
	fmt.Println(unsafe.Sizeof(heartbeatMessageInfo{}))
	fmt.Println("duration", unsafe.Sizeof(heartbeatMessageInfo{}.maxDurationPeerUnresponsive))
	fmt.Println("hbt duration", unsafe.Sizeof(heartbeatMessageInfo{}.maxInactiveTime))
	fmt.Println("time", unsafe.Sizeof(heartbeatMessageInfo{}.lastUptimeDowntime))
	fmt.Println("string", unsafe.Sizeof(heartbeatMessageInfo{}.nodeDisplayName))
	fmt.Println("mutex", unsafe.Sizeof(heartbeatMessageInfo{}.updateMutex))
	fmt.Println("func pointer", unsafe.Sizeof(heartbeatMessageInfo{}.getTimeHandler))
	cxx := make(chan bool)
	fmt.Println("channel", unsafe.Sizeof(cxx))
	var aaa uint32
	fmt.Println("uint32", unsafe.Sizeof(aaa))
	var bbb uint64
	fmt.Println("uint64", unsafe.Sizeof(bbb))
	var ccc int
	fmt.Println("int", unsafe.Sizeof(ccc))
	var ddd float64
	fmt.Println("float64", unsafe.Sizeof(ddd))

	fmt.Println("PubKeyHeartbeat", unsafe.Sizeof(PubKeyHeartbeat{}))

	fmt.Println("Heartbeat dto", unsafe.Sizeof(HeartbeatDTO{}))

	//tt := reflect.TypeOf(HeartbeatDTO{})
	//for i := 0; i < tt.NumField(); i++ {
	//	fmt.Printf("%+v\n", tt.Field(i))
	//}
}

func (m *Monitor) SetMessages(messages map[string]*heartbeatMessageInfo) {
	m.heartbeatMessages = messages
}

func (m *Monitor) GetHbmi(tmstp time.Time) *heartbeatMessageInfo {
	return &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: 0,
		maxInactiveTime:             Duration{},
		totalUpTime:                 Duration{},
		totalDownTime:               Duration{},
		getTimeHandler:              nil,
		timeStamp:                   time.Time{},
		isActive:                    false,
		receivedShardID:             0,
		computedShardID:             0,
		versionNumber:               "",
		nodeDisplayName:             "",
		isValidator:                 false,
		lastUptimeDowntime:          time.Time{},
		genesisTime:                 time.Time{},
	}
}

func (m *Monitor) SendHeartbeatMessage(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

func (m *Monitor) AddHeartbeatMessageToMap(hb *Heartbeat) {
	m.addHeartbeatMessageToMap(hb)
}

func NewHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	isValidator bool,
	genesisTime time.Time,
	timer Timer,
) (*heartbeatMessageInfo, error) {
	return newHeartbeatMessageInfo(
		maxDurationPeerUnresponsive,
		isValidator,
		genesisTime,
		timer,
	)
}

func (hbmi *heartbeatMessageInfo) GetTimeStamp() time.Time {
	return hbmi.timeStamp
}

func (hbmi *heartbeatMessageInfo) GetReceiverShardId() uint32 {
	return hbmi.receivedShardID
}

func (hbmi *heartbeatMessageInfo) GetTotalUpTime() Duration {
	return hbmi.totalUpTime
}

func (hbmi *heartbeatMessageInfo) GetTotalDownTime() Duration {
	return hbmi.totalDownTime
}

func VerifyLengths(hbmi *Heartbeat) error {
	return verifyLengths(hbmi)
}

func GetMaxSizeInBytes() int {
	return maxSizeInBytes
}
