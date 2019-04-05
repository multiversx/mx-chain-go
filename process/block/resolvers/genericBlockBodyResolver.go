package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// GenericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type GenericBlockBodyResolver struct {
	process.TopicResolverSender
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
	marshalizer      marshal.Marshalizer
}

// NewGenericBlockBodyResolver creates a new block body resolver
func NewGenericBlockBodyResolver(
	senderResolver process.TopicResolverSender,
	miniBlockPool storage.Cacher,
	miniBlockStorage storage.Storer,
	marshalizer marshal.Marshalizer) (*GenericBlockBodyResolver, error) {

	if senderResolver == nil {
		return nil, process.ErrNilResolverSender
	}

	if miniBlockPool == nil {
		return nil, process.ErrNilBlockBodyPool
	}

	if miniBlockStorage == nil {
		return nil, process.ErrNilBlockBodyStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	bbResolver := &GenericBlockBodyResolver{
		TopicResolverSender: senderResolver,
		miniBlockPool:       miniBlockPool,
		miniBlockStorage:    miniBlockStorage,
		marshalizer:         marshalizer,
	}

	return bbResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (gbbRes *GenericBlockBodyResolver) ProcessReceivedMessage(message p2p.MessageP2P) ([]byte, error) {
	rd := &process.RequestData{}
	err := rd.Unmarshal(gbbRes.marshalizer, message)
	if err != nil {
		return nil, err
	}

	buff, err := gbbRes.resolveBlockBodyRequest(rd)
	if err != nil {
		return nil, err
	}

	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil, nil
	}

	return nil, gbbRes.Send(buff, message.Peer())
}

func (gbbRes *GenericBlockBodyResolver) resolveBlockBodyRequest(rd *process.RequestData) ([]byte, error) {

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	miniBlockHashes, err := gbbRes.miniBlockHashesFromRequestType(rd)
	if err != nil {
		return nil, err
	}
	miniBlocks := gbbRes.GetMiniBlocks(miniBlockHashes)

	if miniBlocks == nil {
		return nil, process.ErrNilMiniBlocks
	}

	buff, err := gbbRes.marshalizer.Marshal(miniBlocks)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (gbbRes *GenericBlockBodyResolver) miniBlockHashesFromRequestType(requestData *process.RequestData) ([][]byte, error) {
	miniBlockHashes := make([][]byte, 0)

	switch requestData.Type {
	case process.HashType:
		miniBlockHashes = append(miniBlockHashes, requestData.Value)

	case process.HashArrayType:
		err := gbbRes.marshalizer.Unmarshal(&miniBlockHashes, requestData.Value)

		if err != nil {
			return nil, process.ErrUnmarshalMBHashes
		}

	default:
		return nil, process.ErrInvalidRequestType
	}

	return miniBlockHashes, nil
}

// RequestDataFromHash requests a block body from other peers having input the block body hash
func (gbbRes *GenericBlockBodyResolver) RequestDataFromHash(hash []byte) error {
	return gbbRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (gbbRes *GenericBlockBodyResolver) RequestDataFromHashArray(hashes [][]byte) error {
	hash, err := gbbRes.marshalizer.Marshal(hashes)

	if err != nil {
		return err
	}

	return gbbRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.HashArrayType,
		Value: hash,
	})
}

// GetMiniBlocks method returns a list of deserialized mini blocks from a given hash list either from data pool or from storage
func (gbbRes *GenericBlockBodyResolver) GetMiniBlocks(hashes [][]byte) block.MiniBlockSlice {
	miniBlocks := gbbRes.getMiniBlocks(hashes)

	if miniBlocks == nil {
		return nil
	}

	mbLength := len(hashes)
	expandedMiniBlocks := make(block.MiniBlockSlice, mbLength)

	for i := 0; i < mbLength; i++ {
		mb := &block.MiniBlock{}
		err := gbbRes.marshalizer.Unmarshal(mb, miniBlocks[i])

		if err != nil {
			gbbRes.miniBlockPool.Remove(hashes[i])
			err = gbbRes.miniBlockStorage.Remove(hashes[i])
			log.LogIfError(err)
			return nil
		}

		expandedMiniBlocks[i] = mb
	}

	return expandedMiniBlocks
}

// getMiniBlocks method returns a list of serialized mini blocks from a given hash list either from data pool or from storage
func (gbbRes *GenericBlockBodyResolver) getMiniBlocks(hashes [][]byte) [][]byte {
	miniBlocks := gbbRes.getMiniBlocksFromCache(hashes)

	if miniBlocks != nil {
		return miniBlocks
	}

	return gbbRes.getMiniBlocksFromStorer(hashes)
}

// getMiniBlocksFromCache returns a full list of miniblocks from cache.
// If any of the miniblocks is missing the function returns nil
func (gbbRes *GenericBlockBodyResolver) getMiniBlocksFromCache(hashes [][]byte) [][]byte {
	miniBlocksLen := len(hashes)
	miniBlocks := make([][]byte, miniBlocksLen)

	for i := 0; i < miniBlocksLen; i++ {
		cachedMB, _ := gbbRes.miniBlockPool.Peek(hashes[i])

		if cachedMB == nil {
			return nil
		}

		buff, err := gbbRes.marshalizer.Marshal(cachedMB)

		if err != nil {
			log.LogIfError(err)
			return nil
		}

		miniBlocks[i] = buff
	}

	return miniBlocks
}

// getMiniBlocksFromStorer returns a full list of MiniBlocks from the storage unit.
// If any MiniBlock is missing or is invalid, it is removed and the function returns nil
func (gbbRes *GenericBlockBodyResolver) getMiniBlocksFromStorer(hashes [][]byte) [][]byte {
	miniBlocksLen := len(hashes)
	miniBlocks := make([][]byte, miniBlocksLen)

	for i := 0; i < miniBlocksLen; i++ {
		buff, err := gbbRes.miniBlockStorage.Get(hashes[i])

		if buff == nil {
			log.LogIfError(err)
			return nil
		}

		miniBlocks[i] = buff
	}

	return miniBlocks
}
