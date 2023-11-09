package rating

import "github.com/multiversx/mx-chain-go/process"

var _ process.SelectionChance = (*SelectionChance)(nil)

// SelectionChance contains the threshold and chancePercent to be selected
type SelectionChance struct {
	MaxThreshold  uint32
	ChancePercent uint32
}

// GetMaxThreshold returns the maximum threshold for which a chancePercent is active
func (sc *SelectionChance) GetMaxThreshold() uint32 {
	return sc.MaxThreshold
}

// GetChancePercent returns the chances percentage to be selected
func (sc *SelectionChance) GetChancePercent() uint32 {
	return sc.ChancePercent
}
