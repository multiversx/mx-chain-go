package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// genericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type genericBlockBodyResolver struct {
	dataRetriever.TopicResolverSender
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
	marshalizer      marshal.Marshalizer
}

// NewGenericBlockBodyResolver creates a new block body resolver
func NewGenericBlockBodyResolver(
	senderResolver dataRetriever.TopicResolverSender,
	miniBlockPool storage.Cacher,
	miniBlockStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*genericBlockBodyResolver, error) {

	if senderResolver == nil || senderResolver.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverSender
	}

	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilBlockBodyPool
	}

	if miniBlockStorage == nil || miniBlockStorage.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilBlockBodyStorage
	}

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilMarshalizer
	}

	bbResolver := &genericBlockBodyResolver{
		TopicResolverSender: senderResolver,
		miniBlockPool:       miniBlockPool,
		miniBlockStorage:    miniBlockStorage,
		marshalizer:         marshalizer,
	}

	return bbResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (gbbRes *genericBlockBodyResolver) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	rd := &dataRetriever.RequestData{}
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

func (gbbRes *genericBlockBodyResolver) resolveBlockBodyRequest(rd *dataRetriever.RequestData) ([]byte, error) {
	if rd.Value == nil {
		return nil, dataRetriever.ErrNilValue
	}

	hashes, err := gbbRes.miniBlockHashesFromRequestType(rd)
	if err != nil {
		return nil, err
	}

	miniBlocks, _ := gbbRes.GetMiniBlocks(hashes)
	if len(miniBlocks) == 0 {
		return nil, dataRetriever.ErrEmptyMiniBlockSlice
	}

	buff, err := gbbRes.marshalizer.Marshal(miniBlocks)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (gbbRes *genericBlockBodyResolver) miniBlockHashesFromRequestType(requestData *dataRetriever.RequestData) ([][]byte, error) {
	miniBlockHashes := make([][]byte, 0)

	switch requestData.Type {
	case dataRetriever.HashType:
		miniBlockHashes = append(miniBlockHashes, requestData.Value)

	case dataRetriever.HashArrayType:
		err := gbbRes.marshalizer.Unmarshal(&miniBlockHashes, requestData.Value)

		if err != nil {
			return nil, dataRetriever.ErrUnmarshalMBHashes
		}

	default:
		return nil, dataRetriever.ErrInvalidRequestType
	}

	return miniBlockHashes, nil
}

// RequestDataFromHash requests a block body from other peers having input the block body hash
func (gbbRes *genericBlockBodyResolver) RequestDataFromHash(hash []byte) error {
	return gbbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (gbbRes *genericBlockBodyResolver) RequestDataFromHashArray(hashes [][]byte) error {
	hash, err := gbbRes.marshalizer.Marshal(hashes)

	if err != nil {
		return err
	}

	return gbbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: hash,
	})
}

// GetMiniBlocks method returns a list of deserialized mini blocks from a given hash list either from data pool or from storage
func (gbbRes *genericBlockBodyResolver) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks, missingMiniBlocksHashes := gbbRes.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) == 0 {
		return miniBlocks, missingMiniBlocksHashes
	}

	miniBlocksFromStorer, missingMiniBlocksHashes := gbbRes.getMiniBlocksFromStorer(missingMiniBlocksHashes)
	miniBlocks = append(miniBlocks, miniBlocksFromStorer...)

	return miniBlocks, missingMiniBlocksHashes
}

// GetMiniBlocksFromPool method returns a list of deserialized mini blocks from a given hash list from data pool
func (gbbRes *genericBlockBodyResolver) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks := make(block.MiniBlockSlice, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		obj, ok := gbbRes.miniBlockPool.Peek(hashes[i])
		if !ok {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlock, ok := obj.(*block.MiniBlock)
		if !ok {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, missingMiniBlocksHashes
}

// getMiniBlocksFromStorer returns a list of mini blocks from storage and a list of missing hashes
func (gbbRes *genericBlockBodyResolver) getMiniBlocksFromStorer(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks := make(block.MiniBlockSlice, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		buff, err := gbbRes.miniBlockStorage.Get(hashes[i])
		if err != nil {
			log.Debug(err.Error())
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlock := &block.MiniBlock{}
		err = gbbRes.marshalizer.Unmarshal(miniBlock, buff)
		if err != nil {
			log.Debug(err.Error())
			gbbRes.miniBlockPool.Remove([]byte(hashes[i]))
			err = gbbRes.miniBlockStorage.Remove([]byte(hashes[i]))
			if err != nil {
				log.Debug(err.Error())
			}

			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, missingMiniBlocksHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (gbbRes *genericBlockBodyResolver) IsInterfaceNil() bool {
	if gbbRes == nil {
		return true
	}
	return false
}
