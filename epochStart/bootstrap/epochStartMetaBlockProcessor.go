package bootstrap

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

const timeToWaitBeforeCheckingReceivedHeaders = 200 * time.Millisecond
const durationBetweenReRequests = 1 * time.Second

var _ process.InterceptorProcessor = (*epochStartMetaBlockProcessor)(nil)

type epochStartMetaBlockProcessor struct {
	messenger              p2p.Messenger
	requestHandler         process.RequestHandler
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	mutReceivedMetaBlocks  sync.RWMutex
	mapReceivedMetaBlocks  map[string]*block.MetaBlock
	mapMetaBlocksFromPeers map[string][]p2p.PeerID
	chanConsensusReached   chan bool
	metaBlock              *block.MetaBlock
	peerCountTarget        int
}

// NewEpochStartMetaBlockProcessor will return a interceptor processor for epoch start meta block
func NewEpochStartMetaBlockProcessor(
	messenger p2p.Messenger,
	handler process.RequestHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	consensusPercentage float64,
) (process.InterceptorProcessor, error) {
	if check.IfNil(messenger) {
		return nil, epochStart.ErrNilMessenger
	}
	if check.IfNil(handler) {
		return nil, epochStart.ErrNilRequestHandler
	}
	if check.IfNil(marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if !(consensusPercentage > 0.0 && consensusPercentage < 1.0) {
		return nil, epochStart.ErrNilInvalidConsensusThreshold
	}

	peerCountTarget := int(consensusPercentage * float64(len(messenger.ConnectedPeers())))
	return &epochStartMetaBlockProcessor{
		messenger:              messenger,
		requestHandler:         handler,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMetaBlocks:  sync.RWMutex{},
		mapReceivedMetaBlocks:  make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers: make(map[string][]p2p.PeerID),
		peerCountTarget:        peerCountTarget,
		chanConsensusReached:   make(chan bool, 1),
	}, nil
}

// Validate will return nil as there is no need for validation
func (e *epochStartMetaBlockProcessor) Validate(_ process.InterceptedData, _ p2p.PeerID) error {
	return nil
}

// Save will handle the consensus mechanism for the fetched metablocks
func (e *epochStartMetaBlockProcessor) Save(data process.InterceptedData, fromConnectedPeer p2p.PeerID) error {
	log.Info("received header", "type", data.Type(), "hash", data.Hash())
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	metaBlock := interceptedHdr.HeaderHandler().(*block.MetaBlock)

	if !metaBlock.IsStartOfEpochBlock() {
		return epochStart.ErrNotEpochStartBlock
	}

	mbHash, err := core.CalculateHash(e.marshalizer, e.hasher, metaBlock)
	if err != nil {
		return err
	}

	e.mutReceivedMetaBlocks.Lock()
	e.mapReceivedMetaBlocks[string(mbHash)] = metaBlock
	e.addToPeerList(string(mbHash), fromConnectedPeer)
	e.mutReceivedMetaBlocks.Unlock()

	return nil
}

// this func should be called under mutex protection
func (e *epochStartMetaBlockProcessor) addToPeerList(hash string, peer p2p.PeerID) {
	peersListForHash := e.mapMetaBlocksFromPeers[hash]
	for _, pid := range peersListForHash {
		if pid == peer {
			return
		}
	}
	e.mapMetaBlocksFromPeers[hash] = append(e.mapMetaBlocksFromPeers[hash], peer)
}

// GetEpochStartMetaBlock will return the metablock after it is confirmed or an error if the number of tries was exceeded
func (e *epochStartMetaBlockProcessor) GetEpochStartMetaBlock(ctx context.Context) (*block.MetaBlock, error) {
	unknownEpoch := uint32(math.MaxUint32)
	e.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)

	for {
		select {
		case <-e.chanConsensusReached:
			return e.metaBlock, nil
		case <-ctx.Done():
			return nil, epochStart.ErrTimeoutWaitingForMetaBlock
		case <-time.After(durationBetweenReRequests):
			e.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)
		default:
			e.checkMaps()
		}
	}
}

func (e *epochStartMetaBlockProcessor) checkMaps() {
	time.Sleep(timeToWaitBeforeCheckingReceivedHeaders)
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()
	for hash, peersList := range e.mapMetaBlocksFromPeers {
		log.Debug("metablock from peers", "num peers", len(peersList), "target", e.peerCountTarget, "hash", []byte(hash))
		found := e.processEntry(peersList, hash)
		if found {
			break
		}
	}
}

func (e *epochStartMetaBlockProcessor) processEntry(
	peersList []p2p.PeerID,
	hash string,
) bool {
	if len(peersList) >= e.peerCountTarget {
		log.Info("got consensus for epoch start metablock", "len", len(peersList))
		e.metaBlock = e.mapReceivedMetaBlocks[hash]
		e.chanConsensusReached <- true
		return true
	}

	return false
}

// SignalEndOfProcessing won't do anything
func (e *epochStartMetaBlockProcessor) SignalEndOfProcessing(_ []process.InterceptedData) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartMetaBlockProcessor) IsInterfaceNil() bool {
	return e == nil
}
