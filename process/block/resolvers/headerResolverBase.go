package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// HeaderResolverBase is a wrapper over Resolver that is specialized in resolving headers requests by their hash
type HeaderResolverBase struct {
	process.TopicResolverSender
	headers     storage.Cacher
	hdrStorage  storage.Storer
	marshalizer marshal.Marshalizer
}

// NewHeaderResolverBase creates a new base header resolver instance
func NewHeaderResolverBase(
	senderResolver process.TopicResolverSender,
	headers storage.Cacher,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*HeaderResolverBase, error) {

	if senderResolver == nil {
		return nil, process.ErrNilResolverSender
	}
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	if hdrStorage == nil {
		return nil, process.ErrNilHeadersStorage
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	hdrResolver := &HeaderResolverBase{
		TopicResolverSender: senderResolver,
		headers:             headers,
		hdrStorage:          hdrStorage,
		marshalizer:         marshalizer,
	}

	return hdrResolver, nil
}

// ParseReceivedMessage will transform the received p2p.Message in a RequestData object.
func (hrb *HeaderResolverBase) ParseReceivedMessage(message p2p.MessageP2P) (*process.RequestData, error) {
	rd := &process.RequestData{}
	err := rd.Unmarshal(hrb.marshalizer, message)
	if err != nil {
		return nil, err
	}
	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	return rd, nil
}

// ResolveHeaderFromHash resolves a header using its key (header hash)
func (hrb *HeaderResolverBase) ResolveHeaderFromHash(key []byte) ([]byte, error) {
	value, ok := hrb.headers.Peek(key)
	if !ok {
		return hrb.hdrStorage.Get(key)
	}

	buff, err := hrb.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestDataFromHash requests a header from other peers having input the hdr hash
func (hrb *HeaderResolverBase) RequestDataFromHash(hash []byte) error {
	return hrb.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
