package p2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

// topicChannelBufferSize is used to control the object channel buffer size
const topicChannelBufferSize = 10000

// DataReceivedHandler is the signature for the event handler used by Topic struct
type DataReceivedHandler func(name string, data interface{}, msgInfo *MessageInfo)

// Topic struct defines a type of message that can be received and broadcast
// The use of Creator interface gives this struct's a generic use
// It works in the following manner:
//  - the message is received (if it passes the authentication filter)
//  - a new object with the same type of ObjTemplate is created
//  - this new object will be used to unmarshal received data
//  - an async func will call each and every event handler registered on eventBus
//  - the method Broadcast is used to send messages containing object's serialized data to other peers
type Topic struct {
	// Name of the topic
	name string
	// ObjTemplate is used as a template to generate new objects whenever a new message is received
	ObjTemplate             Creator
	marsh                   marshal.Marshalizer
	objChan                 chan MessageInfo
	mutEventBus             sync.RWMutex
	eventBusDataRcvHandlers []DataReceivedHandler
	// SendData will be called by Topic struct whenever a user of this struct tries to send data to other peers
	// It is a function pointer that connects Topic struct with pubsub implementation
	SendData                 func(data []byte) error
	RegisterTopicValidator   func(v pubsub.Validator) error
	UnregisterTopicValidator func() error
	ResolveRequest           func(hash []byte) []byte
	Request                  func(hash []byte) error
	CurrentPeer              peer.ID
}

// MessageInfo will retain additional info about the message, should we care
// when receiving an object on current topic
type MessageInfo struct {
	Data        Creator
	Peer        string
	CurrentPeer string
}

// NewTopic creates a new Topic struct
func NewTopic(name string, objTemplate Creator, marsh marshal.Marshalizer) *Topic {
	topic := Topic{name: name, ObjTemplate: objTemplate, marsh: marsh}
	topic.objChan = make(chan MessageInfo, topicChannelBufferSize)

	go topic.processData()

	return &topic
}

// AddDataReceived registers a new event handler on the eventBus so it can be called async whenever a new object is unmarshaled
func (t *Topic) AddDataReceived(eventHandler DataReceivedHandler) {
	if eventHandler == nil {
		//won't add a nil event handler to list
		return
	}

	t.mutEventBus.Lock()
	defer t.mutEventBus.Unlock()

	t.eventBusDataRcvHandlers = append(t.eventBusDataRcvHandlers, eventHandler)
}

// CreateObject will instantiate a Cloner interface and instantiate its fields
// with the help of a marshalizer implementation
func (t *Topic) CreateObject(data []byte) (Creator, error) {
	// create new instance of the object
	newObj := t.ObjTemplate.Create()

	if data == nil {
		return nil, errors.New("nil message not allowed")
	}

	if len(data) == 0 {
		return nil, errors.New("empty message not allowed")
	}

	//unmarshal data from the message
	err := t.marsh.Unmarshal(newObj, data)
	if err != nil {
		return nil, err
	}

	return newObj, err
}

// NewObjReceived is called from the lower data layer
// it will ignore nils or improper formatted messages
func (t *Topic) NewObjReceived(obj Creator, peerID string) error {
	if obj == nil {
		return errors.New("nil object not allowed")
	}

	//add to the channel so it can be consumed async
	t.objChan <- MessageInfo{Data: obj, Peer: peerID, CurrentPeer: t.CurrentPeer.Pretty()}
	return nil
}

// Broadcast should be called whenever a higher order struct needs to send over the wire an object
// Optionally, the message can be authenticated
func (t *Topic) Broadcast(data interface{}) error {
	if data == nil {
		return errors.New("can not process nil data")
	}

	if t.SendData == nil {
		return errors.New("nil SendData handler")
	}

	//assemble the message
	payload, err := t.marsh.Marshal(data)
	if err != nil {
		return err
	}

	return t.SendData(payload)
}

// BroadcastBuff should be called whenever a higher order struct needs to send over the wire already
// serialized data
// Optionally, the message can be authenticated
func (t *Topic) BroadcastBuff(payload []byte) error {
	if payload == nil {
		return errors.New("can not process nil data")
	}

	if t.SendData == nil {
		return errors.New("send to nil the assembled message?")
	}

	return t.SendData(payload)
}

func (t *Topic) processData() {
	for {
		select {
		case obj := <-t.objChan:
			//a new object is in pump, it has been consumed,
			//call each event handler from the list
			t.mutEventBus.RLock()
			for i := 0; i < len(t.eventBusDataRcvHandlers); i++ {
				t.eventBusDataRcvHandlers[i](t.name, obj.Data, &obj)
			}
			t.mutEventBus.RUnlock()
		}
	}
}

// RegisterValidator adds a validator to this topic
// It delegates the functionality to registerValidator function pointer
func (t *Topic) RegisterValidator(v pubsub.Validator) error {
	if t.RegisterTopicValidator == nil {
		return errors.New("can not delegate registration to parent")
	}

	return t.RegisterTopicValidator(v)
}

// UnregisterValidator removes the validator associated to this topic
// It delegates the functionality to unregisterValidator function pointer
func (t *Topic) UnregisterValidator() error {
	if t.UnregisterTopicValidator == nil {
		return errors.New("can not delegate unregistration to parent")
	}

	return t.UnregisterTopicValidator()
}

// SendRequest sends the hash to all known peers that subscribed to the channel [t.Name]_REQUEST
// It delegates the functionality to sendRequest function pointer
// The object, if exists, should return on the main event bus (regular topic channel)
func (t *Topic) SendRequest(hash []byte) error {
	if hash == nil {
		return errors.New("invalid hash to send")
	}

	if len(hash) == 0 {
		return errors.New("invalid hash to send")
	}

	if t.Request == nil {
		return errors.New("can not delegate request to parent")
	}

	return t.Request(hash)
}

// Name returns the topic name
func (t *Topic) Name() string {
	return t.name
}
