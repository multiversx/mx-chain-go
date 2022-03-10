package metachain

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/state"
)

type validatorList []*state.ValidatorInfo

// Len will return the length of the validatorList
func (v validatorList) Len() int { return len(v) }

// Swap will interchange the objects on input indexes
func (v validatorList) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of validators should be by index and public key
func (v validatorList) Less(i, j int) bool {
	if v[i].TempRating == v[j].TempRating {
		if v[i].Index == v[j].Index {
			return bytes.Compare(v[i].PublicKey, v[j].PublicKey) < 0
		}
		return v[i].Index < v[j].Index
	}
	return v[i].TempRating < v[j].TempRating
}
