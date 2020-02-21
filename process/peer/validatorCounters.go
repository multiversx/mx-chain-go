package peer

type validatorCounters struct {
	leaderIncreaseCount    uint32
	leaderDecreaseCount    uint32
	validatorIncreaseCount uint32
	validatorDecreaseCount uint32
}

type validatorRoundCounters map[string]*validatorCounters

func (vrc *validatorRoundCounters) reset() {
	*vrc = make(validatorRoundCounters)
}

func (vrc validatorRoundCounters) increaseValidator(key []byte) {
	vrc.get(key).validatorIncreaseCount++
}

func (vrc validatorRoundCounters) decreaseValidator(key []byte) {
	vrc.get(key).validatorDecreaseCount++
}

func (vrc validatorRoundCounters) increaseLeader(key []byte) {
	vrc.get(key).leaderIncreaseCount++
}

func (vrc validatorRoundCounters) decreaseLeader(key []byte) {
	vrc.get(key).leaderDecreaseCount++
}

func (vrc validatorRoundCounters) get(key []byte) *validatorCounters {
	vdCounter, ok := vrc[string(key)]
	if !ok {
		vrc[string(key)] = &validatorCounters{}
		return vrc[string(key)]
	}

	return vdCounter
}
