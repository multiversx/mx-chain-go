package interceptor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func createMessengerStub(hasTopic bool, hasTopicValidator bool) *mock.MessengerStub {
	return &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return hasTopicValidator
		},
		HasTopicCalled: func(name string) bool {
			return hasTopic
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(t string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}
}

//------- NewTopicInterceptor

func TestNewTopicInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	ti, err := interceptor.NewTopicInterceptor("", nil, &mock.MarshalizerStub{})

	assert.Nil(t, ti)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewTopicInterceptor_NilMArshalizerShouldErr(t *testing.T) {
	t.Parallel()

	ti, err := interceptor.NewTopicInterceptor("", &mock.MessengerStub{}, nil)

	assert.Nil(t, ti)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewTopicInterceptor_TopicValidatorAlreadySetShouldErr(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(false, false)
	mes.HasTopicValidatorCalled = func(name string) bool {
		if name == topic {
			return true
		}

		return false
	}

	ti, err := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.Nil(t, ti)
	assert.Equal(t, process.ErrValidatorAlreadySet, err)
}

func TestNewTopicInterceptor_CreateTopicErrorsShouldErr(t *testing.T) {
	t.Parallel()

	topic := "test topic"
	errCreateTopic := errors.New("create topic err")

	mes := createMessengerStub(false, false)
	mes.CreateTopicCalled = func(name string, createPipeForTopic bool) error {
		if name == topic {
			return errCreateTopic
		}

		return nil
	}

	ti, err := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.Nil(t, ti)
	assert.Equal(t, errCreateTopic, err)
}

func TestNewTopicInterceptor_OkValsCreateTopicShouldCallSetTopicValidator(t *testing.T) {
	t.Parallel()

	topic := "test topic"
	wasCalled := false

	mes := createMessengerStub(false, false)
	mes.SetTopicValidatorCalled = func(t string, handler func(message p2p.MessageP2P) error) error {
		if t == topic && handler != nil {
			wasCalled = true
		}

		return nil
	}

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.NotNil(t, ti)
	assert.True(t, wasCalled)
}

func TestNewTopicInterceptor_OkValsTopicExistsShouldCallSetTopicValidator(t *testing.T) {
	t.Parallel()

	topic := "test topic"
	wasCalled := false

	mes := createMessengerStub(true, false)
	mes.SetTopicValidatorCalled = func(t string, handler func(message p2p.MessageP2P) error) error {
		if t == topic && handler != nil {
			wasCalled = true
		}

		return nil
	}

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.NotNil(t, ti)
	assert.True(t, wasCalled)
}

func TestNewTopicInterceptor_OkValsShouldRetTheSetTopicValidatorError(t *testing.T) {
	t.Parallel()

	topic := "test topic"
	errCheck := errors.New("check error")

	mes := createMessengerStub(false, false)
	mes.SetTopicValidatorCalled = func(t string, handler func(message p2p.MessageP2P) error) error {
		return errCheck
	}

	ti, err := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.Nil(t, ti)
	assert.Equal(t, errCheck, err)
}

//------- Validator

func TestTopicInterceptor_ValidatorNotSetReceivedMessageHandlerShouldErr(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(true, false)

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})
	err := ti.Validator(nil)

	assert.Equal(t, process.ErrNilReceivedMessageHandler, err)
}

func TestTopicInterceptor_ValidatorCallsHandler(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(true, false)

	wasCalled := false

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})
	ti.SetReceivedMessageHandler(func(message p2p.MessageP2P) error {
		wasCalled = true
		return nil
	})

	_ = ti.Validator(nil)

	assert.True(t, wasCalled)
}

func TestTopicInterceptor_ValidatorReturnsHandlersError(t *testing.T) {
	t.Parallel()

	topic := "test topic"
	errCheck := errors.New("check err")

	mes := createMessengerStub(true, false)

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})
	ti.SetReceivedMessageHandler(func(message p2p.MessageP2P) error {
		return errCheck
	})

	err := ti.Validator(nil)

	assert.Equal(t, errCheck, err)
}

func TestTopicInterceptor_Name(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(true, false)

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.Equal(t, topic, ti.Name())
}

func TestTopicInterceptor_ReceivedMessageHandler(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(true, false)

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, &mock.MarshalizerStub{})

	assert.Nil(t, ti.ReceivedMessageHandler())

	ti.SetReceivedMessageHandler(func(message p2p.MessageP2P) error {
		return nil
	})

	assert.NotNil(t, ti.ReceivedMessageHandler())
}

func TestTopicInterceptor_Marshalizer(t *testing.T) {
	t.Parallel()

	topic := "test topic"

	mes := createMessengerStub(true, false)

	m := &mock.MarshalizerStub{}

	ti, _ := interceptor.NewTopicInterceptor(topic, mes, m)

	assert.True(t, ti.Marshalizer() == m)
}
