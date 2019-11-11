package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
)

// ArgsNewMetaEndOfEpochTrigger defines struct needed to create a new end of epoch trigger
type ArgsNewMetaEndOfEpochTrigger struct {
	Rounder     endOfEpoch.Rounder
	GenesisTime time.Time
	Settings    *config.EndOfEpochConfig
	Epoch       uint32
}

type trigger struct {
	isEndOfEpoch                  bool
	epoch                         uint32
	currentRoundIndex             int64
	epochStartRound               int64
	roundsPerEpoch                int64
	roundsBetweenForcedEndOfEpoch int64
	epochStartTime                time.Time
	rounder                       endOfEpoch.Rounder
}

// NewEndOfEpochTrigger creates a trigger for end of epoch
func NewEndOfEpochTrigger(args *ArgsNewMetaEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewMetaEndOfEpochTrigger
	}
	if check.IfNil(args.Rounder) {
		return nil, endOfEpoch.ErrNilRounder
	}
	if args.Settings == nil {
		return nil, endOfEpoch.ErrNilEndOfEpochSettings
	}
	if args.Settings.RoundsPerEpoch < 1 {
		return nil, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs < 1 {
		return nil, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger
	}
	if args.Settings.MinRoundsBetweenEpochs > args.Settings.RoundsPerEpoch {
		return nil, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger
	}

	return &trigger{
		roundsPerEpoch:                args.Settings.RoundsPerEpoch,
		epochStartTime:                args.GenesisTime,
		epoch:                         args.Epoch,
		roundsBetweenForcedEndOfEpoch: args.Settings.MinRoundsBetweenEpochs,
		rounder:                       args.Rounder,
	}, nil
}

// IsEndOfEpoch return true if conditions are fullfilled for end of epoch
func (t *trigger) IsEndOfEpoch() bool {
	return t.isEndOfEpoch
}

// ForceEndOfEpoch sets the conditions ofr end of epoch to true in case of edge cases
func (t *trigger) ForceEndOfEpoch(round int64) error {
	if t.currentRoundIndex > round {
		return endOfEpoch.ErrSavedRoundIsHigherThanSaved
	}
	if t.currentRoundIndex == round {
		return endOfEpoch.ErrForceEndOfEpochCanBeCalledOnNewRound
	}
	if t.currentRoundIndex+1 != round {
		return endOfEpoch.ErrNewRoundIsNotTheNextRound
	}

	t.currentRoundIndex = round

	if t.currentRoundIndex-t.epochStartRound < t.roundsBetweenForcedEndOfEpoch {
		return endOfEpoch.ErrNotEnoughRoundsBetweenEpochs
	}

	t.epochStartTime = t.rounder.TimeStamp()
	t.epoch += 1
	t.epochStartRound = t.currentRoundIndex
	t.isEndOfEpoch = true

	return nil
}

// Update processes changes in the trigger
func (t *trigger) Update(round int64) {
	if t.currentRoundIndex > round {
		return
	}
	if t.currentRoundIndex == round {
		return
	}
	if t.currentRoundIndex+1 != round {
		return
	}

	t.currentRoundIndex = round

	if t.currentRoundIndex > t.roundsPerEpoch {
		t.epoch += 1
		t.epochStartTime = t.rounder.TimeStamp()
		t.isEndOfEpoch = true
		t.epochStartRound = t.currentRoundIndex
	}
}

// Processed signals end of epoch processing is done
func (t *trigger) Processed() {
	t.isEndOfEpoch = false
}

// Epoch return the current epoch
func (t *trigger) Epoch() uint32 {
	return t.epoch
}

// ReceivedHeader saved the header into pool to verify if end-of-epoch conditions are fullfilled
func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
}

// IsInterfaceNil return true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
