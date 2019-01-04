package resolver

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

// topicResolver is a struct coupled with a p2p.Topic that can process requests
type topicResolver struct {
	messenger   p2p.Messenger
	name        string
	topic       *p2p.Topic
	marshalizer marshal.Marshalizer

	resolveRequest func(rd process.RequestData) ([]byte, error)
}

// NewTopicResolver returns a new topic resolver instance
func NewTopicResolver(
	name string,
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
) (*topicResolver, error) {

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

	resolver := &topicResolver{
		name:        name,
		messenger:   messenger,
		topic:       topic,
		marshalizer: marshalizer,
	}

	topic.ResolveRequest = func(objData []byte) []byte {
		rd := process.RequestData{}

		err := marshalizer.Unmarshal(&rd, objData)
		if err != nil {
			return nil
		}

		if resolver.resolveRequest != nil {
			buff, err := resolver.resolveRequest(rd)
			if err != nil {
				log.Debug(err.Error())
			}

			return buff
		}

		return nil
	}

	return resolver, nil
}

// RequestData is used to request data over channels (topics) from other peers
// This method only sends the request, the received data should be handled by interceptors
func (tr *topicResolver) RequestData(rd process.RequestData) error {
	buff, err := tr.marshalizer.Marshal(&rd)
	if err != nil {
		return err
	}

	if tr.topic.Request == nil {
		return process.ErrTopicNotWiredToMessenger
	}

	return tr.topic.Request(buff)
}

// SetResolverHandler sets the handler that will be called when a new request comes from other peers to
// current node
func (tr *topicResolver) SetResolverHandler(handler func(rd process.RequestData) ([]byte, error)) {
	tr.resolveRequest = handler
}

// ResolverHandler gets the handler that will be called when a new request comes from other peers to
// current node
func (tr *topicResolver) ResolverHandler() func(rd process.RequestData) ([]byte, error) {
	return tr.resolveRequest
}
