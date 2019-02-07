package interceptor

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// TopicInterceptor is a struct coupled with a p2p.Topic that calls CheckReceivedObject whenever Messenger needs to validate
// the data
type topicInterceptor struct {
	messenger   p2p.Messenger
	name        string
	marshalizer marshal.Marshalizer

	receivedMessageHandler func(message p2p.MessageP2P) error
}

// NewTopicInterceptor returns a new data interceptor that runs coupled with p2p.Topics
func NewTopicInterceptor(
	name string,
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
) (*topicInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	intercept := topicInterceptor{
		messenger:   messenger,
		name:        name,
		marshalizer: marshalizer,
	}

	err := intercept.registerValidator()
	if err != nil {
		return nil, err
	}

	return &intercept, nil
}

func (ti *topicInterceptor) registerValidator() error {
	if ti.messenger.HasTopicValidator(ti.name) {
		return process.ErrValidatorAlreadySet
	}

	if !ti.messenger.HasTopic(ti.name) {
		err := ti.messenger.CreateTopic(ti.name, true)
		if err != nil {
			return err
		}
	}

	return ti.messenger.SetTopicValidator(ti.name, ti.validator)
}

func (ti *topicInterceptor) validator(message p2p.MessageP2P) error {
	if ti.receivedMessageHandler == nil {
		return process.ErrNilReceivedMessageHandler
	}

	return ti.receivedMessageHandler(message)
}

// Name returns the name of the interceptor
func (ti *topicInterceptor) Name() string {
	return ti.name
}

// SetReceivedMessageHandler sets the handler that gets called each time new message arrives from peers or from self
func (ti *topicInterceptor) SetReceivedMessageHandler(handler func(message p2p.MessageP2P) error) {
	ti.receivedMessageHandler = handler
}

// ReceivedMessageHandler returns the handler that gets called each time new message arrives from peers or from self
func (ti *topicInterceptor) ReceivedMessageHandler() func(message p2p.MessageP2P) error {
	return ti.receivedMessageHandler
}

func (ti *topicInterceptor) Marshalizer() marshal.Marshalizer {
	return ti.marshalizer
}
