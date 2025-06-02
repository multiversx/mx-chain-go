package chronology

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/consensus"
)

func (chr *chronology) StartRound() {
	chr.startRound(context.Background())
}

func (chr *chronology) SubroundId() int {
	return chr.subroundId
}

func (chr *chronology) SetSubroundId(subroundId int) {
	chr.subroundId = subroundId
}

func (chr *chronology) LoadSubroundHandler(subround int) consensus.SubroundHandler {
	return chr.loadSubroundHandler(subround)
}

func (chr *chronology) SubroundHandlers() []consensus.SubroundHandler {
	return chr.subroundHandlers
}

func (chr *chronology) SetSubroundHandlers(subroundHandlers []consensus.SubroundHandler) {
	chr.subroundHandlers = subroundHandlers
}

func (chr *chronology) UpdateRound() {
	chr.updateRound()
}

func (chr *chronology) InitRound() {
	chr.initRound()
}

// StartRoundsTest calls the unexported startRounds function
func (chr *chronology) StartRoundsTest(ctx context.Context) {
	chr.startRounds(ctx)
}

// SetWatchdog sets the watchdog for chronology object
func (chr *chronology) SetWatchdog(watchdog core.WatchdogTimer) {
	chr.watchdog = watchdog
}

// SetCancelFunc sets cancelFunc for chronology object
func (chr *chronology) SetCancelFunc(cancelFunc func()) {
	chr.cancelFunc = cancelFunc
}
