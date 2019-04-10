package resolvers

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// ShardHeaderResolver is a wrapper over Resolver that is specialized in resolving shard headers requests,
// used by metachain nodes
type ShardHeaderResolver struct {
	*HeaderResolverBase
}

// NewShardHeaderResolver creates a new shard header resolver
func NewShardHeaderResolver(
	senderResolver dataRetriever.TopicResolverSender,
	headers storage.Cacher,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*ShardHeaderResolver, error) {

	hdrResolverBase, err := NewHeaderResolverBase(
		senderResolver,
		headers,
		hdrStorage,
		marshalizer,
	)
	if err != nil {
		return nil, err
	}

	return &ShardHeaderResolver{HeaderResolverBase: hdrResolverBase}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (shdrRes *ShardHeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
	rd, err := shdrRes.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	if rd.Type != dataRetriever.HashType {
		return dataRetriever.ErrResolveTypeUnknown
	}

	buff, err := shdrRes.ResolveHeaderFromHash(rd.Value)
	if err != nil {
		return err
	}
	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return shdrRes.Send(buff, message.Peer())
}
