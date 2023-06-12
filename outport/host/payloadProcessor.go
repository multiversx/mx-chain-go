package host

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

var (
	errNilHandlerFunc = errors.New("err nil handler func")
	errEmptyTopic     = errors.New("empty topic")
)

type payloadProcessor struct {
	handlerFuncs map[string]func() error
	mutex        sync.RWMutex
	log          core.Logger
}

func newPayloadProcessor(log core.Logger) (*payloadProcessor, error) {
	return &payloadProcessor{
		handlerFuncs: make(map[string]func() error),
		log:          log,
	}, nil
}

// ProcessPayload will process the provided payload based on the topic
func (p *payloadProcessor) ProcessPayload(_ []byte, topic string) error {
	p.mutex.RLock()
	handlerFunc, found := p.handlerFuncs[topic]
	p.mutex.RUnlock()

	if !found {
		p.log.Debug("p.ProcessPayload no handler function for the provided topic", "topic", topic)
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

	p.mutex.Lock()
	p.handlerFuncs[topic] = handler
	p.mutex.Unlock()
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
