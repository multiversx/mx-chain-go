package resolver

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// Resolver is a struct coupled with a p2p.Topic that can process requests
type Resolver struct {
	messenger p2p.Messenger
	name      string
}

func NewResolver(
	name string,
	messenger p2p.Messenger,
) (*Resolver, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	topic := messenger.GetTopic(name)
	if topic == nil {
		return nil, process.ErrNilTopic
	}

	topic.ResolveRequest = func(objData []byte) p2p.Newer {

	}
}
