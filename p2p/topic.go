package p2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/pkg/errors"
)

// Cloner interface will be implemented on structs that can create new instances of their type
// We prefer this method as reflection is more costly
type Cloner interface {
	Clone() Cloner
}

// OnTopicReceived is the signature for the event handler used by Topic struct
type OnTopicReceived func(name string, data interface{}, msgInfo *MessageInfo)

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

	mutEventBus sync.RWMutex
	eventBus    []OnTopicReceived

	OnNeedToSendMessage func(mes *Message, flagSign bool) error
}

// MessageInfo will retain additional info about the message, should we care
// when receiving an object on current topic
type MessageInfo struct {
	Object interface{}
	Signed bool
	Hops   int
	Peers  []string
}

// NewTopic creates a new Topic struct
func NewTopic(name string, objTemplate Cloner, marsh marshal.Marshalizer) *Topic {
	topic := Topic{Name: name, ObjTemplate: objTemplate, marsh: marsh}
	topic.objPump = make(chan MessageInfo, 10000)

	go topic.processData()

	return &topic
}

// AddEventHandler registers a new event on the eventBus so it can be called async whenever a new object is unmarshaled
func (t *Topic) AddEventHandler(event OnTopicReceived) {
	if event == nil {
		//won't add a nil event handler to list
		return
	}

	t.mutEventBus.Lock()
	defer t.mutEventBus.Unlock()

	t.eventBus = append(t.eventBus, event)
}

// NewMessageReceived is called from the lower data layer
// it will ignore nils, improper formatted messages or tampered messages
func (t *Topic) NewMessageReceived(message *Message) error {
	// create new instance of the object
	newObj := t.ObjTemplate.Clone()

	if message == nil {
		return errors.New("nil message not allowed")
	}

	//unmarshal data from the message
	err := t.marsh.Unmarshal(newObj, message.Payload)
	if err != nil {
		return err
	}

	//add to the channel so it can be consumed async
	t.objPump <- MessageInfo{Object: newObj, Peers: message.Peers, Signed: message.isSigned, Hops: message.Hops}
	return nil
}

// Broadcast should be called whenever a higher order struct needs to send over the wire an object
// Optionally, the message can be authenticated
func (t *Topic) Broadcast(data interface{}, flagSign bool) error {
	if data == nil {
		return errors.New("can not process nil data")
	}

	//assemble the message
	mes := NewMessage("", nil, t.marsh)
	mes.Type = t.Name
	payload, err := t.marsh.Marshal(data)
	if err != nil {
		return err
	}
	mes.Payload = payload

	if t.OnNeedToSendMessage == nil {
		return errors.New("send to nil the assembled message?")
	}

	return t.OnNeedToSendMessage(mes, flagSign)
}

func (t *Topic) processData() {
	for {
		select {
		case obj := <-t.objPump:
			//a new object is in pump, it has been consumed,
			//call each event handler from the list
			t.mutEventBus.RLock()
			for i := 0; i < len(t.eventBus); i++ {
				t.eventBus[i](t.Name, obj.Object, &obj)
			}
			t.mutEventBus.RUnlock()
		}
	}
}
