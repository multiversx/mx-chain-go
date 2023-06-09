package host

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/data/outport"
)

var errNilHandlerFunc = errors.New("err nil handler func")

type payloadProcessor struct {
	handlerFunc func()
}

func newPayloadProcessor() (*payloadProcessor, error) {
	return &payloadProcessor{
		handlerFunc: func() {},
	}, nil
}

// ProcessPayload will process the provided payload based on the topic
func (p *payloadProcessor) ProcessPayload(_ []byte, topic string) error {
	if topic != outport.TopicSettings {
		return nil
	}

	p.handlerFunc()
	return nil
}

// SetHandlerFunc will set the handler func
func (p *payloadProcessor) SetHandlerFunc(handler func()) error {
	if handler == nil {
		return errNilHandlerFunc
	}

	p.handlerFunc = handler
	return nil
}

// Close will do nothing
func (p *payloadProcessor) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *payloadProcessor) IsInterfaceNil() bool {
	return p == nil
}
