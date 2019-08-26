package heartbeat

import (
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var log = logger.DefaultLogger()

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	singleSigner                crypto.SingleSigner
	maxDurationPeerUnresponsive time.Duration
	keygen                      crypto.KeyGenerator
	marshalizer                 marshal.Marshalizer
	heartbeatMessages           map[string]*heartbeatMessageInfo
	mutHeartbeatMessages        sync.RWMutex
	pubKeysMap                  map[uint32][]string
	mutPubKeysMap               sync.RWMutex
	appStatusHandler            core.AppStatusHandler
}

// NewMonitor returns a new monitor instance
func NewMonitor(
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	maxDurationPeerUnresponsive time.Duration,
	pubKeysMap map[uint32][]string,
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

	pubKeysMapCopy := make(map[uint32][]string, 0)

	mon := &Monitor{
		singleSigner:                singleSigner,
		keygen:                      keygen,
		marshalizer:                 marshalizer,
		heartbeatMessages:           make(map[string]*heartbeatMessageInfo),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		appStatusHandler:            &statusHandler.NilStatusHandler{},
	}

	for shardId, pubKeys := range pubKeysMap {
		for _, pubkey := range pubKeys {
			pubKeysMapCopy[shardId] = append(pubKeysMapCopy[shardId], pubkey)
			mhbi, err := newHeartbeatMessageInfo(maxDurationPeerUnresponsive, true)
			if err != nil {
				return nil, err
			}

			mhbi.computedShardID = shardId
			mon.heartbeatMessages[pubkey] = mhbi
		}
	}
	mon.pubKeysMap = pubKeysMapCopy
	return mon, nil
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
	go func(msg p2p.MessageP2P, hb *Heartbeat) {
		m.mutHeartbeatMessages.Lock()
		defer m.mutHeartbeatMessages.Unlock()

		pe := m.heartbeatMessages[string(hb.Pubkey)]
		if pe == nil {
			pe, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive, false)
			if err != nil {
				log.Error(err.Error())
				return
			}
			m.heartbeatMessages[string(hb.Pubkey)] = pe
		}

		computedShardID := m.computeShardID(string(hb.Pubkey))
		pe.HeartbeatReceived(computedShardID, hb.ShardID, hb.VersionNumber, hb.NodeDisplayName)
		m.updateAllHeartbeatMessages()
	}(message, hbRecv)

	return nil
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
		v.updateFields()

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
			TotalUpTime:     v.totalUpTime,
			TotalDownTime:   v.totalDownTime,
			VersionNumber:   v.versionNumber,
			IsValidator:     v.isValidator,
			NodeDisplayName: v.nodeDisplayName,
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
