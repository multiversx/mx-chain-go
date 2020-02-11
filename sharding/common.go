package sharding

import (
	"bytes"
)

func computeStartIndexAndNumAppearancesForValidator(expEligibleList []Validator, idx int64) (int64, int64) {
	val := expEligibleList[idx].PubKey()
	startIdx := int64(0)
	listLen := int64(len(expEligibleList))

	for i := idx - 1; i >= 0; i-- {
		if !bytes.Equal(expEligibleList[i].PubKey(), val) {
			startIdx = i + 1
			break
		}
	}

	endIdx := listLen - 1
	for i := idx + 1; i < listLen; i++ {
		if !bytes.Equal(expEligibleList[i].PubKey(), val) {
			endIdx = i - 1
			break
		}
	}

	return startIdx, endIdx - startIdx + 1
}
