package chronology

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/ntp"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgChronology holds all dependencies required by the chronology component
type ArgChronology struct {
	Logger           logger.Logger
	GenesisTime      time.Time
	RoundHandler     consensus.RoundHandler
	SyncTimer        ntp.SyncTimer
	Watchdog         core.WatchdogTimer
	AppStatusHandler core.AppStatusHandler
}
