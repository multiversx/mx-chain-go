package metachain

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// ArgsNewMetaEpochStartTrigger defines struct needed to create a new end of epoch trigger
type ArgsNewMetaEpochStartTrigger struct {
	Rounder     epochStart.Rounder
	GenesisTime time.Time
	Settings    *config.EpochStartConfig
	Epoch       uint32
}

type trigger struct {
	isEpochStart                  bool
	epoch                         uint32
	currentRound                  int64
	epochStartRound               int64
	roundsPerEpoch                int64
	roundsBetweenForcedEpochStart int64
	epochStartTime                time.Time
	rounder                       epochStart.Rounder
}

// NewEpochStartTrigger creates a trigger for end of epoch
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
		roundsPerEpoch:                args.Settings.RoundsPerEpoch,
		epochStartTime:                args.GenesisTime,
		epoch:                         args.Epoch,
		roundsBetweenForcedEpochStart: args.Settings.MinRoundsBetweenEpochs,
		rounder:                       args.Rounder,
	}, nil
}

// IsEpochStart return true if conditions are fulfilled for end of epoch
func (t *trigger) IsEpochStart() bool {
	return t.isEpochStart
}

// EpochStartRound returns the start round of the current epoch
func (t *trigger) EpochStartRound() uint64 {
	return uint64(t.epochStartRound)
}

// ForceEpochStart sets the conditions for end of epoch to true in case of edge cases
func (t *trigger) ForceEpochStart(round int64) error {
	if t.currentRound > round {
		return epochStart.ErrSavedRoundIsHigherThanInput
	}
	if t.currentRound == round {
		return epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound
	}

	t.currentRound = round

	if t.currentRound-t.epochStartRound < t.roundsBetweenForcedEpochStart {
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

// Processed sets end of epoch to false and cleans underlying structure
func (t *trigger) Processed() {
	t.isEpochStart = false
}

// Revert sets the end of epoch back to true
func (t *trigger) Revert() {
	t.isEpochStart = true
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
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
