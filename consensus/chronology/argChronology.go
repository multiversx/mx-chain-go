package chronology

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

// ArgChronology holds all dependencies required by the chronology component
type ArgChronology struct {
	GenesisTime      time.Time
	RoundHandler     consensus.RoundHandler
	SyncTimer        ntp.SyncTimer
	Watchdog         core.WatchdogTimer
	AppStatusHandler core.AppStatusHandler
}
