package shardchain

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ process.ValidatorInfoSyncer = (*peerMiniBlockSyncer)(nil)

// waitTime defines the time in seconds to wait after a request has been done
const waitTime = 5 * time.Second

// ArgPeerMiniBlockSyncer holds all dependencies required to create a peerMiniBlockSyncer
type ArgPeerMiniBlockSyncer struct {
	MiniBlocksPool     storage.Cacher
	ValidatorsInfoPool dataRetriever.ShardedDataCacherNotifier
	RequestHandler     epochStart.RequestHandler
}

// peerMiniBlockSyncer implements validator info processing for mini blocks of type PeerMiniBlock
type peerMiniBlockSyncer struct {
	miniBlocksPool     storage.Cacher
	validatorsInfoPool dataRetriever.ShardedDataCacherNotifier
	requestHandler     epochStart.RequestHandler

	mapAllPeerMiniBlocks      map[string]*block.MiniBlock
	mapAllValidatorsInfo      map[string]*state.ShardValidatorInfo
	chRcvAllMiniBlocks        chan struct{}
	chRcvAllValidatorsInfo    chan struct{}
	mutMiniBlocksForBlock     sync.RWMutex
	mutValidatorsInfoForBlock sync.RWMutex
	numMissingPeerMiniBlocks  uint32
	numMissingValidatorsInfo  uint32
}

// NewPeerMiniBlockSyncer creates a new peerMiniBlockSyncer object
func NewPeerMiniBlockSyncer(arguments ArgPeerMiniBlockSyncer) (*peerMiniBlockSyncer, error) {
	if check.IfNil(arguments.MiniBlocksPool) {
		return nil, epochStart.ErrNilMiniBlockPool
	}
	if check.IfNil(arguments.ValidatorsInfoPool) {
		return nil, epochStart.ErrNilValidatorsInfoPool
	}
	if check.IfNil(arguments.RequestHandler) {
		return nil, epochStart.ErrNilRequestHandler
	}

	p := &peerMiniBlockSyncer{
		miniBlocksPool:     arguments.MiniBlocksPool,
		validatorsInfoPool: arguments.ValidatorsInfoPool,
		requestHandler:     arguments.RequestHandler,
	}

	//TODO: change the registerHandler for the miniblockPool to call
	//directly with hash and value - like func (sp *shardProcessor) receivedMetaBlock
	p.miniBlocksPool.RegisterHandler(p.receivedMiniBlock, core.UniqueIdentifier())
	p.validatorsInfoPool.RegisterOnAdded(p.receivedValidatorInfo)

	return p, nil
}

func (p *peerMiniBlockSyncer) initMiniBlocks() {
	p.mutMiniBlocksForBlock.Lock()
	p.mapAllPeerMiniBlocks = make(map[string]*block.MiniBlock)
	p.chRcvAllMiniBlocks = make(chan struct{})
	p.mutMiniBlocksForBlock.Unlock()
}

func (p *peerMiniBlockSyncer) initValidatorsInfo() {
	p.mutValidatorsInfoForBlock.Lock()
	p.mapAllValidatorsInfo = make(map[string]*state.ShardValidatorInfo)
	p.chRcvAllValidatorsInfo = make(chan struct{})
	p.mutValidatorsInfoForBlock.Unlock()
}

// SyncMiniBlocks synchronizes peers mini blocks from an epoch start meta block
func (p *peerMiniBlockSyncer) SyncMiniBlocks(headerHandler data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
	if check.IfNil(headerHandler) {
		return nil, nil, epochStart.ErrNilMetaBlock
	}

	p.initMiniBlocks()

	p.computeMissingPeerBlocks(headerHandler)

	allMissingPeerMiniBlocksHashes, err := p.retrieveMissingMiniBlocks()
	if err != nil {
		return allMissingPeerMiniBlocksHashes, nil, err
	}

	peerBlockBody := p.getAllPeerMiniBlocks(headerHandler)

	return nil, peerBlockBody, nil
}

// SyncValidatorsInfo synchronizes validators info from a block body of an epoch start meta block
func (p *peerMiniBlockSyncer) SyncValidatorsInfo(bodyHandler data.BodyHandler) ([][]byte, map[string]*state.ShardValidatorInfo, error) {
	if check.IfNil(bodyHandler) {
		return nil, nil, epochStart.ErrNilBlockBody
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, epochStart.ErrWrongTypeAssertion
	}

	p.initValidatorsInfo()

	p.computeMissingValidatorsInfo(body)

	allMissingValidatorsInfoHashes, err := p.retrieveMissingValidatorsInfo()
	if err != nil {
		return allMissingValidatorsInfoHashes, nil, err
	}

	validatorsInfo := p.getAllValidatorsInfo(body)

	return nil, validatorsInfo, nil
}

func (p *peerMiniBlockSyncer) receivedMiniBlock(key []byte, val interface{}) {
	peerMiniBlock, ok := val.(*block.MiniBlock)
	if !ok {
		log.Error("receivedMiniBlock", "key", key, "error", epochStart.ErrWrongTypeAssertion)
		return
	}

	if peerMiniBlock.Type != block.PeerBlock {
		return
	}

	log.Debug("peerMiniBlockSyncer.receivedMiniBlock", "mb type", peerMiniBlock.Type)

	p.mutMiniBlocksForBlock.Lock()
	havingPeerMb, ok := p.mapAllPeerMiniBlocks[string(key)]
	if !ok || havingPeerMb != nil {
		p.mutMiniBlocksForBlock.Unlock()
		return
	}

	p.mapAllPeerMiniBlocks[string(key)] = peerMiniBlock
	p.numMissingPeerMiniBlocks--
	numMissingPeerMiniBlocks := p.numMissingPeerMiniBlocks
	p.mutMiniBlocksForBlock.Unlock()

	log.Debug("peerMiniBlockSyncer.receivedMiniBlock", "mb hash", key, "num missing peer mini blocks", numMissingPeerMiniBlocks)

	if numMissingPeerMiniBlocks == 0 {
		p.chRcvAllMiniBlocks <- struct{}{}
	}
}

func (p *peerMiniBlockSyncer) receivedValidatorInfo(key []byte, val interface{}) {
	validatorInfo, ok := val.(*state.ShardValidatorInfo)
	if !ok {
		log.Error("receivedValidatorInfo", "key", key, "error", epochStart.ErrWrongTypeAssertion)
		return
	}

	log.Debug("peerMiniBlockSyncer.receivedValidatorInfo", "pk", validatorInfo.PublicKey)

	p.mutValidatorsInfoForBlock.Lock()
	havingValidatorInfo, ok := p.mapAllValidatorsInfo[string(key)]
	if !ok || havingValidatorInfo != nil {
		p.mutValidatorsInfoForBlock.Unlock()
		return
	}

	p.mapAllValidatorsInfo[string(key)] = validatorInfo
	p.numMissingValidatorsInfo--
	numMissingValidatorsInfo := p.numMissingValidatorsInfo
	p.mutValidatorsInfoForBlock.Unlock()

	log.Debug("peerMiniBlockSyncer.receivedValidatorInfo", "tx hash", key, "num missing validators info", numMissingValidatorsInfo)

	if numMissingValidatorsInfo == 0 {
		p.chRcvAllValidatorsInfo <- struct{}{}
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

		mb := p.mapAllPeerMiniBlocks[string(peerMiniBlock.GetHash())]
		peerBlockBody.MiniBlocks = append(peerBlockBody.MiniBlocks, mb)
	}

	return peerBlockBody
}

func (p *peerMiniBlockSyncer) getAllValidatorsInfo(body *block.Body) map[string]*state.ShardValidatorInfo {
	p.mutValidatorsInfoForBlock.Lock()
	defer p.mutValidatorsInfoForBlock.Unlock()

	validatorsInfo := make(map[string]*state.ShardValidatorInfo)
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.PeerBlock {
			continue
		}

		for _, txHash := range mb.TxHashes {
			validatorInfo := p.mapAllValidatorsInfo[string(txHash)]
			validatorsInfo[string(txHash)] = validatorInfo
		}
	}

	return validatorsInfo
}

func (p *peerMiniBlockSyncer) computeMissingPeerBlocks(metaBlock data.HeaderHandler) {
	p.mutMiniBlocksForBlock.Lock()
	defer p.mutMiniBlocksForBlock.Unlock()

	numMissingPeerMiniBlocks := uint32(0)
	for _, mb := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mb.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		p.mapAllPeerMiniBlocks[string(mb.GetHash())] = nil

		mbObjectFound, ok := p.miniBlocksPool.Peek(mb.GetHash())
		if !ok {
			numMissingPeerMiniBlocks++
			continue
		}

		mbFound, ok := mbObjectFound.(*block.MiniBlock)
		if !ok {
			numMissingPeerMiniBlocks++
			continue
		}

		p.mapAllPeerMiniBlocks[string(mb.GetHash())] = mbFound
	}

	p.numMissingPeerMiniBlocks = numMissingPeerMiniBlocks
}

func (p *peerMiniBlockSyncer) computeMissingValidatorsInfo(body *block.Body) {
	p.mutValidatorsInfoForBlock.Lock()
	defer p.mutValidatorsInfoForBlock.Unlock()

	numMissingValidatorsInfo := uint32(0)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		numMissingValidatorsInfo += p.setMissingValidatorsInfo(miniBlock)
	}

	p.numMissingValidatorsInfo = numMissingValidatorsInfo
}

func (p *peerMiniBlockSyncer) setMissingValidatorsInfo(miniBlock *block.MiniBlock) uint32 {
	numMissingValidatorsInfo := uint32(0)
	for _, txHash := range miniBlock.TxHashes {
		p.mapAllValidatorsInfo[string(txHash)] = nil

		val, ok := p.validatorsInfoPool.SearchFirstData(txHash)
		if !ok {
			numMissingValidatorsInfo++
			continue
		}

		validatorInfo, ok := val.(*state.ShardValidatorInfo)
		if !ok {
			numMissingValidatorsInfo++
			continue
		}

		p.mapAllValidatorsInfo[string(txHash)] = validatorInfo
	}

	return numMissingValidatorsInfo
}

func (p *peerMiniBlockSyncer) retrieveMissingMiniBlocks() ([][]byte, error) {
	p.mutMiniBlocksForBlock.Lock()
	missingMiniBlocks := make([][]byte, 0)
	for mbHash, mb := range p.mapAllPeerMiniBlocks {
		if mb == nil {
			missingMiniBlocks = append(missingMiniBlocks, []byte(mbHash))
		}
	}
	p.numMissingPeerMiniBlocks = uint32(len(missingMiniBlocks))
	p.mutMiniBlocksForBlock.Unlock()

	if len(missingMiniBlocks) == 0 {
		return nil, nil
	}

	go p.requestHandler.RequestMiniBlocks(core.MetachainShardId, missingMiniBlocks)

	select {
	case <-p.chRcvAllMiniBlocks:
		return nil, nil
	case <-time.After(waitTime):
		return p.getAllMissingPeerMiniBlocksHashes(), process.ErrTimeIsOut
	}
}

func (p *peerMiniBlockSyncer) retrieveMissingValidatorsInfo() ([][]byte, error) {
	p.mutValidatorsInfoForBlock.Lock()
	missingValidatorsInfo := make([][]byte, 0)
	for validatorInfoHash, validatorInfo := range p.mapAllValidatorsInfo {
		if validatorInfo == nil {
			missingValidatorsInfo = append(missingValidatorsInfo, []byte(validatorInfoHash))
		}
	}
	p.numMissingValidatorsInfo = uint32(len(missingValidatorsInfo))
	p.mutValidatorsInfoForBlock.Unlock()

	if len(missingValidatorsInfo) == 0 {
		return nil, nil
	}

	go p.requestHandler.RequestValidatorsInfo(missingValidatorsInfo)

	select {
	case <-p.chRcvAllValidatorsInfo:
		return nil, nil
	case <-time.After(waitTime):
		return p.getAllMissingValidatorsInfoHashes(), process.ErrTimeIsOut
	}
}

func (p *peerMiniBlockSyncer) getAllMissingPeerMiniBlocksHashes() [][]byte {
	p.mutMiniBlocksForBlock.RLock()
	defer p.mutMiniBlocksForBlock.RUnlock()

	missingPeerMiniBlocksHashes := make([][]byte, 0)
	for hash, mb := range p.mapAllPeerMiniBlocks {
		if mb == nil {
			missingPeerMiniBlocksHashes = append(missingPeerMiniBlocksHashes, []byte(hash))
		}
	}

	return missingPeerMiniBlocksHashes
}

func (p *peerMiniBlockSyncer) getAllMissingValidatorsInfoHashes() [][]byte {
	p.mutValidatorsInfoForBlock.RLock()
	defer p.mutValidatorsInfoForBlock.RUnlock()

	missingValidatorsInfoHashes := make([][]byte, 0)
	for validatorInfoHash, validatorInfo := range p.mapAllValidatorsInfo {
		if validatorInfo == nil {
			missingValidatorsInfoHashes = append(missingValidatorsInfoHashes, []byte(validatorInfoHash))
		}
	}

	return missingValidatorsInfoHashes
}

// IsInterfaceNil returns true if underlying object is nil
func (p *peerMiniBlockSyncer) IsInterfaceNil() bool {
	return p == nil
}
