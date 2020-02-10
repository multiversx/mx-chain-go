package consensusGroupProviders

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

type reslicingBasedProvider struct {
}

func NewReslicingBasedProvider() *reslicingBasedProvider {
	return &reslicingBasedProvider{}
}

func (rbp *reslicingBasedProvider) Get(randomness uint64, numVal int64, expEligibleList []sharding.Validator) ([]sharding.Validator, error) {
	expEligibleListClone := make([]sharding.Validator, len(expEligibleList))
	copy(expEligibleListClone, expEligibleList)
	valSlice := make([]sharding.Validator, 0)
	for i := int64(0); i < numVal; i++ {
		randomIdx := randomness % uint64(len(expEligibleListClone))
		valSlice = append(valSlice, expEligibleListClone[randomIdx])
		expEligibleListClone = reslice(expEligibleListClone, int64(randomIdx))
	}

	return valSlice, nil
}

func reslice(slice []sharding.Validator, idx int64) []sharding.Validator {
	originalIdx := idx
	val := slice[idx].Address()
	var startIdx int64
	// delete before
	if idx == 0 {
		startIdx = 0
	} else {
		for {
			idx--
			if idx == 0 || !bytes.Equal(slice[idx].Address(), val) {
				startIdx = idx + 1
				break
			}
		}
	}

	idx = originalIdx
	var endIdx int64
	// delete after
	if idx == int64(len(slice)) {
		endIdx = idx
	} else {
		for {
			idx++
			if idx == int64(len(slice)) || !bytes.Equal(slice[idx].Address(), val) {
				endIdx = idx
				break
			}
		}
	}

	retSl := append(slice[:startIdx], slice[endIdx:]...)

	return retSl
}
