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

type testTopicStringNewer struct {
	Data string
}

// New will return a new instance of string. Dummy, just to implement Cloner interface as strings are immutable
func (sc *testTopicStringNewer) New() p2p.Newer {
	return &testTopicStringNewer{}
}

// ID will return the same string as ID
func (sc *testTopicStringNewer) ID() string {
	return sc.Data
}

func TestTopicAddEventHandlerNilShouldNotAddHandler(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.AddDataReceived(nil)

	assert.Equal(t, len(topic.EventBusData()), 0)
}

func TestTopicAddEventHandlerWithArealFuncShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {

	})

	assert.Equal(t, len(topic.EventBusData()), 1)
}

func TestTopicCreateObjectNilDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	_, err := topic.CreateObject(nil)

	assert.NotNil(t, err)
}

func TestTopicCreateObjectEmptyDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	_, err := topic.CreateObject(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopicCreateObjectMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	_, err := topic.CreateObject(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopicNewObjReceivedNilObjShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	err := topic.NewObjReceived(nil, "")

	assert.NotNil(t, err)
}

func TestTopicNewObjReceivedOKMsgShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	wg := sync.WaitGroup{}
	wg.Add(1)

	cnt := int32(0)
	//attach event handler
	topic.AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		assert.Equal(t, name, "test")

		switch data.(type) {
		case p2p.Newer:
			atomic.AddInt32(&cnt, 1)
		default:
			assert.Fail(t, "The data should have been string!")
		}

		wg.Done()

	})

	marsh := mock.MarshalizerMock{}
	payload, err := marsh.Marshal(&testTopicStringNewer{Data: "aaaa"})
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

func TestTopicBroadcastNilDataShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	err := topic.Broadcast(nil)

	assert.NotNil(t, err)
}

func TestTopicBroadcastMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.Marsh().(*mock.MarshalizerMock).Fail = true
	defer func() {
		topic.Marsh().(*mock.MarshalizerMock).Fail = false
	}()

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopicBroadcastNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	err := topic.Broadcast("a string")

	assert.NotNil(t, err)
}

func TestTopicBroadcastSendOkShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

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

func TestTopicSendRequestNilHashShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(nil)

	assert.NotNil(t, err)
}

func TestTopicSendRequestEmptyHashShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(make([]byte, 0))

	assert.NotNil(t, err)
}

func TestTopicSendRequestNoHandlerShouldRetErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})
	err := topic.SendRequest(make([]byte, 1))

	assert.NotNil(t, err)
}

func TestTopicSendRequestShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.SetRequest(func(hash []byte) error {
		if bytes.Equal(hash, []byte("AAAA")) {
			return nil
		}

		return errors.New("should have not got here")
	})
	err := topic.SendRequest([]byte("AAAA"))

	assert.Nil(t, err)
}

func TestTopicRegisterValidatorNoHandlerShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	err := topic.RegisterValidator(nil)
	assert.NotNil(t, err)
}

func TestTopicRegisterValidatorShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	topic.SetRegisterTopicValidator(func(v pubsub.Validator) error {
		return nil
	})

	err := topic.RegisterValidator(nil)
	assert.Nil(t, err)
}

func TestTopicUnregisterValidatorNoHandlerShouldErr(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

	err := topic.UnregisterValidator()
	assert.NotNil(t, err)
}

func TestTopicUnregisterValidatorShouldWork(t *testing.T) {
	t.Parallel()

	topic := p2p.NewTopic("test", &testTopicStringNewer{}, &mock.MarshalizerMock{})

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

func BenchmarkTopicNewObjectCreationPlainInit(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		obj1 = benchmark{}
	}

	obj1.field1 = make([]byte, 0)
}

func BenchmarkTopicNewObjectCreationReflectionNew(b *testing.B) {
	obj1 := benchmark{}

	for i := 0; i < b.N; i++ {
		reflect.New(reflect.TypeOf(obj1)).Interface()
	}
}
