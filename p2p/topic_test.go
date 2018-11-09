package p2p_test

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var mockMarshalizer = mock.MarshalizerMock{}

type testTopicStringCloner struct {
	Data string
}

// Clone will return a new instance of string. Dummy, just to implement Cloner interface as strings are immutable
func (sc *testTopicStringCloner) Clone() p2p.Cloner {
	return &testTopicStringCloner{}
}

// ID will return the same string as ID
func (sc *testTopicStringCloner) ID() string {
	return sc.Data
}

var objTopicStringCloner = testTopicStringCloner{}

func TestTopic_AddEventHandler_Nil_ShouldNotAddHandler(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.AddDataReceived(nil)

	assert.Equal(t, len(topic.EventBusData()), 0)
}

func TestTopic_AddEventHandler_WithARealFunc_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {

	})

	assert.Equal(t, len(topic.EventBusData()), 1)
}

func TestTopic_CreateObject_NilData_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	_, err := topic.CreateObject(nil)

	assert.NotNil(t, err)
}

func TestTopic_CreateObject_EmptyData_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	_, err := topic.CreateObject(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopic_CreateObject_MarshalizerFails_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	_, err := topic.CreateObject(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopic_NewObjReceived_NilObj_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	err := topic.NewObjReceived(nil, "")

	assert.NotNil(t, err)
}

func TestTopic_NewObjReceived_OKMsg_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	wg := sync.WaitGroup{}
	wg.Add(1)

	cnt := int32(0)
	//attach event handler
	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		assert.Equal(t, name, "test")

		switch data.(type) {
		case p2p.Cloner:
			atomic.AddInt32(&cnt, 1)
		default:
			assert.Fail(t, "The data should have been string!")
		}

		wg.Done()

	})

	payload, err := mockMarshalizer.Marshal(&testTopicStringCloner{Data: "aaaa"})
	assert.Nil(t, err)

	obj, err := topic.CreateObject(payload)
	assert.Nil(t, err)
	topic.NewObjReceived(obj, "")

	//start a go routine as watchdog for the wg.Wait()
	go func() {
		time.Sleep(time.Second * 2)
		wg.Done()
	}()

	//wait for the go routine to finish
	wg.Wait()

	assert.Equal(t, atomic.LoadInt32(&cnt), int32(1))
}

func TestTopic_Broadcast_NilData_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	err := topic.Broadcast(nil)

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_MarshalizerFails_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_NoOneToSend_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_SendOK_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.SendData = func(data []byte) error {
		if topic.Name != "test" {
			return errors.New("should have been test")
		}

		if data == nil {
			return errors.New("should have not been nil")
		}

		fmt.Printf("Message: %v\n", data)
		return nil
	}

	err := topic.Broadcast("a string")
	assert.Nil(t, err)
}

func TestTopic_SendRequest_NilHash_ShouldRetErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)
	err := topic.SendRequest(nil)

	assert.NotNil(t, err)
}

func TestTopic_SendRequest_EmptyHash_ShouldRetErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)
	err := topic.SendRequest(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopic_SendRequest_NoHandler_ShouldRetErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)
	err := topic.SendRequest(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopic_SendRequest_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.SetRequest(func(hash []byte) error {
		if bytes.Equal(hash, []byte("AAAA")) {
			return nil
		}

		return errors.New("should have not got here")
	})
	err := topic.SendRequest([]byte("AAAA"))

	assert.Nil(t, err)
}

func TestTopic_RegisterValidator_NoHandler_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	err := topic.RegisterValidator(nil)
	assert.NotNil(t, err)
}

func TestTopic_RegisterValidator_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.SetRegisterTopicValidator(func(v pubsub.Validator) error {
		return nil
	})

	err := topic.RegisterValidator(nil)
	assert.Nil(t, err)
}

func TestTopic_UnregisterValidator_NoHandler_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	err := topic.UnregisterValidator()
	assert.NotNil(t, err)
}

func TestTopic_UnregisterValidator_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objTopicStringCloner, &mockMarshalizer)

	topic.SetUnregisterTopicValidator(func() error {
		return nil
	})

	err := topic.UnregisterValidator()
	assert.Nil(t, err)
}

type benchmark struct {
	field1  []byte
	field2  []byte
	field3  []byte
	field4  []byte
	field5  []byte
	field6  []byte
	field7  uint64
	field8  uint64
	field9  uint64
	field10 int64
	field11 int64
	field12 string
	field13 string
	field14 string
}

func BenchmarkTopicNewObjectCreation_PlainInit(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		obj1 = benchmark{}
	}

	obj1.field1 = make([]byte, 0)
}

func BenchmarkTopicNewObjectCreation_ReflectionNew(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		reflect.New(reflect.TypeOf(obj1)).Interface()
	}
}
