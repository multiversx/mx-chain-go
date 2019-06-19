package chronology

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
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
