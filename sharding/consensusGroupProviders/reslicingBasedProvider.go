package consensusGroupProviders

import (
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
	valSlice := make([]sharding.Validator, 0, numVal)
	for i := int64(0); i < numVal; i++ {
		randomIdx := randomness % uint64(len(expEligibleListClone))
		valSlice = append(valSlice, expEligibleListClone[randomIdx])
		expEligibleListClone = reslice(expEligibleListClone, int64(randomIdx))
	}

	return valSlice, nil
}

func reslice(slice []sharding.Validator, idx int64) []sharding.Validator {
	startIdx, nbEntries := computeNumAppearancesForValidator(slice, idx)
	endIdx := startIdx + nbEntries - 1

	retSl := append(slice[:startIdx], slice[endIdx:]...)

	return retSl
}
