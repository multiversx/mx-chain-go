package host

import (
	"errors"
)

var (
	errNilHandlerFunc = errors.New("err nil handler func")
	errEmptyTopic     = errors.New("empty topic")
)

type payloadProcessor struct {
	handlerFuncs map[string]func() error
}

func newPayloadProcessor() (*payloadProcessor, error) {
	return &payloadProcessor{
		handlerFuncs: make(map[string]func() error),
	}, nil
}

// ProcessPayload will process the provided payload based on the topic
func (p *payloadProcessor) ProcessPayload(_ []byte, topic string) error {
	handlerFunc, found := p.handlerFuncs[topic]
	if !found {
		return nil
	}

	return handlerFunc()
}

// SetHandlerFuncForTopic will set the handler func for the provided topic
func (p *payloadProcessor) SetHandlerFuncForTopic(handler func() error, topic string) error {
	if handler == nil {
		return errNilHandlerFunc
	}
	if topic == "" {
		return errEmptyTopic
	}

	p.handlerFuncs[topic] = handler
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
