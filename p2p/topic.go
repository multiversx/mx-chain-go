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

// Cloner interface will be implemented on structs that can create new instances of their type
// We prefer this method as reflection is more costly
type Cloner interface {
	Clone() Cloner
	ID() string
}

// DataReceived is the signature for the event handler used by Topic struct
type DataReceived func(name string, data interface{}, msgInfo *MessageInfo)

// Topic struct defines a type of message that can be received and broadcast
// The use of empty interface gives this struct's a generic use
// It works in the following manner:
//  - the message is received (if it passes the authentication filter)
//  - a new object with the same type of ObjTemplate is created
//  - this new object will be used to unmarshal received data
//  - an async func will call each and every event handler registered on eventBus
//  - there is a method to be called to broadcast new messages
type Topic struct {
	Name        string
	ObjTemplate Cloner
	marsh       marshal.Marshalizer

	objPump chan MessageInfo

	mutEventBus  sync.RWMutex
	eventBusData []DataReceived

	SendData func(data []byte) error

	registerTopicValidator   func(v pubsub.Validator) error
	unregisterTopicValidator func() error

	ResolveRequest func(hash []byte) Cloner
	request        func(hash []byte) error

	CurrentPeer peer.ID
}

// MessageInfo will retain additional info about the message, should we care
// when receiving an object on current topic
type MessageInfo struct {
	Object      interface{}
	Peer        string
	CurrentPeer string
}

// NewTopic creates a new Topic struct
func NewTopic(name string, objTemplate Cloner, marsh marshal.Marshalizer) *Topic {
	topic := Topic{Name: name, ObjTemplate: objTemplate, marsh: marsh}
	topic.objPump = make(chan MessageInfo, topicChannelBufferSize)

	go topic.processData()

	return &topic
}

// AddDataReceived registers a new event on the eventBus so it can be called async whenever a new object is unmarshaled
func (t *Topic) AddDataReceived(event DataReceived) {
	if event == nil {
		//won't add a nil event handler to list
		return
	}

	t.mutEventBus.Lock()
	defer t.mutEventBus.Unlock()

	t.eventBusData = append(t.eventBusData, event)
}

// CreateObject will instantiate a Cloner interface and instantiate its fields
// with the help of a marshalizer implementation
func (t *Topic) CreateObject(data []byte) (Cloner, error) {
	// create new instance of the object
	newObj := t.ObjTemplate.Clone()

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
func (t *Topic) NewObjReceived(obj Cloner, peerID string) error {
	if obj == nil {
		return errors.New("nil object not allowed")
	}

	//add to the channel so it can be consumed async
	t.objPump <- MessageInfo{Object: obj, Peer: peerID, CurrentPeer: t.CurrentPeer.Pretty()}
	return nil
}

// Broadcast should be called whenever a higher order struct needs to send over the wire an object
// Optionally, the message can be authenticated
func (t *Topic) Broadcast(data interface{}) error {
	if data == nil {
		return errors.New("can not process nil data")
	}

	//assemble the message
	payload, err := t.marsh.Marshal(data)
	if err != nil {
		return err
	}

	if t.SendData == nil {
		return errors.New("send to nil the assembled message?")
	}

	return t.SendData(payload)
}

func (t *Topic) processData() {
	for {
		select {
		case obj := <-t.objPump:
			//a new object is in pump, it has been consumed,
			//call each event handler from the list
			t.mutEventBus.RLock()
			for i := 0; i < len(t.eventBusData); i++ {
				t.eventBusData[i](t.Name, obj.Object, &obj)
			}
			t.mutEventBus.RUnlock()
		}
	}
}

// RegisterValidator adds a validator to this topic
// It delegates the functionality to registerValidator function pointer
func (t *Topic) RegisterValidator(v pubsub.Validator) error {
	if t.registerTopicValidator == nil {
		return errors.New("can not delegate registration to parent")
	}

	return t.registerTopicValidator(v)
}

// UnregisterValidator removes the validator associated to this topic
// It delegates the functionality to unregisterValidator function pointer
func (t *Topic) UnregisterValidator() error {
	if t.unregisterTopicValidator == nil {
		return errors.New("can not delegate unregistration to parent")
	}

	return t.unregisterTopicValidator()
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

	if t.request == nil {
		return errors.New("can not delegate request to parent")
	}

	return t.request(hash)
}
