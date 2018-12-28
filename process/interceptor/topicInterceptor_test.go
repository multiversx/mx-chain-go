package interceptor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewTopicInterceptor

func TestNewTopicInterceptor_NilMessengerShouldErr(t *testing.T) {
	ti, err := interceptor.NewTopicInterceptor("", nil, &mock.StringCreator{})
	assert.Equal(t, process.ErrNilMessenger, err)
	assert.Nil(t, ti)
}

func TestNewTopicInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	ti, err := interceptor.NewTopicInterceptor("", &mock.MessengerStub{}, nil)
	assert.Equal(t, process.ErrNilNewer, err)
	assert.Nil(t, ti)
}

func TestNewTopicInterceptor_ErrMessengerAddTopicShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return errors.New("failure")
	}

	ti, err := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})
	assert.NotNil(t, err)
	assert.Nil(t, ti)
}

func TestNewTopicInterceptor_ErrMessengerRegistrationValidatorShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return nil
	}

	ti, err := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})
	assert.Equal(t, process.ErrRegisteringValidator, err)
	assert.Nil(t, ti)
}

func TestNewTopicInterceptor_OkValsShouldWork(t *testing.T) {
	mes := mock.NewMessengerStub()

	wasCalled := false
	mes.AddTopicCalled = func(t *p2p.Topic) error {
		t.RegisterTopicValidator = func(v pubsub.Validator) error {
			wasCalled = true

			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	ti, err := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.NotNil(t, ti)
}

func TestNewTopicInterceptor_WithExistingTopicShouldWork(t *testing.T) {
	mes := mock.NewMessengerStub()

	wasCalled := false

	topicName := "test"

	topic := p2p.NewTopic(topicName, &mock.StringCreator{}, mes.Marshalizer())
	topic.RegisterTopicValidator = func(v pubsub.Validator) error {
		wasCalled = true

		return nil
	}

	mes.AddTopicCalled = func(topic *p2p.Topic) error {
		assert.Fail(t, "should have not reached this point")

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == topicName {
			return topic
		}

		return nil
	}

	ti, err := interceptor.NewTopicInterceptor(topicName, mes, &mock.StringCreator{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.NotNil(t, ti)
}

//------- Validation

func TestTopicInterceptor_ValidationMalfunctionMarshalizerReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	ti, _ := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})

	//we have the validator func, let's test with a broken marshalizer
	mes.Marshalizer().(*mock.MarshalizerMock).Fail = true

	blankMessage := pubsub.Message{}

	assert.False(t, ti.Validator(nil, &blankMessage))
}

func TestTopicInterceptor_ValidationNilCheckReceivedObjectReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	ti, _ := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringCreator{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.False(t, ti.Validator(nil, message))
}

func TestTopicInterceptor_ValidationCheckReceivedObjectFalseReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	intercept, _ := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})

	wasCalled := false

	intercept.SetCheckReceivedObjectHandler(func(newer p2p.Creator, rawData []byte) error {
		wasCalled = true
		return errors.New("err1")
	})

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringCreator{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.False(t, intercept.Validator(nil, message))
	assert.True(t, wasCalled)
}

func TestTopicInterceptor_ValidationCheckReceivedObjectTrueReturnTrue(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	intercept, _ := interceptor.NewTopicInterceptor("", mes, &mock.StringCreator{})

	wasCalled := false

	intercept.SetCheckReceivedObjectHandler(func(newer p2p.Creator, rawData []byte) error {
		wasCalled = true
		return nil
	})

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringCreator{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.True(t, intercept.Validator(nil, message))
	assert.True(t, wasCalled)
}
