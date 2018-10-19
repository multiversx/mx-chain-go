package p2p_test

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var mockMarshalizer = mock.MarshalizerMock{}

type testStringCloner struct {
	Data string
}

// Clone will return a new instance of string. Dummy, just to implement Cloner interface as strings are immutable
func (sc *testStringCloner) Clone() p2p.Cloner {
	return &testStringCloner{}
}

var objStringCloner = testStringCloner{}

func TestTopic_AddEventHandler_Nil_ShouldNotAddHandler(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	topic.AddEventHandler(nil)

	assert.Equal(t, len(topic.EventBus()), 0)
}

func TestTopic_AddEventHandler_WithARealFunc_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	topic.AddEventHandler(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {

	})

	assert.Equal(t, len(topic.EventBus()), 1)
}

func TestTopic_NewMessageReceived_NilMsg_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	err := topic.NewMessageReceived(nil)

	assert.NotNil(t, err)
}

func TestTopic_NewMessageReceived_MarshalizerFails_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	err := topic.NewMessageReceived(&p2p.Message{})

	assert.NotNil(t, err)
}

func TestTopic_NewMessageReceived_OKMsg_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	wg := sync.WaitGroup{}
	wg.Add(1)

	cnt := int32(0)
	//attach event handler
	topic.AddEventHandler(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		assert.Equal(t, name, "test")

		switch data.(type) {
		case p2p.Cloner:
			atomic.AddInt32(&cnt, 1)
		default:
			assert.Fail(t, "The data should have been string!")
		}

		wg.Done()

	})

	//create a new Message
	buff, err := topic.Marsh().Marshal(testStringCloner{Data: "a string"})
	assert.Nil(t, err)

	mes := p2p.NewMessage("aaa", buff, &mockMarshalizer)
	topic.NewMessageReceived(mes)

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
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	err := topic.Broadcast(nil, false)

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_MarshalizerFails_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	err := topic.Broadcast("a string", false)

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_NoOneToSend_ShouldErr(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	err := topic.Broadcast("a string", false)

	assert.NotNil(t, err)
}

func TestTopic_Broadcast_SendOK_ShouldWork(t *testing.T) {
	topic := p2p.NewTopic("test", &objStringCloner, &mockMarshalizer)

	topic.OnNeedToSendMessage = func(mes *p2p.Message, flagSign bool) error {
		if topic.Name != "test" {
			return errors.New("should have been test")
		}

		if mes == nil {
			return errors.New("should have not been nil")
		}

		fmt.Printf("Message: %v\n", mes)
		return nil
	}

	err := topic.Broadcast("a string", false)
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
		//reflect.New(reflect.TypeOf(obj1)).Interface()
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
