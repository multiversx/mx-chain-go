package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// GenericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type GenericBlockBodyResolver struct {
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
	marshalizer marshal.Marshalizer) (*GenericBlockBodyResolver, error) {

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
func (gbbRes *GenericBlockBodyResolver) ProcessReceivedMessage(message p2p.MessageP2P) error {
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

func (gbbRes *GenericBlockBodyResolver) resolveBlockBodyRequest(rd *dataRetriever.RequestData) ([]byte, error) {
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

func (gbbRes *GenericBlockBodyResolver) miniBlockHashesFromRequestType(requestData *dataRetriever.RequestData) ([][]byte, error) {
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
func (gbbRes *GenericBlockBodyResolver) RequestDataFromHash(hash []byte) error {
	return gbbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (gbbRes *GenericBlockBodyResolver) RequestDataFromHashArray(hashes [][]byte) error {
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
func (gbbRes *GenericBlockBodyResolver) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	marshalizedMiniBlocks, missingMiniBlocksHashes := gbbRes.getMiniBlocks(hashes)
	miniBlocks := make(block.MiniBlockSlice, 0)

	for hash, marshalizedMiniBlock := range marshalizedMiniBlocks {
		miniBlock := &block.MiniBlock{}
		err := gbbRes.marshalizer.Unmarshal(miniBlock, marshalizedMiniBlock)
		if err != nil {
			log.Debug(err.Error())
			gbbRes.miniBlockPool.Remove([]byte(hash))
			err = gbbRes.miniBlockStorage.Remove([]byte(hash))
			if err != nil {
				log.Debug(err.Error())
			}

			missingMiniBlocksHashes = append(missingMiniBlocksHashes, []byte(hash))
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, missingMiniBlocksHashes
}

// getMiniBlocks method returns a list of serialized mini blocks from a given hash list either from data pool or from storage
func (gbbRes *GenericBlockBodyResolver) getMiniBlocks(hashes [][]byte) (map[string][]byte, [][]byte) {
	marshalizedMiniBlocks, missingMiniBlocksHashes := gbbRes.getMiniBlocksFromCache(hashes)
	if len(missingMiniBlocksHashes) == 0 {
		return marshalizedMiniBlocks, missingMiniBlocksHashes
	}

	marshalizedMiniBlocksFromStorer, missingMiniBlocksHashes := gbbRes.getMiniBlocksFromStorer(missingMiniBlocksHashes)
	for hash, marshalizedMiniBlockFromStorer := range marshalizedMiniBlocksFromStorer {
		marshalizedMiniBlocks[hash] = marshalizedMiniBlockFromStorer
	}

	return marshalizedMiniBlocks, missingMiniBlocksHashes
}

// getMiniBlocksFromCache returns a list of marshalized mini blocks from cache and a list of missing hashes
func (gbbRes *GenericBlockBodyResolver) getMiniBlocksFromCache(hashes [][]byte) (map[string][]byte, [][]byte) {
	marshalizedMiniBlocks := make(map[string][]byte)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		miniBlock, ok := gbbRes.miniBlockPool.Peek(hashes[i])
		if !ok {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		buff, err := gbbRes.marshalizer.Marshal(miniBlock)
		if err != nil {
			log.Debug(err.Error())
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		marshalizedMiniBlocks[string(hashes[i])] = buff
	}

	return marshalizedMiniBlocks, missingMiniBlocksHashes
}

// getMiniBlocksFromStorer returns a list of marshalized mini blocks from the storage unit and a list of missing hashes
func (gbbRes *GenericBlockBodyResolver) getMiniBlocksFromStorer(hashes [][]byte) (map[string][]byte, [][]byte) {
	marshalizedMiniBlocks := make(map[string][]byte)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		buff, err := gbbRes.miniBlockStorage.Get(hashes[i])
		if err != nil {
			log.Debug(err.Error())
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		marshalizedMiniBlocks[string(hashes[i])] = buff
	}

	return marshalizedMiniBlocks, missingMiniBlocksHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (gbbRes *GenericBlockBodyResolver) IsInterfaceNil() bool {
	if gbbRes == nil {
		return true
	}
	return false
}
