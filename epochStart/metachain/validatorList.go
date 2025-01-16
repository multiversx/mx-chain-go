package metachain

import (
	"bytes"

	"github.com/multiversx/mx-chain-go/state"
)

type validatorList []state.ValidatorInfoHandler

// Len will return the length of the validatorList
func (v validatorList) Len() int { return len(v) }

// Swap will interchange the objects on input indexes
func (v validatorList) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of validators should be by index and public key
func (v validatorList) Less(i, j int) bool {
	if v[i].GetTempRating() == v[j].GetTempRating() {
		if v[i].GetIndex() == v[j].GetIndex() {
			return bytes.Compare(v[i].GetPublicKey(), v[j].GetPublicKey()) < 0
		}
		return v[i].GetIndex() < v[j].GetIndex()
	}
	return v[i].GetTempRating() < v[j].GetTempRating()
}
