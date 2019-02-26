package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// GenericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type GenericBlockBodyResolver struct {
	process.TopicResolverSender
	blockBodyPool storage.Cacher
	blockStorage  storage.Storer
	marshalizer   marshal.Marshalizer
}

// NewGenericBlockBodyResolver creates a new block body resolver
func NewGenericBlockBodyResolver(
	senderResolver process.TopicResolverSender,
	blockBodyPool storage.Cacher,
	blockBodyStorage storage.Storer,
	marshalizer marshal.Marshalizer) (*GenericBlockBodyResolver, error) {

	if senderResolver == nil {
		return nil, process.ErrNilResolverSender
	}

	if blockBodyPool == nil {
		return nil, process.ErrNilBlockBodyPool
	}

	if blockBodyStorage == nil {
		return nil, process.ErrNilBlockBodyStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	bbResolver := &GenericBlockBodyResolver{
		TopicResolverSender: senderResolver,
		blockBodyPool:       blockBodyPool,
		blockStorage:        blockBodyStorage,
		marshalizer:         marshalizer,
	}

	return bbResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (gbbRes *GenericBlockBodyResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
	rd := &process.RequestData{}
	err := rd.Unmarshal(gbbRes.marshalizer, message)
	if err != nil {
		return err
	}

	buff, err := gbbRes.resolveBlockBodyRequest(rd)
	if err != nil {
		return err
	}

	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return gbbRes.Send(buff, message.Peer())
}

func (gbbRes *GenericBlockBodyResolver) resolveBlockBodyRequest(rd *process.RequestData) ([]byte, error) {
	if rd.Type != process.HashType {
		return nil, process.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	blockBody, _ := gbbRes.blockBodyPool.Get(rd.Value)
	if blockBody != nil {
		buff, err := gbbRes.marshalizer.Marshal(blockBody)
		if err != nil {
			return nil, err
		}

		return buff, nil
	}

	return gbbRes.blockStorage.Get(rd.Value)
}

// RequestDataFromHash requests a block body from other peers having input the block body hash
func (gbbRes *GenericBlockBodyResolver) RequestDataFromHash(hash []byte) error {
	return gbbRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
