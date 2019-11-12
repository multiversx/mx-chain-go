package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// maxBuffToSendTrieNodes represents max buffer size to send in bytes
var maxBuffToSendTrieNodes = uint64(2 << 17) //128KB

// TrieNodeResolver is a wrapper over Resolver that is specialized in resolving trie node requests
type TrieNodeResolver struct {
	dataRetriever.TopicResolverSender
	trieDataGetter dataRetriever.TrieDataGetter
	marshalizer    marshal.Marshalizer
}

// NewTrieNodeResolver creates a new trie node resolver
func NewTrieNodeResolver(
	senderResolver dataRetriever.TopicResolverSender,
	trieDataGetter dataRetriever.TrieDataGetter,
	marshalizer marshal.Marshalizer,
) (*TrieNodeResolver, error) {
	if check.IfNil(senderResolver) {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(trieDataGetter) {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}
	if check.IfNil(marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}

	return &TrieNodeResolver{
		TopicResolverSender: senderResolver,
		trieDataGetter:      trieDataGetter,
		marshalizer:         marshalizer,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (tnRes *TrieNodeResolver) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	rd := &dataRetriever.RequestData{}
	err := rd.Unmarshal(tnRes.marshalizer, message)
	if err != nil {
		return err
	}

	if rd.Value == nil {
		return dataRetriever.ErrNilValue
	}

	switch rd.Type {
	case dataRetriever.HashType:
		serializedNodes, err := tnRes.trieDataGetter.GetSerializedNodes(rd.Value, maxBuffToSendTrieNodes)
		if err != nil {
			return err
		}

		buff, err := tnRes.marshalizer.Marshal(serializedNodes)
		if err != nil {
			return err
		}

		return tnRes.Send(buff, message.Peer())
	default:
		return dataRetriever.ErrRequestTypeNotImplemented
	}
}

// RequestDataFromHash requests trie nodes from other peers having input a trie node hash
func (tnRes *TrieNodeResolver) RequestDataFromHash(hash []byte) error {
	return tnRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnRes *TrieNodeResolver) IsInterfaceNil() bool {
	if tnRes == nil {
		return true
	}
	return false
}
