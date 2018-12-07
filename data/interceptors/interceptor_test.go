package interceptors_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

//------- NewInterceptor

func TestNewInterceptorNilMessengerShouldErr(t *testing.T) {
	_, err := interceptors.NewInterceptor("", nil, &mock.StringNewer{})
	assert.Equal(t, interceptors.ErrNilMessenger, err)
}

func TestNewInterceptorNilTemplateObjectShouldErr(t *testing.T) {
	_, err := interceptors.NewInterceptor("", &mock.MessengerStub{}, nil)
	assert.Equal(t, interceptors.ErrNilNewer, err)
}

func TestNewInterceptorErrMessengerAddTopicShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return errors.New("failure")
	}

	_, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
	assert.NotNil(t, err)
}

func TestNewInterceptorErrMessengerRegistrationValidatorShouldErr(t *testing.T) {
	mes := mock.NewMessengerStub()
	mes.AddTopicCalled = func(t *p2p.Topic) error {
		return nil
	}

	_, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Equal(t, interceptors.ErrRegisteringValidator, err)
}

func TestNewInterceptorOkValsShouldWork(t *testing.T) {
	mes := mock.NewMessengerStub()

	wasCalled := false
	mes.AddTopicCalled = func(t *p2p.Topic) error {
		t.RegisterTopicValidator = func(v pubsub.Validator) error {
			wasCalled = true

			return nil
		}

		return nil
	}

	_, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

//------- Validation

func TestInterceptorValidationMalfunctionMarshalizerReturnFalse(t *testing.T) {
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

	_, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	//we have the validator func, let's test with a broken marshalizer
	mes.Marshalizer().(*mock.MarshalizerMock).Fail = true

	blankMessage := pubsub.Message{}

	assert.False(t, validator(nil, &blankMessage))
}

func TestInterceptorValidationNilCheckReceivedObjectReturnFalse(t *testing.T) {
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

	_, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
	assert.Nil(t, err)

	//we have the validator func, let's test with a message
	objToMarshalizeUnmarshalize := &mock.StringNewer{Data: "test data"}

	message := &pubsub.Message{Message: &pubsub_pb.Message{}}
	data, err := mes.Marshalizer().Marshal(objToMarshalizeUnmarshalize)
	assert.Nil(t, err)
	message.Data = data

	assert.False(t, validator(nil, message))
}

func TestInterceptorValidationCheckReceivedObjectFalseReturnFalse(t *testing.T) {
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

	intercept, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
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

func TestInterceptorValidationCheckReceivedObjectTrueReturnTrue(t *testing.T) {
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

	intercept, err := interceptors.NewInterceptor("", mes, &mock.StringNewer{})
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
