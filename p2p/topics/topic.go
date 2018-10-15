package topics

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

// OnTopicReceived is the signature for the event handler used by Topic struct
type OnTopicReceived func(name string, data interface{})

// Topic struct defines a type of message that can be received and broadcast
// The use of empty interface gives this struct's a generic use
// It works in the following manner:
//  - the message is received (if it passes the authentication filter)
//  - a new object with the same type of ObjTemplate is created
//  - this new object will be used to unmarshal received data
//  - an async func will call each and every event registered on eventBus
//  - there is a method to be called to broadcast new messages
type Topic struct {
	Name        string
	ObjTemplate interface{}
	marsh       marshal.Marshalizer

	objPump chan interface{}

	mutEventBus sync.RWMutex
	eventBus    []OnTopicReceived

	OnNeedToSendMessage func(topic string, mes *p2p.Message) error
}

// NewTopic creates a new Topic struct
func NewTopic(name string, objTemplate interface{}, marsh marshal.Marshalizer) *Topic {
	topic := Topic{Name: name, ObjTemplate: objTemplate, marsh: marsh}
	topic.objPump = make(chan interface{}, 10000)

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
func (t *Topic) NewMessageReceived(message p2p.Message) {
	// create new instance of the object
	newObj := reflect.New(reflect.TypeOf(t.ObjTemplate)).Interface()

	//unmarshal data from the message
	err := t.marsh.Unmarshal(newObj, message.Payload)

	if err != nil {
		t.log(err)
		return
	}

	//add to the channel so it can be consumed async
	t.objPump <- newObj
}

// Broadcast should be called whenever a hight order struct needs to send over the wire an object
// Optionally, the message can be authenticated with a provided privateKey
func (t *Topic) Broadcast(data interface{}, peerID peer.ID, skForSigning crypto2.PrivKey) error {
	//assemle the message
	mes := p2p.Message{}
	mes.Type = t.Name
	mes.AddHop(peerID.Pretty())
	payload, err := t.marsh.Marshal(data)
	if err != nil {
		return err
	}
	mes.Payload = payload

	//if sk was provided, sign the message
	if skForSigning != nil {
		err = mes.SignWithPrivateKey(skForSigning)
		if err != nil {
			return err
		}
	}

	if t.OnNeedToSendMessage == nil {
		return errors.New("send to nil the assembled message?")
	}

	return t.OnNeedToSendMessage(t.Name, &mes)
}

func (t *Topic) processData() {
	for {
		select {
		case obj := <-t.objPump:
			//a new object is in pump, it has been consumed,
			//call each event handler from the list
			t.mutEventBus.RLock()
			for i := 0; i < len(t.eventBus); i++ {
				t.eventBus[i](t.Name, obj)
			}
			t.mutEventBus.RUnlock()
		}
	}
}

func (t *Topic) log(err error) {
	fmt.Printf("Error processing received message: %v\n", err.Error())
}
