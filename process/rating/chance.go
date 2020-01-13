package rating

type Chance struct {
	maxThreshold     uint32
	chancePercentage uint32
}

//GetMaxThreshold returns the maxThreshold until this ChancePercentage holds
func (bsr *Chance) GetMaxThreshold() uint32 {
	return bsr.maxThreshold
}

//GetChancePercentage returns the percentage for the RatingChance
func (bsr *Chance) GetChancePercentage() uint32 {
	return bsr.chancePercentage
}
