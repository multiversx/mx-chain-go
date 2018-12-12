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

//------- NewInterceptor

func TestNewInterceptor_NilMessengerShouldErr(t *testing.T) {
	_, err := interceptor.NewInterceptor("", nil, &mock.StringNewer{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	_, err := interceptor.NewInterceptor("", &mock.MessengerStub{}, nil)
	assert.Equal(t, process.ErrNilNewer, err)
}

func TestNewInterceptor_ErrMessengerAddTopicShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return errors.New("failure")
	}

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.NotNil(t, err)
}

func TestNewInterceptor_ErrMessengerRegistrationValidatorShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return nil
	}

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Equal(t, process.ErrRegisteringValidator, err)
}

func TestNewInterceptor_OkValsShouldWork(t *testing.T) {
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

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestNewInterceptor_WithExistingTopicShouldWork(t *testing.T) {
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

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- Validation

func TestInterceptorValidation_MalfunctionMarshalizerReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic
	var validator pubsub.Validator

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			validator = v
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	//we have the validator func, let's test with a broken marshalizer
	mes.Marshalizer().(*mock.MarshalizerMock).Fail = true

	blankMessage := pubsub.Message{}

	assert.False(t, validator(nil, &blankMessage))
}

func TestInterceptorValidation_NilCheckReceivedObjectReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic
	var validator pubsub.Validator

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			validator = v
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	_, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.False(t, validator(nil, message))
}

func TestInterceptorValidation_CheckReceivedObjectFalseReturnFalse(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic
	var validator pubsub.Validator

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			validator = v
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	intercept, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	wasCalled := false

	intercept.CheckReceivedObject = func(newer p2p.Newer, rawData []byte) bool {
		wasCalled = true
		return false
	}

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.False(t, validator(nil, message))
	assert.True(t, wasCalled)
}

func TestInterceptorValidation_CheckReceivedObjectTrueReturnTrue(t *testing.T) {
	mes := mock.NewMessengerStub()

	var topic *p2p.Topic
	var validator pubsub.Validator

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		topic = t
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			validator = v
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	intercept, err := interceptor.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	wasCalled := false

	intercept.CheckReceivedObject = func(newer p2p.Newer, rawData []byte) bool {
		wasCalled = true
		return true
	}

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.True(t, validator(nil, message))
	assert.True(t, wasCalled)
}
