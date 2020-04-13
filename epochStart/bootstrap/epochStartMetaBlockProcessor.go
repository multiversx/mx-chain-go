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
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const durationBetweenChecks = 200 * time.Millisecond
const durationBetweenReRequests = 1 * time.Second
const durationBetweenCheckingNumConnectedPeers = 500 * time.Millisecond
const minNumConnectedPeers = 6
const minNumOfPeersToConsiderBlockValid = 3

var _ process.InterceptorProcessor = (*epochStartMetaBlockProcessor)(nil)

type epochStartMetaBlockProcessor struct {
	messenger              Messenger
	requestHandler         RequestHandler
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
	messenger Messenger,
	handler RequestHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	consensusPercentage uint8,
) (*epochStartMetaBlockProcessor, error) {
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
	if !(consensusPercentage > 0 && consensusPercentage <= 100) {
		return nil, epochStart.ErrInvalidConsensusThreshold
	}

	processor := &epochStartMetaBlockProcessor{
		messenger:              messenger,
		requestHandler:         handler,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		mutReceivedMetaBlocks:  sync.RWMutex{},
		mapReceivedMetaBlocks:  make(map[string]*block.MetaBlock),
		mapMetaBlocksFromPeers: make(map[string][]p2p.PeerID),
		chanConsensusReached:   make(chan bool, 1),
	}

	processor.waitForEnoughNumConnectedPeers(messenger)
	percentage := float64(consensusPercentage) / 100.0
	peerCountTarget := int(percentage * float64(len(messenger.ConnectedPeers())))
	processor.peerCountTarget = peerCountTarget

	log.Debug("consensus percentage for epoch start meta block ", "value (%)", consensusPercentage, "peerCountTarget", peerCountTarget)
	return processor, nil
}

// Validate will return nil as there is no need for validation
func (e *epochStartMetaBlockProcessor) Validate(_ process.InterceptedData, _ p2p.PeerID) error {
	return nil
}

func (e *epochStartMetaBlockProcessor) waitForEnoughNumConnectedPeers(messenger Messenger) {
	for {
		numConnectedPeers := len(messenger.ConnectedPeers())
		if numConnectedPeers >= minNumConnectedPeers {
			break
		}

		log.Debug("epoch bootstrapper: not enough connected peers",
			"wanted", minNumConnectedPeers,
			"actual", numConnectedPeers)
		time.Sleep(durationBetweenCheckingNumConnectedPeers)
	}
}

// Save will handle the consensus mechanism for the fetched metablocks
// All errors are just logged because if this function returns an error, the processing is finished. This way, we ignore
// wrong received data and wait for relevant intercepted data
func (e *epochStartMetaBlockProcessor) Save(data process.InterceptedData, fromConnectedPeer p2p.PeerID) error {
	if check.IfNil(data) {
		log.Debug("epoch bootstrapper: nil intercepted data")
		return nil
	}

	log.Info("received header", "type", data.Type(), "hash", data.Hash())
	interceptedHdr, ok := data.(process.HdrValidatorHandler)
	if !ok {
		log.Warn("saving epoch start meta block error", "error", epochStart.ErrWrongTypeAssertion)
		return nil
	}

	metaBlock := interceptedHdr.HeaderHandler().(*block.MetaBlock)

	if !metaBlock.IsStartOfEpochBlock() {
		log.Warn("saving epoch start meta block error", "error", epochStart.ErrNotEpochStartBlock)
		return nil
	}

	mbHash, err := core.CalculateHash(e.marshalizer, e.hasher, metaBlock)
	if err != nil {
		log.Warn("saving epoch start meta block error", "error", err)
		return nil
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
// This is a blocking method which will end after the consensus for the meta block is obtained or the context is done
func (e *epochStartMetaBlockProcessor) GetEpochStartMetaBlock(ctx context.Context) (*block.MetaBlock, error) {
	originalIntra, originalCross, err := e.requestHandler.GetNumPeersToQuery(factory.MetachainBlocksTopic)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = e.requestHandler.SetNumPeersToQuery(factory.MetachainBlocksTopic, originalIntra, originalCross)
		if err != nil {
			log.Warn("epoch bootstrapper: error setting num of peers intra/cross for resolver",
				"resolver", factory.MetachainBlocksTopic,
				"error", err)
		}
	}()

	err = e.requestMetaBlock()
	if err != nil {
		return nil, err
	}

	chanRequests := time.After(durationBetweenReRequests)
	chanCheckMaps := time.After(durationBetweenChecks)
	for {
		select {
		case <-e.chanConsensusReached:
			return e.metaBlock, nil
		case <-ctx.Done():
			return e.getMostReceivedMetaBlock()
		case <-chanRequests:
			err = e.requestMetaBlock()
			if err != nil {
				return nil, err
			}
			chanRequests = time.After(durationBetweenReRequests)
		case <-chanCheckMaps:
			e.checkMaps()
			chanCheckMaps = time.After(durationBetweenChecks)
		}
	}
}

func (e *epochStartMetaBlockProcessor) getMostReceivedMetaBlock() (*block.MetaBlock, error) {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	const hashNotAvailable = "N/A"
	mostReceivedHash := hashNotAvailable
	maxLength := minNumOfPeersToConsiderBlockValid - 1
	for hash, entry := range e.mapMetaBlocksFromPeers {
		if len(entry) > maxLength {
			maxLength = len(entry)
			mostReceivedHash = hash
		}
	}

	if mostReceivedHash == hashNotAvailable {
		return nil, epochStart.ErrTimeoutWaitingForMetaBlock
	}

	return e.mapReceivedMetaBlocks[mostReceivedHash], nil
}

func (e *epochStartMetaBlockProcessor) requestMetaBlock() error {
	numConnectedPeers := len(e.messenger.ConnectedPeers())
	err := e.requestHandler.SetNumPeersToQuery(factory.MetachainBlocksTopic, numConnectedPeers, numConnectedPeers)
	if err != nil {
		return err
	}

	unknownEpoch := uint32(math.MaxUint32)
	e.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)

	return nil
}

func (e *epochStartMetaBlockProcessor) checkMaps() {
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
