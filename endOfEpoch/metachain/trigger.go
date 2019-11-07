package metachain

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

var log = logger.DefaultLogger()

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

func NewEndOfEpochTrigger(args *ArgsNewMetaEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewMetaEndOfEpochTrigger
	}
	if check.IfNil(args.Rounder) {
		return nil, endOfEpoch.ErrNilRounder
	}
	if args.Settings == nil {
		return nil, endOfEpoch.ErrNilSettingsHandler
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

func (t *trigger) ForceEndOfEpoch() {
	t.rounder.UpdateRound(t.epochStartTime, t.syncTimer.CurrentTime())
	currRoundIndex := t.rounder.Index()

	if currRoundIndex < t.roundsBetweenForcedEndOfEpoch {
		log.Info("Tried to force end of epoch before passing of enough rounds")
		return
	}

	t.epochStartTime = t.rounder.TimeStamp()
	t.epoch += 1
}

func (t *trigger) Epoch() uint32 {
	return t.epoch
}

func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
}

func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
