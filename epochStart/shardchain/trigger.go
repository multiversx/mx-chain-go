package shardchain

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("epochStart/shardchain")

// ArgsShardEpochStartTrigger struct { defines the arguments needed for new start of epoch trigger
type ArgsShardEpochStartTrigger struct {
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher

	HeaderValidator epochStart.HeaderValidator
	Uint64Converter typeConverters.Uint64ByteSliceConverter

	DataPool               dataRetriever.PoolsHolder
	Storage                dataRetriever.StorageService
	RequestHandler         epochStart.RequestHandler
	EpochStartNotifier     epochStart.Notifier
	ValidatorInfoProcessor process.ValidatorInfoProcessorHandler

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
	epochStartShardHeader       *block.Header
	epochStartMeta              *block.MetaBlock

	mutTrigger        sync.RWMutex
	mapHashHdr        map[string]*block.MetaBlock
	mapNonceHashes    map[uint64][]string
	mapEpochStartHdrs map[string]*block.MetaBlock

	headersPool         dataRetriever.HeadersPool
	shardHdrStorage     storage.Storer
	metaHdrStorage      storage.Storer
	triggerStorage      storage.Storer
	metaNonceHdrStorage storage.Storer

	uint64Converter typeConverters.Uint64ByteSliceConverter

	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	headerValidator epochStart.HeaderValidator

	requestHandler     epochStart.RequestHandler
	epochStartNotifier epochStart.Notifier

	epoch uint32

	newEpochHdrReceived bool
	isEpochStart        bool

	validatorInfoProcessor process.ValidatorInfoProcessorHandler
}

type metaInfo struct {
	hdr  *block.MetaBlock
	hash string
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
	if check.IfNil(args.ValidatorInfoProcessor) {
		return nil, epochStart.ErrNilValidatorInfoProcessor
	}
	if check.IfNil(args.Uint64Converter) {
		return nil, epochStart.ErrNilUint64Converter
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}

	metaHdrStorage := args.Storage.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaHdrStorage) {
		return nil, epochStart.ErrNilMetaBlockStorage
	}

	triggerStorage := args.Storage.GetStorer(dataRetriever.BootstrapUnit)
	if check.IfNil(triggerStorage) {
		return nil, epochStart.ErrNilTriggerStorage
	}

	metaHdrNoncesStorage := args.Storage.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if check.IfNil(metaHdrNoncesStorage) {
		return nil, epochStart.ErrNilMetaNonceHashStorage
	}

	shardHdrStorage := args.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
	if check.IfNil(shardHdrStorage) {
		return nil, epochStart.ErrNilShardHeaderStorage
	}

	trigggerStateKey := fmt.Sprintf("initial_value_epoch%d", args.Epoch)

	newTrigger := &trigger{
		triggerStateKey:             []byte(trigggerStateKey),
		epoch:                       args.Epoch,
		currentRoundIndex:           0,
		epochStartRound:             0,
		epochFinalityAttestingRound: 0,
		isEpochStart:                false,
		validity:                    args.Validity,
		finality:                    args.Finality,
		newEpochHdrReceived:         false,
		mutTrigger:                  sync.RWMutex{},
		mapHashHdr:                  make(map[string]*block.MetaBlock),
		mapNonceHashes:              make(map[uint64][]string),
		mapEpochStartHdrs:           make(map[string]*block.MetaBlock),
		headersPool:                 args.DataPool.Headers(),
		metaHdrStorage:              metaHdrStorage,
		shardHdrStorage:             shardHdrStorage,
		triggerStorage:              triggerStorage,
		metaNonceHdrStorage:         metaHdrNoncesStorage,
		uint64Converter:             args.Uint64Converter,
		marshalizer:                 args.Marshalizer,
		hasher:                      args.Hasher,
		headerValidator:             args.HeaderValidator,
		requestHandler:              args.RequestHandler,
		epochMetaBlockHash:          nil,
		epochStartNotifier:          args.EpochStartNotifier,
		epochStartMeta:              &block.MetaBlock{},
		epochStartShardHeader:       &block.Header{},
		validatorInfoProcessor:      args.ValidatorInfoProcessor,
	}

	newTrigger.headersPool.RegisterHandler(newTrigger.receivedMetaBlock)

	err := newTrigger.saveState(newTrigger.triggerStateKey)
	if err != nil {
		return nil, err
	}

	return newTrigger, nil
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

// ForceEpochStart sets the conditions for start of epoch to true in case of edge cases
func (t *trigger) ForceEpochStart(_ uint64) error {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	return nil
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

func (t *trigger) setLowerEpochFinalityAttestingRoundIfNeeded(metaHdr *block.MetaBlock) {
	if metaHdr.GetRound() >= t.epochFinalityAttestingRound {
		return
	}
	err := t.headerValidator.IsHeaderConstructionValid(metaHdr, t.epochStartMeta)
	if err != nil {
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

	if t.isEpochStart && headerHandler.GetEpoch() == t.epoch {
		t.setLowerEpochFinalityAttestingRoundIfNeeded(metaHdr)
		return
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

	t.updateTriggerFromMeta(metaHdr, metaBlockHash)
}

// call only if mutex is locked before
func (t *trigger) updateTriggerFromMeta(metaHdr *block.MetaBlock, hdrHash []byte) {
	if metaHdr.IsStartOfEpochBlock() {
		t.newEpochHdrReceived = true
		t.mapEpochStartHdrs[string(hdrHash)] = metaHdr
	}

	t.mapHashHdr[string(hdrHash)] = metaHdr
	t.mapNonceHashes[metaHdr.Nonce] = append(t.mapNonceHashes[metaHdr.Nonce], string(hdrHash))

	sortedMetaInfo := make([]*metaInfo, 0, len(t.mapEpochStartHdrs))
	for hash, hdr := range t.mapEpochStartHdrs {
		currMetaInfo := &metaInfo{
			hdr:  hdr,
			hash: hash,
		}
		sortedMetaInfo = append(sortedMetaInfo, currMetaInfo)
	}

	sort.Slice(sortedMetaInfo, func(i, j int) bool {
		return sortedMetaInfo[i].hdr.Epoch < sortedMetaInfo[j].hdr.Epoch
	})

	for _, currMetaInfo := range sortedMetaInfo {
		canActivateEpochStart, finalityAttestingRound := t.checkIfTriggerCanBeActivated(currMetaInfo.hash, currMetaInfo.hdr)
		if canActivateEpochStart && t.epoch < currMetaInfo.hdr.Epoch {
			t.epoch = currMetaInfo.hdr.Epoch
			t.isEpochStart = true
			t.epochStartRound = currMetaInfo.hdr.Round
			t.epochFinalityAttestingRound = finalityAttestingRound
			t.epochMetaBlockHash = []byte(currMetaInfo.hash)
			t.epochStartMeta = currMetaInfo.hdr
			t.saveCurrentState(currMetaInfo.hdr.GetRound())

			msg := fmt.Sprintf("EPOCH %d BEGINS IN ROUND (%d)", t.epoch, t.epochStartRound)
			log.Debug(display.Headline(msg, "", "#"))
			log.Debug("trigger.updateTriggerFromMeta", "isEpochStart", t.isEpochStart)
			logger.SetCorrelationEpoch(t.epoch)
		}

		// save all final-valid epoch start blocks
		if canActivateEpochStart {
			metaBuff, err := t.marshalizer.Marshal(currMetaInfo.hdr)
			if err != nil {
				log.Debug("updateTriggerFromMeta marshal", "error", err.Error())
				continue
			}

			epochStartIdentifier := core.EpochStartIdentifier(currMetaInfo.hdr.Epoch)
			err = t.metaHdrStorage.Put([]byte(epochStartIdentifier), metaBuff)
			if err != nil {
				log.Debug("updateTriggerMeta put into metaHdrStorage", "error", err.Error())
				continue
			}
		}
	}
}

// call only if mutex is locked before
func (t *trigger) isMetaBlockValid(_ string, metaHdr *block.MetaBlock) bool {
	currHdr := metaHdr
	for i := metaHdr.Nonce - 1; i >= metaHdr.Nonce-t.validity; i-- {
		neededHdr, err := t.getHeaderWithNonceAndHash(i, currHdr.PrevHash)
		if err != nil {
			return false
		}

		err = t.headerValidator.IsHeaderConstructionValid(currHdr, neededHdr)
		if err != nil {
			return false
		}
	}

	return true
}

func (t *trigger) isMetaBlockFinal(_ string, metaHdr *block.MetaBlock) (bool, uint64) {
	nextBlocksVerified := uint64(0)
	finalityAttestingRound := metaHdr.Round
	currHdr := metaHdr
	for nonce := metaHdr.Nonce + 1; nonce <= metaHdr.Nonce+t.finality; nonce++ {
		currHash, err := core.CalculateHash(t.marshalizer, t.hasher, currHdr)
		if err != nil {
			continue
		}

		neededHdr, err := t.getHeaderWithNonceAndPrevHash(nonce, currHash)
		if err != nil {
			continue
		}

		err = t.headerValidator.IsHeaderConstructionValid(neededHdr, currHdr)
		if err != nil {
			continue
		}

		currHdr = neededHdr

		finalityAttestingRound = currHdr.GetRound()
		nextBlocksVerified += 1
	}

	if nextBlocksVerified < t.finality {
		for nonce := currHdr.Nonce + 1; nonce <= currHdr.Nonce+t.finality; nonce++ {
			go t.requestHandler.RequestMetaHeaderByNonce(nonce)
		}
		return false, 0
	}

	return true, finalityAttestingRound
}

// call only if mutex is locked before
func (t *trigger) checkIfTriggerCanBeActivated(hash string, metaHdr *block.MetaBlock) (bool, uint64) {
	isMetaHdrValid := t.isMetaBlockValid(hash, metaHdr)
	if !isMetaHdrValid {
		return false, 0
	}

	if metaHdr.IsStartOfEpochBlock() {
		err := t.validatorInfoProcessor.ProcessMetaBlock(metaHdr, []byte(hash))
		if err != nil {
			log.Warn("processMetablock failed", "error", err)
			return false, 0
		}

		t.epochStartNotifier.NotifyAllPrepare(metaHdr)
	}

	isMetaHdrFinal, finalityAttestingRound := t.isMetaBlockFinal(hash, metaHdr)
	return isMetaHdrFinal, finalityAttestingRound
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndHashFromMaps(nonce uint64, neededHash []byte) *block.MetaBlock {
	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		if !bytes.Equal(neededHash, []byte(hash)) {
			continue
		}

		neededHdr := t.mapHashHdr[hash]
		if neededHdr != nil {
			return neededHdr
		}
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithHashFromPool(neededHash []byte) *block.MetaBlock {
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
func (t *trigger) getHeaderWithHashFromStorage(neededHash []byte) *block.MetaBlock {
	storageData, err := t.metaHdrStorage.Get(neededHash)
	if err == nil {
		var neededHdr block.MetaBlock
		err = t.marshalizer.Unmarshal(&neededHdr, storageData)
		if err == nil {
			t.mapHashHdr[string(neededHash)] = &neededHdr
			t.mapNonceHashes[neededHdr.Nonce] = append(t.mapNonceHashes[neededHdr.Nonce], string(neededHash))
			return &neededHdr
		}
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndHash(nonce uint64, neededHash []byte) (*block.MetaBlock, error) {
	metaHdr := t.getHeaderWithNonceAndHashFromMaps(nonce, neededHash)
	if metaHdr != nil {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithHashFromPool(neededHash)
	if metaHdr != nil {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithHashFromStorage(neededHash)
	if metaHdr != nil {
		return metaHdr, nil
	}

	go t.requestHandler.RequestMetaHeader(neededHash)

	return nil, epochStart.ErrMetaHdrNotFound
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithNonceAndPrevHashFromMaps(nonce uint64, prevHash []byte) *block.MetaBlock {
	lowestRound := uint64(math.MaxUint64)
	chosenMeta := &block.MetaBlock{}

	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		hdrWithNonce := t.mapHashHdr[hash]
		if check.IfNil(hdrWithNonce) {
			continue
		}
		if !bytes.Equal(hdrWithNonce.PrevHash, prevHash) {
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
func (t *trigger) getHeaderWithNonceAndPrevHashFromCache(nonce uint64, prevHash []byte) *block.MetaBlock {
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
func (t *trigger) getHeaderWithNonceAndPrevHash(nonce uint64, prevHash []byte) (*block.MetaBlock, error) {
	metaHdr := t.getHeaderWithNonceAndPrevHashFromMaps(nonce, prevHash)
	if metaHdr != nil {
		return metaHdr, nil
	}

	metaHdr = t.getHeaderWithNonceAndPrevHashFromCache(nonce, prevHash)
	if metaHdr != nil {
		return metaHdr, nil
	}

	nonceToByteSlice := t.uint64Converter.ToByteSlice(nonce)
	dataHdr, err := t.metaNonceHdrStorage.Get(nonceToByteSlice)
	if err != nil || len(dataHdr) == 0 {
		go t.requestHandler.RequestMetaHeaderByNonce(nonce)
		return nil, err
	}

	b := batch.Batch{}
	err = t.marshalizer.Unmarshal(&b, dataHdr)
	if err != nil {
		return nil, err
	}

	if len(b.Data) != 1 {
		return nil, marshal.ErrUnmarshallingBadSize
	}

	return t.getHeaderWithNonceAndHash(nonce, b.Data[0])
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (t *trigger) SetProcessed(header data.HeaderHandler) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	shardHdr, ok := header.(*block.Header)
	if !ok {
		return
	}

	if !shardHdr.IsStartOfEpochBlock() {
		return
	}

	t.epoch = shardHdr.Epoch
	t.isEpochStart = false
	t.newEpochHdrReceived = false
	t.epochMetaBlockHash = shardHdr.EpochStartMetaHash
	t.epochStartShardHeader = shardHdr

	t.epochStartNotifier.NotifyAll(shardHdr)

	t.mapHashHdr = make(map[string]*block.MetaBlock)
	t.mapNonceHashes = make(map[uint64][]string)
	t.mapEpochStartHdrs = make(map[string]*block.MetaBlock)

	t.saveCurrentState(header.GetRound())

	log.Debug("trigger.SetProcessed", "isEpochStart", t.isEpochStart)

	shardHdrBuff, errNotCritical := t.marshalizer.Marshal(shardHdr)
	if errNotCritical != nil {
		log.Warn("SetProcessed marshal error", "error", errNotCritical)
		return
	}

	epochStartIdentifier := core.EpochStartIdentifier(shardHdr.Epoch)
	errNotCritical = t.shardHdrStorage.Put([]byte(epochStartIdentifier), shardHdrBuff)
	if errNotCritical != nil {
		log.Warn("SetProcessed put to shard header storage error", "error", errNotCritical)
		return
	}
}

// RevertStateToBlock will revert the state of the trigger to the current block
func (t *trigger) RevertStateToBlock(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return epochStart.ErrNilHeaderHandler
	}

	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	currentHeaderHash, err := core.CalculateHash(t.marshalizer, t.hasher, header)
	if err != nil {
		log.Warn("RevertStateToBlock error on hashing", "error", err)
		return err
	}

	if !bytes.Equal(t.epochStartShardHeader.GetPrevHash(), currentHeaderHash) {
		return nil
	}

	log.Debug("trigger.RevertStateToBlock behind start of epoch block")

	if t.epochStartShardHeader.Epoch <= 1 {
		t.epochStartShardHeader = &block.Header{}
		t.isEpochStart = true
		t.newEpochHdrReceived = true
		log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)

		return nil
	}

	prevEpochStartIdentifier := core.EpochStartIdentifier(t.epochStartShardHeader.Epoch - 1)
	shardHdrBuff, err := t.shardHdrStorage.SearchFirst([]byte(prevEpochStartIdentifier))
	if err != nil {
		log.Warn("RevertStateToBlock get header from storage error", "err", err)
		return err
	}

	shardHdr := &block.Header{}
	err = t.marshalizer.Unmarshal(shardHdr, shardHdrBuff)
	if err != nil {
		log.Warn("RevertStateToBlock unmarshal error", "err", err)
		return err
	}

	epochStartIdentifier := core.EpochStartIdentifier(t.epochStartShardHeader.Epoch)
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

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochMetaBlockHash
}

// EpochStartMeta returns the epoch start announcing meta header
func (t *trigger) EpochStartMeta() *block.MetaBlock {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochStartMeta
}

// GetSavedStateKey returns the last saved trigger state key
func (t *trigger) GetSavedStateKey() []byte {
	return t.triggerStateKey
}

// Update updates the end-of-epoch trigger
func (t *trigger) Update(_ uint64) {
}

// SetFinalityAttestingRound sets the round which finalized the start of epoch block
func (t *trigger) SetFinalityAttestingRound(_ uint64) {
}

// SetCurrentEpochStartRound sets the round when the current epoch started
func (t *trigger) SetCurrentEpochStartRound(_ uint64) {
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}

// needs to be called under locked mutex
func (t *trigger) saveCurrentState(round uint64) {
	t.triggerStateKey = []byte(fmt.Sprint(round))
	err := t.saveState(t.triggerStateKey)
	if err != nil {
		log.Debug("error saving trigger state", "error", err, "key", t.triggerStateKey)
	}
}
