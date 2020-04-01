package shardchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// waitTime defines the time in seconds to wait after a request has been done
const waitTime = 5 * time.Second

// ArgValidatorInfoProcessor holds all dependencies required to create a validatorInfoProcessor
type ArgValidatorInfoProcessor struct {
	MiniBlocksPool               storage.Cacher
	Marshalizer                  marshal.Marshalizer
	ValidatorStatisticsProcessor epochStart.ValidatorStatisticsProcessorHandler
	Requesthandler               epochStart.RequestHandler
	Hasher                       hashing.Hasher
}

// ValidatorInfoProcessor implements validator info processing for miniblocks of type peerMiniblock
type ValidatorInfoProcessor struct {
	miniBlocksPool               storage.Cacher
	marshalizer                  marshal.Marshalizer
	hasher                       hashing.Hasher
	validatorStatisticsProcessor epochStart.ValidatorStatisticsProcessorHandler
	requestHandler               epochStart.RequestHandler

	mapAllMisingPeerMiniblocks  map[string]*block.MiniBlock
	headerHash                  []byte
	metaHeader                  data.HeaderHandler
	chRcvAllMiniblocks          chan struct{}
	mutAllMissingPeerMiniblocks sync.RWMutex
	numMissingPeerMiniblocks    uint32
}

// NewValidatorInfoProcessor creates a new ValidatorInfoProcessor object
func NewValidatorInfoProcessor(arguments ArgValidatorInfoProcessor) (*ValidatorInfoProcessor, error) {
	if check.IfNil(arguments.ValidatorStatisticsProcessor) {
		return nil, epochStart.ErrNilValidatorStatistics
	}
	if check.IfNil(arguments.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(arguments.MiniBlocksPool) {
		return nil, epochStart.ErrNilMiniBlockPool
	}
	if check.IfNil(arguments.Requesthandler) {
		return nil, epochStart.ErrNilRequestHandler
	}

	vip := &ValidatorInfoProcessor{
		miniBlocksPool:               arguments.MiniBlocksPool,
		marshalizer:                  arguments.Marshalizer,
		validatorStatisticsProcessor: arguments.ValidatorStatisticsProcessor,
		requestHandler:               arguments.Requesthandler,
		hasher:                       arguments.Hasher,
	}

	//TODO: change the registerHandler for the miniblockPool to call
	//directly with hash and value - like func (sp *shardProcessor) receivedMetaBlock
	vip.miniBlocksPool.RegisterHandler(vip.receivedMiniBlock)

	return vip, nil
}

func (vip *ValidatorInfoProcessor) init(metaBlock *block.MetaBlock, metablockHash []byte) {
	vip.mutAllMissingPeerMiniblocks.Lock()
	vip.metaHeader = metaBlock
	vip.mapAllMisingPeerMiniblocks = make(map[string]*block.MiniBlock)
	vip.headerHash = metablockHash
	vip.chRcvAllMiniblocks = make(chan struct{})
	vip.mutAllMissingPeerMiniblocks.Unlock()
}

// ProcessMetaBlock processes an epochstart block asyncrhonous, processing the PeerMiniblocks
func (vip *ValidatorInfoProcessor) ProcessMetaBlock(metaBlock *block.MetaBlock, metablockHash []byte) (map[string]*block.MiniBlock, error) {
	vip.init(metaBlock, metablockHash)

	vip.computeMissingPeerBlocks(metaBlock)

	allMissingPeerMiniblocks, err := vip.retrieveMissingBlocks()
	if err != nil {
		return allMissingPeerMiniblocks, err
	}

	err = vip.processAllPeerMiniBlocks(metaBlock)
	if err != nil {
		return allMissingPeerMiniblocks, err
	}

	return allMissingPeerMiniblocks, nil
}

func (vip *ValidatorInfoProcessor) receivedMiniBlock(key []byte) {
	mb, ok := vip.miniBlocksPool.Get(key)
	if !ok {
		return
	}

	peerMb, ok := mb.(*block.MiniBlock)
	if !ok || peerMb.Type != block.PeerBlock {
		return
	}

	log.Trace(fmt.Sprintf("received miniblock of type %s", peerMb.Type))

	vip.mutAllMissingPeerMiniblocks.Lock()
	havingPeerMb, ok := vip.mapAllMisingPeerMiniblocks[string(key)]
	if !ok || havingPeerMb != nil {
		vip.mutAllMissingPeerMiniblocks.Unlock()
		return
	}

	vip.mapAllMisingPeerMiniblocks[string(key)] = peerMb
	vip.numMissingPeerMiniblocks--
	numMissingPeerMiniblocks := vip.numMissingPeerMiniblocks
	vip.mutAllMissingPeerMiniblocks.Unlock()

	if numMissingPeerMiniblocks == 0 {
		vip.chRcvAllMiniblocks <- struct{}{}
	}
}

func (vip *ValidatorInfoProcessor) processAllPeerMiniBlocks(metaBlock *block.MetaBlock) error {
	for _, peerMiniBlock := range metaBlock.MiniBlockHeaders {
		if peerMiniBlock.Type != block.PeerBlock {
			continue
		}

		mb := vip.mapAllMisingPeerMiniblocks[string(peerMiniBlock.Hash)]
		for _, txHash := range mb.TxHashes {
			vid := &state.ShardValidatorInfo{}
			err := vip.marshalizer.Unmarshal(vid, txHash)
			if err != nil {
				return err
			}

			err = vip.validatorStatisticsProcessor.Process(vid)
			if err != nil {
				return err
			}
		}
	}

	_, err := vip.validatorStatisticsProcessor.Commit()

	return err
}

func (vip *ValidatorInfoProcessor) computeMissingPeerBlocks(metaBlock *block.MetaBlock) {
	numMissingPeerMiniblocks := uint32(0)
	vip.mutAllMissingPeerMiniblocks.Lock()

	for _, mb := range metaBlock.MiniBlockHeaders {
		if mb.Type != block.PeerBlock {
			continue
		}

		vip.mapAllMisingPeerMiniblocks[string(mb.Hash)] = nil

		mbObjectFound, ok := vip.miniBlocksPool.Peek(mb.Hash)
		if !ok {
			numMissingPeerMiniblocks++
			continue
		}

		mbFound, ok := mbObjectFound.(*block.MiniBlock)
		if !ok {
			numMissingPeerMiniblocks++
			continue
		}

		vip.mapAllMisingPeerMiniblocks[string(mb.Hash)] = mbFound
	}

	vip.numMissingPeerMiniblocks = numMissingPeerMiniblocks
	vip.mutAllMissingPeerMiniblocks.Unlock()
}

func (vip *ValidatorInfoProcessor) retrieveMissingBlocks() (map[string]*block.MiniBlock, error) {
	vip.mutAllMissingPeerMiniblocks.Lock()
	missingMiniblocks := make([][]byte, 0)
	for mbHash, mb := range vip.mapAllMisingPeerMiniblocks {
		if mb == nil {
			missingMiniblocks = append(missingMiniblocks, []byte(mbHash))
		}
	}
	vip.numMissingPeerMiniblocks = uint32(len(missingMiniblocks))
	vip.mutAllMissingPeerMiniblocks.Unlock()

	if len(missingMiniblocks) == 0 {
		return nil, nil
	}

	go vip.requestHandler.RequestMiniBlocks(core.MetachainShardId, missingMiniblocks)

	select {
	case <-vip.chRcvAllMiniblocks:
		return nil, nil
	case <-time.After(waitTime):
		return vip.getCopyOfAllMissingPeerMiniblocks(), process.ErrTimeIsOut
	}
}

func (vip *ValidatorInfoProcessor) getCopyOfAllMissingPeerMiniblocks() map[string]*block.MiniBlock {
	vip.mutAllMissingPeerMiniblocks.RLock()
	defer vip.mutAllMissingPeerMiniblocks.RUnlock()

	mapMiniBlocks := make(map[string]*block.MiniBlock, len(vip.mapAllMisingPeerMiniblocks))
	for hash, mb := range vip.mapAllMisingPeerMiniblocks {
		mapMiniBlocks[hash] = mb
	}

	return mapMiniBlocks
}

// IsInterfaceNil returns true if underlying object is nil
func (vip *ValidatorInfoProcessor) IsInterfaceNil() bool {
	return vip == nil
}
