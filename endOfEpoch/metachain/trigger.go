package metachain

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

// ArgsNewMetaEndOfEpochTrigger defines struct needed to create a new end of epoch trigger
type ArgsNewMetaEndOfEpochTrigger struct {
	Rounder     endOfEpoch.Rounder
	SyncTimer   ntp.SyncTimer
	GenesisTime time.Time
	Settings    *config.EndOfEpochConfig
	Epoch       uint32
}

type trigger struct {
	epoch                         uint32
	rounder                       endOfEpoch.Rounder
	roundsPerEpoch                int64
	roundsBetweenForcedEndOfEpoch int64
	epochStartTime                time.Time
	syncTimer                     ntp.SyncTimer
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
	if check.IfNil(args.SyncTimer) {
		return nil, endOfEpoch.ErrNilSyncTimer
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
		rounder:                       args.Rounder,
		roundsPerEpoch:                args.Settings.RoundsPerEpoch,
		epochStartTime:                args.GenesisTime,
		syncTimer:                     args.SyncTimer,
		epoch:                         args.Epoch,
		roundsBetweenForcedEndOfEpoch: args.Settings.MinRoundsBetweenEpochs,
	}, nil
}

// IsEndOfEpoch return true if conditions are fullfilled for end of epoch
func (t *trigger) IsEndOfEpoch() bool {
	t.rounder.UpdateRound(t.epochStartTime, t.syncTimer.CurrentTime())
	currRoundIndex := t.rounder.Index()

	if currRoundIndex == 0 {
		return true
	}

	if currRoundIndex > t.roundsPerEpoch {
		t.epoch += 1
		t.epochStartTime = t.rounder.TimeStamp()
		return true
	}

	return false
}

// ForceEndOfEpoch sets the conditions ofr end of epoch to true in case of edge cases
func (t *trigger) ForceEndOfEpoch() error {
	oldRoundIndex := t.rounder.Index()
	t.rounder.UpdateRound(t.epochStartTime, t.syncTimer.CurrentTime())
	currRoundIndex := t.rounder.Index()

	if currRoundIndex < t.roundsBetweenForcedEndOfEpoch {
		return endOfEpoch.ErrNotEnoughRoundsBetweenEpochs
	}

	if oldRoundIndex != currRoundIndex {
		return endOfEpoch.ErrForceEndOfEpochCanBeNotCalledOnNewRound
	}

	t.epochStartTime = t.rounder.TimeStamp()
	t.epoch += 1

	return nil
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
