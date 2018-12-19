package resolver

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// Resolver is a struct coupled with a p2p.Topic that can process requests
type Resolver struct {
	messenger   p2p.Messenger
	name        string
	topic       *p2p.Topic
	marshalizer marshal.Marshalizer

	ResolveRequest func(rd *RequestData) []byte
}

func NewResolver(
	name string,
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
) (*Resolver, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	topic := messenger.GetTopic(name)
	if topic == nil {
		return nil, process.ErrNilTopic
	}

	if topic.ResolveRequest != nil {
		return nil, process.ErrResolveRequestAlreadyAssigned
	}

	resolver := &Resolver{
		name:        name,
		messenger:   messenger,
		topic:       topic,
		marshalizer: marshalizer,
	}

	topic.ResolveRequest = func(objData []byte) []byte {
		rd := &RequestData{}

		err := marshalizer.Unmarshal(rd, objData)
		if err != nil {
			return nil
		}

		if resolver.ResolveRequest != nil {
			return resolver.ResolveRequest(rd)
		}

		return nil
	}

	return resolver, nil
}

func (r *Resolver) RequestData(rd *RequestData) error {
	if rd == nil {
		return process.ErrNilRequestData
	}

	buff, err := r.marshalizer.Marshal(rd)
	if err != nil {
		return err
	}

	return r.topic.Request(buff)
}
