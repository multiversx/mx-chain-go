package metablock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// ShardHeaderResolver is a wrapper over Resolver that is specialized in resolving shard headers requests,
// used by metachain nodes
type ShardHeaderResolver struct {
	process.TopicResolverSender
	headers     storage.Cacher
	hdrStorage  storage.Storer
	marshalizer marshal.Marshalizer
}

// NewShardHeaderResolver creates a new shard header resolver
func NewShardHeaderResolver(
	senderResolver process.TopicResolverSender,
	headers storage.Cacher,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*ShardHeaderResolver, error) {

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

	hdrResolver := &ShardHeaderResolver{
		TopicResolverSender: senderResolver,
		headers:             headers,
		hdrStorage:          hdrStorage,
		marshalizer:         marshalizer,
	}

	return hdrResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (shdrRes *ShardHeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
	rd := &process.RequestData{}
	err := rd.Unmarshal(shdrRes.marshalizer, message)
	if err != nil {
		return err
	}
	buff, err := shdrRes.resolveHdrRequest(rd)
	if err != nil {
		return err
	}
	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return shdrRes.Send(buff, message.Peer())
}

func (shdrRes *ShardHeaderResolver) resolveHdrRequest(rd *process.RequestData) ([]byte, error) {
	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	var buff []byte
	var err error

	if rd.Type == process.HashType {
		buff, err = shdrRes.resolveHeaderFromHash(rd.Value)
	} else {
		return nil, process.ErrResolveTypeUnknown
	}

	return buff, err
}

func (shdrRes *ShardHeaderResolver) resolveHeaderFromHash(key []byte) ([]byte, error) {
	value, ok := shdrRes.headers.Peek(key)

	if !ok {
		return shdrRes.hdrStorage.Get(key)
	}

	buff, err := shdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestDataFromHash requests a header from other peers having input the hdr hash
func (shdrRes *ShardHeaderResolver) RequestDataFromHash(hash []byte) error {
	return shdrRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
