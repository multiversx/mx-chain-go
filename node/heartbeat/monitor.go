package heartbeat

import (
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// Monitor represents the heartbeat component that processes received heartbeat messages
type Monitor struct {
	p2pMessenger                P2PMessenger
	singleSigner                crypto.SingleSigner
	maxDurationPeerUnresponsive time.Duration
	keygen                      crypto.KeyGenerator
	marshalizer                 marshal.Marshalizer
	m                           map[string]*PubkeyElement
	mut                         sync.Mutex
}

// NewMonitor returns a new monitor instance
func NewMonitor(
	p2pMessenger P2PMessenger,
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	maxDurationPeerUnresponsive time.Duration,
	genesisPubKeyList []string,
) (*Monitor, error) {

	if p2pMessenger == nil {
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
	if len(genesisPubKeyList) == 0 {
		return nil, ErrEmptyGenesisList
	}

	mon := &Monitor{
		p2pMessenger:                p2pMessenger,
		singleSigner:                singleSigner,
		keygen:                      keygen,
		marshalizer:                 marshalizer,
		m:                           make(map[string]*PubkeyElement),
		maxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
	}

	for _, pubkey := range genesisPubKeyList {
		mon.m[pubkey] = NewPubkeyElement(maxDurationPeerUnresponsive)
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
		m.mut.Lock()
		defer m.mut.Unlock()

		addr := m.p2pMessenger.PeerAddress(msg.Peer())
		if addr == "" {
			//address is not known for the peer that emitted the message
			addr = msg.Peer().Pretty()
		}

		pe := m.m[string(hb.Pubkey)]
		if pe == nil {
			pe = NewPubkeyElement(m.maxDurationPeerUnresponsive)
			m.m[string(hb.Pubkey)] = pe
		}

		pe.HeartbeatArrived(addr)
	}(message, hbRecv)

	return nil
}

// GetHeartbeats returns the heartbeat status
func (m *Monitor) GetHeartbeats() []PubkeyHeartbeat {
	m.mut.Lock()
	defer m.mut.Unlock()

	status := make([]PubkeyHeartbeat, len(m.m))

	idx := 0
	for k, v := range m.m {
		status[idx] = PubkeyHeartbeat{
			HexPublicKey:   hex.EncodeToString([]byte(k)),
			PeerHeartBeats: v.GetPeerHeartbeats(),
		}
		idx++
	}

	sort.Slice(status, func(i, j int) bool {
		return strings.Compare(status[i].HexPublicKey, status[j].HexPublicKey) < 0
	})

	return status
}
