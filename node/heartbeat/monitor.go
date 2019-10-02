package heartbeat

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var log = logger.DefaultLogger()

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	maxDurationPeerUnresponsive time.Duration
	marshalizer                 marshal.Marshalizer
	heartbeatMessages           map[string]*heartbeatMessageInfo
	mutHeartbeatMessages        sync.RWMutex
	pubKeysMap                  map[uint32][]string
	fullPeersSlice              [][]byte
	mutPubKeysMap               sync.RWMutex
	appStatusHandler            core.AppStatusHandler
	genesisTime                 time.Time
	messageHandler              MessageHandler
	storer                      HeartbeatStorageHandler
	timeHandler                 func() time.Time
}

// NewMonitor returns a new monitor instance
func NewMonitor(
	marshalizer marshal.Marshalizer,
	maxDurationPeerUnresponsive time.Duration,
	pubKeysMap map[uint32][]string,
	genesisTime time.Time,
	messageHandler MessageHandler,
	storer HeartbeatStorageHandler,
	timeHandler func() time.Time,
) (*Monitor, error) {

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if len(pubKeysMap) == 0 {
		return nil, ErrEmptyPublicKeysMap
	}
	if messageHandler == nil || messageHandler.IsInterfaceNil() {
		return nil, ErrNilMessageHandler
	}
	if storer == nil || storer.IsInterfaceNil() {
		return nil, ErrNilHeartbeatStorer
	}
	if timeHandler == nil {
		return nil, ErrNilGetTimeHandler
	}

	mon := &Monitor{
		marshalizer:                 marshalizer,
		heartbeatMessages:           make(map[string]*heartbeatMessageInfo),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		appStatusHandler:            &statusHandler.NilStatusHandler{},
		genesisTime:                 genesisTime,
		messageHandler:              messageHandler,
		storer:                      storer,
		timeHandler:                 timeHandler,
	}

	err := mon.storer.UpdateGenesisTime(genesisTime)
	if err != nil {
		return nil, err
	}

	err = mon.initializeHeartbeatMessagesInfo(pubKeysMap)
	if err != nil {
		return nil, err
	}

	err = mon.loadRestOfPubKeysFromStorage()
	if err != nil {
		log.Warn(fmt.Sprintf("heartbeat can't load public keys from storage: %s", err.Error()))
	}

	return mon, nil
}

func (m *Monitor) initializeHeartbeatMessagesInfo(pubKeysMap map[uint32][]string) error {
	pubKeysMapCopy := make(map[uint32][]string, 0)
	for shardId, pubKeys := range pubKeysMap {
		for _, pubkey := range pubKeys {
			err := m.loadHbmiFromStorer(pubkey)
			if err != nil { // if pubKey not found in DB, create a new instance
				getTimeHandler := func() time.Time { return time.Now() }
				mhbi, errNewHbmi := newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, true, m.genesisTime, getTimeHandler)
				if errNewHbmi != nil {
					return errNewHbmi
				}

				mhbi.genesisTime = m.genesisTime
				mhbi.computedShardID = shardId
				m.heartbeatMessages[pubkey] = mhbi
			}
			pubKeysMapCopy[shardId] = append(pubKeysMapCopy[shardId], pubkey)
		}
	}

	m.pubKeysMap = pubKeysMapCopy
	return nil
}

func (m *Monitor) loadRestOfPubKeysFromStorage() error {
	peersSlice, err := m.storer.LoadKeys()
	if err != nil {
		return err
	}

	for _, peer := range peersSlice {
		if _, ok := m.heartbeatMessages[string(peer)]; !ok { // peer not in nodes map
			err = m.loadHbmiFromStorer(string(peer))
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (m *Monitor) loadHbmiFromStorer(pubKey string) error {
	hbmiDTO, err := m.storer.LoadHbmiDTO(pubKey)
	if err != nil {
		return err
	}

	receivedHbmi := m.convertFromExportedStruct(*hbmiDTO, m.maxDurationPeerUnresponsive)
	receivedHbmi.getTimeHandler = func() time.Time { return time.Now() }
	receivedHbmi.lastUptimeDowntime = time.Now()
	receivedHbmi.genesisTime = m.genesisTime
	if receivedHbmi.timeStamp == m.genesisTime && time.Now().Sub(m.genesisTime) > 0 {
		receivedHbmi.totalDownTime = Duration{time.Now().Sub(m.genesisTime)}
	}

	m.heartbeatMessages[pubKey] = &receivedHbmi

	return nil
}

// SetAppStatusHandler will set the AppStatusHandler which will be used for monitoring
func (m *Monitor) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return ErrNilAppStatusHandler
	}

	m.appStatusHandler = ash
	return nil
}

// ProcessReceivedMessage satisfies the p2p.MessageProcessor interface so it can be called
// by the p2p subsystem each time a new heartbeat message arrives
func (m *Monitor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hbRecv, err := m.messageHandler.CreateHeartbeatFromP2pMessage(message)
	if err != nil {
		return err
	}

	//message is validated, process should be done async, method can return nil
	go m.addHeartbeatMessageToMap(hbRecv)

	return nil
}

func (m *Monitor) addHeartbeatMessageToMap(hb *Heartbeat) {
	m.mutHeartbeatMessages.Lock()
	defer m.mutHeartbeatMessages.Unlock()

	pubKeyStr := string(hb.Pubkey)
	pe, ok := m.heartbeatMessages[pubKeyStr]
	if pe == nil || !ok {
		var err error
		getTimeHandler := func() time.Time { return time.Now() }
		pe, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, false, m.genesisTime, getTimeHandler)
		if err != nil {
			log.Error(err.Error())
			return
		}
		m.heartbeatMessages[pubKeyStr] = pe
	}

	computedShardID := m.computeShardID(pubKeyStr)
	pe.HeartbeatReceived(computedShardID, hb.ShardID, hb.VersionNumber, hb.NodeDisplayName)
	//m.storeHeartbeat(hb.Pubkey)
	hbDTO := m.convertToExportedStruct(pe)
	err := m.storer.SavePubkeyData(hb.Pubkey, &hbDTO)
	if err != nil {
		log.Warn(fmt.Sprintf("cannot save heartbeat to db: %s", err.Error()))
	}
	m.updateAllHeartbeatMessages()
}

func (m *Monitor) addPeerToFullPeersSlice(pubKey []byte) {
	if !m.isPeerInFullPeersSlice(pubKey) {
		m.fullPeersSlice = append(m.fullPeersSlice, pubKey)
		err := m.storer.SaveKeys(m.fullPeersSlice)
		if err != nil {
			log.Warn(fmt.Sprintf("can't store the keys slice: %s", err.Error()))
		}
	}
}

func (m *Monitor) isPeerInFullPeersSlice(pubKey []byte) bool {
	for _, peer := range m.fullPeersSlice {
		if bytes.Equal(peer, pubKey) {
			return true
		}
	}

	return false
}

func (m *Monitor) computeShardID(pubkey string) uint32 {
	// TODO : the shard ID will be recomputed at the end of an epoch / beginning of a new one.
	//  For the moment, just find the shard ID from a copy of the initial pub keys map
	m.mutPubKeysMap.RLock()
	defer m.mutPubKeysMap.RUnlock()
	for shardID, pubKeysSlice := range m.pubKeysMap {
		for _, pKey := range pubKeysSlice {
			if pKey == pubkey {
				return shardID
			}
		}
	}

	// if not found, return the latest known computed shard ID
	return m.heartbeatMessages[pubkey].computedShardID
}

func (m *Monitor) updateAllHeartbeatMessages() {
	counterActiveValidators := 0
	counterConnectedNodes := 0
	for _, v := range m.heartbeatMessages {
		//TODO change here
		v.updateFields(time.Now())

		if v.isActive {
			counterConnectedNodes++

			if v.isValidator {
				counterActiveValidators++
			}
		}
	}

	m.appStatusHandler.SetUInt64Value(core.MetricLiveValidatorNodes, uint64(counterActiveValidators))
	m.appStatusHandler.SetUInt64Value(core.MetricConnectedNodes, uint64(counterConnectedNodes))
}

// GetHeartbeats returns the heartbeat status
func (m *Monitor) GetHeartbeats() []PubKeyHeartbeat {
	m.mutHeartbeatMessages.RLock()
	status := make([]PubKeyHeartbeat, len(m.heartbeatMessages))

	idx := 0
	for k, v := range m.heartbeatMessages {
		status[idx] = PubKeyHeartbeat{
			HexPublicKey:    hex.EncodeToString([]byte(k)),
			TimeStamp:       v.timeStamp,
			MaxInactiveTime: v.maxInactiveTime,
			IsActive:        v.isActive,
			ReceivedShardID: v.receivedShardID,
			ComputedShardID: v.computedShardID,
			TotalUpTime:     int(v.totalUpTime.Seconds()),
			TotalDownTime:   int(v.totalDownTime.Seconds()),
			VersionNumber:   v.versionNumber,
			IsValidator:     v.isValidator,
			NodeDisplayName: v.nodeDisplayName,
		}
		if status[idx].TimeStamp == m.genesisTime && time.Now().Sub(m.genesisTime) > 0 {
			status[idx].TotalDownTime = int(time.Now().Sub(m.genesisTime).Seconds())
		}
		idx++
	}
	m.mutHeartbeatMessages.RUnlock()

	sort.Slice(status, func(i, j int) bool {
		return strings.Compare(status[i].HexPublicKey, status[j].HexPublicKey) < 0
	})

	return status
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *Monitor) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

func (m *Monitor) convertToExportedStruct(v *heartbeatMessageInfo) HeartbeatDTO {
	return HeartbeatDTO{
		TimeStamp:          v.timeStamp,
		MaxInactiveTime:    v.maxInactiveTime,
		IsActive:           v.isActive,
		ReceivedShardID:    v.receivedShardID,
		ComputedShardID:    v.computedShardID,
		TotalUpTime:        v.totalUpTime,
		TotalDownTime:      v.totalDownTime,
		VersionNumber:      v.versionNumber,
		IsValidator:        v.isValidator,
		NodeDisplayName:    v.nodeDisplayName,
		LastUptimeDowntime: v.lastUptimeDowntime,
		GenesisTime:        v.genesisTime,
	}
}

func (m *Monitor) convertFromExportedStruct(hbDTO HeartbeatDTO, maxDuration time.Duration) heartbeatMessageInfo {
	hbmi := heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDuration,
		maxInactiveTime:             hbDTO.MaxInactiveTime,
		timeStamp:                   hbDTO.TimeStamp,
		isActive:                    hbDTO.IsActive,
		totalUpTime:                 hbDTO.TotalUpTime,
		totalDownTime:               hbDTO.TotalDownTime,
		receivedShardID:             hbDTO.ReceivedShardID,
		computedShardID:             hbDTO.ComputedShardID,
		versionNumber:               hbDTO.VersionNumber,
		nodeDisplayName:             hbDTO.NodeDisplayName,
		isValidator:                 hbDTO.IsValidator,
		lastUptimeDowntime:          hbDTO.LastUptimeDowntime,
		genesisTime:                 hbDTO.GenesisTime,
	}

	return hbmi
}
