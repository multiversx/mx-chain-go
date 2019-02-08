package chronology

import (
	"time"
)

func (chr *chronology) StartRound() {
	chr.startRound()
}

func (chr *chronology) SubroundId() int {
	return chr.subroundId
}

func (chr *chronology) SetSubroundId(subroundId int) {
	chr.subroundId = subroundId
}

func (chr *chronology) LoadSubroundHandler(subround int) SubroundHandler {
	return chr.loadSubroundHandler(subround)
}

func (chr *chronology) SubroundHandlers() []SubroundHandler {
	return chr.subroundHandlers
}

func (chr *chronology) SetSubroundHandlers(subroundHandlers []SubroundHandler) {
	chr.subroundHandlers = subroundHandlers
}

func (chr *chronology) UpdateRound() {
	chr.updateRound()
}

func (chr *chronology) HaveTimeInCurrentRound() time.Duration {
	return chr.haveTimeInCurrentRound()
}
