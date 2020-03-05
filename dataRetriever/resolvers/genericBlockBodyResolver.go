package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgGenericBlockBodyResolver is the argument structure used to create new GenericBlockBodyResolver instance
type ArgGenericBlockBodyResolver struct {
	SenderResolver   dataRetriever.TopicResolverSender
	MiniBlockPool    storage.Cacher
	MiniBlockStorage storage.Storer
	Marshalizer      marshal.Marshalizer
	AntifloodHandler dataRetriever.P2PAntifloodHandler
	Throttler        dataRetriever.ResolverThrottler
}

// genericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type genericBlockBodyResolver struct {
	dataRetriever.TopicResolverSender
	messageProcessor
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
}

// NewGenericBlockBodyResolver creates a new block body resolver
func NewGenericBlockBodyResolver(arg ArgGenericBlockBodyResolver) (*genericBlockBodyResolver, error) {
	if check.IfNil(arg.SenderResolver) {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.MiniBlockPool) {
		return nil, dataRetriever.ErrNilBlockBodyPool
	}
	if check.IfNil(arg.MiniBlockStorage) {
		return nil, dataRetriever.ErrNilBlockBodyStorage
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return nil, dataRetriever.ErrNilThrottler
	}

	bbResolver := &genericBlockBodyResolver{
		TopicResolverSender: arg.SenderResolver,
		miniBlockPool:       arg.MiniBlockPool,
		miniBlockStorage:    arg.MiniBlockStorage,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.Topic(),
			throttler:        arg.Throttler,
		},
	}

	return bbResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (gbbRes *genericBlockBodyResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	err := gbbRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	gbbRes.throttler.StartProcessing()
	defer gbbRes.throttler.EndProcessing()

	rd, err := gbbRes.parseReceivedMessage(message)
	if err != nil {
		return err
	}

	buff, err := gbbRes.resolveBlockBodyRequest(rd)
	if err != nil {
		return err
	}

	if buff == nil {
		log.Trace("missing data", "data", rd)
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
func (gbbRes *genericBlockBodyResolver) RequestDataFromHash(hash []byte, epoch uint32) error {
	return gbbRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
		Epoch: epoch,
	})
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (gbbRes *genericBlockBodyResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
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
		buff, err := gbbRes.miniBlockStorage.SearchFirst(hashes[i])
		if err != nil {
			log.Trace("missing miniblock",
				"error", err.Error(),
				"hash", hashes[i])
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlock := &block.MiniBlock{}
		err = gbbRes.marshalizer.Unmarshal(miniBlock, buff)
		if err != nil {
			log.Debug("marshal error", "error", err.Error())
			gbbRes.miniBlockPool.Remove(hashes[i])
			err = gbbRes.miniBlockStorage.Remove(hashes[i])
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
	return gbbRes == nil
}
