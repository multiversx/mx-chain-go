package rating

import "github.com/multiversx/mx-chain-go/process"

var _ process.RatingChanceHandler = (*selectionChance)(nil)

type selectionChance struct {
	maxThreshold     uint32
	chancePercentage uint32
}

//GetMaxThreshold returns the maxThreshold until this ChancePercentage holds
func (bsr *selectionChance) GetMaxThreshold() uint32 {
	return bsr.maxThreshold
}

//GetChancePercentage returns the percentage for the RatingChance
func (bsr *selectionChance) GetChancePercentage() uint32 {
	return bsr.chancePercentage
}

//IsInterfaceNil verifies if the interface is nil
func (bsr *selectionChance) IsInterfaceNil() bool {
	return bsr == nil
}
