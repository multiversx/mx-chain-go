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

func (p *payloadProcessor) ProcessPayload(_ []byte, topic string) error {
	if topic != outport.TopicSettings {
		return nil
	}

	p.handlerFunc()
	return nil
}

func (p *payloadProcessor) SetHandlerFunc(handler func()) error {
	if handler == nil {
		return errNilHandlerFunc
	}

	p.handlerFunc = handler
	return nil
}

func (p *payloadProcessor) Close() error {
	return nil
}

func (p *payloadProcessor) IsInterfaceNil() bool {
	return p == nil
}
