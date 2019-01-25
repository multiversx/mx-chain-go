package chronology

import (
	"time"
)

func (chr *Chronology) InitRound() {
	chr.initRound()
}

func (chr *Chronology) UpdateSelfSubroundIfNeeded(subRoundId SubroundId) {
	chr.updateSelfSubroundIfNeeded(subRoundId)
}

func MaxDiffAccepted() time.Duration {
	return maxDifAccepted
}
