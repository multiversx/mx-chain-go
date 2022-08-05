package chronology

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// ArgChronology holds all dependencies required by the chronology component
type ArgChronology struct {
	GenesisTime      time.Time
	RoundHandler     consensus.RoundHandler
	SyncTimer        consensus.SyncTimer
	Watchdog         core.WatchdogTimer
	AppStatusHandler core.AppStatusHandler
}
