package heartbeat

import (
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var log = logger.DefaultLogger()

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	peerMessenger               PeerMessenger
	singleSigner                crypto.SingleSigner
	maxDurationPeerUnresponsive time.Duration
	keygen                      crypto.KeyGenerator
	marshalizer                 marshal.Marshalizer
	heartbeatMessages           map[string]*heartbeatMessageInfo
	mutHeartbeatMessages        sync.RWMutex
	shardID                     uint32
}

// NewMonitor returns a new monitor instance
func NewMonitor(
	peerMessenger PeerMessenger,
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	maxDurationPeerUnresponsive time.Duration,
	pubKeyList []string,
	shardID uint32,
) (*Monitor, error) {

	if peerMessenger == nil {
		return nil, ErrNilMessenger
	}
	if singleSigner == nil {
		return nil, ErrNilSingleSigner
	}
	if keygen == nil {
		return nil, ErrNilKeyGenerator
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if len(pubKeyList) == 0 {
		return nil, ErrEmptyPublicKeyList
	}

	mon := &Monitor{
		peerMessenger:               peerMessenger,
		singleSigner:                singleSigner,
		keygen:                      keygen,
		marshalizer:                 marshalizer,
		heartbeatMessages:           make(map[string]*heartbeatMessageInfo),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		shardID:                     shardID,
	}

	var err error
	for _, pubkey := range pubKeyList {
		mon.heartbeatMessages[pubkey], err = newHeartbeatMessageInfo(maxDurationPeerUnresponsive)
		if err != nil {
			return nil, err
		}
	}

	return mon, nil
}

// ProcessReceivedMessage satisfies the p2p.MessageProcessor interface so it can be called
// by the p2p subsystem each time a new heartbeat message arrives
func (m *Monitor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
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

	senderPubkey, err := m.keygen.PublicKeyFromByteArray(hbRecv.Pubkey)
	if err != nil {
		return err
	}

	err = m.singleSigner.Verify(senderPubkey, hbRecv.Payload, hbRecv.Signature)
	if err != nil {
		return err
	}

	//message is validated, process should be done async, method can return nil
	go func(msg p2p.MessageP2P, hb *Heartbeat) {
		m.mutHeartbeatMessages.Lock()
		defer m.mutHeartbeatMessages.Unlock()

		addr := m.peerMessenger.PeerAddress(msg.Peer())
		if addr == "" {
			//address is not known for the peer that emitted the message
			addr = msg.Peer().Pretty()
		}

		pe := m.heartbeatMessages[string(hb.Pubkey)]
		if pe == nil {
			pe, err = newHeartbeatMessageInfo(m.maxDurationPeerUnresponsive)
			if err != nil {
				log.Error(err.Error())
				return
			}
			m.heartbeatMessages[string(hb.Pubkey)] = pe
		}

		pe.HeartbeatReceived(m.shardID)
	}(message, hbRecv)

	return nil
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
			ShardID:         v.shardID,
		}
		idx++

	}
	m.mutHeartbeatMessages.RUnlock()

	sort.Slice(status, func(i, j int) bool {
		return strings.Compare(status[i].HexPublicKey, status[j].HexPublicKey) < 0
	})

	return status
}
