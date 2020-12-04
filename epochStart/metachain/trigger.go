package metachain

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("epochStart/metachain")

var _ dataRetriever.EpochHandler = (*trigger)(nil)
var _ epochStart.TriggerHandler = (*trigger)(nil)
var _ process.EpochStartTriggerHandler = (*trigger)(nil)
var _ process.EpochBootstrapper = (*trigger)(nil)
var _ closing.Closer = (*trigger)(nil)

const minimumNonceToStartEpoch = 4
const disabledRoundForForceEpochStart = math.MaxUint64

// ArgsNewMetaEpochStartTrigger defines struct needed to create a new start of epoch trigger
type ArgsNewMetaEpochStartTrigger struct {
	GenesisTime        time.Time
	Settings           *config.EpochStartConfig
	Epoch              uint32
	EpochStartRound    uint64
	EpochStartNotifier epochStart.Notifier
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	Storage            dataRetriever.StorageService
}

type trigger struct {
	isEpochStart                bool
	epoch                       uint32
	epochStartMeta              *block.MetaBlock
	currentRound                uint64
	epochFinalityAttestingRound uint64
	currEpochStartRound         uint64
	prevEpochStartRound         uint64
	nextEpochStartRound         uint64
	roundsPerEpoch              uint64
	minRoundsBetweenEpochs      uint64
	epochStartMetaHash          []byte
	triggerStateKey             []byte
	epochStartTime              time.Time
	mutTrigger                  sync.RWMutex
	epochStartNotifier          epochStart.Notifier
	metaHeaderStorage           storage.Storer
	triggerStorage              storage.Storer
	marshalizer                 marshal.Marshalizer
	hasher                      hashing.Hasher
	appStatusHandler            core.AppStatusHandler
}

// NewEpochStartTrigger creates a trigger for start of epoch
func NewEpochStartTrigger(args *ArgsNewMetaEpochStartTrigger) (*trigger, error) {
	if args == nil {
		return nil, epochStart.ErrNilArgsNewMetaEpochStartTrigger
	}
	if args.Settings == nil {
		return nil, epochStart.ErrNilEpochStartSettings
	}
	if args.Settings.RoundsPerEpoch < 1 {
		return nil, fmt.Errorf("%w, RoundsPerEpoch < 1", epochStart.ErrInvalidSettingsForEpochStartTrigger)
	}
	if args.Settings.MinRoundsBetweenEpochs < 1 {
		return nil, fmt.Errorf("%w, MinRoundsBetweenEpochs < 1", epochStart.ErrInvalidSettingsForEpochStartTrigger)
	}
	if args.Settings.MinRoundsBetweenEpochs > args.Settings.RoundsPerEpoch {
		return nil, fmt.Errorf("%w, MinRoundsBetweenEpochs > RoundsPerEpoch", epochStart.ErrInvalidSettingsForEpochStartTrigger)
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Storage) {
		return nil, epochStart.ErrNilStorageService
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}

	triggerStorage := args.Storage.GetStorer(dataRetriever.BootstrapUnit)
	if check.IfNil(triggerStorage) {
		return nil, epochStart.ErrNilTriggerStorage
	}

	metaBlockStorage := args.Storage.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(triggerStorage) {
		return nil, epochStart.ErrNilMetaBlockStorage
	}

	trigggerStateKey := core.TriggerRegistryInitialKeyPrefix + fmt.Sprintf("%d", args.Epoch)
	trig := &trigger{
		triggerStateKey:             []byte(trigggerStateKey),
		roundsPerEpoch:              uint64(args.Settings.RoundsPerEpoch),
		epochStartTime:              args.GenesisTime,
		currEpochStartRound:         args.EpochStartRound,
		prevEpochStartRound:         args.EpochStartRound,
		epoch:                       args.Epoch,
		minRoundsBetweenEpochs:      uint64(args.Settings.MinRoundsBetweenEpochs),
		mutTrigger:                  sync.RWMutex{},
		epochFinalityAttestingRound: args.EpochStartRound,
		epochStartNotifier:          args.EpochStartNotifier,
		metaHeaderStorage:           metaBlockStorage,
		triggerStorage:              triggerStorage,
		marshalizer:                 args.Marshalizer,
		hasher:                      args.Hasher,
		epochStartMeta:              &block.MetaBlock{},
		appStatusHandler:            &statusHandler.NilStatusHandler{},
		nextEpochStartRound:         disabledRoundForForceEpochStart,
	}

	err := trig.saveState(trig.triggerStateKey)
	if err != nil {
		return nil, err
	}

	return trig, nil
}

// IsEpochStart return true if conditions are fulfilled for start of epoch
func (t *trigger) IsEpochStart() bool {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.isEpochStart
}

// EpochStartRound returns the start round of the current epoch
func (t *trigger) EpochStartRound() uint64 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.currEpochStartRound
}

// EpochFinalityAttestingRound returns the round when epoch start block was finalized
func (t *trigger) EpochFinalityAttestingRound() uint64 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epochFinalityAttestingRound
}

// ForceEpochStart sets the round at which the new epoch will start
func (t *trigger) ForceEpochStart(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.nextEpochStartRound = round
	if t.nextEpochStartRound >= t.currEpochStartRound+t.roundsPerEpoch {
		t.nextEpochStartRound = disabledRoundForForceEpochStart
		log.Debug("can not force epoch start because the resulting round is in the next epoch")

		return
	}
	if t.nextEpochStartRound-t.currEpochStartRound < t.minRoundsBetweenEpochs {
		t.nextEpochStartRound = t.currEpochStartRound + t.minRoundsBetweenEpochs
		log.Debug("can not force epoch start on provided round",
			"provided round", round, "computed round", t.nextEpochStartRound)
	}

	log.Debug("set new epoch start round", "round", t.nextEpochStartRound)
}

// Update processes changes in the trigger
func (t *trigger) Update(round uint64, nonce uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.currentRound = round

	if t.isEpochStart {
		return
	}

	isZeroEpochEdgeCase := nonce < minimumNonceToStartEpoch
	isNormalEpochStart := t.currentRound > t.currEpochStartRound+t.roundsPerEpoch
	isWithEarlyEndOfEpoch := t.currentRound >= t.nextEpochStartRound
	shouldTriggerEpochStart := (isNormalEpochStart || isWithEarlyEndOfEpoch) && !isZeroEpochEdgeCase
	if shouldTriggerEpochStart {
		t.epoch += 1
		t.isEpochStart = true
		t.prevEpochStartRound = t.currEpochStartRound
		t.currEpochStartRound = t.currentRound
		t.saveCurrentState(round)

		msg := fmt.Sprintf("EPOCH %d BEGINS IN ROUND (%d)", t.epoch, t.currentRound)
		log.Debug(display.Headline(msg, "", "#"))
		log.Debug("trigger.Update", "isEpochStart", t.isEpochStart)
		logger.SetCorrelationEpoch(t.epoch)
		t.nextEpochStartRound = disabledRoundForForceEpochStart
	}
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (t *trigger) SetProcessed(header data.HeaderHandler, body data.BodyHandler) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	metaBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return
	}
	if !metaBlock.IsStartOfEpochBlock() {
		return
	}

	metaBuff, errNotCritical := t.marshalizer.Marshal(metaBlock)
	if errNotCritical != nil {
		log.Debug("SetProcessed marshal", "error", errNotCritical.Error())
	}

	t.appStatusHandler.SetUInt64Value(core.MetricRoundAtEpochStart, metaBlock.Round)
	t.appStatusHandler.SetUInt64Value(core.MetricNonceAtEpochStart, metaBlock.Nonce)

	metaHash := t.hasher.Compute(string(metaBuff))

	t.currEpochStartRound = metaBlock.Round
	t.epoch = metaBlock.Epoch
	t.isEpochStart = false
	t.currentRound = metaBlock.Round
	t.epochStartMeta = metaBlock
	t.epochStartMetaHash = metaHash

	t.epochStartNotifier.NotifyAllPrepare(metaBlock, body)
	t.epochStartNotifier.NotifyAll(metaBlock)

	t.saveCurrentState(metaBlock.Round)

	log.Debug("trigger.SetProcessed", "isEpochStart", t.isEpochStart)

	epochStartIdentifier := core.EpochStartIdentifier(metaBlock.Epoch)
	errNotCritical = t.triggerStorage.Put([]byte(epochStartIdentifier), metaBuff)
	if errNotCritical != nil {
		log.Warn("SetProcessed put into triggerStorage", "error", errNotCritical.Error())
	}

	errNotCritical = t.metaHeaderStorage.Put([]byte(epochStartIdentifier), metaBuff)
	if errNotCritical != nil {
		log.Warn("SetProcessed put into metaHdrStorage", "error", errNotCritical.Error())
	}
}

// SetFinalityAttestingRound sets the round which finalized the start of epoch block
func (t *trigger) SetFinalityAttestingRound(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if round > t.currEpochStartRound {
		t.epochFinalityAttestingRound = round
		t.saveCurrentState(round)
		t.epochStartNotifier.NotifyEpochChangeConfirmed(t.epoch)
	}
}

// RevertStateToBlock will revert the state of the trigger to the current block
func (t *trigger) RevertStateToBlock(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return epochStart.ErrNilHeaderHandler
	}

	if header.IsStartOfEpochBlock() {
		log.Debug("RevertStateToBlock with epoch start block called")
		t.SetProcessed(header, nil)
		return nil
	}

	t.mutTrigger.RLock()
	prevMeta := t.epochStartMeta
	t.mutTrigger.RUnlock()

	currentHeaderHash, err := core.CalculateHash(t.marshalizer, t.hasher, header)
	if err != nil {
		log.Warn("RevertStateToBlock error on hashing", "error", err)
		return err
	}

	if !bytes.Equal(prevMeta.GetPrevHash(), currentHeaderHash) {
		return nil
	}

	log.Debug("RevertStateToBlock to revert behind epoch start block is called")
	err = t.revert(prevMeta)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.currentRound = header.GetRound()
	t.mutTrigger.Unlock()

	return nil
}

// SetAppStatusHandler will set the satus handler for the trigger
func (t *trigger) SetAppStatusHandler(handler core.AppStatusHandler) error {
	if check.IfNil(handler) {
		return epochStart.ErrNilStatusHandler
	}

	t.appStatusHandler = handler
	return nil
}

func (t *trigger) revert(header data.HeaderHandler) error {
	if check.IfNil(header) || !header.IsStartOfEpochBlock() || header.GetEpoch() == 0 {
		return nil
	}

	metaHdr, ok := header.(*block.MetaBlock)
	if !ok {
		log.Warn("wrong type assertion in Revert metachain trigger")
		return epochStart.ErrWrongTypeAssertion
	}

	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	prevEpochStartIdentifier := core.EpochStartIdentifier(metaHdr.Epoch - 1)
	epochStartMetaBuff, err := t.metaHeaderStorage.SearchFirst([]byte(prevEpochStartIdentifier))
	if err != nil {
		log.Warn("Revert get previous meta from storage", "error", err)
		return err
	}

	epochStartMeta := &block.MetaBlock{}
	err = t.marshalizer.Unmarshal(epochStartMeta, epochStartMetaBuff)
	if err != nil {
		log.Warn("Revert unmarshal previous meta", "error", err)
		return err
	}

	epochStartIdentifier := core.EpochStartIdentifier(metaHdr.Epoch)
	errNotCritical := t.triggerStorage.Remove([]byte(epochStartIdentifier))
	if errNotCritical != nil {
		log.Debug("Revert remove from triggerStorage", "error", errNotCritical.Error())
	}

	errNotCritical = t.metaHeaderStorage.Remove([]byte(epochStartIdentifier))
	if errNotCritical != nil {
		log.Debug("Revert remove from triggerStorage", "error", errNotCritical.Error())
	}

	t.currEpochStartRound = metaHdr.EpochStart.Economics.PrevEpochStartRound
	t.epoch = metaHdr.Epoch - 1
	t.isEpochStart = false
	t.epochStartMeta = epochStartMeta

	log.Debug("trigger.revert",
		"isEpochStart", t.isEpochStart,
		"epoch", t.epoch,
		"epochStartRound", t.currEpochStartRound)

	return nil
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// MetaEpoch return the current epoch
func (t *trigger) MetaEpoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// RequestEpochStartIfNeeded request the needed epoch start block if metablock with new epoch was received
func (t *trigger) RequestEpochStartIfNeeded(_ data.HeaderHandler) {
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	return t.epochStartMetaHash
}

// GetSavedStateKey returns the last saved trigger state key
func (t *trigger) GetSavedStateKey() []byte {
	return t.triggerStateKey
}

// SetEpochStartMetaHdrHash sets the epoch start meta header has
func (t *trigger) SetEpochStartMetaHdrHash(metaHdrHash []byte) {
	t.epochStartMetaHash = metaHdrHash
}

// SetCurrentEpochStartRound sets the round when the current epoch started
func (t *trigger) SetCurrentEpochStartRound(round uint64) {
	t.mutTrigger.Lock()
	t.currEpochStartRound = round
	t.currentRound = round
	t.saveCurrentState(round)
	t.mutTrigger.Unlock()
}

// Close will close the endless running go routine
func (t *trigger) Close() error {
	return nil
}

// IsInterfaceNil return true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}

// needs to be called under locked mutex
func (t *trigger) saveCurrentState(round uint64) {
	t.triggerStateKey = []byte(fmt.Sprint(round))
	err := t.saveState(t.triggerStateKey)
	if err != nil {
		log.Warn("error saving trigger state", "error", err, "key", t.triggerStateKey)
	}
}
