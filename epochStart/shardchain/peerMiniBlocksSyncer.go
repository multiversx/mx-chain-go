package shardchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.ValidatorInfoSyncer = (*peerMiniBlockSyncer)(nil)

// waitTime defines the time in seconds to wait after a request has been done
const waitTime = 5 * time.Second

// ArgPeerMiniBlockSyncer holds all dependencies required to create a peerMiniBlockSyncer
type ArgPeerMiniBlockSyncer struct {
	MiniBlocksPool storage.Cacher
	Requesthandler epochStart.RequestHandler
}

// peerMiniBlockSyncer implements validator info processing for miniblocks of type peerMiniblock
type peerMiniBlockSyncer struct {
	miniBlocksPool storage.Cacher
	requestHandler epochStart.RequestHandler

	mapAllPeerMiniblocks     map[string]*block.MiniBlock
	chRcvAllMiniblocks       chan struct{}
	mutMiniBlocksForBlock    sync.RWMutex
	numMissingPeerMiniblocks uint32
}

// NewPeerMiniBlockSyncer creates a new peerMiniBlockSyncer object
func NewPeerMiniBlockSyncer(arguments ArgPeerMiniBlockSyncer) (*peerMiniBlockSyncer, error) {
	if check.IfNil(arguments.MiniBlocksPool) {
		return nil, epochStart.ErrNilMiniBlockPool
	}
	if check.IfNil(arguments.Requesthandler) {
		return nil, epochStart.ErrNilRequestHandler
	}

	p := &peerMiniBlockSyncer{
		miniBlocksPool: arguments.MiniBlocksPool,
		requestHandler: arguments.Requesthandler,
	}

	//TODO: change the registerHandler for the miniblockPool to call
	//directly with hash and value - like func (sp *shardProcessor) receivedMetaBlock
	p.miniBlocksPool.RegisterHandler(p.receivedMiniBlock, core.UniqueIdentifier())

	return p, nil
}

func (p *peerMiniBlockSyncer) init() {
	p.mutMiniBlocksForBlock.Lock()
	p.mapAllPeerMiniblocks = make(map[string]*block.MiniBlock)
	p.chRcvAllMiniblocks = make(chan struct{})
	p.mutMiniBlocksForBlock.Unlock()
}

// SyncMiniBlocks processes an epochstart block asyncrhonous, processing the PeerMiniblocks
func (p *peerMiniBlockSyncer) SyncMiniBlocks(metaBlock data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
	if check.IfNil(metaBlock) {
		return nil, nil, epochStart.ErrNilMetaBlock
	}

	p.init()

	p.computeMissingPeerBlocks(metaBlock)

	allMissingPeerMiniblocksHashes, err := p.retrieveMissingBlocks()
	if err != nil {
		return allMissingPeerMiniblocksHashes, nil, err
	}

	peerBlockBody := p.getAllPeerMiniBlocks(metaBlock)

	return nil, peerBlockBody, nil
}

func (p *peerMiniBlockSyncer) receivedMiniBlock(key []byte, val interface{}) {
	peerMb, ok := val.(*block.MiniBlock)
	if !ok || peerMb.Type != block.PeerBlock {
		return
	}

	log.Trace(fmt.Sprintf("received miniblock of type %s", peerMb.Type))

	p.mutMiniBlocksForBlock.Lock()
	havingPeerMb, ok := p.mapAllPeerMiniblocks[string(key)]
	if !ok || havingPeerMb != nil {
		p.mutMiniBlocksForBlock.Unlock()
		return
	}

	p.mapAllPeerMiniblocks[string(key)] = peerMb
	p.numMissingPeerMiniblocks--
	numMissingPeerMiniblocks := p.numMissingPeerMiniblocks
	p.mutMiniBlocksForBlock.Unlock()

	if numMissingPeerMiniblocks == 0 {
		p.chRcvAllMiniblocks <- struct{}{}
	}
}

func (p *peerMiniBlockSyncer) getAllPeerMiniBlocks(metaBlock data.HeaderHandler) data.BodyHandler {
	p.mutMiniBlocksForBlock.Lock()
	defer p.mutMiniBlocksForBlock.Unlock()

	peerBlockBody := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}
	for _, peerMiniBlock := range metaBlock.GetMiniBlockHeaderHandlers() {
		if peerMiniBlock.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		mb := p.mapAllPeerMiniblocks[string(peerMiniBlock.GetHash())]
		peerBlockBody.MiniBlocks = append(peerBlockBody.MiniBlocks, mb)
	}

	return peerBlockBody
}

func (p *peerMiniBlockSyncer) computeMissingPeerBlocks(metaBlock data.HeaderHandler) {
	numMissingPeerMiniblocks := uint32(0)
	p.mutMiniBlocksForBlock.Lock()

	for _, mb := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mb.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		p.mapAllPeerMiniblocks[string(mb.GetHash())] = nil

		mbObjectFound, ok := p.miniBlocksPool.Peek(mb.GetHash())
		if !ok {
			numMissingPeerMiniblocks++
			continue
		}

		mbFound, ok := mbObjectFound.(*block.MiniBlock)
		if !ok {
			numMissingPeerMiniblocks++
			continue
		}

		p.mapAllPeerMiniblocks[string(mb.GetHash())] = mbFound
	}

	p.numMissingPeerMiniblocks = numMissingPeerMiniblocks
	p.mutMiniBlocksForBlock.Unlock()
}

func (p *peerMiniBlockSyncer) retrieveMissingBlocks() ([][]byte, error) {
	p.mutMiniBlocksForBlock.Lock()
	missingMiniblocks := make([][]byte, 0)
	for mbHash, mb := range p.mapAllPeerMiniblocks {
		if mb == nil {
			missingMiniblocks = append(missingMiniblocks, []byte(mbHash))
		}
	}
	p.numMissingPeerMiniblocks = uint32(len(missingMiniblocks))
	p.mutMiniBlocksForBlock.Unlock()

	if len(missingMiniblocks) == 0 {
		return nil, nil
	}

	go p.requestHandler.RequestMiniBlocks(core.MetachainShardId, missingMiniblocks)

	select {
	case <-p.chRcvAllMiniblocks:
		return nil, nil
	case <-time.After(waitTime):
		return p.getAllMissingPeerMiniblocksHashes(), process.ErrTimeIsOut
	}
}

func (p *peerMiniBlockSyncer) getAllMissingPeerMiniblocksHashes() [][]byte {
	p.mutMiniBlocksForBlock.RLock()
	defer p.mutMiniBlocksForBlock.RUnlock()

	missingPeerMiniBlocksHashes := make([][]byte, 0)
	for hash, mb := range p.mapAllPeerMiniblocks {
		if mb == nil {
			missingPeerMiniBlocksHashes = append(missingPeerMiniBlocksHashes, []byte(hash))
		}
	}

	return missingPeerMiniBlocksHashes
}

// IsInterfaceNil returns true if underlying object is nil
func (p *peerMiniBlockSyncer) IsInterfaceNil() bool {
	return p == nil
}
