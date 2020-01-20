package txcache

const oneMilion = 1000000
const oneTrilion = oneMilion * oneMilion
const delta = 0.00000001

func toMicroERD(erd int64) int64 {
	return erd * 1000000
}

func (cache *TxCache) getRawScoreOfSender(sender string) float64 {
	list, ok := cache.txListBySender.getListForSender(sender)
	if !ok {
		panic("sender not in cache")
	}

	rawScore := list.computeRawScore()
	return rawScore
}
