package heartbeat

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

const peersKeysDbEntry = "keys"

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	singleSigner                crypto.SingleSigner
	maxDurationPeerUnresponsive time.Duration
	keygen                      crypto.KeyGenerator
	marshalizer                 marshal.Marshalizer
	heartbeatMessages           map[string]*heartbeatMessageInfo
	mutHeartbeatMessages        sync.RWMutex
	pubKeysMap                  map[uint32][]string
	fullPeersSlice              [][]byte
	mutPubKeysMap               sync.RWMutex
	appStatusHandler            core.AppStatusHandler
	monitorDB                   storage.Storer
	genesisTime                 time.Time
}

// NewMonitor returns a new monitor instance
func NewMonitor(
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	maxDurationPeerUnresponsive time.Duration,
	pubKeysMap map[uint32][]string,
	monitorDB storage.Storer,
	genesisTime time.Time,
) (*Monitor, error) {

	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return nil, ErrNilSingleSigner
	}
	if keygen == nil || keygen.IsInterfaceNil() {
		return nil, ErrNilKeyGenerator
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if len(pubKeysMap) == 0 {
		return nil, ErrEmptyPublicKeysMap
	}
	if monitorDB == nil {
		return nil, errors.New("nil monitor db")
	}

	mon := &Monitor{
		singleSigner:                singleSigner,
		keygen:                      keygen,
		marshalizer:                 marshalizer,
		heartbeatMessages:           make(map[string]*heartbeatMessageInfo),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		appStatusHandler:            &statusHandler.NilStatusHandler{},
		monitorDB:                   monitorDB,
		genesisTime:                 genesisTime,
	}

	err := mon.fetchGenesisTimeFromDb()
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
			err := m.loadHbmiFromDbIfExists(pubkey)
			if err != nil { // if pubKey not found in DB, create a new instance
				mhbi, err := newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, true, m.genesisTime, nil)
				if err != nil {
					return err
				}

				mhbi.genesisTime = m.genesisTime
				m.heartbeatMessages[pubkey] = mhbi
			}
			pubKeysMapCopy[shardId] = append(pubKeysMapCopy[shardId], pubkey)
		}
	}

	m.pubKeysMap = pubKeysMapCopy
	return nil
}

func (m *Monitor) fetchGenesisTimeFromDb() error {
	if m.monitorDB.Has([]byte("genesisTime")) != nil {
		err := m.saveGenesisTimeToDb()
		if err != nil {
			return err
		}
	} else {
		genesisTimeFromDbBytes, err := m.monitorDB.Get([]byte("genesisTime"))
		if err != nil {
			return errors.New("monitor: can't get genesis time from db")
		}

		var genesisTimeFromDb time.Time
		err = m.marshalizer.Unmarshal(&genesisTimeFromDb, genesisTimeFromDbBytes)
		if err != nil {
			return errors.New("monitor: can't unmarshal genesis time")
		}

		if genesisTimeFromDb != m.genesisTime {
			log.Info(fmt.Sprintf("updated heartbeat's genesis time to %s", genesisTimeFromDb))
		}

		err = m.saveGenesisTimeToDb()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) saveGenesisTimeToDb() error {
	genesisTimeBytes, err := m.marshalizer.Marshal(m.genesisTime)
	if err != nil {
		return errors.New("monitor: can't marshal genesis time")
	}

	err = m.monitorDB.Put([]byte("genesisTime"), genesisTimeBytes)
	if err != nil {
		return errors.New("monitor: can't store genesis time")
	}

	return nil
}

func (m *Monitor) loadRestOfPubKeysFromStorage() error {
	allKeysBytes, err := m.monitorDB.Get([]byte(peersKeysDbEntry))
	if err != nil {
		return err
	}

	var peersSlice [][]byte
	err = m.marshalizer.Unmarshal(&peersSlice, allKeysBytes)
	if err != nil {
		return err
	}

	for _, peer := range peersSlice {
		if _, ok := m.heartbeatMessages[string(peer)]; !ok { // peer not in nodes map
			err = m.loadHbmiFromDbIfExists(string(peer))
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (m *Monitor) loadHbmiFromDbIfExists(pubKey string) error {
	pkbytes := []byte(pubKey)

	hbFromDB, err := m.monitorDB.Get(pkbytes)
	if err != nil {
		return err
	}

	receivedHbmi, err := m.getHbmiFromDbBytes(hbFromDB)
	if err != nil {
		return err
	}

	receivedHbmi.getTimeHandler = nil
	receivedHbmi.lastUptimeDowntime = time.Now()
	receivedHbmi.genesisTime = m.genesisTime
	if receivedHbmi.timeStamp == m.genesisTime && time.Now().Sub(m.genesisTime) > 0 {
		receivedHbmi.totalDownTime = Duration{time.Now().Sub(m.genesisTime)}
	}

	m.heartbeatMessages[pubKey] = receivedHbmi

	return nil
}

func (m *Monitor) getHbmiFromDbBytes(message []byte) (*heartbeatMessageInfo, error) {
	heartbeatDto := HeartbeatDTO{}
	err := m.marshalizer.Unmarshal(&heartbeatDto, message)
	if err != nil {
		return nil, err
	}

	hbmi := m.convertFromExportedStruct(heartbeatDto)

	return &hbmi, nil
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
	if message == nil || message.IsInterfaceNil() {
		return ErrNilMessage
	}
	if message.Data() == nil {
		return ErrNilDataToProcess
	}

	hbRecv := &Heartbeat{}

	err := m.marshalizer.Unmarshal(hbRecv, message.Data())
	if err != nil {
		return err
	}

	err = m.verifySignature(hbRecv)
	if err != nil {
		return err
	}

	//message is validated, process should be done async, method can return nil
	go m.sendHeartbeatMessage(hbRecv)

	return nil
}

func (m *Monitor) sendHeartbeatMessage(hb *Heartbeat) {
	m.mutHeartbeatMessages.Lock()
	defer m.mutHeartbeatMessages.Unlock()

	pubKeyStr := string(hb.Pubkey)
	pe := m.heartbeatMessages[pubKeyStr]
	if pe == nil {
		var err error
		pe, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, false, m.genesisTime, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}
		m.heartbeatMessages[pubKeyStr] = pe
	}

	computedShardID := m.computeShardID(pubKeyStr)
	pe.HeartbeatReceived(computedShardID, hb.ShardID, hb.VersionNumber, hb.NodeDisplayName)
	m.storeHeartbeat(hb.Pubkey)
	m.updateAllHeartbeatMessages()
}

func (m *Monitor) storeHeartbeat(pubKey []byte) {
	hbInfo := m.heartbeatMessages[string(pubKey)]
	marshalizedHeartBeat, err := m.marshalizer.Marshal(m.convertToExportedStruct(hbInfo))
	if err != nil {
		log.Warn(fmt.Sprintf("can't marshal heartbeat message for pubKey %s: %s", pubKey, err.Error()))
		return
	}

	errStore := m.monitorDB.Put(pubKey, marshalizedHeartBeat)
	if errStore != nil {
		log.Warn(fmt.Sprintf("can't save heartbeat to db: %s", errStore.Error()))
	}

	m.addPeerToFullPeersSlice(pubKey)
}

func (m *Monitor) addPeerToFullPeersSlice(pubKey []byte) {
	if !m.isPeerInFullPeersSlice(pubKey) {
		m.fullPeersSlice = append(m.fullPeersSlice, pubKey)
		marshalizedFullPeersSlice, errMarsh := m.marshalizer.Marshal(m.fullPeersSlice)
		if errMarsh != nil {
			log.Warn(fmt.Sprintf("can't marshal peers keys slice: %s", errMarsh.Error()))
		}

		errStoreKeys := m.monitorDB.Put([]byte(peersKeysDbEntry), marshalizedFullPeersSlice)
		if errStoreKeys != nil {
			log.Warn(fmt.Sprintf("can't store the keys slice: %s", errStoreKeys.Error()))
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
	}
}

func (m *Monitor) convertFromExportedStruct(hbDTO HeartbeatDTO) heartbeatMessageInfo {
	hbmi := heartbeatMessageInfo{
		maxDurationPeerUnresponsive: m.maxDurationPeerUnresponsive,
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
	}

	return hbmi
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

func (m *Monitor) verifySignature(hbRecv *Heartbeat) error {
	senderPubKey, err := m.keygen.PublicKeyFromByteArray(hbRecv.Pubkey)
	if err != nil {
		return err
	}

	copiedHeartbeat := *hbRecv
	copiedHeartbeat.Signature = nil
	buffCopiedHeartbeat, err := m.marshalizer.Marshal(copiedHeartbeat)
	if err != nil {
		return err
	}

	return m.singleSigner.Verify(senderPubKey, buffCopiedHeartbeat, hbRecv.Signature)
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
