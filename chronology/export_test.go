package chronology

func (chr *Chronology) InitRound() {
	chr.initRound()
}

func (chr *Chronology) UpdateSelfSubroundIfNeeded(subRoundId SubroundId) {
	chr.updateSelfSubroundIfNeeded(subRoundId)
}
