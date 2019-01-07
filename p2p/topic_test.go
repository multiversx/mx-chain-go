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

type testTopicStringCreator struct {
	Data string
}

// New will return a new instance of string. Dummy, just to implement Cloner interface as strings are immutable
func (sc *testTopicStringCreator) Create() p2p.Creator {
	return &testTopicStringCreator{}
}

// ID will return the same string as ID
func (sc *testTopicStringCreator) ID() string {
	return sc.Data
}

func TestTopic_AddEventHandlerNilShouldNotAddHandler(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.AddDataReceived(nil)

	assert.Equal(t, len(topic.EventBusData()), 0)
}

func TestTopic_AddEventHandlerWithArealFuncShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {

	})

	assert.Equal(t, len(topic.EventBusData()), 1)
}

func TestTopic_CreateObjectNilDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	_, err := topic.CreateObject(nil)

	assert.NotNil(t, err)
}

func TestTopic_CreateObjectEmptyDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	_, err := topic.CreateObject(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopic_CreateObjectMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, marshalizer)

	marshalizer.Fail = true

	_, err := topic.CreateObject(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopic_NewObjReceivedNilObjShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.NewObjReceived(nil, "")

	assert.NotNil(t, err)
}

func TestTopic_NewObjReceivedOKMsgShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	wg := sync.WaitGroup{}
	wg.Add(1)

	cnt := int32(0)
	//attach event handler
	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		assert.Equal(t, name, "test")

		switch data.(type) {
		case p2p.Creator:
			atomic.AddInt32(&cnt, 1)
		default:
			assert.Fail(t, "The data should have been string!")
		}

		wg.Done()

	})

	marsh := mock.MarshalizerMock{}
	payload, err := marsh.Marshal(&testTopicStringCreator{Data: "aaaa"})
	assert.Nil(t, err)

	obj, err := topic.CreateObject(payload)
	assert.Nil(t, err)
	err = topic.NewObjReceived(obj, "")
	assert.Nil(t, err)

	//start a go routine as watchdog for the wg.Wait()
	go func() {
		time.Sleep(time.Second * 2)
		wg.Done()
	}()

	//wait for the go routine to finish
	wg.Wait()

	assert.Equal(t, atomic.LoadInt32(&cnt), int32(1))
}

//------- Broadcast

func TestTopic_BroadcastNilDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.Broadcast(nil)

	assert.NotNil(t, err)
}

func TestTopic_BroadcastMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, marshalizer)

	topic.SendData = func(data []byte) error {
		return nil
	}

	marshalizer.Fail = true

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopic_BroadcastNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopic_BroadcastSendOkShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.SendData = func(data []byte) error {
		if topic.Name() != "test" {
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

//------- BroadcastBuff

func TestTopic_BroadcastBuffNilDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.BroadcastBuff(nil)

	assert.NotNil(t, err)
}

func TestTopic_BroadcastBuffNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.BroadcastBuff([]byte("a string"))

	assert.NotNil(t, err)
}

func TestTopic_BroadcastBuffSendOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	topic := p2p.NewTopic("test", &testTopicStringCreator{}, marshalizer)

	topic.SendData = func(data []byte) error {
		if topic.Name() != "test" {
			return errors.New("should have been test")
		}

		if data == nil {
			return errors.New("should have not been nil")
		}

		fmt.Printf("Message: %v\n", string(data))
		return nil
	}

	buff, err := marshalizer.Marshal(testTopicStringCreator{"AAA"})
	assert.Nil(t, err)

	err = topic.BroadcastBuff(buff)
	assert.Nil(t, err)
}

//------- SendRequest

func TestTopic_SendRequestNilHashShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(nil)

	assert.NotNil(t, err)
}

func TestTopic_SendRequestEmptyHashShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopic_SendRequestNoHandlerShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopic_SendRequestShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.Request = func(hash []byte) error {
		if bytes.Equal(hash, []byte("AAAA")) {
			return nil
		}

		return errors.New("should have not got here")
	}
	err := topic.SendRequest([]byte("AAAA"))

	assert.Nil(t, err)
}

//------- RegisterValidator

func TestTopic_RegisterValidatorNoHandlerShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.RegisterValidator(nil)
	assert.NotNil(t, err)
}

func TestTopic_RegisterValidatorShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.RegisterTopicValidator = func(v pubsub.Validator) error {
		return nil
	}

	err := topic.RegisterValidator(nil)
	assert.Nil(t, err)
}

//------- UnregisterValidator

func TestTopic_UnregisterValidatorNoHandlerShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	err := topic.UnregisterValidator()
	assert.NotNil(t, err)
}

func TestTopic_UnregisterValidatorShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringCreator{}, &mock.MarshalizerMock{})

	topic.UnregisterTopicValidator = func() error {
		return nil
	}

	err := topic.UnregisterValidator()
	assert.Nil(t, err)
}

//------- Benchmarks

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

func BenchmarkTopic_NewObjectCreationPlainInit(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		obj1 = benchmark{}
	}

	obj1.field1 = make([]byte, 0)
}

func BenchmarkTopic_NewObjectCreationReflectionNew(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		reflect.New(reflect.TypeOf(obj1)).Interface()
	}
}
