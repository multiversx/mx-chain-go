package consensusGroupProviders

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

func computeNumAppearancesForValidator(expEligibleList []sharding.Validator, idx int64) (int64, int64) {
	val := expEligibleList[idx].PubKey()
	startIdx := idx
	listLen := int64(len(expEligibleList))

	for i := idx - 1; i >= 0; i-- {
		if !bytes.Equal(expEligibleList[idx].PubKey(), val) {
			startIdx = idx + 1
			break
		}
	}

	var endIdx int64
	for i := idx + 1; i < listLen; i++ {
		if !bytes.Equal(expEligibleList[idx].PubKey(), val) {
			endIdx = idx
			break
		}
	}

	return startIdx, endIdx - startIdx + 1
}
