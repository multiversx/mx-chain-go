package process

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// heartbeatMessageInfo retain the message info received from another node (identified by a public key)
type heartbeatMessageInfo struct {
	maxDurationPeerUnresponsive time.Duration
	maxInactiveTime             time.Duration
	totalUpTime                 time.Duration
	totalDownTime               time.Duration
	timeStamp                   time.Time
	lastUptimeDowntime          time.Time
	genesisTime                 time.Time
	versionNumber               string
	nodeDisplayName             string
	identity                    string
	peerType                    string
	receivedShardID             uint32
	computedShardID             uint32
	updateMutex                 sync.Mutex
	getTimeHandler              func() time.Time
	isActive                    bool
	nonce                       uint64
	numInstances                uint64
	peerSubType                 uint32
	pidString                   string
}

// newHeartbeatMessageInfo returns a new instance of a heartbeatMessageInfo
func newHeartbeatMessageInfo(
	maxDurationPeerUnresponsive time.Duration,
	peerType string,
	genesisTime time.Time,
	timer heartbeat.Timer,
) (*heartbeatMessageInfo, error) {

	if maxDurationPeerUnresponsive == 0 {
		return nil, heartbeat.ErrInvalidMaxDurationPeerUnresponsive
	}
	if check.IfNil(timer) {
		return nil, heartbeat.ErrNilTimer
	}

	hbmi := &heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		maxInactiveTime:             time.Duration(0),
		isActive:                    false,
		receivedShardID:             uint32(0),
		timeStamp:                   genesisTime,
		lastUptimeDowntime:          timer.Now(),
		totalUpTime:                 time.Duration(0),
		totalDownTime:               time.Duration(0),
		versionNumber:               "",
		nodeDisplayName:             "",
		identity:                    "",
		peerType:                    peerType,
		genesisTime:                 genesisTime,
		getTimeHandler:              timer.Now,
		nonce:                       0,
		numInstances:                0,
		peerSubType:                 0,
		pidString:                   "",
	}

	return hbmi, nil
}

// ComputeActive will update the isActive field
func (hbmi *heartbeatMessageInfo) ComputeActive(crtTime time.Time) {
	hbmi.updateMutex.Lock()
	defer hbmi.updateMutex.Unlock()
	validDuration := computeValidDuration(crtTime, hbmi)
	hbmi.isActive = hbmi.isActive && validDuration
	hbmi.updateTimes(crtTime)
}

func (hbmi *heartbeatMessageInfo) updateTimes(crtTime time.Time) {
	if crtTime.Sub(hbmi.genesisTime) < 0 {
		return
	}
	hbmi.updateMaxInactiveTimeDuration(crtTime)
	hbmi.updateUpAndDownTime(crtTime)
}

func computeValidDuration(crtTime time.Time, hbmi *heartbeatMessageInfo) bool {
	crtDuration := crtTime.Sub(hbmi.timeStamp)
	crtDuration = maxDuration(0, crtDuration)
	validDuration := crtDuration <= hbmi.maxDurationPeerUnresponsive
	return validDuration
}

// Will update the total time a node was up and down
func (hbmi *heartbeatMessageInfo) updateUpAndDownTime(crtTime time.Time) {
	if hbmi.lastUptimeDowntime.Sub(hbmi.genesisTime) < 0 {
		hbmi.lastUptimeDowntime = hbmi.genesisTime
	}

	if crtTime.Sub(hbmi.timeStamp) < 0 {
		return
	}

	lastDuration := crtTime.Sub(hbmi.lastUptimeDowntime)
	lastDuration = maxDuration(0, lastDuration)

	if lastDuration == 0 {
		return
	}

	uptime, downTime := hbmi.computeUptimeDowntime(crtTime, lastDuration)

	hbmi.totalUpTime += uptime
	hbmi.totalDownTime += downTime

	hbmi.isActive = uptime == lastDuration
	hbmi.lastUptimeDowntime = crtTime
}

func (hbmi *heartbeatMessageInfo) computeUptimeDowntime(
	crtTime time.Time,
	lastDuration time.Duration,
) (time.Duration, time.Duration) {
	durationSinceLastHeartbeat := crtTime.Sub(hbmi.timeStamp)
	insideActiveWindowAfterHeartheat := durationSinceLastHeartbeat <= hbmi.maxDurationPeerUnresponsive
	noHeartbeatReceived := hbmi.timeStamp == hbmi.genesisTime && !hbmi.isActive
	outSideActiveWindowAfterHeartbeat := durationSinceLastHeartbeat-lastDuration > hbmi.maxDurationPeerUnresponsive

	uptime := time.Duration(0)
	downtime := time.Duration(0)

	if noHeartbeatReceived || outSideActiveWindowAfterHeartbeat {
		downtime = lastDuration
		return uptime, downtime
	}

	if insideActiveWindowAfterHeartheat {
		uptime = lastDuration
		return uptime, downtime
	}

	downtime = durationSinceLastHeartbeat - hbmi.maxDurationPeerUnresponsive
	uptime = lastDuration - downtime

	return uptime, downtime
}

// HeartbeatReceived processes a new message arrived from a peer
func (hbmi *heartbeatMessageInfo) HeartbeatReceived(
	computedShardID uint32,
	receivedShardID uint32,
	version string,
	nodeDisplayName string,
	identity string,
	peerType string,
	nonce uint64,
	numInstances uint64,
	peerSubType uint32,
	pidString string,
) {
	hbmi.updateMutex.Lock()
	defer hbmi.updateMutex.Unlock()
	crtTime := hbmi.getTimeHandler()

	hbmi.computedShardID = computedShardID
	hbmi.receivedShardID = receivedShardID
	hbmi.versionNumber = version
	hbmi.nodeDisplayName = nodeDisplayName
	hbmi.identity = identity
	hbmi.peerType = peerType

	hbmi.updateTimes(crtTime)
	hbmi.timeStamp = crtTime
	hbmi.isActive = true
	hbmi.nonce = nonce
	hbmi.numInstances = numInstances
	hbmi.peerSubType = peerSubType
	hbmi.pidString = pidString
}

// UpdateShardAndPeerType - updates the shard and peerType only for a heartbeat message info
func (hbmi *heartbeatMessageInfo) UpdateShardAndPeerType(
	computedShardID uint32,
	peerType string,
) {
	hbmi.updateMutex.Lock()
	defer hbmi.updateMutex.Unlock()

	hbmi.computedShardID = computedShardID
	hbmi.peerType = peerType
}

func (hbmi *heartbeatMessageInfo) updateMaxInactiveTimeDuration(currentTime time.Time) {
	crtDuration := currentTime.Sub(hbmi.timeStamp)
	crtDuration = maxDuration(0, crtDuration)

	greaterDurationThanMax := hbmi.maxInactiveTime < crtDuration
	currentTimeAfterGenesis := hbmi.genesisTime.Sub(currentTime) < 0

	if greaterDurationThanMax && currentTimeAfterGenesis {
		hbmi.maxInactiveTime = crtDuration
	}
}

func maxDuration(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}

	return second
}

// GetIsActive will return true if the peer is set as active
func (hbmi *heartbeatMessageInfo) GetIsActive() bool {
	hbmi.updateMutex.Lock()
	defer hbmi.updateMutex.Unlock()
	isActive := hbmi.isActive
	return isActive
}

// GetIsValidator will return true is the peer is a validator
func (hbmi *heartbeatMessageInfo) GetIsValidator() bool {
	hbmi.updateMutex.Lock()
	defer hbmi.updateMutex.Unlock()
	return hbmi.peerType == string(core.EligibleList) || hbmi.peerType == string(core.WaitingList)
}
