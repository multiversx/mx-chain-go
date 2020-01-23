package metachain

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("epochStart/metachain")

// ArgsNewMetaEpochStartTrigger defines struct needed to create a new start of epoch trigger
type ArgsNewMetaEpochStartTrigger struct {
	GenesisTime        time.Time
	Settings           *config.EpochStartConfig
	Epoch              uint32
	EpochStartRound    uint64
	EpochStartNotifier epochStart.StartOfEpochNotifier
	Marshalizer        marshal.Marshalizer
	Storage            dataRetriever.StorageService
}

type trigger struct {
	isEpochStart                bool
	epoch                       uint32
	currentRound                uint64
	epochFinalityAttestingRound uint64
	currEpochStartRound         uint64
	prevEpochStartRound         uint64
	roundsPerEpoch              uint64
	minRoundsBetweenEpochs      uint64
	epochStartMetaHash          []byte
	epochStartTime              time.Time
	mutTrigger                  sync.RWMutex
	epochStartNotifier          epochStart.StartOfEpochNotifier
	metaHdrStorage              storage.Storer
	marshalizer                 marshal.Marshalizer
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
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs < 1 {
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs > args.Settings.RoundsPerEpoch {
		return nil, epochStart.ErrInvalidSettingsForEpochStartTrigger
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

	metaHdrStorage := args.Storage.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaHdrStorage) {
		return nil, epochStart.ErrNilMetaHdrStorage
	}

	return &trigger{
		roundsPerEpoch:              uint64(args.Settings.RoundsPerEpoch),
		epochStartTime:              args.GenesisTime,
		currEpochStartRound:         args.EpochStartRound,
		prevEpochStartRound:         args.EpochStartRound,
		epoch:                       args.Epoch,
		minRoundsBetweenEpochs:      uint64(args.Settings.MinRoundsBetweenEpochs),
		mutTrigger:                  sync.RWMutex{},
		epochFinalityAttestingRound: args.EpochStartRound,
		epochStartNotifier:          args.EpochStartNotifier,
		metaHdrStorage:              metaHdrStorage,
		marshalizer:                 args.Marshalizer,
	}, nil
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

// ForceEpochStart sets the conditions for start of epoch to true in case of edge cases
func (t *trigger) ForceEpochStart(round uint64) error {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.currentRound > round {
		return epochStart.ErrSavedRoundIsHigherThanInput
	}
	if t.currentRound == round {
		return epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound
	}

	t.currentRound = round

	if t.currentRound-t.currEpochStartRound < t.minRoundsBetweenEpochs {
		return epochStart.ErrNotEnoughRoundsBetweenEpochs
	}

	if !t.isEpochStart {
		t.epoch += 1
	}

	t.currEpochStartRound = t.currentRound
	t.isEpochStart = true

	return nil
}

// Update processes changes in the trigger
func (t *trigger) Update(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.isEpochStart {
		return
	}

	t.currentRound = round

	if t.currentRound > t.currEpochStartRound+t.roundsPerEpoch {
		t.epoch += 1
		t.isEpochStart = true
		t.prevEpochStartRound = t.currEpochStartRound
		t.currEpochStartRound = t.currentRound
	}
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (t *trigger) SetProcessed(header data.HeaderHandler) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	metaBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return
	}
	if !metaBlock.IsStartOfEpochBlock() {
		return
	}

	metaBuff, err := t.marshalizer.Marshal(metaBlock)
	if err != nil {
		log.Debug("SetProcessed marshal", "error", err.Error())
	}

	epochStartIdentifier := core.EpochStartIdentifier(metaBlock.Epoch)
	err = t.metaHdrStorage.Put([]byte(epochStartIdentifier), metaBuff)
	if err != nil {
		log.Debug("SetProcessed put into metaHdrStorage", "error", err.Error())
	}

	t.currEpochStartRound = metaBlock.Round
	t.epoch = metaBlock.Epoch
	t.epochStartNotifier.NotifyAll(metaBlock)
	t.isEpochStart = false
}

// SetFinalityAttestingRound sets the round which finalized the start of epoch block
func (t *trigger) SetFinalityAttestingRound(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if round > t.currEpochStartRound {
		t.epochFinalityAttestingRound = round
	}
}

// Revert sets the start of epoch back to true
func (t *trigger) Revert(round uint64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	epochStartIdentifier := core.EpochStartIdentifier(t.epoch)
	err := t.metaHdrStorage.Remove([]byte(epochStartIdentifier))
	if err != nil {
		log.Debug("Revert remove from metaHdrStorage", "error", err.Error())
	}

	t.currEpochStartRound = t.prevEpochStartRound
	t.epoch = t.epoch - 1
	t.isEpochStart = false
	t.currentRound = round
	if t.currentRound > 0 {
		t.currentRound = t.currentRound - 1
	}

	log.Debug("epoch trigger revert called", "epoch", t.epoch, "epochStartRound", t.currEpochStartRound)
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// ReceivedHeader saved the header into pool to verify if end-of-epoch conditions are fulfilled
func (t *trigger) ReceivedHeader(_ data.HeaderHandler) {
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	return t.epochStartMetaHash
}

// SetEpochStartMetaHdrHash sets the epoch start meta header hase
func (t *trigger) SetEpochStartMetaHdrHash(metaHdrHash []byte) {
	t.epochStartMetaHash = metaHdrHash
}

// SetCurrentEpochStartRound sets the round when the current epoch started
func (t *trigger) SetCurrentEpochStartRound(round uint64) {
	t.mutTrigger.Lock()
	t.currentRound = round
	t.mutTrigger.Unlock()
}

// IsInterfaceNil return true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
