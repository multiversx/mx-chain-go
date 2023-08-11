package shardchain

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("epochStart/shardchain")

var _ dataRetriever.EpochHandler = (*trigger)(nil)
var _ epochStart.TriggerHandler = (*trigger)(nil)
var _ process.EpochStartTriggerHandler = (*trigger)(nil)
var _ process.EpochBootstrapper = (*trigger)(nil)
var _ closing.Closer = (*trigger)(nil)

// sleepTime defines the time in milliseconds between each iteration made in requestMissingMiniBlocks method
const sleepTime = 1 * time.Second

// ArgsShardEpochStartTrigger struct { defines the arguments needed for new start of epoch trigger
type ArgsShardEpochStartTrigger struct {
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher

	HeaderValidator epochStart.HeaderValidator
	Uint64Converter typeConverters.Uint64ByteSliceConverter

	DataPool             dataRetriever.PoolsHolder
	Storage              dataRetriever.StorageService
	RequestHandler       epochStart.RequestHandler
	EpochStartNotifier   epochStart.Notifier
	PeerMiniBlocksSyncer process.ValidatorInfoSyncer
	RoundHandler         process.RoundHandler
	AppStatusHandler     core.AppStatusHandler
	EnableEpochsHandler  common.EnableEpochsHandler

	Epoch    uint32
	Validity uint64
	Finality uint64
}

type trigger struct {
	currentRoundIndex           int64
	epochStartRound             uint64
	epochMetaBlockHash          []byte
	triggerStateKey             []byte
	finality                    uint64
	validity                    uint64
	epochFinalityAttestingRound uint64
	epochStartShardHeader       data.HeaderHandler
	epochStartMeta              data.HeaderHandler

	mutTrigger         sync.RWMutex
	mapHashHdr         map[string]data.HeaderHandler
	mapNonceHashes     map[uint64][]string
	mapEpochStartHdrs  map[string]data.HeaderHandler
	mapFinalizedEpochs map[uint32]string

	headersPool                   dataRetriever.HeadersPool
	miniBlocksPool                storage.Cacher
	validatorInfoPool             dataRetriever.ShardedDataCacherNotifier
	currentEpochValidatorInfoPool epochStart.ValidatorInfoCacher
	shardHdrStorage               storage.Storer
	metaHdrStorage                storage.Storer
	triggerStorage                storage.Storer
	metaNonceHdrStorage           storage.Storer
	epochStartStaticStorage       storage.Storer

	uint64Converter typeConverters.Uint64ByteSliceConverter

	marshaller      marshal.Marshalizer
	hasher          hashing.Hasher
	headerValidator epochStart.HeaderValidator

	requestHandler     epochStart.RequestHandler
	epochStartNotifier epochStart.Notifier
	roundHandler       process.RoundHandler

	epoch                           uint32
	metaEpoch                       uint32
	newEpochHdrReceived             bool
	isEpochStart                    bool
	requestedFinalityAttestingBlock atomic.Flag

	peerMiniBlocksSyncer process.ValidatorInfoSyncer

	appStatusHandler    core.AppStatusHandler
	enableEpochsHandler common.EnableEpochsHandler

	mapMissingMiniBlocks     map[string]uint32
	mapMissingValidatorsInfo map[string]uint32
	mutMissingMiniBlocks     sync.RWMutex
	mutMissingValidatorsInfo sync.RWMutex
	cancelFunc               func()
}

type metaInfo struct {
	hdr  data.HeaderHandler
	hash string
}

type metaInfoSlice []*metaInfo

// Len will return the length of the metaInfoSlice
func (m metaInfoSlice) Len() int { return len(m) }

// Swap will interchange the objects on input indexes
func (m metaInfoSlice) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of headers should be by epoch, by nonce and by hash in ascending order
// this will ensure that in case of equality for epoch, the metaHdr with higher nonce will
// be processed last - that is  the correct one - as it finalizes the previous nonce
func (m metaInfoSlice) Less(i, j int) bool {
	if m[i].hdr.GetEpoch() == m[j].hdr.GetEpoch() {
		if m[i].hdr.GetNonce() == m[j].hdr.GetNonce() {
			return m[i].hash < m[j].hash
		}
		return m[i].hdr.GetNonce() < m[j].hdr.GetNonce()
	}
	return m[i].hdr.GetEpoch() < m[j].hdr.GetEpoch()
}

// NewEpochStartTrigger creates a trigger to signal start of epoch
func NewEpochStartTrigger(args *ArgsShardEpochStartTrigger) (*trigger, error) {
	if args == nil {
		return nil, epochStart.ErrNilArgsNewShardEpochStartTrigger
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, epochStart.ErrNilHeaderValidator
	}
	if check.IfNil(args.DataPool) {
		return nil, epochStart.ErrNilDataPoolsHolder
	}
	if check.IfNil(args.Storage) {
		return nil, epochStart.ErrNilStorageService
	}
	if check.IfNil(args.RequestHandler) {
		return nil, epochStart.ErrNilRequestHandler
	}
	if check.IfNil(args.DataPool.Headers()) {
		return nil, epochStart.ErrNilMetaBlocksPool
	}
	if check.IfNil(args.DataPool.MiniBlocks()) {
		return nil, epochStart.ErrNilMiniBlockPool
	}
	if check.IfNil(args.DataPool.ValidatorsInfo()) {
		return nil, epochStart.ErrNilValidatorsInfoPool
	}
	if check.IfNil(args.DataPool.CurrentEpochValidatorInfo()) {
		return nil, epochStart.ErrNilCurrentEpochValidatorsInfoPool
	}
	if check.IfNil(args.PeerMiniBlocksSyncer) {
		return nil, epochStart.ErrNilValidatorInfoProcessor
	}
	if check.IfNil(args.Uint64Converter) {
		return nil, epochStart.ErrNilUint64Converter
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.RoundHandler) {
		return nil, epochStart.ErrNilRoundHandler
	}
	if check.IfNil(args.AppStatusHandler) {
		return nil, epochStart.ErrNilStatusHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, epochStart.ErrNilEnableEpochsHandler
	}

	metaHdrStorage, err := args.Storage.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	triggerStorage, err := args.Storage.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return nil, err
	}

	metaHdrNoncesStorage, err := args.Storage.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	shardHdrStorage, err := args.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	epochStartStaticStorage, err := args.Storage.GetStorer(dataRetriever.EpochStartStaticUnit)
	if err != nil {
		return nil, err
	}

	trigggerStateKey := common.TriggerRegistryInitialKeyPrefix + fmt.Sprintf("%d", args.Epoch)

	t := &trigger{
		triggerStateKey:               []byte(trigggerStateKey),
		epoch:                         args.Epoch,
		metaEpoch:                     args.Epoch,
		currentRoundIndex:             0,
		epochStartRound:               0,
		epochFinalityAttestingRound:   0,
		isEpochStart:                  false,
		validity:                      args.Validity,
		finality:                      args.Finality,
		newEpochHdrReceived:           false,
		mutTrigger:                    sync.RWMutex{},
		mapHashHdr:                    make(map[string]data.HeaderHandler),
		mapNonceHashes:                make(map[uint64][]string),
		mapEpochStartHdrs:             make(map[string]data.HeaderHandler),
		mapFinalizedEpochs:            make(map[uint32]string),
		headersPool:                   args.DataPool.Headers(),
		miniBlocksPool:                args.DataPool.MiniBlocks(),
		validatorInfoPool:             args.DataPool.ValidatorsInfo(),
		currentEpochValidatorInfoPool: args.DataPool.CurrentEpochValidatorInfo(),
		metaHdrStorage:                metaHdrStorage,
		shardHdrStorage:               shardHdrStorage,
		triggerStorage:                triggerStorage,
		metaNonceHdrStorage:           metaHdrNoncesStorage,
		uint64Converter:               args.Uint64Converter,
		marshaller:                    args.Marshalizer,
		hasher:                        args.Hasher,
		headerValidator:               args.HeaderValidator,
		requestHandler:                args.RequestHandler,
		epochMetaBlockHash:            nil,
		epochStartNotifier:            args.EpochStartNotifier,
		epochStartMeta:                &block.MetaBlock{},
		epochStartShardHeader:         &block.Header{},
		peerMiniBlocksSyncer:          args.PeerMiniBlocksSyncer,
		appStatusHandler:              args.AppStatusHandler,
		roundHandler:                  args.RoundHandler,
		enableEpochsHandler:           args.EnableEpochsHandler,
		epochStartStaticStorage:       epochStartStaticStorage,
	}

	t.headersPool.RegisterHandler(t.receivedMetaBlock)

	err = t.saveState(t.triggerStateKey)
	if err != nil {
		return nil, err
	}

	t.mapMissingMiniBlocks = make(map[string]uint32)
	t.mapMissingValidatorsInfo = make(map[string]uint32)

	var ctx context.Context
	ctx, t.cancelFunc = context.WithCancel(context.Background())
	go t.requestMissingMiniBlocks(ctx)
	go t.requestMissingValidatorsInfo(ctx)

	return t, nil
}

func (t *trigger) clearMissingMiniBlocksMap(epoch uint32) {
	t.mutMissingMiniBlocks.Lock()
	defer t.mutMissingMiniBlocks.Unlock()

	for hash, epochOfMissingMb := range t.mapMissingMiniBlocks {
		if epochOfMissingMb <= epoch {
			delete(t.mapMissingMiniBlocks, hash)
		}
	}
}

func (t *trigger) clearMissingValidatorsInfoMap(epoch uint32) {
	t.mutMissingValidatorsInfo.Lock()
	defer t.mutMissingValidatorsInfo.Unlock()

	for hash, epochOfMissingValidatorInfo := range t.mapMissingValidatorsInfo {
		if epochOfMissingValidatorInfo <= epoch {
			delete(t.mapMissingValidatorsInfo, hash)
		}
	}
}

func (t *trigger) requestMissingMiniBlocks(ctx context.Context) {
	timer := time.NewTimer(sleepTime)
	defer timer.Stop()

	for {
		timer.Reset(sleepTime)

		select {
		case <-ctx.Done():
			log.Debug("requestMissingMiniBlocks: trigger's go routine is stopping...")
			return
		case <-timer.C:
		}

		t.mutMissingMiniBlocks.RLock()
		if len(t.mapMissingMiniBlocks) == 0 {
			t.mutMissingMiniBlocks.RUnlock()
			continue
		}

		missingMiniBlocks := make([][]byte, 0, len(t.mapMissingMiniBlocks))
		for hash, epoch := range t.mapMissingMiniBlocks {
			missingMiniBlocks = append(missingMiniBlocks, []byte(hash))
			log.Debug("trigger.requestMissingMiniBlocks", "epoch", epoch, "hash", []byte(hash))
		}
		t.mutMissingMiniBlocks.RUnlock()

		go t.requestHandler.RequestMiniBlocks(core.MetachainShardId, missingMiniBlocks)

		timer.Reset(waitTime)

		select {
		case <-ctx.Done():
			log.Debug("requestMissingMiniBlocks: trigger's go routine is stopping...")
			return
		case <-timer.C:
		}

		t.updateMissingMiniBlocks()
	}
}

func (t *trigger) requestMissingValidatorsInfo(ctx context.Context) {
	timer := time.NewTimer(sleepTime)
	defer timer.Stop()

	for {
		timer.Reset(sleepTime)

		select {
		case <-ctx.Done():
			log.Debug("requestMissingValidatorsInfo: trigger's go routine is stopping...")
			return
		case <-timer.C:
		}

		t.mutMissingValidatorsInfo.RLock()
		if len(t.mapMissingValidatorsInfo) == 0 {
			t.mutMissingValidatorsInfo.RUnlock()
			continue
		}

		missingValidatorsInfo := make([][]byte, 0, len(t.mapMissingValidatorsInfo))
		for hash, epoch := range t.mapMissingValidatorsInfo {
			missingValidatorsInfo = append(missingValidatorsInfo, []byte(hash))
			log.Debug("trigger.requestMissingValidatorsInfo", "epoch", epoch, "hash", []byte(hash))
		}
		t.mutMissingValidatorsInfo.RUnlock()

		go t.requestHandler.RequestValidatorsInfo(missingValidatorsInfo)

		timer.Reset(waitTime)

		select {
		case <-ctx.Done():
			log.Debug("requestMissingValidatorsInfo: trigger's go routine is stopping...")
			return
		case <-timer.C:
		}

		t.updateMissingValidatorsInfo()
	}
}

func (t *trigger) updateMissingMiniBlocks() {
	t.mutMissingMiniBlocks.Lock()
	for hash := range t.mapMissingMiniBlocks {
		if t.miniBlocksPool.Has([]byte(hash)) {
			delete(t.mapMissingMiniBlocks, hash)
		}
	}
	numMissingMiniBlocks := len(t.mapMissingMiniBlocks)
	t.mutMissingMiniBlocks.Unlock()

	if numMissingMiniBlocks == 0 {
		log.Debug("trigger.updateMissingMiniBlocks -> updateTriggerFromMeta")
		t.mutTrigger.Lock()
		t.updateTriggerFromMeta()
		t.mutTrigger.Unlock()
	}
}

func (t *trigger) updateMissingValidatorsInfo() {
	t.mutMissingValidatorsInfo.Lock()
	for hash := range t.mapMissingValidatorsInfo {
		_, isValidatorInfoFound := t.validatorInfoPool.SearchFirstData([]byte(hash))
		if isValidatorInfoFound {
			delete(t.mapMissingValidatorsInfo, hash)
		}
	}
	numMissingValidatorsInfo := len(t.mapMissingValidatorsInfo)
	t.mutMissingValidatorsInfo.Unlock()

	if numMissingValidatorsInfo == 0 {
		log.Debug("trigger.updateMissingValidatorsInfo -> updateTriggerFromMeta")
		t.mutTrigger.Lock()
		t.updateTriggerFromMeta()
		t.mutTrigger.Unlock()
	}
}

// IsEpochStart returns true if conditions are fulfilled for start of epoch
func (t *trigger) IsEpochStart() bool {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.isEpochStart
}

// Epoch returns the current epoch number
func (t *trigger) Epoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// MetaEpoch returns the highest finalized meta epoch number
func (t *trigger) MetaEpoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.metaEpoch
}

// EpochStartRound returns the start round of the current epoch
func (t *trigger) EpochStartRound() uint64 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochStartRound
}

// EpochFinalityAttestingRound returns the round when epoch start block was finalized
func (t *trigger) EpochFinalityAttestingRound() uint64 {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	return t.epochFinalityAttestingRound
}

// ForceEpochStart does nothing in this implementation
func (t *trigger) ForceEpochStart(_ uint64) {
}

// RequestEpochStartIfNeeded request the needed epoch start block if metablock with new epoch was received
func (t *trigger) RequestEpochStartIfNeeded(interceptedHeader data.HeaderHandler) {
	if interceptedHeader.IsStartOfEpochBlock() {
		return
	}
	if interceptedHeader.GetEpoch() <= t.Epoch() {
		return
	}
	_, ok := interceptedHeader.(*block.MetaBlock)
	if !ok {
		return
	}

	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	found := false
	for _, header := range t.mapEpochStartHdrs {
		if header.GetEpoch() >= interceptedHeader.GetEpoch() {
			found = true
			break
		}
	}

	if !found {
		t.requestHandler.RequestStartOfEpochMetaBlock(interceptedHeader.GetEpoch())
	}
}

func (t *trigger) changeEpochFinalityAttestingRoundIfNeeded(
	metaHdr *block.MetaBlock,
	receivedHash []byte,
) {
	hash := t.mapFinalizedEpochs[metaHdr.Epoch]
	epochStartMetaHdr := t.mapEpochStartHdrs[hash]
	if check.IfNil(epochStartMetaHdr) {
		return
	}

	isHeaderOnTopOfFinalityAttestingRound := metaHdr.Nonce == epochStartMetaHdr.GetNonce()+t.finality+1
	if isHeaderOnTopOfFinalityAttestingRound {
		metaHdrWithFinalityAttestingRound, err := t.getHeaderWithNonceAndHash(epochStartMetaHdr.GetNonce()+t.finality, metaHdr.PrevHash)
		if err != nil {
			log.Debug("searched metaHeader was not found")
			_ = t.requestedFinalityAttestingBlock.SetReturningPrevious()
			return
		}

		t.epochFinalityAttestingRound = metaHdrWithFinalityAttestingRound.GetRound()
		return
	}

	isFinalityAttestingBlock := metaHdr.Nonce == epochStartMetaHdr.GetNonce()+t.finality
	if !isFinalityAttestingBlock {
		return
	}

	err := t.headerValidator.IsHeaderConstructionValid(metaHdr, t.epochStartMeta)
	if err != nil {
		return
	}

	if t.requestedFinalityAttestingBlock.IsSet() {
		_, err = t.getHeaderWithNonceAndPrevHash(metaHdr.Nonce+1, receivedHash)
		if err != nil {
			return
		}

		t.epochFinalityAttestingRound = metaHdr.GetRound()
		t.requestedFinalityAttestingBlock.Reset()
		return
	}

	if metaHdr.GetRound() >= t.epochFinalityAttestingRound {
		return
	}

	t.epochFinalityAttestingRound = metaHdr.GetRound()
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving checks if trigger can be updated
func (t *trigger) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	metaHdr, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return
	}

	if !t.isPreviousEpochStartMetaBlock(metaHdr, metaBlockHash) {
		_, ok = t.mapFinalizedEpochs[metaHdr.Epoch]
		if t.metaEpoch == headerHandler.GetEpoch() && ok {
			t.changeEpochFinalityAttestingRoundIfNeeded(metaHdr, metaBlockHash)
			return
		}
	}

	if !t.newEpochHdrReceived && !metaHdr.IsStartOfEpochBlock() {
		return
	}

	isMetaStartOfEpochForCurrentEpoch := metaHdr.Epoch == t.epoch && metaHdr.IsStartOfEpochBlock()
	if isMetaStartOfEpochForCurrentEpoch {
		return
	}

	if _, ok = t.mapHashHdr[string(metaBlockHash)]; ok {
		return
	}
	if _, ok = t.mapEpochStartHdrs[string(metaBlockHash)]; ok {
		return
	}

	if metaHdr.IsStartOfEpochBlock() {
		t.newEpochHdrReceived = true
		t.mapEpochStartHdrs[string(metaBlockHash)] = metaHdr
		// waiting for late broadcast of mini blocks and transactions to be done and received
		wait := common.ExtraDelayForRequestBlockInfo
		roundDifferences := t.roundHandler.Index() - int64(headerHandler.GetRound())
		if roundDifferences > 1 {
			wait = 0
		}

		time.Sleep(wait)
	}

	t.mapHashHdr[string(metaBlockHash)] = metaHdr
	t.mapNonceHashes[metaHdr.Nonce] = append(t.mapNonceHashes[metaHdr.Nonce], string(metaBlockHash))

	t.updateTriggerFromMeta()
}

// call only if mutex is locked before
func (t *trigger) isPreviousEpochStartMetaBlock(metaBlock *block.MetaBlock, metaBlockHash []byte) bool {
	metaHdrHashesWithNonce := t.mapNonceHashes[metaBlock.Nonce+1]
	for _, hash := range metaHdrHashesWithNonce {
		epochStartMetaBlock, ok := t.mapEpochStartHdrs[hash]
		if !ok {
			continue
		}
		if !bytes.Equal(metaBlockHash, epochStartMetaBlock.GetPrevHash()) {
			continue
		}

		return true
	}

	return false
}

// call only if mutex is locked before
func (t *trigger) updateTriggerFromMeta() {
	sortedMetaInfo := make(metaInfoSlice, 0, len(t.mapEpochStartHdrs))
	for hash, hdr := range t.mapEpochStartHdrs {
		if _, ok := t.mapFinalizedEpochs[hdr.GetEpoch()]; ok {
			continue
		}

		currMetaInfo := &metaInfo{
			hdr:  hdr,
			hash: hash,
		}
		sortedMetaInfo = append(sortedMetaInfo, currMetaInfo)
	}

	sort.Sort(sortedMetaInfo)

	for _, currMetaInfo := range sortedMetaInfo {
		if _, ok := t.mapFinalizedEpochs[currMetaInfo.hdr.GetEpoch()]; ok {
			continue
		}

		canActivateEpochStart, finalityAttestingRound := t.checkIfTriggerCanBeActivated(currMetaInfo.hash, currMetaInfo.hdr)
		if canActivateEpochStart && t.metaEpoch < currMetaInfo.hdr.GetEpoch() {
			t.metaEpoch = currMetaInfo.hdr.GetEpoch()
			t.isEpochStart = true
			t.epochStartRound = currMetaInfo.hdr.GetRound()
			t.epochFinalityAttestingRound = finalityAttestingRound
			t.epochMetaBlockHash = []byte(currMetaInfo.hash)
			t.epochStartMeta = currMetaInfo.hdr
			t.saveCurrentState(currMetaInfo.hdr.GetRound())
			t.epochStartNotifier.NotifyEpochChangeConfirmed(t.metaEpoch)

			msg := fmt.Sprintf("EPOCH %d BEGINS IN ROUND (%d)", t.metaEpoch, t.epochStartRound)
			log.Debug(display.Headline(msg, "", "#"))
			log.Debug("trigger.updateTriggerFromMeta", "isEpochStart", t.isEpochStart)
			logger.SetCorrelationEpoch(t.metaEpoch)
			t.clearMissingMiniBlocksMap(t.metaEpoch)
			t.clearMissingValidatorsInfoMap(t.metaEpoch)
		}

		// save all final-valid epoch start blocks
		if canActivateEpochStart {
			t.mapFinalizedEpochs[currMetaInfo.hdr.GetEpoch()] = currMetaInfo.hash
			t.saveEpochStartMeta(currMetaInfo.hdr)
		}
	}
}

func (t *trigger) saveEpochStartMeta(metaHdr data.HeaderHandler) {
	if check.IfNil(metaHdr) {
		return
	}

	epochStartIdentifier := core.EpochStartIdentifier(metaHdr.GetEpoch())

	metaBuff, err := t.marshaller.Marshal(metaHdr)
	if err != nil {
		log.Debug("updateTriggerFromMeta marshal", "error", err.Error())
		return
	}

	err = t.metaHdrStorage.Put([]byte(epochStartIdentifier), metaBuff)
	if err != nil {
		log.Debug("updateTriggerMeta put into metaHdrStorage", "error", err.Error())
	}

	err = t.triggerStorage.Put([]byte(epochStartIdentifier), metaBuff)
	if err != nil {
		log.Debug("updateTriggerMeta put into triggerStorage", "error", err.Error())
	}
}

// call only if mutex is locked before
func (t *trigger) isMetaBlockValid(hash string, metaHdr data.HeaderHandler) bool {
	currHdr := metaHdr
	for i := metaHdr.GetNonce() - 1; i >= metaHdr.GetNonce()-t.validity; i-- {
		neededHdr, err := t.getHeaderWithNonceAndHash(i, currHdr.GetPrevHash())
		if err != nil {
			log.Debug("isMetaBlockValid.getHeaderWithNonceAndHash", "hash", hash, "error", err.Error())
			return false
		}

		err = t.headerValidator.IsHeaderConstructionValid(currHdr, neededHdr)
		if err != nil {
			log.Debug("isMetaBlockValid.IsHeaderConstructionValid", "hash", hash, "error", err.Error())
			return false
		}

		currHdr = neededHdr
	}

	return true
}

func (t *trigger) isMetaBlockFinal(_ string, metaHdr data.HeaderHandler) (bool, uint64) {
	nextBlocksVerified := uint64(0)
	finalityAttestingRound := metaHdr.GetRound()
	currHdr := metaHdr
	for nonce := metaHdr.GetNonce() + 1; nonce <= metaHdr.GetNonce()+t.finality; nonce++ {
		currHash, err := core.CalculateHash(t.marshaller, t.hasher, currHdr)
		if err != nil {
			continue
		}

		neededHdr, err := t.getHeaderWithNonceAndPrevHash(nonce, currHash)
		if err != nil {
			continue
		}

		currHdr = neededHdr

		finalityAttestingRound = currHdr.GetRound()
		nextBlocksVerified += 1
	}

	if nextBlocksVerified < t.finality {
		log.Debug("isMetaBlockFinal", "nextBlocksVerified", nextBlocksVerified, "finality", t.finality)
		for nonce := currHdr.GetNonce() + 1; nonce <= currHdr.GetNonce()+t.finality; nonce++ {
			go t.requestHandler.RequestMetaHeaderByNonce(nonce)
		}
		return false, 0
	}

	return true, finalityAttestingRound
}

// call only if mutex is locked before
func (t *trigger) checkIfTriggerCanBeActivated(hash string, metaHdr data.HeaderHandler) (bool, uint64) {
	isMetaHdrValid := t.isMetaBlockValid(hash, metaHdr)
	if !isMetaHdrValid {
		return false, 0
	}

	missingMiniBlocksHashes, blockBody, err := t.peerMiniBlocksSyncer.SyncMiniBlocks(metaHdr)
	if err != nil {
		t.addMissingMiniBlocks(metaHdr.GetEpoch(), missingMiniBlocksHashes)
		log.Debug("checkIfTriggerCanBeActivated.SyncMiniBlocks", "num missing mini blocks", len(missingMiniBlocksHashes), "error", err)
		return false, 0
	}

	if metaHdr.GetEpoch() >= t.enableEpochsHandler.RefactorPeersMiniBlocksEnableEpoch() {
		missingValidatorsInfoHashes, validatorsInfo, err := t.peerMiniBlocksSyncer.SyncValidatorsInfo(blockBody)
		if err != nil {
			t.addMissingValidatorsInfo(metaHdr.GetEpoch(), missingValidatorsInfoHashes)
			log.Debug("checkIfTriggerCanBeActivated.SyncValidatorsInfo", "num missing validators info", len(missingValidatorsInfoHashes), "error", err)
			return false, 0
		}

		for validatorInfoHash, validatorInfo := range validatorsInfo {
			t.currentEpochValidatorInfoPool.AddValidatorInfo([]byte(validatorInfoHash), validatorInfo)
		}
	}

	t.epochStartNotifier.NotifyAllPrepare(metaHdr, blockBody)

	isMetaHdrFinal, finalityAttestingRound := t.isMetaBlockFinal(hash, metaHdr)
	return isMetaHdrFinal, finalityAttestingRound
}

func (t *trigger) addMissingMiniBlocks(epoch uint32, missingMiniBlocksHashes [][]byte) {
	t.mutMissingMiniBlocks.Lock()
	defer t.mutMissingMiniBlocks.Unlock()

	for _, hash := range missingMiniBlocksHashes {
		t.mapMissingMiniBlocks[string(hash)] = epoch
		log.Debug("trigger.addMissingMiniBlocks", "epoch", epoch, "hash", hash)
	}
}

func (t *trigger) addMissingValidatorsInfo(epoch uint32, missingValidatorsInfoHashes [][]byte) {
	t.mutMissingValidatorsInfo.Lock()
	defer t.mutMissingValidatorsInfo.Unlock()

	for _, hash := range missingValidatorsInfoHashes {
		t.mapMissingValidatorsInfo[string(hash)] = epoch
		log.Debug("trigger.addMissingValidatorsInfo", "epoch", epoch, "hash", hash)
	}
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndHashFromMaps(nonce uint64, neededHash []byte) data.HeaderHandler {
	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		if !bytes.Equal(neededHash, []byte(hash)) {
			continue
		}

		neededHdr := t.mapHashHdr[hash]
		if !check.IfNil(neededHdr) {
			return neededHdr
		}
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithHashFromPool(neededHash []byte) data.HeaderHandler {
	peekedData, _ := t.headersPool.GetHeaderByHash(neededHash)
	neededHdr, ok := peekedData.(*block.MetaBlock)
	if ok {
		t.mapHashHdr[string(neededHash)] = neededHdr
		t.mapNonceHashes[neededHdr.Nonce] = append(t.mapNonceHashes[neededHdr.Nonce], string(neededHash))
		return neededHdr
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithHashFromStorage(neededHash []byte) data.HeaderHandler {
	storageData, err := t.metaHdrStorage.Get(neededHash)
	if err == nil {
		var neededHdr block.MetaBlock
		err = t.marshaller.Unmarshal(&neededHdr, storageData)
		if err == nil {
			t.mapHashHdr[string(neededHash)] = &neededHdr
			t.mapNonceHashes[neededHdr.Nonce] = append(t.mapNonceHashes[neededHdr.Nonce], string(neededHash))
			return &neededHdr
		}
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndHash(nonce uint64, neededHash []byte) (data.HeaderHandler, error) {
	metaHdr := t.getHeaderWithNonceAndHashFromMaps(nonce, neededHash)
	if !check.IfNil(metaHdr) {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithHashFromPool(neededHash)
	if !check.IfNil(metaHdr) {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithHashFromStorage(neededHash)
	if !check.IfNil(metaHdr) {
		return metaHdr, nil
	}

	go t.requestHandler.RequestMetaHeader(neededHash)

	return nil, epochStart.ErrMetaHdrNotFound
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndPrevHashFromMaps(nonce uint64, prevHash []byte) data.HeaderHandler {
	lowestRound := uint64(math.MaxUint64)
	var chosenMeta data.HeaderHandler = &block.MetaBlock{}

	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		hdrWithNonce := t.mapHashHdr[hash]
		if check.IfNil(hdrWithNonce) {
			continue
		}
		if !bytes.Equal(hdrWithNonce.GetPrevHash(), prevHash) {
			continue
		}

		if lowestRound > hdrWithNonce.GetRound() {
			lowestRound = hdrWithNonce.GetRound()
			chosenMeta = hdrWithNonce
		}
	}

	if lowestRound == uint64(math.MaxUint64) {
		return nil
	}

	return chosenMeta
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndPrevHashFromCache(nonce uint64, prevHash []byte) data.HeaderHandler {
	headers, hashes, err := t.headersPool.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
	if err != nil {
		return nil
	}

	lowestRound := uint64(math.MaxUint64)
	chosenMeta := &block.MetaBlock{}
	for i, header := range headers {
		if !bytes.Equal(header.GetPrevHash(), prevHash) {
			continue
		}

		hdrWithNonce, ok := header.(*block.MetaBlock)
		if !ok {
			continue
		}

		if lowestRound > hdrWithNonce.GetRound() {
			lowestRound = hdrWithNonce.GetRound()
			chosenMeta = hdrWithNonce
		}

		t.mapHashHdr[string(hashes[i])] = hdrWithNonce
		t.mapNonceHashes[hdrWithNonce.Nonce] = append(t.mapNonceHashes[hdrWithNonce.Nonce], string(hashes[i]))
	}

	if lowestRound == uint64(math.MaxUint64) {
		return nil
	}

	return chosenMeta
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndPrevHash(nonce uint64, prevHash []byte) (data.HeaderHandler, error) {
	metaHdr := t.getHeaderWithNonceAndPrevHashFromMaps(nonce, prevHash)
	if !check.IfNil(metaHdr) {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithNonceAndPrevHashFromCache(nonce, prevHash)
	if !check.IfNil(metaHdr) {
		return metaHdr, nil
	}

	return nil, epochStart.ErrMetaHdrNotFound
}

func (t *trigger) getAllFinishedStartOfEpochMetaHdrs() []data.HeaderHandler {
	finishedMetaHdrs := make([]data.HeaderHandler, 0, len(t.mapFinalizedEpochs))
	for _, hash := range t.mapFinalizedEpochs {
		metaHdr := t.mapEpochStartHdrs[hash]
		finishedMetaHdrs = append(finishedMetaHdrs, metaHdr)
	}
	return finishedMetaHdrs
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (t *trigger) SetProcessed(header data.HeaderHandler, _ data.BodyHandler) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	shardHdr, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return
	}

	if !shardHdr.IsStartOfEpochBlock() {
		return
	}

	t.appStatusHandler.SetUInt64Value(common.MetricRoundAtEpochStart, shardHdr.GetRound())
	t.appStatusHandler.SetUInt64Value(common.MetricNonceAtEpochStart, shardHdr.GetNonce())

	t.epoch = shardHdr.GetEpoch()
	if t.metaEpoch < t.epoch {
		t.metaEpoch = t.epoch
		t.epochMetaBlockHash = shardHdr.GetEpochStartMetaHash()
	}

	t.isEpochStart = false
	t.newEpochHdrReceived = false
	t.epochStartShardHeader = shardHdr
	finishedStartOfEpochMetaHdrs := t.getAllFinishedStartOfEpochMetaHdrs()

	err := t.saveEpochStartInfoToStaticStorer()
	if err != nil {
		log.Warn("SetProcessed: failed to save epoch start info to static storage", "error", err)
	}

	t.epochStartNotifier.NotifyAll(shardHdr)

	t.mapHashHdr = make(map[string]data.HeaderHandler)
	t.mapNonceHashes = make(map[uint64][]string)
	t.mapEpochStartHdrs = make(map[string]data.HeaderHandler)
	t.mapFinalizedEpochs = make(map[uint32]string)

	t.saveCurrentState(header.GetRound())

	log.Debug("trigger.SetProcessed", "isEpochStart", t.isEpochStart)

	shardHdrBuff, errNotCritical := t.marshaller.Marshal(shardHdr)
	if errNotCritical != nil {
		log.Warn("SetProcessed marshal error", "error", errNotCritical)
	}

	epochStartIdentifier := core.EpochStartIdentifier(shardHdr.GetEpoch())
	errNotCritical = t.shardHdrStorage.Put([]byte(epochStartIdentifier), shardHdrBuff)
	if errNotCritical != nil {
		logLevel := logger.LogWarning
		if core.IsClosingError(errNotCritical) {
			logLevel = logger.LogDebug
		}
		log.Log(logLevel, "SetProcessed put to shard header storage error", "error", errNotCritical)
	}

	// save finished start of epoch meta hdrs to current storage
	for _, metaHdr := range finishedStartOfEpochMetaHdrs {
		t.saveEpochStartMeta(metaHdr)
	}
}

// RevertStateToBlock will revert the state of the trigger to the current block
func (t *trigger) RevertStateToBlock(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return epochStart.ErrNilHeaderHandler
	}

	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	currentHeaderHash, err := core.CalculateHash(t.marshaller, t.hasher, header)
	if err != nil {
		log.Warn("RevertStateToBlock error on hashing", "error", err)
		return err
	}

	if !bytes.Equal(t.epochStartShardHeader.GetPrevHash(), currentHeaderHash) {
		return nil
	}

	log.Debug("trigger.RevertStateToBlock behind start of epoch block")

	if t.epochStartShardHeader.GetEpoch() <= 1 {
		t.epochStartShardHeader = &block.Header{}
		t.isEpochStart = true
		t.newEpochHdrReceived = true
		log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)

		return nil
	}

	shardHdrBuff := make([]byte, 0)
	epoch := t.epochStartShardHeader.GetEpoch() - 1
	for ; epoch > 0; epoch-- {
		prevEpochStartIdentifier := core.EpochStartIdentifier(epoch)
		shardHdrBuff, err = t.shardHdrStorage.SearchFirst([]byte(prevEpochStartIdentifier))
		if err != nil {
			log.Debug("RevertStateToBlock get header from storage error", "err", err)
			continue
		}

		break
	}

	if epoch == 0 {
		t.epochStartShardHeader = &block.Header{}
		t.isEpochStart = true
		t.newEpochHdrReceived = true
		log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)

		return nil
	}

	shardHdr, err := process.UnmarshalShardHeader(t.marshaller, shardHdrBuff)
	if err != nil {
		log.Warn("RevertStateToBlock unmarshal error", "err", err)
		return err
	}

	epochStartIdentifier := core.EpochStartIdentifier(t.epochStartShardHeader.GetEpoch())
	errNotCritical := t.shardHdrStorage.Remove([]byte(epochStartIdentifier))
	if errNotCritical != nil {
		log.Warn("RevertStateToBlock remove from header storage error", "err", errNotCritical)
	}

	t.epochStartShardHeader = shardHdr
	t.isEpochStart = true
	t.newEpochHdrReceived = true
	log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)

	return nil
}

// this has to be run under trigger mutex
func (t *trigger) saveEpochStartInfoToStaticStorer() error {
	epoch := t.metaEpoch

	epochStartMetablock, err := t.saveEpochStartBlockToStaticStorer(epoch)
	if err != nil {
		log.Debug("saveEpochStartBlockToStaticStorage failed", "error", err.Error())
		return err
	}

	err = t.savePeerMiniBlocksToStaticStorer(epochStartMetablock)
	if err != nil {
		log.Debug("savePeerMiniBlocksToStaticStorer failed", "error", err.Error())
		return err
	}

	return nil
}

func (t *trigger) saveEpochStartBlockToStaticStorer(epoch uint32) (data.HeaderHandler, error) {
	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	marshalledMetaBlock, err := t.metaHdrStorage.Get([]byte(epochStartIdentifier))
	if err != nil {
		return nil, err
	}

	epochStartMetablock := &block.MetaBlock{}
	err = t.marshaller.Unmarshal(epochStartMetablock, marshalledMetaBlock)
	if err != nil {
		return nil, err
	}

	epochStartBootstrapKey := append([]byte(common.EpochStartStaticBootstrapKeyPrefix), []byte(fmt.Sprint(epoch))...)
	err = t.epochStartStaticStorage.Put(epochStartBootstrapKey, marshalledMetaBlock)
	if err != nil {
		return nil, err
	}

	log.Debug("saveEpochStartInfoToStaticStorage put metaBlock into epochStartStaticStorage", "epoch", epoch) // TODO: remove log

	return epochStartMetablock, nil
}

func (t *trigger) fetchAndSavePeerMiniBlockToStaticStorer(
	mbHeader data.MiniBlockHeaderHandler,
) (*block.MiniBlock, error) {
	miniBlockHash := mbHeader.GetHash()
	value, ok := t.miniBlocksPool.Get(miniBlockHash)
	if !ok {
		log.Warn("failed to get miniblock from pool", "miniblock hash", miniBlockHash)
		return nil, fmt.Errorf("failed to get miniblock from cache")
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		return nil, errors.ErrWrongTypeAssertion
	}

	marshalizedMiniBlock, err := t.marshaller.Marshal(miniBlock)
	if err != nil {
		return nil, err
	}

	err = t.epochStartStaticStorage.Put(miniBlockHash, marshalizedMiniBlock)
	if err != nil {
		return nil, err
	}

	log.Debug("saveEpochStartInfoToStaticStorage put miniBlock into epochStartStaticStorage", "miniBlock hash", miniBlockHash) // TODO: remove log?

	return miniBlock, nil
}

func (t *trigger) fetchAndSaveValidatorInfoToStaticStorer(validatorInfoHash []byte) error {
	validatorInfo, err := t.currentEpochValidatorInfoPool.GetValidatorInfo(validatorInfoHash)
	if err != nil {
		return err
	}

	validatorInfoMarshalled, err := t.marshaller.Marshal(validatorInfo)
	if err != nil {
		return err
	}

	err = t.epochStartStaticStorage.Put(validatorInfoHash, validatorInfoMarshalled)
	if err != nil {
		return err
	}

	log.Debug("saveEpochStartInfoToStaticStorage put validatorInfo into epochStartStaticStorage", "validatorInfo hash", validatorInfoHash)

	return nil
}

func (t *trigger) savePeerMiniBlocksToStaticStorer(header data.HeaderHandler) error {
	for _, miniBlockHeader := range header.GetMiniBlockHeaderHandlers() {
		if miniBlockHeader.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		miniBlock, err := t.fetchAndSavePeerMiniBlockToStaticStorer(miniBlockHeader)
		if err != nil {
			return err
		}

		if header.GetEpoch() >= t.enableEpochsHandler.RefactorPeersMiniBlocksEnableEpoch() {
			for _, txHash := range miniBlock.TxHashes {
				err := t.fetchAndSaveValidatorInfoToStaticStorer(txHash)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochMetaBlockHash
}

// GetSavedStateKey returns the last saved trigger state key
func (t *trigger) GetSavedStateKey() []byte {
	return t.triggerStateKey
}

// Update updates the end-of-epoch trigger
func (t *trigger) Update(_ uint64, _ uint64) {
}

// SetFinalityAttestingRound sets the round which finalized the start of epoch block
func (t *trigger) SetFinalityAttestingRound(_ uint64) {
}

// SetCurrentEpochStartRound sets the round when the current epoch started
func (t *trigger) SetCurrentEpochStartRound(_ uint64) {
}

// needs to be called under locked mutex
func (t *trigger) saveCurrentState(round uint64) {
	t.triggerStateKey = []byte(fmt.Sprint(round))
	err := t.saveState(t.triggerStateKey)
	if err != nil {
		log.Debug("error saving trigger state", "error", err, "key", t.triggerStateKey)
	}
}

// Close will close the endless running go routine
func (t *trigger) Close() error {
	if t.cancelFunc != nil {
		t.cancelFunc()
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
