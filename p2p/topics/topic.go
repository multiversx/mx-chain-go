package topics

import (
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
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
//  - an async func will call each and every event handler registered on eventBus
//  - there is a method to be called to broadcast new messages
type Topic struct {
	Name        string
	ObjTemplate interface{}
	marsh       marshal.Marshalizer
	hasher      hashing.Hasher

	objPump chan interface{}

	mutEventBus sync.RWMutex
	eventBus    []OnTopicReceived

	OnNeedToSendMessage func(topic string, buff []byte) error

	queue *p2p.MessageQueue
}

// NewTopic creates a new Topic struct
func NewTopic(name string, objTemplate interface{}, marsh marshal.Marshalizer, hasher hashing.Hasher, maxCapacity int) *Topic {
	topic := Topic{Name: name, ObjTemplate: objTemplate, marsh: marsh, hasher: hasher}
	topic.objPump = make(chan interface{}, 10000)
	topic.queue = p2p.NewMessageQueue(maxCapacity)

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
func (t *Topic) NewMessageReceived(buff []byte) error {
	//buffer sanity checks
	if len(buff) == 0 {
		return nil
	}

	if (len(buff) == 1) && (buff[0] == byte('\n')) {
		return nil
	}

	if buff[len(buff)-1] == byte('\n') {
		buff = buff[0 : len(buff)-1]
	}

	//unmarshaling the message
	message, err := p2p.CreateFromByteArray(t.marsh, buff)

	if err != nil {
		return nil
	}

	//check authentication
	err = message.VerifyAndSetSigned()
	if err != nil {
		return nil
	}

	//check hash for multiple receives
	hash := t.hasher.Compute(string(message.Payload))
	if t.queue.ContainsAndAdd(string(hash)) {
		return nil
	}

	// create new instance of the object
	newObj := reflect.New(reflect.TypeOf(t.ObjTemplate)).Interface()

	//unmarshal data from the message
	err = t.marsh.Unmarshal(newObj, message.Payload)

	if err != nil {
		return err
	}

	//add to the channel so it can be consumed async
	t.objPump <- newObj
	return nil
}

// Broadcast should be called whenever a hight order struct needs to send over the wire an object
// Optionally, the message can be authenticated with a provided privateKey
func (t *Topic) Broadcast(data interface{}, peerID peer.ID, skForSigning crypto2.PrivKey) error {
	if data == nil {
		return errors.New("can not process nil data")
	}

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

	buff, err := mes.ToByteArray()
	if err != nil {
		return err
	}

	return t.OnNeedToSendMessage(t.Name, buff)
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
