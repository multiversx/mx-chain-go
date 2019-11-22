package metachain

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// ArgsNewMetaEpochStartTrigger defines struct needed to create a new start of epoch trigger
type ArgsNewMetaEpochStartTrigger struct {
	Rounder     epochStart.Rounder
	GenesisTime time.Time
	Settings    *config.EpochStartConfig
	Epoch       uint32
}

type trigger struct {
	isEpochStart           bool
	epoch                  uint32
	currentRound           int64
	epochStartRound        int64
	roundsPerEpoch         int64
	minRoundsBetweenEpochs int64
	epochStartTime         time.Time
	rounder                epochStart.Rounder
	mutTrigger             sync.RWMutex
}

// NewEpochStartTrigger creates a trigger for start of epoch
func NewEpochStartTrigger(args *ArgsNewMetaEpochStartTrigger) (*trigger, error) {
	if args == nil {
		return nil, epochStart.ErrNilArgsNewMetaEpochStartTrigger
	}
	if check.IfNil(args.Rounder) {
		return nil, epochStart.ErrNilRounder
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

	return &trigger{
		roundsPerEpoch:         args.Settings.RoundsPerEpoch,
		epochStartTime:         args.GenesisTime,
		epoch:                  args.Epoch,
		minRoundsBetweenEpochs: args.Settings.MinRoundsBetweenEpochs,
		rounder:                args.Rounder,
		mutTrigger:             sync.RWMutex{},
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

	return uint64(t.epochStartRound)
}

// ForceEpochStart sets the conditions for start of epoch to true in case of edge cases
func (t *trigger) ForceEpochStart(round int64) error {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.currentRound > round {
		return epochStart.ErrSavedRoundIsHigherThanInput
	}
	if t.currentRound == round {
		return epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound
	}

	t.currentRound = round

	if t.currentRound-t.epochStartRound < t.minRoundsBetweenEpochs {
		return epochStart.ErrNotEnoughRoundsBetweenEpochs
	}

	t.epochStartTime = t.rounder.TimeStamp()
	t.epoch += 1
	t.epochStartRound = t.currentRound
	t.isEpochStart = true

	return nil
}

// Update processes changes in the trigger
func (t *trigger) Update(round int64) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	if t.currentRound+1 != round {
		return
	}

	t.currentRound = round

	if t.currentRound > t.epochStartRound+t.roundsPerEpoch {
		t.epoch += 1
		t.epochStartTime = t.rounder.TimeStamp()
		t.isEpochStart = true
		t.epochStartRound = t.currentRound
	}
}

// Processed sets start of epoch to false and cleans underlying structure
func (t *trigger) Processed() {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.isEpochStart = false
}

// Revert sets the start of epoch back to true
func (t *trigger) Revert() {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	t.isEpochStart = true
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
	t.mutTrigger.RLock()
	defer t.mutTrigger.RUnlock()

	return t.epoch
}

// ReceivedHeader saved the header into pool to verify if end-of-epoch conditions are fulfilled
func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
}

// EpochStartMetaHdrHash returns the announcing meta header hash which created the new epoch
func (t *trigger) EpochStartMetaHdrHash() []byte {
	return nil
}

// IsInterfaceNil return true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
