package memp2p

import "github.com/ElrondNetwork/elrond-go/p2p"

func (messenger *Messenger) TopicValidator(name string) p2p.MessageProcessor {
	messenger.topicsMutex.RLock()
	processor, _ := messenger.topicValidators[name]
	messenger.topicsMutex.RUnlock()

	return processor
}
