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
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/storage"
)

var log = logger.GetOrCreate("epochStart/shardchain")

var _ dataRetriever.EpochHandler = (*trigger)(nil)
var _ epochStart.TriggerHandler = (*trigger)(nil)
var _ process.EpochStartTriggerHandler = (*trigger)(nil)
var _ process.EpochBootstrapper = (*trigger)(nil)
var _ epochStart.BootstrapCompletedNotifier = (*trigger)(nil)
var _ closing.Closer = (*trigger)(nil)

// sleepTime defines the time in milliseconds between each iteration made in requestMissingMiniBlocks method
const sleepTime = 200 * time.Millisecond

const numRoundsWithoutReceivedMetaBlocks = 5

const defaultPendingProofRetryInterval = time.Second
const maxPendingProofRequestsPerPass = 16

// ArgsShardEpochStartTrigger struct { defines the arguments needed for new start of epoch trigger
type ArgsShardEpochStartTrigger struct {
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher

	HeaderValidator epochStart.HeaderValidator
	Uint64Converter typeConverters.Uint64ByteSliceConverter

	DataPool                                    dataRetriever.PoolsHolder
	Storage                                     dataRetriever.StorageService
	RequestHandler                              epochStart.RequestHandler
	EpochStartNotifier                          epochStart.Notifier
	PeerMiniBlocksSyncer                        process.ValidatorInfoSyncer
	RoundHandler                                process.RoundHandler
	AppStatusHandler                            core.AppStatusHandler
	EnableEpochsHandler                         common.EnableEpochsHandler
	ExtraDelayForRequestBlockInfoInMilliseconds int
	WaitForBootstrapCompletion                  bool

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

	mutTrigger                sync.RWMutex
	mapHashHdr                map[string]data.HeaderHandler
	mapNonceHashes            map[uint64][]string
	mapEpochStartHdrs         map[string]data.HeaderHandler
	mapFinalizedEpochs        map[uint32]string
	mapPreparedEpochStartHdrs map[string]struct{}

	headersPool                   dataRetriever.HeadersPool
	proofsPool                    dataRetriever.ProofsPool
	metaFinalityView              process.MetaFinalityView
	miniBlocksPool                storage.Cacher
	validatorInfoPool             dataRetriever.ShardedDataCacherNotifier
	currentEpochValidatorInfoPool epochStart.ValidatorInfoCacher
	shardHdrStorage               storage.Storer
	metaHdrStorage                storage.Storer
	triggerStorage                storage.Storer
	metaNonceHdrStorage           storage.Storer

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

	appStatusHandler              core.AppStatusHandler
	enableEpochsHandler           common.EnableEpochsHandler
	extraDelayForRequestBlockInfo time.Duration

	mapMissingMiniBlocks     map[string]uint32
	mapMissingValidatorsInfo map[string]uint32
	mutMissingMiniBlocks     sync.RWMutex
	mutMissingValidatorsInfo sync.RWMutex
	cancelFunc               func()

	chanMetaBlockReceived chan struct{}

	mutPendingEpochStartData     sync.Mutex
	pendingEpochStartProofs      map[string]pendingEpochStartProof
	pendingEpochStartHeaders     map[uint32]struct{}
	pendingFinalityEvidence      map[string]finalityEvidenceRequest
	epochStartRecoveryCandidates map[string]uint32
	chanPendingEpochStartData    chan struct{}
	pendingProofRetryInterval    time.Duration
	nextProofRequestSequence     uint64
	recoveryGeneration           uint64
	recoveryClosed               bool
	recoveryRequestCursors       [numRecoveryRequestClasses]string
	finalityCandidateCursor      string
	callbackAdmission            atomic.Flag
}

type pendingEpochStartProof struct {
	epoch           uint32
	requestSequence uint64
}

// finalityEvidenceRequest identifies the epoch start meta block whose neighbourhood the node still
// needs before it can hold it final: the parent settles it when non contended, a proofed child
// settles it in every case
type finalityEvidenceRequest struct {
	epoch    uint32
	nonce    uint64
	round    uint64
	hash     []byte
	prevHash []byte
}

type recoveryRequestClass uint8

const (
	finalityRecoveryRequest recoveryRequestClass = iota
	proofRecoveryRequest
	headerRecoveryRequest
	numRecoveryRequestClasses
)

type recoveryRequestKind uint8

const (
	equivalentProofByHashRequest recoveryRequestKind = iota
	metaHeaderByHashRequest
	startOfEpochMetaBlockRequest
	metaHeaderByNonceRequest
)

type recoveryRequestOperation struct {
	class        recoveryRequestClass
	kind         recoveryRequestKind
	requestKey   string
	ownerKeys    []string
	epoch        uint32
	requestEpoch uint32
	sequence     uint64
	nonce        uint64
	hash         []byte
}

type pendingFinalityCandidate struct {
	key     string
	sortKey string
	info    finalityEvidenceRequest
}

type metaBlockValidity uint8

const (
	metaBlockValidityIncomplete metaBlockValidity = iota
	metaBlockValidityValid
	metaBlockValidityInvalid
)

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
	if check.IfNil(args.DataPool.Proofs()) {
		return nil, epochStart.ErrNilProofsPool
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
	if args.ExtraDelayForRequestBlockInfoInMilliseconds < 0 {
		return nil, process.ErrNegativeValue
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.RefactorPeersMiniBlocksFlag,
	})
	if err != nil {
		return nil, err
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

	// built over the trigger's own pools so activation shares the node-wide finality definition
	metaFinalityView, err := track.NewMetaFinalityView(track.ArgsMetaFinalityView{
		HeadersPool: args.DataPool.Headers(),
		ProofsPool:  args.DataPool.Proofs(),
	})
	if err != nil {
		return nil, err
	}

	triggerStateKey := common.TriggerRegistryInitialKeyPrefix + fmt.Sprintf("%d", args.Epoch)
	t := &trigger{
		triggerStateKey:               []byte(triggerStateKey),
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
		proofsPool:                    args.DataPool.Proofs(),
		metaFinalityView:              metaFinalityView,
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
		extraDelayForRequestBlockInfo: time.Duration(args.ExtraDelayForRequestBlockInfoInMilliseconds) * time.Millisecond,
		chanMetaBlockReceived:         make(chan struct{}, 1),
		pendingEpochStartProofs:       make(map[string]pendingEpochStartProof),
		pendingEpochStartHeaders:      make(map[uint32]struct{}),
		pendingFinalityEvidence:       make(map[string]finalityEvidenceRequest),
		epochStartRecoveryCandidates:  make(map[string]uint32),
		chanPendingEpochStartData:     make(chan struct{}, 1),
		pendingProofRetryInterval:     defaultPendingProofRetryInterval,
	}
	t.callbackAdmission.SetValue(!args.WaitForBootstrapCompletion)

	t.headersPool.RegisterHandler(t.receivedMetaBlock)
	t.proofsPool.RegisterHandler(t.receivedProof)

	err = t.saveState(t.triggerStateKey)
	if err != nil {
		return nil, err
	}

	t.mapMissingMiniBlocks = make(map[string]uint32)
	t.mapMissingValidatorsInfo = make(map[string]uint32)
	t.mapPreparedEpochStartHdrs = make(map[string]struct{})

	var ctx context.Context
	ctx, t.cancelFunc = context.WithCancel(context.Background())
	go t.requestMissingMiniBlocks(ctx)
	go t.requestMissingValidatorsInfo(ctx)
	go t.watchdogRequestEpochStartMetaBlock(ctx)
	go t.requestPendingEpochStartProofs(ctx)

	return t, nil
}

func (t *trigger) getExtraDelayForRequestsBlockInfo() time.Duration {
	return t.extraDelayForRequestBlockInfo
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

			t.mutTrigger.Lock()
			if !t.isEpochStart {
				t.updateTriggerFromMeta()
			}
			t.mutTrigger.Unlock()

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

// processedEpochLocked returns the epoch of the last committed shard epoch-start header.
// The caller must hold mutTrigger.
func (t *trigger) processedEpochLocked() uint32 {
	if check.IfNil(t.epochStartShardHeader) {
		return t.epoch
	}
	// An initial bootstrap registry can contain an empty placeholder header.
	if t.epochStartShardHeader.GetEpoch() == 0 && t.epoch > 0 &&
		t.epochStartShardHeader.GetNonce() == 0 && t.epochStartShardHeader.GetRound() == 0 &&
		len(t.epochStartShardHeader.GetPrevHash()) == 0 &&
		bytes.HasPrefix(t.triggerStateKey, []byte(common.TriggerRegistryInitialKeyPrefix)) {
		return t.epoch
	}

	return t.epochStartShardHeader.GetEpoch()
}

func (t *trigger) processedEpoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.processedEpochLocked()
}

func (t *trigger) actionableEpoch() (uint32, bool) {
	processedEpoch := t.processedEpoch()
	if processedEpoch == math.MaxUint32 {
		return 0, false
	}

	return processedEpoch + 1, true
}

func (t *trigger) isActionableSupernovaEpochStart(header data.HeaderHandler) bool {
	if !header.IsStartOfEpochBlock() || !t.isSupernovaEpochStartFinalityEnabled(header.GetEpoch()) {
		return true
	}

	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.isActionableSupernovaEpochStartLocked(header)
}

func (t *trigger) isActionableSupernovaEpochStartLocked(header data.HeaderHandler) bool {
	if !header.IsStartOfEpochBlock() || !t.isSupernovaEpochStartFinalityEnabled(header.GetEpoch()) {
		return true
	}

	processedEpoch := t.processedEpochLocked()
	return processedEpoch != math.MaxUint32 && header.GetEpoch() == processedEpoch+1
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
	shardHeader, isShardHeader := interceptedHeader.(data.ShardHeaderHandler)
	if isShardHeader && shardHeader.IsStartOfEpochBlock() {
		t.requestEpochStartForShardHeader(shardHeader)
		return
	}

	if interceptedHeader.IsStartOfEpochBlock() {
		return
	}
	currentEpoch := t.Epoch()
	if t.isSupernovaEpochStartFinalityEnabled(interceptedHeader.GetEpoch()) {
		currentEpoch = t.processedEpoch()
	}
	if interceptedHeader.GetEpoch() <= currentEpoch {
		return
	}
	if t.isSupernovaEpochStartFinalityEnabled(interceptedHeader.GetEpoch()) &&
		(currentEpoch == math.MaxUint32 || interceptedHeader.GetEpoch() != currentEpoch+1) {
		return
	}
	_, ok := interceptedHeader.(data.MetaHeaderHandler)
	if !ok {
		return
	}

	t.mutTrigger.RLock()
	found := false
	for _, header := range t.mapEpochStartHdrs {
		if header.GetEpoch() >= interceptedHeader.GetEpoch() {
			found = true
			break
		}
	}
	t.mutTrigger.RUnlock()

	if !found {
		t.requestHandler.RequestStartOfEpochMetaBlock(interceptedHeader.GetEpoch())
	}
}

func (t *trigger) requestEpochStartForShardHeader(header data.ShardHeaderHandler) {
	targetEpoch := header.GetEpoch()
	metaBlockHash := header.GetEpochStartMetaHash()
	if len(metaBlockHash) == 0 {
		return
	}

	if !t.isSupernovaEpochStartFinalityEnabled(targetEpoch) {
		go t.requestHandler.RequestMetaHeader(metaBlockHash)
		return
	}

	processedEpoch := t.processedEpoch()
	if processedEpoch == math.MaxUint32 || targetEpoch != processedEpoch+1 {
		return
	}

	hash := bytes.Clone(metaBlockHash)
	go t.requestEpochStartFromProposal(hash, targetEpoch)
}

func (t *trigger) requestEpochStartFromProposal(metaBlockHash []byte, targetEpoch uint32) {
	actionableEpoch, ok := t.actionableEpoch()
	if !ok || targetEpoch != actionableEpoch {
		return
	}

	header, err := t.headersPool.GetHeaderByHash(metaBlockHash)
	if err != nil || check.IfNil(header) {
		if !t.isCurrentActionableEpoch(targetEpoch) {
			return
		}
		t.requestHandler.RequestMetaHeaderForEpoch(metaBlockHash, targetEpoch)
		header, err = t.headersPool.GetHeaderByHash(metaBlockHash)
		if err != nil || check.IfNil(header) {
			return
		}
	}

	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok || !metaHeader.IsStartOfEpochBlock() || header.GetEpoch() != targetEpoch {
		return
	}

	_, err = t.proofsPool.GetProof(core.MetachainShardId, metaBlockHash)
	if err != nil {
		if !t.isCurrentActionableEpoch(targetEpoch) {
			return
		}
		t.requestHandler.RequestEquivalentProofByHashForEpoch(core.MetachainShardId, metaBlockHash, targetEpoch)
		return
	}

	if !t.isCurrentActionableEpoch(targetEpoch) {
		return
	}
	t.processMetaHeaderWithProof(header, metaBlockHash)
}

func (t *trigger) registerEpochStartRecoveryCandidate(hash []byte, epoch uint32) (bool, uint64) {
	key := string(hash)

	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	if t.recoveryClosed {
		return false, 0
	}
	if _, found := t.epochStartRecoveryCandidates[key]; found {
		return false, 0
	}
	if _, found := t.pendingEpochStartProofs[key]; found {
		return false, 0
	}
	if _, found := t.pendingFinalityEvidence[key]; found {
		return false, 0
	}
	if _, found := t.pendingEpochStartHeaders[epoch]; found {
		return false, 0
	}

	t.epochStartRecoveryCandidates[key] = epoch
	return true, t.recoveryGeneration
}

func (t *trigger) recoverEpochStartCandidate(hash []byte, epoch uint32, generation uint64) {
	if !t.isRecoveryGenerationCurrent(generation) {
		return
	}

	processedEpoch := t.processedEpoch()
	if processedEpoch == math.MaxUint32 || epoch != processedEpoch+1 {
		t.removeEpochStartRecoveryCandidateForGeneration(string(hash), epoch, generation)
		return
	}

	header, err := t.headersPool.GetHeaderByHash(hash)
	if err == nil && !check.IfNil(header) {
		if !t.isRecoveryGenerationCurrent(generation) {
			return
		}
		t.recoverEpochStartWithHeader(header, hash, epoch, generation)
		return
	}

	if !t.moveCandidateToPendingHeader(hash, epoch, generation) {
		return
	}

	header, err = t.headersPool.GetHeaderByHash(hash)
	if err == nil && !check.IfNil(header) {
		if !t.isRecoveryGenerationCurrent(generation) {
			return
		}
		t.recoverEpochStartWithHeader(header, hash, epoch, generation)
		return
	}

	if !t.isRecoveryGenerationCurrent(generation) {
		return
	}
	t.requestHandler.RequestMetaHeaderForEpoch(hash, epoch)
}

func (t *trigger) recoverEpochStartWithHeader(header data.HeaderHandler, hash []byte, epoch uint32, generation uint64) {
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok || !metaHeader.IsStartOfEpochBlock() || header.GetEpoch() != epoch {
		t.discardEpochStartRecoveryCandidate(string(hash), epoch, generation)
		return
	}

	_, err := t.proofsPool.GetProof(core.MetachainShardId, hash)
	if err != nil {
		if t.moveRecoveryToPendingProof(hash, epoch, generation) {
			t.requestHandler.RequestEquivalentProofByHashForEpoch(core.MetachainShardId, hash, header.GetEpoch())
		}
		return
	}

	if t.processMetaHeaderWithProofForGeneration(header, hash, generation) {
		t.removePendingEpochStartHeaderForGeneration(epoch, generation)
	}
}

func (t *trigger) changeEpochFinalityAttestingRoundIfNeeded(
	metaHdr data.MetaHeaderHandler,
	receivedHash []byte,
) {
	hash := t.mapFinalizedEpochs[metaHdr.GetEpoch()]
	epochStartMetaHdr := t.mapEpochStartHdrs[hash]
	if check.IfNil(epochStartMetaHdr) {
		return
	}

	isHeaderOnTopOfFinalityAttestingRound := metaHdr.GetNonce() == epochStartMetaHdr.GetNonce()+t.finality+1
	if isHeaderOnTopOfFinalityAttestingRound {
		metaHdrWithFinalityAttestingRound, err := t.getHeaderWithNonceAndHash(epochStartMetaHdr.GetNonce()+t.finality, metaHdr.GetPrevHash())
		if err != nil {
			log.Debug("searched metaHeader was not found")
			_ = t.requestedFinalityAttestingBlock.SetReturningPrevious()
			return
		}

		t.epochFinalityAttestingRound = metaHdrWithFinalityAttestingRound.GetRound()
		return
	}

	isFinalityAttestingBlock := metaHdr.GetNonce() == epochStartMetaHdr.GetNonce()+t.finality
	if !isFinalityAttestingBlock {
		return
	}

	err := t.headerValidator.IsHeaderConstructionValid(metaHdr, t.epochStartMeta)
	if err != nil {
		return
	}

	if t.requestedFinalityAttestingBlock.IsSet() {
		_, err = t.getHeaderWithNonceAndPrevHash(metaHdr.GetNonce()+1, receivedHash)
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

func (t *trigger) receivedProof(headerProof data.HeaderProofHandler) {
	if check.IfNil(headerProof) {
		return
	}
	if headerProof.GetHeaderShardId() != core.MetachainShardId {
		return
	}
	if !t.callbackAdmission.IsSet() {
		return
	}

	log.Debug("received proof in trigger", "proof for header hash", headerProof.GetHeaderHash())
	header, err := t.headersPool.GetHeaderByHash(headerProof.GetHeaderHash())
	if err != nil || check.IfNil(header) {
		return
	}

	usesSupernovaFinality := t.isSupernovaEpochStartFinalityEnabled(header.GetEpoch())
	if !usesSupernovaFinality {
		t.removePendingEpochStartHeader(header.GetEpoch())
	}
	if !t.processMetaHeaderWithProof(header, headerProof.GetHeaderHash()) {
		return
	}
	t.removePendingEpochStartProof(string(headerProof.GetHeaderHash()), header.GetEpoch())
	if usesSupernovaFinality && header.IsStartOfEpochBlock() {
		t.removePendingEpochStartHeader(header.GetEpoch())
	}
}

func (t *trigger) processMetaHeaderWithProof(header data.HeaderHandler, metaBlockHash []byte) bool {
	return t.processMetaHeaderWithProofInternal(header, metaBlockHash, nil)
}

func (t *trigger) processMetaHeaderWithProofForGeneration(
	header data.HeaderHandler,
	metaBlockHash []byte,
	generation uint64,
) bool {
	return t.processMetaHeaderWithProofInternal(header, metaBlockHash, &generation)
}

func (t *trigger) processMetaHeaderWithProofInternal(
	header data.HeaderHandler,
	metaBlockHash []byte,
	generation *uint64,
) bool {
	if generation != nil && !t.isRecoveryGenerationCurrent(*generation) {
		return false
	}
	if generation == nil && !t.callbackAdmission.IsSet() {
		return false
	}
	if !t.isActionableSupernovaEpochStart(header) {
		return true
	}

	isRecoveryCandidate := t.prepareRecoveryProcessingOwner(header, metaBlockHash, generation)
	isMissingFinalityEvidence := isRecoveryCandidate &&
		!t.metaFinalityView.IsMetaHeaderHeldFinal(header, metaBlockHash)

	t.mutTrigger.Lock()
	if generation == nil && !t.callbackAdmission.IsSet() {
		t.mutTrigger.Unlock()
		return false
	}
	if generation != nil && !t.isRecoveryGenerationCurrent(*generation) {
		t.mutTrigger.Unlock()
		return false
	}
	if !t.isActionableSupernovaEpochStartLocked(header) {
		t.mutTrigger.Unlock()
		return true
	}
	if isMissingFinalityEvidence {
		if generation == nil {
			t.trackMissingFinalityEvidence(header, string(metaBlockHash))
		} else {
			t.trackMissingFinalityEvidenceForGeneration(header, string(metaBlockHash), *generation)
		}
	}
	t.checkMetaHeaderForEpochTriggerEquivalentProofs(header, metaBlockHash)
	_, retained := t.mapEpochStartHdrs[string(metaBlockHash)]
	constructionInvalid := isRecoveryCandidate && t.metaBlockValidity(string(metaBlockHash), header) == metaBlockValidityInvalid
	t.mutTrigger.Unlock()

	if isRecoveryCandidate && (!retained || constructionInvalid) {
		key := string(metaBlockHash)
		if generation == nil {
			t.removeEpochStartRecoveryCandidate(key, header.GetEpoch())
			t.removePendingFinalityEvidence(key)
		} else {
			t.removeEpochStartRecoveryCandidateForGeneration(key, header.GetEpoch(), *generation)
			t.removePendingFinalityEvidenceForGeneration(key, *generation)
		}
	}

	return true
}

func (t *trigger) prepareRecoveryProcessingOwner(
	header data.HeaderHandler,
	metaBlockHash []byte,
	generation *uint64,
) bool {
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok || !metaHeader.IsStartOfEpochBlock() {
		return false
	}
	if !t.isSupernovaEpochStartFinalityEnabled(header.GetEpoch()) {
		return false
	}
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	processedEpoch := t.processedEpochLocked()
	if header.GetEpoch() <= processedEpoch ||
		processedEpoch == math.MaxUint32 || header.GetEpoch() != processedEpoch+1 {
		return false
	}

	key := string(metaBlockHash)
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	if t.recoveryClosed || (generation != nil && t.recoveryGeneration != *generation) {
		return false
	}

	_, candidateFound := t.epochStartRecoveryCandidates[key]
	_, proofFound := t.pendingEpochStartProofs[key]
	_, finalityFound := t.pendingFinalityEvidence[key]
	_, headerFound := t.pendingEpochStartHeaders[header.GetEpoch()]
	if !candidateFound && !proofFound && !finalityFound && !headerFound {
		return false
	}

	t.epochStartRecoveryCandidates[key] = header.GetEpoch()
	return true
}

func (t *trigger) moveCandidateToPendingHeader(metaBlockHash []byte, epoch uint32, generation uint64) bool {
	key := string(metaBlockHash)

	t.mutPendingEpochStartData.Lock()
	currentEpoch, found := t.epochStartRecoveryCandidates[key]
	if t.recoveryClosed || t.recoveryGeneration != generation || !found || currentEpoch != epoch {
		t.mutPendingEpochStartData.Unlock()
		return false
	}

	delete(t.epochStartRecoveryCandidates, key)
	t.pendingEpochStartHeaders[epoch] = struct{}{}
	t.mutPendingEpochStartData.Unlock()

	t.signalPendingEpochStartData()
	return true
}

func (t *trigger) addPendingProof(metaBlockHash []byte, epoch uint32) bool {
	return t.addPendingProofInternal(metaBlockHash, epoch, nil, false)
}

func (t *trigger) addPrioritizedPendingProof(metaBlockHash []byte, epoch uint32) bool {
	return t.addPendingProofInternal(metaBlockHash, epoch, nil, true)
}

func (t *trigger) addActionablePendingProof(
	metaBlockHash []byte,
	epoch uint32,
	generation *uint64,
	prioritize bool,
) bool {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	processedEpoch := t.processedEpochLocked()
	if processedEpoch == math.MaxUint32 || epoch != processedEpoch+1 {
		return false
	}

	return t.addPendingProofInternal(metaBlockHash, epoch, generation, prioritize)
}

func (t *trigger) addPendingProofForGeneration(metaBlockHash []byte, epoch uint32, generation *uint64) bool {
	return t.addPendingProofInternal(metaBlockHash, epoch, generation, false)
}

func (t *trigger) addPendingProofInternal(
	metaBlockHash []byte,
	epoch uint32,
	generation *uint64,
	prioritize bool,
) bool {
	key := string(metaBlockHash)

	t.mutPendingEpochStartData.Lock()
	if t.recoveryClosed || (generation != nil && t.recoveryGeneration != *generation) {
		t.mutPendingEpochStartData.Unlock()
		return false
	}

	_, exists := t.pendingEpochStartProofs[key]
	if !exists {
		requestSequence := uint64(0)
		if !prioritize {
			t.nextProofRequestSequence++
			requestSequence = t.nextProofRequestSequence
		}
		t.pendingEpochStartProofs[key] = pendingEpochStartProof{
			epoch:           epoch,
			requestSequence: requestSequence,
		}
	}
	delete(t.epochStartRecoveryCandidates, key)
	t.mutPendingEpochStartData.Unlock()

	if exists {
		return false
	}

	t.signalPendingEpochStartData()
	return true
}

func (t *trigger) moveRecoveryToPendingProof(metaBlockHash []byte, epoch uint32, generation uint64) bool {
	return t.addPendingProofForGeneration(metaBlockHash, epoch, &generation)
}

func (t *trigger) signalPendingEpochStartData() {
	select {
	case t.chanPendingEpochStartData <- struct{}{}:
	default:
	}
}

func (t *trigger) removeEpochStartRecoveryCandidate(key string, epoch uint32) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryClosed {
		t.mutPendingEpochStartData.Unlock()
		return
	}
	currentEpoch, found := t.epochStartRecoveryCandidates[key]
	if found && currentEpoch == epoch {
		delete(t.epochStartRecoveryCandidates, key)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removeEpochStartRecoveryCandidateForGeneration(key string, epoch uint32, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryGeneration == generation {
		currentEpoch, found := t.epochStartRecoveryCandidates[key]
		if found && currentEpoch == epoch {
			delete(t.epochStartRecoveryCandidates, key)
		}
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) discardEpochStartRecoveryCandidate(key string, epoch uint32, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryGeneration != generation {
		t.mutPendingEpochStartData.Unlock()
		return
	}
	if currentEpoch, found := t.epochStartRecoveryCandidates[key]; found && currentEpoch == epoch {
		delete(t.epochStartRecoveryCandidates, key)
	}
	if pendingProof, found := t.pendingEpochStartProofs[key]; found && pendingProof.epoch == epoch {
		delete(t.pendingEpochStartProofs, key)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removePendingEpochStartDataForEpochOrOlder(epoch uint32) {
	t.mutPendingEpochStartData.Lock()
	for key, info := range t.pendingEpochStartProofs {
		if info.epoch <= epoch {
			delete(t.pendingEpochStartProofs, key)
		}
	}
	for pendingEpoch := range t.pendingEpochStartHeaders {
		if pendingEpoch <= epoch {
			delete(t.pendingEpochStartHeaders, pendingEpoch)
		}
	}
	for key, info := range t.pendingFinalityEvidence {
		if info.epoch <= epoch {
			delete(t.pendingFinalityEvidence, key)
		}
	}
	for key, candidateEpoch := range t.epochStartRecoveryCandidates {
		if candidateEpoch <= epoch {
			delete(t.epochStartRecoveryCandidates, key)
		}
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removeRecoveryCandidateByHash(hash []byte) {
	t.mutPendingEpochStartData.Lock()
	if !t.recoveryClosed {
		delete(t.epochStartRecoveryCandidates, string(hash))
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) resetPendingEpochStartData() {
	t.mutPendingEpochStartData.Lock()
	t.recoveryGeneration++
	t.pendingEpochStartProofs = make(map[string]pendingEpochStartProof)
	t.pendingEpochStartHeaders = make(map[uint32]struct{})
	t.pendingFinalityEvidence = make(map[string]finalityEvidenceRequest)
	t.epochStartRecoveryCandidates = make(map[string]uint32)
	t.recoveryRequestCursors = [numRecoveryRequestClasses]string{}
	t.finalityCandidateCursor = ""
	t.mutPendingEpochStartData.Unlock()
}

// removePendingEpochStartProof deletes the entry only if it still holds the same epoch, so a
// stale snapshot cannot remove state recreated by a concurrent callback
func (t *trigger) removePendingEpochStartProof(key string, epoch uint32) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryClosed {
		t.mutPendingEpochStartData.Unlock()
		return
	}
	current, found := t.pendingEpochStartProofs[key]
	if found && current.epoch == epoch {
		delete(t.pendingEpochStartProofs, key)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removePendingEpochStartHeader(epoch uint32) {
	t.mutPendingEpochStartData.Lock()
	if !t.recoveryClosed {
		delete(t.pendingEpochStartHeaders, epoch)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removePendingEpochStartHeaderForGeneration(epoch uint32, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryGeneration == generation {
		delete(t.pendingEpochStartHeaders, epoch)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removePendingEpochStartProofForGeneration(key string, epoch uint32, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryGeneration == generation {
		current, found := t.pendingEpochStartProofs[key]
		if found && current.epoch == epoch {
			delete(t.pendingEpochStartProofs, key)
		}
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) movePendingProofToHeaderRecoveryForGeneration(key string, epoch uint32, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if !t.recoveryClosed && t.recoveryGeneration == generation {
		current, found := t.pendingEpochStartProofs[key]
		if found && current.epoch == epoch {
			delete(t.pendingEpochStartProofs, key)
			t.pendingEpochStartHeaders[epoch] = struct{}{}
		}
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) pendingEpochStartDataSnapshot() (
	uint32,
	map[string]pendingEpochStartProof,
	map[uint32]struct{},
	uint64,
) {
	for {
		processedEpochBefore := t.processedEpoch()

		t.mutPendingEpochStartData.Lock()
		generation := t.recoveryGeneration
		if t.recoveryClosed {
			t.mutPendingEpochStartData.Unlock()
			return processedEpochBefore, nil, nil, generation
		}

		pending := make(map[string]pendingEpochStartProof, len(t.pendingEpochStartProofs))
		for key, info := range t.pendingEpochStartProofs {
			pending[key] = info
		}

		pendingHeaders := make(map[uint32]struct{}, len(t.pendingEpochStartHeaders))
		for epoch := range t.pendingEpochStartHeaders {
			pendingHeaders[epoch] = struct{}{}
		}
		t.mutPendingEpochStartData.Unlock()

		processedEpochAfter := t.processedEpoch()
		t.mutPendingEpochStartData.Lock()
		closed := t.recoveryClosed
		isCurrentGeneration := t.recoveryGeneration == generation
		t.mutPendingEpochStartData.Unlock()

		if closed {
			return processedEpochAfter, nil, nil, generation
		}
		if processedEpochBefore == processedEpochAfter && isCurrentGeneration {
			return processedEpochAfter, pending, pendingHeaders, generation
		}
	}
}

func (t *trigger) isRecoveryGenerationCurrent(generation uint64) bool {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	return !t.recoveryClosed && t.recoveryGeneration == generation
}

func (t *trigger) hasPendingEpochStartData() bool {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	return len(t.pendingEpochStartProofs) > 0 ||
		len(t.pendingEpochStartHeaders) > 0 ||
		len(t.pendingFinalityEvidence) > 0
}

// trackMissingFinalityEvidence records an epoch start meta block the node holds proofed but not yet
// final, and asks for the neighbour data that settles it. Arrival of that data re-enters the trigger
// through the pool callbacks, so registering the request is all that is needed to close the loop.
func (t *trigger) trackMissingFinalityEvidence(metaHdr data.HeaderHandler, hash string) {
	t.trackMissingFinalityEvidenceInternal(metaHdr, hash, nil)
}

func (t *trigger) trackMissingFinalityEvidenceForGeneration(
	metaHdr data.HeaderHandler,
	hash string,
	generation uint64,
) {
	t.trackMissingFinalityEvidenceInternal(metaHdr, hash, &generation)
}

func (t *trigger) trackMissingFinalityEvidenceInternal(
	metaHdr data.HeaderHandler,
	hash string,
	generation *uint64,
) {
	info := finalityEvidenceRequest{
		epoch:    metaHdr.GetEpoch(),
		nonce:    metaHdr.GetNonce(),
		round:    metaHdr.GetRound(),
		hash:     []byte(hash),
		prevHash: metaHdr.GetPrevHash(),
	}

	if !t.addPendingFinalityEvidenceForGeneration(info, generation) {
		return
	}

	log.Debug("trigger.trackMissingFinalityEvidence: epoch start meta block not held final, requesting neighbours",
		"epoch", info.epoch,
		"nonce", info.nonce,
		"hash", info.hash,
	)

	// requested off the trigger's mutex, the request handler may block on the network
	go t.requestFinalityEvidence(info)
}

// addPendingFinalityEvidence returns true if the entry was newly added
func (t *trigger) addPendingFinalityEvidence(info finalityEvidenceRequest) bool {
	return t.addPendingFinalityEvidenceForGeneration(info, nil)
}

func (t *trigger) addPendingFinalityEvidenceForGeneration(
	info finalityEvidenceRequest,
	generation *uint64,
) bool {
	key := string(info.hash)

	t.mutPendingEpochStartData.Lock()
	if t.recoveryClosed || (generation != nil && t.recoveryGeneration != *generation) {
		t.mutPendingEpochStartData.Unlock()
		return false
	}

	_, exists := t.pendingFinalityEvidence[key]
	if !exists {
		t.pendingFinalityEvidence[key] = info
	}
	delete(t.epochStartRecoveryCandidates, key)
	t.mutPendingEpochStartData.Unlock()

	if exists {
		return false
	}

	select {
	case t.chanPendingEpochStartData <- struct{}{}:
	default:
	}

	return true
}

func (t *trigger) removePendingFinalityEvidence(hash string) {
	t.mutPendingEpochStartData.Lock()
	if !t.recoveryClosed {
		delete(t.pendingFinalityEvidence, hash)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) removePendingFinalityEvidenceForGeneration(hash string, generation uint64) {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryGeneration == generation {
		delete(t.pendingFinalityEvidence, hash)
	}
	t.mutPendingEpochStartData.Unlock()
}

func (t *trigger) pendingFinalityEvidenceSnapshot() map[string]finalityEvidenceRequest {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	pending := make(map[string]finalityEvidenceRequest, len(t.pendingFinalityEvidence))
	for key, info := range t.pendingFinalityEvidence {
		pending[key] = info
	}

	return pending
}

// requestFinalityEvidence asks only for what the pools are still missing, recomputed on every pass
// so that a partially answered request is not repeated in full
func (t *trigger) requestFinalityEvidence(info finalityEvidenceRequest) {
	operations := t.finalityEvidenceRequestOperations(info)
	for _, operation := range operations {
		t.executeRecoveryRequest(operation)
	}
}

func (t *trigger) finalityEvidenceRequestOperations(info finalityEvidenceRequest) []recoveryRequestOperation {
	operations := make([]recoveryRequestOperation, 0, 2)
	ownerKey := string(info.hash)
	parent, err := t.headersPool.GetHeaderByHash(info.prevHash)
	if err != nil || check.IfNil(parent) {
		parentEpoch := uint32(0)
		if info.epoch > 0 {
			parentEpoch = info.epoch - 1
		}
		operations = append(operations, recoveryRequestOperation{
			class:        finalityRecoveryRequest,
			kind:         metaHeaderByHashRequest,
			requestKey:   recoveryOperationKey(metaHeaderByHashRequest, parentEpoch, 0, info.prevHash),
			ownerKeys:    []string{ownerKey},
			epoch:        info.epoch,
			requestEpoch: parentEpoch,
			hash:         info.prevHash,
		})
	} else if parentCanSettle(info, parent) && !t.proofsPool.HasProof(core.MetachainShardId, info.prevHash) {
		operations = append(operations, recoveryRequestOperation{
			class:        finalityRecoveryRequest,
			kind:         equivalentProofByHashRequest,
			requestKey:   recoveryOperationKey(equivalentProofByHashRequest, parent.GetEpoch(), 0, info.prevHash),
			ownerKeys:    []string{ownerKey},
			epoch:        info.epoch,
			requestEpoch: parent.GetEpoch(),
			hash:         info.prevHash,
		})
	}

	childNonce := info.nonce + 1
	children, childrenHashes, err := t.headersPool.GetHeadersByNonceAndShardId(childNonce, core.MetachainShardId)
	if err != nil {
		operations = append(operations, recoveryRequestOperation{
			class:      finalityRecoveryRequest,
			kind:       metaHeaderByNonceRequest,
			requestKey: recoveryOperationKey(metaHeaderByNonceRequest, info.epoch, childNonce, nil),
			ownerKeys:  []string{ownerKey},
			epoch:      info.epoch,
			nonce:      childNonce,
		})
		return operations
	}

	foundChildOnBranch := false
	for i, child := range children {
		if i >= len(childrenHashes) || check.IfNil(child) || !bytes.Equal(child.GetPrevHash(), info.hash) {
			continue
		}

		foundChildOnBranch = true
		if t.proofsPool.HasProof(core.MetachainShardId, childrenHashes[i]) {
			continue
		}

		operations = append(operations, recoveryRequestOperation{
			class:        finalityRecoveryRequest,
			kind:         equivalentProofByHashRequest,
			requestKey:   recoveryOperationKey(equivalentProofByHashRequest, child.GetEpoch(), childNonce, childrenHashes[i]),
			ownerKeys:    []string{ownerKey},
			epoch:        info.epoch,
			requestEpoch: child.GetEpoch(),
			hash:         childrenHashes[i],
		})
	}

	if !foundChildOnBranch {
		operations = append(operations, recoveryRequestOperation{
			class:      finalityRecoveryRequest,
			kind:       metaHeaderByNonceRequest,
			requestKey: recoveryOperationKey(metaHeaderByNonceRequest, info.epoch, childNonce, nil),
			ownerKeys:  []string{ownerKey},
			epoch:      info.epoch,
			nonce:      childNonce,
		})
	}

	return operations
}

func recoveryOperationKey(kind recoveryRequestKind, epoch uint32, nonce uint64, hash []byte) string {
	return fmt.Sprintf("%010d:%d:%020d:%x", epoch, kind, nonce, hash)
}

func (t *trigger) executeRecoveryRequest(operation recoveryRequestOperation) {
	switch operation.kind {
	case metaHeaderByHashRequest:
		t.requestHandler.RequestMetaHeaderForEpoch(operation.hash, operation.requestEpoch)
	case equivalentProofByHashRequest:
		t.requestHandler.RequestEquivalentProofByHashForEpoch(core.MetachainShardId, operation.hash, operation.requestEpoch)
	case startOfEpochMetaBlockRequest:
		t.requestHandler.RequestStartOfEpochMetaBlock(operation.epoch)
	case metaHeaderByNonceRequest:
		t.requestHandler.RequestMetaHeaderByNonceForEpoch(operation.nonce, operation.epoch)
	}
}

// parentCanSettle mirrors the non contended clause of the finality view: over a round gap the parent
// settles nothing however well proofed, so its proof is not worth asking for
func parentCanSettle(info finalityEvidenceRequest, parent data.HeaderHandler) bool {
	if parent.GetNonce()+1 != info.nonce {
		return false
	}

	return !common.IsContendedRound(info.round, parent.GetRound())
}

func (t *trigger) getPendingProofRetryInterval() time.Duration {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	return t.pendingProofRetryInterval
}

// requestPendingEpochStartProofs re-requests proofs for pending epoch-start meta headers,
// independently of metablock traffic; dormant while there is nothing pending
func (t *trigger) requestPendingEpochStartProofs(ctx context.Context) {
	// created stopped; armed only when pending work appears
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	timerActive := false
	for {
		select {
		case <-ctx.Done():
			log.Debug("requestPendingEpochStartProofs: trigger's go routine is stopping...")
			return
		case <-t.chanPendingEpochStartData:
			// Wait one interval before processing newly pending work.
			if !timerActive {
				timer.Reset(t.getPendingProofRetryInterval())
				timerActive = true
			}
		case <-timer.C:
			// ctx.Done may be ready in the same select round; do not run a pass after cancellation
			if ctx.Err() != nil {
				log.Debug("requestPendingEpochStartProofs: trigger's go routine is stopping...")
				return
			}
			timerActive = false
			if t.retryPendingEpochStartProofs() {
				timer.Reset(t.getPendingProofRetryInterval())
				timerActive = true
			}
		}
	}
}

// retryPendingEpochStartProofs runs one retry pass and returns true if entries remain pending
func (t *trigger) retryPendingEpochStartProofs() bool {
	currentEpoch, pendingProofs, pendingHeaders, generation := t.pendingEpochStartDataSnapshot()
	operations := [numRecoveryRequestClasses][]recoveryRequestOperation{}
	if currentEpoch == math.MaxUint32 {
		return t.hasPendingEpochStartData()
	}
	actionableEpoch := currentEpoch + 1

	for key, info := range pendingProofs {
		if !t.isRecoveryGenerationCurrent(generation) {
			return t.hasPendingEpochStartData()
		}

		epoch := info.epoch
		if epoch <= currentEpoch {
			t.removePendingEpochStartProofForGeneration(key, epoch, generation)
			continue
		}
		if epoch != actionableEpoch {
			continue
		}

		metaBlockHash := []byte(key)
		header, err := t.headersPool.GetHeaderByHash(metaBlockHash)
		if err != nil || check.IfNil(header) {
			t.movePendingProofToHeaderRecoveryForGeneration(key, epoch, generation)

			// Close the race in which the header returned between the failed lookup and the
			// state transition; duplicate pool insertion would not invoke receivedMetaBlock.
			header, err = t.headersPool.GetHeaderByHash(metaBlockHash)
			if err == nil && !check.IfNil(header) {
				t.addPendingProofForGeneration(metaBlockHash, epoch, &generation)
				continue
			}

			if !t.isRecoveryGenerationCurrent(generation) {
				return t.hasPendingEpochStartData()
			}
			operations[headerRecoveryRequest] = append(operations[headerRecoveryRequest], recoveryRequestOperation{
				class:      headerRecoveryRequest,
				kind:       startOfEpochMetaBlockRequest,
				requestKey: recoveryOperationKey(startOfEpochMetaBlockRequest, epoch, 0, nil),
				epoch:      epoch,
			})
			continue
		}

		_, err = t.proofsPool.GetProof(core.MetachainShardId, metaBlockHash)
		if err != nil {
			operations[proofRecoveryRequest] = append(operations[proofRecoveryRequest], recoveryRequestOperation{
				class:        proofRecoveryRequest,
				kind:         equivalentProofByHashRequest,
				requestKey:   recoveryOperationKey(equivalentProofByHashRequest, epoch, 0, metaBlockHash),
				ownerKeys:    []string{key},
				epoch:        epoch,
				requestEpoch: epoch,
				sequence:     info.requestSequence,
				hash:         metaBlockHash,
			})
			continue
		}

		if !t.isRecoveryGenerationCurrent(generation) {
			return t.hasPendingEpochStartData()
		}
		if !t.processMetaHeaderWithProofForGeneration(header, metaBlockHash, generation) {
			return t.hasPendingEpochStartData()
		}
		t.removePendingEpochStartProofForGeneration(key, epoch, generation)
		t.removePendingEpochStartHeaderForGeneration(epoch, generation)
	}

	for epoch := range pendingHeaders {
		if !t.isRecoveryGenerationCurrent(generation) {
			return t.hasPendingEpochStartData()
		}

		if epoch <= currentEpoch {
			t.removePendingEpochStartHeaderForGeneration(epoch, generation)
			continue
		}
		if epoch != actionableEpoch {
			continue
		}

		operations[headerRecoveryRequest] = append(operations[headerRecoveryRequest], recoveryRequestOperation{
			class:      headerRecoveryRequest,
			kind:       startOfEpochMetaBlockRequest,
			requestKey: recoveryOperationKey(startOfEpochMetaBlockRequest, epoch, 0, nil),
			epoch:      epoch,
		})
	}

	if t.isRecoveryGenerationCurrent(generation) {
		operations[finalityRecoveryRequest] = t.retryPendingFinalityEvidence(actionableEpoch, generation)
		t.executePendingRecoveryRequests(operations, generation)
	}

	return t.hasPendingEpochStartData()
}

// retryPendingFinalityEvidence reconstructs retained candidates and collects their missing requests.
func (t *trigger) retryPendingFinalityEvidence(actionableEpoch uint32, generation uint64) []recoveryRequestOperation {
	pendingEvidence := t.pendingFinalityEvidenceSnapshot()
	for key, info := range pendingEvidence {
		if info.epoch < actionableEpoch {
			t.removePendingFinalityEvidenceForGeneration(key, generation)
			delete(pendingEvidence, key)
		}
	}
	candidates := t.selectPendingFinalityCandidates(pendingEvidence, actionableEpoch)

	operations := make([]recoveryRequestOperation, 0, len(candidates))
	needsReevaluation := false
	for _, candidate := range candidates {
		key := candidate.key
		info := candidate.info
		if !t.markFinalityCandidateInspected(candidate, generation) {
			return operations
		}
		if !t.isPendingFinalityEvidence(key, info.epoch, generation) {
			continue
		}

		header, err := t.headersPool.GetHeaderByHash(info.hash)
		if err != nil || check.IfNil(header) {
			operations = append(operations, recoveryRequestOperation{
				class:        finalityRecoveryRequest,
				kind:         metaHeaderByHashRequest,
				requestKey:   recoveryOperationKey(metaHeaderByHashRequest, info.epoch, 0, info.hash),
				ownerKeys:    []string{key},
				epoch:        info.epoch,
				requestEpoch: info.epoch,
				hash:         info.hash,
			})
			continue
		}
		if !t.proofsPool.HasProof(core.MetachainShardId, info.hash) {
			operations = append(operations, recoveryRequestOperation{
				class:        finalityRecoveryRequest,
				kind:         equivalentProofByHashRequest,
				requestKey:   recoveryOperationKey(equivalentProofByHashRequest, info.epoch, 0, info.hash),
				ownerKeys:    []string{key},
				epoch:        info.epoch,
				requestEpoch: info.epoch,
				hash:         info.hash,
			})
			continue
		}

		if t.isEpochStartCandidateRetained(key) {
			needsReevaluation = true
		} else if !t.processMetaHeaderWithProofForGeneration(header, info.hash, generation) {
			return operations
		}
		if !t.isPendingFinalityEvidence(key, info.epoch, generation) {
			continue
		}

		operations = append(operations, t.finalityEvidenceRequestOperations(info)...)
	}
	if needsReevaluation && t.isRecoveryGenerationCurrent(generation) {
		t.mutTrigger.Lock()
		if t.isRecoveryGenerationCurrent(generation) {
			t.updateTriggerFromMeta()
		}
		t.mutTrigger.Unlock()
	}

	return operations
}

func (t *trigger) selectPendingFinalityCandidates(
	pending map[string]finalityEvidenceRequest,
	actionableEpoch uint32,
) []pendingFinalityCandidate {
	candidates := make([]pendingFinalityCandidate, 0, len(pending))
	for key, info := range pending {
		if info.epoch != actionableEpoch {
			continue
		}

		candidate := pendingFinalityCandidate{
			key:     key,
			sortKey: fmt.Sprintf("%010d:%020d:%x", info.epoch, info.nonce, info.hash),
			info:    info,
		}
		candidates = append(candidates, candidate)
	}

	candidates = t.orderFinalityCandidates(candidates)
	return candidates[:min(maxPendingProofRequestsPerPass, len(candidates))]
}

func (t *trigger) orderFinalityCandidates(candidates []pendingFinalityCandidate) []pendingFinalityCandidate {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].sortKey < candidates[j].sortKey
	})
	if len(candidates) < 2 {
		return candidates
	}

	t.mutPendingEpochStartData.Lock()
	cursor := t.finalityCandidateCursor
	t.mutPendingEpochStartData.Unlock()
	start := sort.Search(len(candidates), func(index int) bool {
		return candidates[index].sortKey > cursor
	})
	if start == 0 {
		return candidates
	}
	if start == len(candidates) {
		start = 0
	}

	ordered := make([]pendingFinalityCandidate, 0, len(candidates))
	ordered = append(ordered, candidates[start:]...)
	ordered = append(ordered, candidates[:start]...)
	return ordered
}

func (t *trigger) markFinalityCandidateInspected(
	candidate pendingFinalityCandidate,
	generation uint64,
) bool {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	if t.recoveryClosed || t.recoveryGeneration != generation {
		return false
	}
	info, found := t.pendingFinalityEvidence[candidate.key]
	if !found || info.epoch != candidate.info.epoch {
		return true
	}

	t.finalityCandidateCursor = candidate.sortKey
	return true
}

func (t *trigger) isEpochStartCandidateRetained(key string) bool {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	_, found := t.mapEpochStartHdrs[key]
	return found
}

func (t *trigger) executePendingRecoveryRequests(
	operations [numRecoveryRequestClasses][]recoveryRequestOperation,
	generation uint64,
) {
	operations = deduplicateRecoveryRequests(operations)
	for class := recoveryRequestClass(0); class < numRecoveryRequestClasses; class++ {
		operations[class] = t.orderedRecoveryRequests(operations[class], class)
	}

	indices := [numRecoveryRequestClasses]int{}
	t.executeRecoveryRequestSet(operations, &indices, maxPendingProofRequestsPerPass, generation)
}

func (t *trigger) executeRecoveryRequestSet(
	operations [numRecoveryRequestClasses][]recoveryRequestOperation,
	indices *[numRecoveryRequestClasses]int,
	limit int,
	generation uint64,
) int {
	numRequests := 0
	for numRequests < limit {
		madeProgress := false
		for class := recoveryRequestClass(0); class < numRecoveryRequestClasses; class++ {
			for indices[class] < len(operations[class]) {
				operation := operations[class][indices[class]]
				indices[class]++
				if !t.markRecoveryRequest(operation, generation) {
					continue
				}

				t.executeRecoveryRequest(operation)
				numRequests++
				madeProgress = true
				break
			}

			if numRequests == limit {
				break
			}
		}

		if !madeProgress {
			break
		}
	}

	return numRequests
}

func (t *trigger) orderedRecoveryRequests(
	operations []recoveryRequestOperation,
	class recoveryRequestClass,
) []recoveryRequestOperation {
	sort.Slice(operations, func(i, j int) bool {
		if class == proofRecoveryRequest && operations[i].sequence != operations[j].sequence {
			return operations[i].sequence < operations[j].sequence
		}
		return operations[i].requestKey < operations[j].requestKey
	})
	if class == proofRecoveryRequest {
		return operations
	}
	if len(operations) < 2 {
		return operations
	}

	t.mutPendingEpochStartData.Lock()
	cursor := t.recoveryRequestCursors[class]
	t.mutPendingEpochStartData.Unlock()
	start := sort.Search(len(operations), func(index int) bool {
		return operations[index].requestKey > cursor
	})
	if start == 0 || start == len(operations) {
		if start == len(operations) {
			start = 0
		} else {
			return operations
		}
	}

	ordered := make([]recoveryRequestOperation, 0, len(operations))
	ordered = append(ordered, operations[start:]...)
	ordered = append(ordered, operations[:start]...)
	return ordered
}

func deduplicateRecoveryRequests(
	operations [numRecoveryRequestClasses][]recoveryRequestOperation,
) [numRecoveryRequestClasses][]recoveryRequestOperation {
	result := [numRecoveryRequestClasses][]recoveryRequestOperation{}
	for class := recoveryRequestClass(0); class < numRecoveryRequestClasses; class++ {
		indicesByRequest := make(map[string]int, len(operations[class]))
		for _, operation := range operations[class] {
			index, found := indicesByRequest[operation.requestKey]
			if !found {
				indicesByRequest[operation.requestKey] = len(result[class])
				result[class] = append(result[class], operation)
				continue
			}

			existing := &result[class][index]
			existing.ownerKeys = appendUniqueStrings(existing.ownerKeys, operation.ownerKeys...)
			if operation.sequence < existing.sequence {
				existing.sequence = operation.sequence
			}
		}
	}

	return result
}

func appendUniqueStrings(destination []string, values ...string) []string {
	for _, value := range values {
		found := false
		for _, existing := range destination {
			if existing == value {
				found = true
				break
			}
		}
		if !found {
			destination = append(destination, value)
		}
	}

	return destination
}

func (t *trigger) markRecoveryRequest(operation recoveryRequestOperation, generation uint64) bool {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	if t.recoveryClosed || t.recoveryGeneration != generation {
		return false
	}

	switch operation.class {
	case finalityRecoveryRequest:
		if !hasPendingFinalityOwner(t.pendingFinalityEvidence, operation.ownerKeys) {
			return false
		}
	case proofRecoveryRequest:
		ownerKey, info, found := pendingProofOwner(t.pendingEpochStartProofs, operation.ownerKeys)
		if !found {
			return false
		}
		t.nextProofRequestSequence++
		info.requestSequence = t.nextProofRequestSequence
		t.pendingEpochStartProofs[ownerKey] = info
	case headerRecoveryRequest:
		if _, found := t.pendingEpochStartHeaders[operation.epoch]; !found {
			return false
		}
	}

	t.recoveryRequestCursors[operation.class] = operation.requestKey
	return true
}

func hasPendingFinalityOwner(
	pending map[string]finalityEvidenceRequest,
	ownerKeys []string,
) bool {
	for _, ownerKey := range ownerKeys {
		if _, found := pending[ownerKey]; found {
			return true
		}
	}

	return false
}

func pendingProofOwner(
	pending map[string]pendingEpochStartProof,
	ownerKeys []string,
) (string, pendingEpochStartProof, bool) {
	for _, ownerKey := range ownerKeys {
		info, found := pending[ownerKey]
		if found {
			return ownerKey, info, true
		}
	}

	return "", pendingEpochStartProof{}, false
}

func (t *trigger) isPendingFinalityEvidence(key string, epoch uint32, generation uint64) bool {
	t.mutPendingEpochStartData.Lock()
	defer t.mutPendingEpochStartData.Unlock()

	if t.recoveryClosed || t.recoveryGeneration != generation {
		return false
	}
	info, found := t.pendingFinalityEvidence[key]
	return found && info.epoch == epoch
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving checks if trigger can be updated
func (t *trigger) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	t.processReceivedMetaBlock(headerHandler, metaBlockHash, true)
}

func (t *trigger) processReceivedMetaBlock(
	headerHandler data.HeaderHandler,
	metaBlockHash []byte,
	requestProofImmediately bool,
) bool {
	if headerHandler.GetShardID() != core.MetachainShardId {
		return false
	}
	if !t.callbackAdmission.IsSet() {
		return false
	}

	select {
	case t.chanMetaBlockReceived <- struct{}{}:
	default:
	}

	log.Debug("received meta header in trigger", "header hash", metaBlockHash)
	if t.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, headerHandler.GetEpoch()) {
		proof, err := t.proofsPool.GetProof(headerHandler.GetShardID(), metaBlockHash)
		if err != nil {
			metaHdr, ok := headerHandler.(data.MetaHeaderHandler)
			if !ok || !metaHdr.IsStartOfEpochBlock() {
				return false
			}
			currentEpoch := t.Epoch()
			usesSupernovaFinality := t.isSupernovaEpochStartFinalityEnabled(headerHandler.GetEpoch())
			if usesSupernovaFinality {
				currentEpoch = t.processedEpoch()
			}
			if metaHdr.GetEpoch() > currentEpoch {
				if usesSupernovaFinality &&
					(currentEpoch == math.MaxUint32 || metaHdr.GetEpoch() != currentEpoch+1) {
					return false
				}
				log.Debug("proof not found for epoch start meta header, tracking recovery",
					"header hash", metaBlockHash,
					"epoch", headerHandler.GetEpoch(),
				)
				// record before requesting, so a fast response cannot complete unpended
				shouldRequestImmediately := requestProofImmediately &&
					currentEpoch != math.MaxUint32 && metaHdr.GetEpoch() == currentEpoch+1
				var added bool
				if usesSupernovaFinality {
					added = t.addActionablePendingProof(
						metaBlockHash,
						metaHdr.GetEpoch(),
						nil,
						!shouldRequestImmediately,
					)
				} else if shouldRequestImmediately {
					added = t.addPendingProof(metaBlockHash, metaHdr.GetEpoch())
				} else {
					added = t.addPrioritizedPendingProof(metaBlockHash, metaHdr.GetEpoch())
				}
				if added {
					if !shouldRequestImmediately {
						return false
					}
					// stamp the target epoch: the requester drops requests labeled before Andromeda activation
					go t.requestHandler.RequestEquivalentProofByHashForEpoch(core.MetachainShardId, metaBlockHash, metaHdr.GetEpoch())
					return true
				}
			}
			return false
		}

		usesSupernovaFinality := t.isSupernovaEpochStartFinalityEnabled(headerHandler.GetEpoch())
		if !usesSupernovaFinality {
			t.removePendingEpochStartHeader(headerHandler.GetEpoch())
		}
		if !t.processMetaHeaderWithProof(headerHandler, proof.GetHeaderHash()) {
			return false
		}
		if usesSupernovaFinality && headerHandler.IsStartOfEpochBlock() {
			t.removePendingEpochStartHeader(headerHandler.GetEpoch())
		}
		return false
	}

	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()
	if !t.callbackAdmission.IsSet() {
		return false
	}

	t.checkMetaHeaderForEpochTriggerLegacy(headerHandler, metaBlockHash)

	return false
}

// OnBootstrapCompleted enables trigger callbacks and replays headers received during bootstrap.
func (t *trigger) OnBootstrapCompleted() {
	t.mutPendingEpochStartData.Lock()
	if t.recoveryClosed || t.callbackAdmission.IsSet() {
		t.mutPendingEpochStartData.Unlock()
		return
	}
	t.callbackAdmission.SetValue(true)
	t.mutPendingEpochStartData.Unlock()

	numReplayed := t.replayMetaHeadersFromPool()
	log.Debug("trigger callback admission opened",
		"replayed meta headers", numReplayed,
		"epoch", t.Epoch(),
	)
}

type metaHeaderForReplay struct {
	header data.HeaderHandler
	hash   []byte
}

type pooledEpochStartCandidate struct {
	header data.HeaderHandler
	hash   []byte
}

type triggerRollbackState struct {
	epochStartShardHeader       data.HeaderHandler
	epochStartMeta              data.HeaderHandler
	metaEpoch                   uint32
	epochStartRound             uint64
	epochFinalityAttestingRound uint64
	epochMetaBlockHash          []byte
	isEpochStart                bool
	newEpochHdrReceived         bool
}

func (t *trigger) replayMetaHeadersFromPool() int {
	nonces := t.headersPool.Nonces(core.MetachainShardId)
	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	numReplayed := 0
	numImmediateProofRequests := 0
	for _, nonce := range nonces {
		headers, hashes, err := t.headersPool.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
		if err != nil {
			continue
		}

		competitors := make([]metaHeaderForReplay, 0, len(headers))
		for index, header := range headers {
			if index >= len(hashes) || check.IfNil(header) {
				continue
			}

			competitors = append(competitors, metaHeaderForReplay{
				header: header,
				hash:   hashes[index],
			})
		}
		sort.Slice(competitors, func(i, j int) bool {
			if competitors[i].header.GetRound() == competitors[j].header.GetRound() {
				return bytes.Compare(competitors[i].hash, competitors[j].hash) < 0
			}

			return competitors[i].header.GetRound() < competitors[j].header.GetRound()
		})

		for _, competitor := range competitors {
			requestProofImmediately := numImmediateProofRequests < maxPendingProofRequestsPerPass
			if t.processReceivedMetaBlock(competitor.header, competitor.hash, requestProofImmediately) {
				numImmediateProofRequests++
			}
			numReplayed++
		}
	}

	return numReplayed
}

func (t *trigger) checkMetaHeaderForEpochTriggerEquivalentProofs(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaHdr, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return
	}

	log.Debug("trigger.checkMetaHeaderForEpochTriggerEquivalentProofs", "metaHdr epoch", metaHdr.GetEpoch(), "metaBlockHash", metaBlockHash)
	if !t.shouldUpdateTrigger(metaHdr, metaBlockHash) {
		return
	}

	log.Debug("trigger.updateTriggerHeaderData")
	t.updateTriggerHeaderData(metaHdr, metaBlockHash)
	t.updateTriggerFromMeta()
}

func (t *trigger) checkMetaHeaderForEpochTriggerLegacy(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaHdr, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return
	}

	if !t.isPreviousEpochStartMetaBlock(metaHdr, metaBlockHash) {
		_, ok = t.mapFinalizedEpochs[metaHdr.GetEpoch()]
		if t.metaEpoch == headerHandler.GetEpoch() && ok {
			t.changeEpochFinalityAttestingRoundIfNeeded(metaHdr, metaBlockHash)
			return
		}
	}

	if !t.shouldUpdateTrigger(metaHdr, metaBlockHash) {
		return
	}

	t.updateTriggerHeaderData(metaHdr, metaBlockHash)
	t.updateTriggerFromMeta()
}

func (t *trigger) shouldUpdateTrigger(metaHdr data.MetaHeaderHandler, metaBlockHash []byte) bool {
	if !t.newEpochHdrReceived && !metaHdr.IsStartOfEpochBlock() {
		return false
	}

	currentEpoch := t.epoch
	if metaHdr.IsStartOfEpochBlock() && t.isSupernovaEpochStartFinalityEnabled(metaHdr.GetEpoch()) {
		currentEpoch = t.processedEpochLocked()
	}
	isMetaStartOfEpochForCurrentOrOlderEpoch := metaHdr.GetEpoch() <= currentEpoch && metaHdr.IsStartOfEpochBlock()
	if isMetaStartOfEpochForCurrentOrOlderEpoch {
		return false
	}

	_, foundHdrInMap := t.mapHashHdr[string(metaBlockHash)]
	_, foundHdrInEpochStartMap := t.mapEpochStartHdrs[string(metaBlockHash)]

	finalizedMetaBlockHash, ok := t.mapFinalizedEpochs[metaHdr.GetEpoch()]
	foundHdrInFinalizedMap := ok && bytes.Equal(metaBlockHash, []byte(finalizedMetaBlockHash))

	if foundHdrInMap && foundHdrInEpochStartMap && foundHdrInFinalizedMap {
		return false
	}

	return true
}

func (t *trigger) updateTriggerHeaderData(metaHdr data.MetaHeaderHandler, metaBlockHash []byte) {
	if metaHdr.IsStartOfEpochBlock() {
		t.newEpochHdrReceived = true
		t.mapEpochStartHdrs[string(metaBlockHash)] = metaHdr
		// waiting for late broadcast of mini blocks and transactions to be done and received
		wait := t.getExtraDelayForRequestsBlockInfo()
		roundDifferences := t.roundHandler.Index() - int64(metaHdr.GetRound())
		if roundDifferences > 1 {
			wait = 0
		}

		log.Debug("updateTriggerHeaderData: waiting for late broadcast of mini blocks and transactions",
			"wait time", wait,
		)

		time.Sleep(wait)
	}

	t.mapHashHdr[string(metaBlockHash)] = metaHdr
	t.mapNonceHashes[metaHdr.GetNonce()] = append(t.mapNonceHashes[metaHdr.GetNonce()], string(metaBlockHash))
}

// call only if mutex is locked before
func (t *trigger) isPreviousEpochStartMetaBlock(metaBlock data.MetaHeaderHandler, metaBlockHash []byte) bool {
	metaHdrHashesWithNonce := t.mapNonceHashes[metaBlock.GetNonce()+1]
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
	processedEpoch := t.processedEpochLocked()
	actionableEpoch := uint32(0)
	hasActionableEpoch := processedEpoch != math.MaxUint32
	if hasActionableEpoch {
		actionableEpoch = processedEpoch + 1
	}

	sortedMetaInfo := make(metaInfoSlice, 0, len(t.mapEpochStartHdrs))
	for hash, hdr := range t.mapEpochStartHdrs {
		if t.isSupernovaEpochStartFinalityEnabled(hdr.GetEpoch()) &&
			(!hasActionableEpoch || hdr.GetEpoch() != actionableEpoch) {
			continue
		}
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
		log.Debug("trigger.updateTriggerFromMeta", "canActivateEpochStart", canActivateEpochStart, "finalityAttestingRound", finalityAttestingRound)
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
			// the epoch is settled, so no candidate of it needs its neighbourhood any more
			t.removePendingEpochStartDataForEpochOrOlder(currMetaInfo.hdr.GetEpoch())
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
	return t.metaBlockValidity(hash, metaHdr) == metaBlockValidityValid
}

// call only if mutex is locked before
func (t *trigger) metaBlockValidity(hash string, metaHdr data.HeaderHandler) metaBlockValidity {
	currHdr := metaHdr
	numSteps := min(metaHdr.GetNonce(), t.validity)
	for step := uint64(0); step < numSteps; step++ {
		nonce := metaHdr.GetNonce() - step - 1
		neededHdr, err := t.getHeaderWithNonceAndHash(nonce, currHdr.GetPrevHash())
		if err != nil {
			log.Debug("isMetaBlockValid.getHeaderWithNonceAndHash", "hash", hash, "error", err.Error())
			return metaBlockValidityIncomplete
		}

		err = t.headerValidator.IsHeaderConstructionValid(currHdr, neededHdr)
		if err != nil {
			log.Debug("isMetaBlockValid.IsHeaderConstructionValid", "hash", hash, "error", err.Error())
			return metaBlockValidityInvalid
		}

		currHdr = neededHdr
	}

	return metaBlockValidityValid
}

func (t *trigger) isMetaBlockFinal(hash string, metaHdr data.HeaderHandler) (bool, uint64) {
	if !t.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, metaHdr.GetEpoch()) {
		return t.isMetaBlockFinalLegacy(hash, metaHdr)
	}

	// under Supernova a contended epoch start must not
	// activate the trigger until the node holds it final
	if t.isSupernovaEpochStartFinalityEnabled(metaHdr.GetEpoch()) {
		if !t.metaFinalityView.IsMetaHeaderHeldFinal(metaHdr, []byte(hash)) {
			// this verdict is decided by the pools alone, so a node whose sync fell behind the
			// epoch start meta block's neighbourhood never recovers it unless it asks for it
			t.trackMissingFinalityEvidence(metaHdr, hash)
			return false, 0
		}

		return true, metaHdr.GetRound()
	}

	hasProof := t.proofsPool.HasProof(metaHdr.GetShardID(), []byte(hash))
	if !hasProof {
		return false, 0
	}

	return true, metaHdr.GetRound()
}

func (t *trigger) isSupernovaEpochStartFinalityEnabled(epoch uint32) bool {
	return t.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, epoch) &&
		epoch > t.enableEpochsHandler.GetActivationEpoch(common.SupernovaFlag)
}

func (t *trigger) isMetaBlockFinalLegacy(_ string, metaHdr data.HeaderHandler) (bool, uint64) {
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

	if t.enableEpochsHandler.IsFlagEnabledInEpoch(common.RefactorPeersMiniBlocksFlag, metaHdr.GetEpoch()) {
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

	if _, alreadyPrepared := t.mapPreparedEpochStartHdrs[hash]; !alreadyPrepared {
		t.epochStartNotifier.NotifyAllPrepare(metaHdr, blockBody)
		t.mapPreparedEpochStartHdrs[hash] = struct{}{}
	}

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
	neededHdr, ok := peekedData.(data.MetaHeaderHandler)
	if ok {
		t.mapHashHdr[string(neededHash)] = neededHdr
		t.mapNonceHashes[neededHdr.GetNonce()] = append(t.mapNonceHashes[neededHdr.GetNonce()], string(neededHash))
		return neededHdr
	}

	return nil
}

// call only if mutex is locked before
func (t *trigger) getHeaderWithHashFromStorage(neededHash []byte) data.HeaderHandler {
	storageData, err := t.metaHdrStorage.Get(neededHash)
	if err == nil {
		neededHdr, err := process.UnmarshalMetaHeader(t.marshaller, storageData)
		if err == nil {
			t.mapHashHdr[string(neededHash)] = neededHdr
			t.mapNonceHashes[neededHdr.GetNonce()] = append(t.mapNonceHashes[neededHdr.GetNonce()], string(neededHash))
			return neededHdr
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
	var chosenMeta data.MetaHeaderHandler
	for i, header := range headers {
		if !bytes.Equal(header.GetPrevHash(), prevHash) {
			continue
		}

		hdrWithNonce, ok := header.(data.MetaHeaderHandler)
		if !ok {
			continue
		}

		if lowestRound > hdrWithNonce.GetRound() {
			lowestRound = hdrWithNonce.GetRound()
			chosenMeta = hdrWithNonce
		}

		t.mapHashHdr[string(hashes[i])] = hdrWithNonce
		t.mapNonceHashes[hdrWithNonce.GetNonce()] = append(t.mapNonceHashes[hdrWithNonce.GetNonce()], string(hashes[i]))
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

	shardHdr, ok := header.(data.ShardHeaderHandler)
	if !ok {
		t.mutTrigger.Unlock()
		return
	}

	if !shardHdr.IsStartOfEpochBlock() {
		t.mutTrigger.Unlock()
		return
	}
	previousEpoch := t.epoch

	t.appStatusHandler.SetUInt64Value(common.MetricRoundAtEpochStart, shardHdr.GetRound())
	t.appStatusHandler.SetUInt64Value(common.MetricNonceAtEpochStart, shardHdr.GetNonce())

	t.epoch = shardHdr.GetEpoch()
	t.removePendingEpochStartDataForEpochOrOlder(t.epoch)
	if t.metaEpoch < t.epoch {
		t.metaEpoch = t.epoch
		t.epochMetaBlockHash = shardHdr.GetEpochStartMetaHash()
	}

	t.isEpochStart = false
	t.newEpochHdrReceived = false
	t.epochStartShardHeader = shardHdr
	finishedStartOfEpochMetaHdrs := t.getAllFinishedStartOfEpochMetaHdrs()

	t.epochStartNotifier.NotifyAll(shardHdr)

	t.mapHashHdr = make(map[string]data.HeaderHandler)
	t.mapNonceHashes = make(map[uint64][]string)
	t.mapEpochStartHdrs = make(map[string]data.HeaderHandler)
	t.mapFinalizedEpochs = make(map[uint32]string)
	t.mapPreparedEpochStartHdrs = make(map[string]struct{})

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

	shouldReconsider := shardHdr.GetEpoch() > previousEpoch && shardHdr.GetEpoch() < math.MaxUint32 &&
		t.isSupernovaEpochStartFinalityEnabled(shardHdr.GetEpoch()+1)
	var recoveryGeneration uint64
	if shouldReconsider {
		t.mutPendingEpochStartData.Lock()
		shouldReconsider = !t.recoveryClosed
		recoveryGeneration = t.recoveryGeneration
		t.mutPendingEpochStartData.Unlock()
	}
	var actionableEpoch uint32
	if shouldReconsider {
		actionableEpoch = shardHdr.GetEpoch() + 1
	}
	t.mutTrigger.Unlock()

	if shouldReconsider {
		go t.reconsiderPooledEpochStart(actionableEpoch, recoveryGeneration)
	}
}

func (t *trigger) reconsiderPooledEpochStart(actionableEpoch uint32, generation uint64) {
	if !t.isRecoveryGenerationCurrent(generation) || !t.isCurrentActionableEpoch(actionableEpoch) {
		return
	}

	var candidates [3]pooledEpochStartCandidate
	var candidateFound [3]bool
	for _, nonce := range t.headersPool.Nonces(core.MetachainShardId) {
		headers, hashes, err := t.headersPool.GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)
		if err != nil {
			continue
		}

		for index, header := range headers {
			if index >= len(hashes) || check.IfNil(header) || !header.IsStartOfEpochBlock() ||
				header.GetEpoch() != actionableEpoch {
				continue
			}
			if _, ok := header.(data.MetaHeaderHandler); !ok {
				continue
			}

			group := 2
			if t.proofsPool.HasProof(core.MetachainShardId, hashes[index]) {
				group = 1
				if t.metaFinalityView.IsMetaHeaderHeldFinal(header, hashes[index]) {
					group = 0
				}
			}

			if !candidateFound[group] ||
				isPreferredPooledCandidate(header, hashes[index], candidates[group]) {
				candidates[group] = pooledEpochStartCandidate{
					header: header,
					hash:   bytes.Clone(hashes[index]),
				}
				candidateFound[group] = true
			}
		}
	}

	var selected *pooledEpochStartCandidate
	for index := range candidates {
		if candidateFound[index] {
			selected = &candidates[index]
			break
		}
	}
	if selected == nil || !t.isRecoveryGenerationCurrent(generation) ||
		!t.isCurrentActionableEpoch(actionableEpoch) {
		return
	}

	_, err := t.proofsPool.GetProof(core.MetachainShardId, selected.hash)
	if err != nil {
		t.addActionablePendingProof(selected.hash, actionableEpoch, &generation, false)
		return
	}

	t.processMetaHeaderWithProofForGeneration(selected.header, selected.hash, generation)
}

func isPreferredPooledCandidate(
	header data.HeaderHandler,
	hash []byte,
	current pooledEpochStartCandidate,
) bool {
	if header.GetNonce() != current.header.GetNonce() {
		return header.GetNonce() < current.header.GetNonce()
	}
	if header.GetRound() != current.header.GetRound() {
		return header.GetRound() < current.header.GetRound()
	}

	return bytes.Compare(hash, current.hash) < 0
}

func (t *trigger) isCurrentActionableEpoch(epoch uint32) bool {
	actionableEpoch, ok := t.actionableEpoch()
	return ok && epoch == actionableEpoch
}

// RevertStateToBlock will revert the state of the trigger to the current block
func (t *trigger) RevertStateToBlock(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return epochStart.ErrNilHeaderHandler
	}

	t.mutTrigger.Lock()

	currentHeaderHash, err := core.CalculateHash(t.marshaller, t.hasher, header)
	if err != nil {
		log.Warn("RevertStateToBlock error on hashing", "error", err)
		t.mutTrigger.Unlock()
		return err
	}

	if !bytes.Equal(t.epochStartShardHeader.GetPrevHash(), currentHeaderHash) {
		t.mutTrigger.Unlock()
		return nil
	}

	log.Debug("trigger.RevertStateToBlock behind start of epoch block")

	// the revert target's epoch start block is necessarily its stored ancestor: skipped empty
	// epochs resolve exactly, and a miss means corruption that fabricated state would hide
	rolledBackEpochStartHeader, isShardEpochStart := t.epochStartShardHeader.(data.ShardHeaderHandler)
	usesSupernovaRecovery := isShardEpochStart &&
		t.isSupernovaEpochStartFinalityEnabled(rolledBackEpochStartHeader.GetEpoch())

	prevEpoch := header.GetEpoch()
	if prevEpoch == 0 && !usesSupernovaRecovery {
		t.epochStartShardHeader = &block.Header{}
		t.isEpochStart = true
		t.newEpochHdrReceived = true
		log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)
		t.mutTrigger.Unlock()

		return nil
	}

	var shardHdr data.ShardHeaderHandler
	if prevEpoch == 0 {
		shardHdr = &block.Header{}
	} else {
		prevEpochStartIdentifier := core.EpochStartIdentifier(prevEpoch)
		shardHdrBuff, err := t.shardHdrStorage.SearchFirst([]byte(prevEpochStartIdentifier))
		if err != nil {
			log.Warn("RevertStateToBlock previous epoch start header not found", "epoch", prevEpoch, "err", err)
			t.mutTrigger.Unlock()
			return err
		}

		shardHdr, err = process.UnmarshalShardHeader(t.marshaller, shardHdrBuff)
		if err != nil {
			log.Warn("RevertStateToBlock unmarshal error", "err", err)
			t.mutTrigger.Unlock()
			return err
		}
	}

	if !usesSupernovaRecovery {
		t.removeRolledBackEpochStartHeaderFromStorageForEpoch(t.epochStartShardHeader.GetEpoch())
		t.epochStartShardHeader = shardHdr
		t.isEpochStart = true
		t.newEpochHdrReceived = true
		log.Debug("trigger.RevertStateToBlock", "isEpochStart", t.isEpochStart)
		t.mutTrigger.Unlock()

		return nil
	}

	recoveryHash := bytes.Clone(rolledBackEpochStartHeader.GetEpochStartMetaHash())
	recoveryEpoch := rolledBackEpochStartHeader.GetEpoch()
	previousState := t.rollbackStateSnapshot()

	t.epochStartShardHeader = shardHdr
	t.restorePreActivationStateForEpoch(shardHdr.GetEpoch(), shardHdr)
	err = t.saveState(t.triggerStateKey)
	if err != nil {
		t.restoreRollbackState(previousState)
		t.mutTrigger.Unlock()
		return err
	}

	t.removeRolledBackEpochStartHeaderFromStorageForEpoch(recoveryEpoch)
	t.mapHashHdr = make(map[string]data.HeaderHandler)
	t.mapNonceHashes = make(map[uint64][]string)
	t.mapEpochStartHdrs = make(map[string]data.HeaderHandler)
	t.mapFinalizedEpochs = make(map[uint32]string)
	t.mapPreparedEpochStartHdrs = make(map[string]struct{})
	t.resetMissingPreparationData()
	t.resetPendingEpochStartData()
	var registered bool
	var generation uint64
	if len(recoveryHash) > 0 {
		registered, generation = t.registerEpochStartRecoveryCandidate(recoveryHash, recoveryEpoch)
	}
	t.mutTrigger.Unlock()

	if registered {
		go t.recoverEpochStartCandidate(recoveryHash, recoveryEpoch, generation)
	}

	return nil
}

func (t *trigger) removeRolledBackEpochStartHeaderFromStorageForEpoch(epoch uint32) {
	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	errNotCritical := t.shardHdrStorage.Remove([]byte(epochStartIdentifier))
	if errNotCritical != nil {
		log.Warn("RevertStateToBlock remove from header storage error", "err", errNotCritical)
	}
}

func (t *trigger) rollbackStateSnapshot() triggerRollbackState {
	return triggerRollbackState{
		epochStartShardHeader:       t.epochStartShardHeader,
		epochStartMeta:              t.epochStartMeta,
		metaEpoch:                   t.metaEpoch,
		epochStartRound:             t.epochStartRound,
		epochFinalityAttestingRound: t.epochFinalityAttestingRound,
		epochMetaBlockHash:          bytes.Clone(t.epochMetaBlockHash),
		isEpochStart:                t.isEpochStart,
		newEpochHdrReceived:         t.newEpochHdrReceived,
	}
}

func (t *trigger) restoreRollbackState(state triggerRollbackState) {
	t.epochStartShardHeader = state.epochStartShardHeader
	t.epochStartMeta = state.epochStartMeta
	t.metaEpoch = state.metaEpoch
	t.epochStartRound = state.epochStartRound
	t.epochFinalityAttestingRound = state.epochFinalityAttestingRound
	t.epochMetaBlockHash = state.epochMetaBlockHash
	t.isEpochStart = state.isEpochStart
	t.newEpochHdrReceived = state.newEpochHdrReceived
}

func (t *trigger) resetMissingPreparationData() {
	t.mutMissingMiniBlocks.Lock()
	t.mapMissingMiniBlocks = make(map[string]uint32)
	t.mutMissingMiniBlocks.Unlock()

	t.mutMissingValidatorsInfo.Lock()
	t.mapMissingValidatorsInfo = make(map[string]uint32)
	t.mutMissingValidatorsInfo.Unlock()
}

// DisarmDeadEpochStartActivation reverts an activation armed by a dead epoch start meta block so
// the canonical sibling can re-arm; covers received-time arming only. Returns true if disarmed.
func (t *trigger) DisarmDeadEpochStartActivation(epoch uint32, deadEpochStartHash []byte) bool {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	finalizedHash, ok := t.mapFinalizedEpochs[epoch]
	if !ok || finalizedHash != string(deadEpochStartHash) {
		return false
	}

	log.Warn("trigger.DisarmDeadEpochStartActivation",
		"epoch", epoch,
		"dead epoch start hash", deadEpochStartHash,
	)

	delete(t.mapFinalizedEpochs, epoch)
	t.forgetEpochStartHeader(deadEpochStartHash)
	t.removeStoredEpochStartMeta(epoch)
	// the canonical sibling registers its own evidence request when it is evaluated
	t.removePendingFinalityEvidence(string(deadEpochStartHash))
	t.removeRecoveryCandidateByHash(deadEpochStartHash)

	if bytes.Equal(t.epochMetaBlockHash, deadEpochStartHash) {
		t.restorePreActivationState()
	}

	err := t.saveState(t.triggerStateKey)
	if err != nil {
		log.Warn("DisarmDeadEpochStartActivation saveState", "error", err)
	}

	return true
}

// call only if mutex is locked before
func (t *trigger) forgetEpochStartHeader(hash []byte) {
	hdr := t.mapHashHdr[string(hash)]
	delete(t.mapEpochStartHdrs, string(hash))
	delete(t.mapHashHdr, string(hash))
	if check.IfNil(hdr) {
		return
	}

	hashes := t.mapNonceHashes[hdr.GetNonce()]
	remaining := make([]string, 0, len(hashes))
	for _, current := range hashes {
		if current == string(hash) {
			continue
		}

		remaining = append(remaining, current)
	}

	if len(remaining) == 0 {
		delete(t.mapNonceHashes, hdr.GetNonce())
		return
	}

	t.mapNonceHashes[hdr.GetNonce()] = remaining
}

// call only if mutex is locked before
func (t *trigger) removeStoredEpochStartMeta(epoch uint32) {
	epochStartIdentifier := []byte(core.EpochStartIdentifier(epoch))
	errNotCritical := t.metaHdrStorage.Remove(epochStartIdentifier)
	if errNotCritical != nil {
		log.Debug("removeStoredEpochStartMeta metaHdrStorage remove", "error", errNotCritical)
	}

	errNotCritical = t.triggerStorage.Remove(epochStartIdentifier)
	if errNotCritical != nil {
		log.Debug("removeStoredEpochStartMeta triggerStorage remove", "error", errNotCritical)
	}
}

// call only if mutex is locked before
func (t *trigger) restorePreActivationState() {
	newEpochHdrReceived := t.newEpochHdrReceived
	t.restorePreActivationStateForEpoch(t.epoch, t.epochStartShardHeader)
	t.newEpochHdrReceived = newEpochHdrReceived
}

func (t *trigger) restorePreActivationStateForEpoch(epoch uint32, epochStartHeader data.HeaderHandler) {
	t.metaEpoch = epoch
	t.isEpochStart = false
	t.newEpochHdrReceived = false
	t.epochStartMeta = &block.MetaBlock{}
	t.epochStartRound = 0
	t.epochFinalityAttestingRound = 0

	t.epochMetaBlockHash = nil
	shardHdr, ok := epochStartHeader.(data.ShardHeaderHandler)
	if ok && len(shardHdr.GetEpochStartMetaHash()) > 0 {
		t.epochMetaBlockHash = bytes.Clone(shardHdr.GetEpochStartMetaHash())
	}

	epochStartIdentifier := []byte(core.EpochStartIdentifier(epoch))
	metaBuff, err := t.metaHdrStorage.SearchFirst(epochStartIdentifier)
	if err != nil {
		log.Debug("restorePreActivationState epoch start meta not in storage", "epoch", epoch, "error", err)
		return
	}

	prevStartMeta, err := process.UnmarshalMetaHeader(t.marshaller, metaBuff)
	if err != nil {
		log.Warn("restorePreActivationState unmarshal", "error", err)
		return
	}

	t.epochStartMeta = prevStartMeta
	t.epochStartRound = prevStartMeta.GetRound()
	t.epochFinalityAttestingRound = prevStartMeta.GetRound()
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochMetaBlockHash
}

// LastCommitedEpochStartHdr returns the epoch start header
func (t *trigger) LastCommitedEpochStartHdr() (data.HeaderHandler, error) {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	// marshal + unmarshal deep copy
	headerBytes, err := t.marshaller.Marshal(t.epochStartShardHeader)
	if err != nil {
		return nil, err
	}

	return process.UnmarshalShardHeader(t.marshaller, headerBytes)
}

// GetEpochStartHdrFromStorage returns the epoch start header from storage
func (t *trigger) GetEpochStartHdrFromStorage(epoch uint32) (data.HeaderHandler, error) {
	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	shardHdrBuff, err := t.shardHdrStorage.SearchFirst([]byte(epochStartIdentifier))
	if err != nil {
		log.Warn("GetEpochStartHdrFromStorage search first", "epoch", epoch, "identifier", epochStartIdentifier, "error", err)
		return nil, err
	}

	return process.UnmarshalShardHeader(t.marshaller, shardHdrBuff)
}

// GetSavedStateKey returns the last saved trigger state key
func (t *trigger) GetSavedStateKey() []byte {
	return t.triggerStateKey
}

// Update updates the end-of-epoch trigger
func (t *trigger) Update(_ uint64, _ uint64) {
}

// SetEpochChange will do nothing
func (t *trigger) SetEpochChange(_ uint64) {
}

// ShouldProposeEpochChange will always return false
func (t *trigger) ShouldProposeEpochChange(_ uint64, _ uint64) bool {
	return false
}

// SetEpochChangeProposed will do nothing
func (t *trigger) SetEpochChangeProposed(_ bool) {
	// no implementation needed
}

// GetEpochChangeProposed will always return false
func (t *trigger) GetEpochChangeProposed() bool {
	return false
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

func (t *trigger) computeWatchdogTimeout() time.Duration {
	timeout := t.roundHandler.TimeDuration() * numRoundsWithoutReceivedMetaBlocks
	if timeout <= 0 {
		return 0
	}
	return timeout
}

func (t *trigger) watchdogRequestEpochStartMetaBlock(ctx context.Context) {
	watchdogTimeout := t.computeWatchdogTimeout()
	if watchdogTimeout == 0 {
		return
	}

	timer := time.NewTimer(watchdogTimeout)
	defer timer.Stop()

	resetTimer := func(d time.Duration) {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(d)
	}

	for {
		select {
		case <-ctx.Done():
			log.Debug("watchdogRequestEpochStartMetaBlock: trigger's go routine is stopping...")
			return
		case <-t.chanMetaBlockReceived:
			resetTimer(t.resetWatchdogTimeout(watchdogTimeout))
		case <-timer.C:
			t.handleWatchdogTimeout()
			resetTimer(t.resetWatchdogTimeout(watchdogTimeout))
		}
	}
}

func (t *trigger) resetWatchdogTimeout(fallback time.Duration) time.Duration {
	timeout := t.computeWatchdogTimeout()
	if timeout == 0 {
		return fallback
	}
	return timeout
}

func (t *trigger) handleWatchdogTimeout() {
	t.mutTrigger.RLock()
	epoch := t.processedEpochLocked()
	isEpochStart := t.isEpochStart
	t.mutTrigger.RUnlock()

	if isEpochStart || epoch == math.MaxUint32 {
		return
	}

	log.Debug("watchdog: no metablock received for too long, requesting epoch start metablock",
		"current epoch", epoch,
		"requesting epoch", epoch+1,
	)
	go t.requestHandler.RequestStartOfEpochMetaBlock(epoch + 1)
}

// Close will close the endless running go routine
func (t *trigger) Close() error {
	t.mutPendingEpochStartData.Lock()
	t.recoveryClosed = true
	t.callbackAdmission.Reset()
	t.recoveryGeneration++
	t.pendingEpochStartProofs = make(map[string]pendingEpochStartProof)
	t.pendingEpochStartHeaders = make(map[uint32]struct{})
	t.pendingFinalityEvidence = make(map[string]finalityEvidenceRequest)
	t.epochStartRecoveryCandidates = make(map[string]uint32)
	t.recoveryRequestCursors = [numRecoveryRequestClasses]string{}
	t.finalityCandidateCursor = ""
	t.mutPendingEpochStartData.Unlock()

	if t.cancelFunc != nil {
		t.cancelFunc()
	}

	t.mutTrigger.Lock()
	t.mutTrigger.Unlock()

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
