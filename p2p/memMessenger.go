package p2p

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/btcsuite/btcutil/base58"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/multiformats/go-multiaddr"
)

const signPrefix = "libp2p-pubsub:"

var mutGloballyRegPeers *sync.Mutex

// globallyRegisteredPeers is the main map used for in memory communication
var globallyRegisteredPeers map[peer.ID]*MemMessenger

func init() {
	mutGloballyRegPeers = &sync.Mutex{}
	globallyRegisteredPeers = make(map[peer.ID]*MemMessenger)
}

// MemMessenger is a fake memory Messenger used for testing
// TODO keep up with NetMessenger modifications
type MemMessenger struct {
	peerID            peer.ID
	privKey           crypto.PrivKey
	mutConnectedPeers sync.Mutex
	connectedPeers    map[peer.ID]*MemMessenger
	marsh             marshal.Marshalizer
	hasher            hashing.Hasher
	rt                *RoutingTable
	seqNo             uint64
	mutClosed         sync.RWMutex
	closed            bool
	chSend            chan *pubsub.Message
	mutSeenMessages   sync.Mutex
	seenMessages      *TimeCache
	mutValidators     sync.RWMutex
	validators        map[string]pubsub.Validator
	mutTopics         sync.RWMutex
	topics            map[string]*Topic
	mutGossipCache    sync.Mutex
	gossipCache       *TimeCache
}

// NewMemMessenger creates a memory Messenger with the same behaviour as the NetMessenger.
// Should be used in tests!
func NewMemMessenger(marsh marshal.Marshalizer, hasher hashing.Hasher,
	cp *ConnectParams) (*MemMessenger, error) {

	if marsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create messenger")
	}

	if hasher == nil {
		return nil, errors.New("hasher is nil! Can't create messenger")
	}

	mm := MemMessenger{
		seqNo:             0,
		marsh:             marsh,
		closed:            false,
		hasher:            hasher,
		chSend:            make(chan *pubsub.Message),
		privKey:           cp.PrivKey,
		peerID:            cp.ID,
		mutConnectedPeers: sync.Mutex{},
		mutClosed:         sync.RWMutex{},
		mutSeenMessages:   sync.Mutex{},
		seenMessages:      NewTimeCache(time.Second * 120),
		mutValidators:     sync.RWMutex{},
		validators:        make(map[string]pubsub.Validator),
		mutTopics:         sync.RWMutex{},
		topics:            make(map[string]*Topic),
		mutGossipCache:    sync.Mutex{},
		gossipCache:       NewTimeCache(durTimeCache),
	}

	mm.mutConnectedPeers.Lock()
	mm.connectedPeers = make(map[peer.ID]*MemMessenger)
	mm.connectedPeers[mm.peerID] = &mm
	mm.mutConnectedPeers.Unlock()

	mm.rt = NewRoutingTable(mm.peerID)

	mutGloballyRegPeers.Lock()
	globallyRegisteredPeers[mm.peerID] = &mm
	mutGloballyRegPeers.Unlock()

	go mm.processLoop()

	return &mm, nil
}

// Closes a MemMessenger. Receiving and sending data is no longer possible
func (mm *MemMessenger) Close() error {
	mm.mutClosed.Lock()
	defer mm.mutClosed.Unlock()

	mm.closed = true

	return nil
}

// ID returns the current id
func (mm *MemMessenger) ID() peer.ID {
	return mm.peerID
}

// Peers returns the connected peers list
func (mm *MemMessenger) Peers() []peer.ID {
	peers := make([]peer.ID, 0)

	mm.mutConnectedPeers.Lock()
	for p := range mm.connectedPeers {
		peers = append(peers, p)
	}
	mm.mutConnectedPeers.Unlock()

	return peers
}

// Conns return the connections made by this memory messenger
func (mm *MemMessenger) Conns() []net.Conn {
	conns := make([]net.Conn, 0)

	mm.mutConnectedPeers.Lock()
	for p := range mm.connectedPeers {
		c := &mock.ConnMock{LocalP: mm.peerID, RemoteP: p}
		conns = append(conns, c)
	}
	mm.mutConnectedPeers.Unlock()

	return conns
}

// Marshalizer returns the used marshalizer object
func (mm *MemMessenger) Marshalizer() marshal.Marshalizer {
	return mm.marsh
}

// Hasher returns the used marshalizer object
func (mm *MemMessenger) Hasher() hashing.Hasher {
	return mm.hasher
}

// RouteTable will return the RoutingTable object
func (mm *MemMessenger) RouteTable() *RoutingTable {
	return mm.rt
}

// Addresses will return all addresses bound to current messenger
func (mm *MemMessenger) Addresses() []string {
	return []string{string(mm.peerID.Pretty())}
}

// ConnectToInitialAddresses is used to explicitly connect to a well known set of addresses
func (mm *MemMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
	for i := 0; i < len(addresses); i++ {
		addr := peer.ID(base58.Decode(addresses[i]))

		mutGloballyRegPeers.Lock()
		val, ok := globallyRegisteredPeers[addr]
		mutGloballyRegPeers.Unlock()

		if !ok {
			log.Error(fmt.Sprintf("Bootstrapping the peer '%v' failed! [not found]\n", addresses[i]))
			continue
		}

		if mm.peerID == addr {
			//won't add self
			continue
		}

		mm.mutConnectedPeers.Lock()
		//connect this to other peer
		mm.connectedPeers[addr] = val
		//connect other the other peer to this
		val.connectedPeers[mm.peerID] = mm
		mm.mutConnectedPeers.Unlock()
	}
}

// Bootstrap will try to connect to as many peers as possible
func (mm *MemMessenger) Bootstrap(ctx context.Context) {
	go mm.doBootstrap()
}

func (mm *MemMessenger) doBootstrap() {
	for {
		mm.mutClosed.RLock()
		if mm.closed {
			mm.mutClosed.RUnlock()
			return
		}
		mm.mutClosed.RUnlock()

		temp := make(map[peer.ID]*MemMessenger, 0)

		mutGloballyRegPeers.Lock()
		for k, v := range globallyRegisteredPeers {
			if !mm.rt.Has(k) {
				mm.rt.Update(k)

				temp[k] = v
			}
		}
		mutGloballyRegPeers.Unlock()

		mm.mutConnectedPeers.Lock()
		for k, v := range temp {
			mm.connectedPeers[k] = v
		}
		mm.mutConnectedPeers.Unlock()

		time.Sleep(time.Second)
	}

}

// PrintConnected displays the connected peers
func (mm *MemMessenger) PrintConnected() {
	conns := mm.Conns()

	connectedTo := fmt.Sprintf("Node %s is connected to: \n", mm.ID().Pretty())
	for i := 0; i < len(conns); i++ {
		connectedTo = connectedTo + fmt.Sprintf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(mm.ID(), conns[i].RemotePeer()))
	}

	log.Debug(connectedTo)
}

// AddAddress adds a new address to peer store
func (mm *MemMessenger) AddAddress(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	mutGloballyRegPeers.Lock()
	val, ok := globallyRegisteredPeers[p]
	mutGloballyRegPeers.Unlock()

	if !ok {
		val = nil
	}

	mm.mutConnectedPeers.Lock()
	mm.connectedPeers[p] = val
	mm.mutConnectedPeers.Unlock()
}

// Connectedness tests for a connection between self and another peer
func (mm *MemMessenger) Connectedness(pid peer.ID) net.Connectedness {
	mm.mutConnectedPeers.Lock()
	_, ok := mm.connectedPeers[pid]
	mm.mutConnectedPeers.Unlock()

	if ok {
		return net.Connected
	} else {
		return net.NotConnected
	}
}

// GetTopic returns the topic from its name or nil if no topic with that name
// was ever registered
func (mm *MemMessenger) GetTopic(topicName string) *Topic {
	mm.mutTopics.RLock()
	defer mm.mutTopics.RUnlock()

	if t, ok := mm.topics[topicName]; ok {
		return t
	}

	return nil
}

// AddTopic registers a new topic to this messenger
func (mm *MemMessenger) AddTopic(t *Topic) error {
	//sanity checks
	if t == nil {
		return errors.New("topic can not be nil")
	}

	if strings.Contains(t.Name, requestTopicSuffix) {
		return errors.New("topic name contains request suffix")
	}

	mm.mutTopics.Lock()

	if _, ok := mm.topics[t.Name]; ok {
		mm.mutTopics.Unlock()
		return errors.New("topic already exists")
	}

	mm.topics[t.Name] = t
	t.CurrentPeer = mm.ID()
	mm.mutTopics.Unlock()

	// func that publishes on network from Topic object
	t.SendData = func(data []byte) error {
		return mm.publish(t.Name, data)
	}

	// validator registration func
	t.registerTopicValidator = func(v pubsub.Validator) error {
		return mm.registerValidator(t.Name, v)
	}

	// validator unregistration func
	t.unregisterTopicValidator = func() error {
		return mm.unregisterValidator(t.Name)
	}

	//wire-up a plain func for publishing on request channel
	t.request = func(hash []byte) error {
		return mm.publish(t.Name+requestTopicSuffix, hash)
	}

	return nil
}

func (mm *MemMessenger) nextSeqno() []byte {
	seqno := make([]byte, 8)
	counter := atomic.AddUint64(&mm.seqNo, 1)
	binary.BigEndian.PutUint64(seqno, counter)
	return seqno
}

func (mm *MemMessenger) gotNewMessage(mes *pubsub.Message) {
	err := verifyMessageSignature(mes.Message)

	if err != nil {
		log.Error(err.Error())
		return
	}

	mm.mutSeenMessages.Lock()
	if mm.seenMessages.Has(msgID(mes)) {
		mm.mutSeenMessages.Unlock()
		return
	}
	mm.seenMessages.Add(msgID(mes))
	mm.mutSeenMessages.Unlock()

	if len(mes.TopicIDs) == 0 {
		log.Error("no topic")
		return
	}

	v := pubsub.Validator(nil)
	mm.mutValidators.RLock()
	v, _ = mm.validators[mes.TopicIDs[0]]
	mm.mutValidators.RUnlock()

	if v != nil {
		if !v(context.Background(), mes) {
			//message is not valid, dropping
			return
		}
	}

	if !mm.gotNewData(mes.GetData(), mes.GetFrom(), mes.TopicIDs[0]) {
		return
	}

	//broadcast to peers
	mm.chSend <- mes
}

func (mm *MemMessenger) gotNewData(data []byte, peerID peer.ID, topic string) bool {
	splt := strings.Split(topic, requestTopicSuffix)

	regularTopic := splt[0]

	t := mm.GetTopic(regularTopic)

	if t == nil {
		//not subscribed to regular topic, drop
		return false
	}

	if len(splt) == 2 {
		//request message

		//resolver has not been set up, let the message go to the other peers, maybe they can resolve the request
		if t.ResolveRequest == nil {
			return true
		}

		//payload == hash
		obj := t.ResolveRequest(data)

		if obj == nil {
			//object not found
			return true
		}

		//found object, no need to resend the request message to peers
		//test whether we also should broadcast the message (others might have broadcast it just before us)
		has := false

		mm.mutGossipCache.Lock()
		has = mm.gossipCache.Has(obj.ID())
		mm.mutGossipCache.Unlock()

		if !has {
			//only if the current peer did not receive an equal object to cloner,
			//then it shall broadcast it
			err := t.Broadcast(obj)
			if err != nil {
				log.Error(err.Error())
			}
		}

		return false
	}

	//regular message
	obj, err := t.CreateObject(data)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	mm.mutGossipCache.Lock()
	if mm.gossipCache.Has(obj.ID()) {
		//duplicate object, skip
		mm.mutGossipCache.Unlock()
		return false
	}

	mm.gossipCache.Add(obj.ID())
	mm.mutGossipCache.Unlock()

	err = t.NewObjReceived(obj, peerID.Pretty())
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (mm *MemMessenger) publish(topic string, data []byte) error {
	mm.mutClosed.Lock()
	if mm.closed {
		mm.mutClosed.Unlock()
		return errors.New("messenger is closed")
	}
	mm.mutClosed.Unlock()

	seqno := mm.nextSeqno()
	mes := pubsub_pb.Message{}
	mes.Data = data
	mes.TopicIDs = []string{topic}
	mes.From = []byte(mm.ID())
	mes.Seqno = seqno

	pbsMessage := pubsub.Message{Message: &mes}

	if mm.privKey != nil {
		mes.From = []byte(mm.peerID)
		err := signMessage(&mes, mm.privKey, mm.peerID)
		if err != nil {
			return err
		}
	}

	v := pubsub.Validator(nil)
	mm.mutValidators.RLock()
	v, _ = mm.validators[mes.TopicIDs[0]]
	mm.mutValidators.RUnlock()

	if v != nil {
		if !v(context.Background(), &pbsMessage) {
			//message is not valid, dropping
			return errors.New("invalid message")
		}
	}

	mm.gotNewMessage(&pbsMessage)

	return nil
}

func (mm *MemMessenger) processLoop() {
	for {
		select {
		case mes := <-mm.chSend:

			//send to connected peers
			peers := mm.Peers()

			for i := 0; i < len(peers); i++ {
				peerID := peer.ID(peers[i]).Pretty()

				//do not send to originator
				if mes.GetFrom() == peers[i] {
					continue
				}

				if peerID == mm.ID().Pretty() {
					//broadcast to self allowed
					mm.gotNewMessage(mes)
					continue
				}

				mm.mutConnectedPeers.Lock()
				val, ok := mm.connectedPeers[peer.ID(base58.Decode(peerID))]
				mm.mutConnectedPeers.Unlock()

				if !ok || val == nil {
					log.Error("invalid peer")
					continue
				}

				go val.gotNewMessage(mes)
			}
		}
	}
}

func (mm *MemMessenger) registerValidator(topic string, v pubsub.Validator) error {
	mm.mutValidators.Lock()
	defer mm.mutValidators.Unlock()

	_, ok := mm.validators[topic]

	if !ok {
		mm.validators[topic] = v
		return nil
	}

	return errors.New(fmt.Sprintf("topic %v already has a validator set", topic))
}

func (mm *MemMessenger) unregisterValidator(topic string) error {
	mm.mutValidators.Lock()
	defer mm.mutValidators.Unlock()

	_, ok := mm.validators[topic]

	if !ok {
		return errors.New(fmt.Sprintf("topic %v does not have a validator set", topic))
	}

	delete(mm.validators, topic)
	return nil
}

//Helper funcs

// msgID returns a unique ID of the passed Message
func msgID(pmsg *pubsub.Message) string {
	return string(pmsg.GetFrom()) + string(pmsg.GetSeqno())
}

func signMessage(mes *pubsub_pb.Message, key crypto.PrivKey, pid peer.ID) error {
	buff, err := mes.Marshal()
	if err != nil {
		return err
	}

	buff = withSignPrefix(buff)

	sig, err := key.Sign(buff)
	if err != nil {
		return err
	}

	mes.Signature = sig

	pk, _ := pid.ExtractPublicKey()
	if pk == nil {
		pubk, err := key.GetPublic().Bytes()
		if err != nil {
			return err
		}
		mes.Key = pubk
	}

	return nil
}

func verifyMessageSignature(m *pubsub_pb.Message) error {
	pubk, err := messagePubKey(m)
	if err != nil {
		return err
	}

	xm := *m
	xm.Signature = nil
	xm.Key = nil
	buff, err := xm.Marshal()
	if err != nil {
		return err
	}

	buff = withSignPrefix(buff)

	valid, err := pubk.Verify(buff, m.Signature)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func withSignPrefix(bytes []byte) []byte {
	return append([]byte(signPrefix), bytes...)
}

func messagePubKey(m *pubsub_pb.Message) (crypto.PubKey, error) {
	var pubk crypto.PubKey

	pid, err := peer.IDFromBytes(m.From)
	if err != nil {
		return nil, err
	}

	if m.Key == nil {
		// no attached key, it must be extractable from the source ID
		pubk, err = pid.ExtractPublicKey()
		if err != nil {
			return nil, fmt.Errorf("cannot extract signing key: %s", err.Error())
		}
		if pubk == nil {
			return nil, fmt.Errorf("cannot extract signing key")
		}
	} else {
		pubk, err = crypto.UnmarshalPublicKey(m.Key)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal signing key: %s", err.Error())
		}

		// verify that the source ID matches the attached key
		if !pid.MatchesPublicKey(pubk) {
			return nil, fmt.Errorf("bad signing key; source ID %s doesn't match key", pid)
		}
	}

	return pubk, nil
}
