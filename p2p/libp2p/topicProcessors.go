package libp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type topicProcessors struct {
	processors    map[string]p2p.MessageProcessor
	identifiers   []string
	mutProcessors sync.RWMutex
}

func newTopicProcessors() *topicProcessors {
	return &topicProcessors{
		processors:  make(map[string]p2p.MessageProcessor),
		identifiers: make([]string, 0),
	}
}

func (tp *topicProcessors) addTopicProcessor(identifier string, processor p2p.MessageProcessor) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if alreadyExists {
		return fmt.Errorf("%w, in addTopicProcessor, identifier %s",
			p2p.ErrMessageProcessorAlreadyDefined,
			identifier,
		)
	}

	tp.processors[identifier] = processor
	tp.identifiers = append(tp.identifiers, identifier)

	return nil
}

func (tp *topicProcessors) removeTopicProcessor(identifier string) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if !alreadyExists {
		return fmt.Errorf("%w, in removeTopicProcessor, identifier %s",
			p2p.ErrMessageProcessorDoesNotExists,
			identifier,
		)
	}

	delete(tp.processors, identifier)
	for i := 0; i < len(tp.identifiers); i++ {
		if identifier == tp.identifiers[i] {
			tp.identifiers = append(tp.identifiers[:i], tp.identifiers[i+1:]...)
			break
		}
	}

	return nil
}

func (tp *topicProcessors) getList() ([]string, []p2p.MessageProcessor) {
	tp.mutProcessors.RLock()
	defer tp.mutProcessors.RUnlock()

	list := make([]p2p.MessageProcessor, 0, len(tp.identifiers))
	identifiersCopy := make([]string, len(tp.identifiers))
	copy(identifiersCopy, tp.identifiers)

	for _, identifier := range tp.identifiers {
		list = append(list, tp.processors[identifier])
	}

	return identifiersCopy, list
}
