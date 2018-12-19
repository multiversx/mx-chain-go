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
	_, err := interceptor.NewTopicInterceptor("", nil, &mock.StringNewer{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewTopicInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	_, err := interceptor.NewTopicInterceptor("", &mock.MessengerStub{}, nil)
	assert.Equal(t, process.ErrNilNewer, err)
}

func TestNewTopicInterceptor_ErrMessengerAddTopicShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return errors.New("failure")
	}

	_, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.NotNil(t, err)
}

func TestNewTopicInterceptor_ErrMessengerRegistrationValidatorShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return nil
	}

	_, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Equal(t, process.ErrRegisteringValidator, err)
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

	_, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestNewTopicInterceptor_WithExistingTopicShouldWork(t *testing.T) {
	mes := mock.NewMessengerStub()

	wasCalled := false

	topic := p2p.NewTopic("test", &mock.StringNewer{}, mes.Marshalizer())
	topic.RegisterTopicValidator = func(v pubsub.Validator) error {
		wasCalled = true

		return nil
	}

	mes.AddTopicCalled = func(topic *p2p.Topic) error {
		assert.Fail(t, "should have not reached this point")

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return topic
	}

	_, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
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

	ti, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

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

	ti, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

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

	intercept, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	wasCalled := false

	intercept.SetCheckReceivedObjectHandler(func(newer p2p.Newer, rawData []byte) bool {
		wasCalled = true
		return false
	})

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

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

	intercept, err := interceptor.NewTopicInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	wasCalled := false

	intercept.SetCheckReceivedObjectHandler(func(newer p2p.Newer, rawData []byte) bool {
		wasCalled = true
		return true
	})

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.True(t, intercept.Validator(nil, message))
	assert.True(t, wasCalled)
}
