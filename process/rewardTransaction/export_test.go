package rewardTransaction

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
)

// Hasher will return the hasher of InterceptedRewardTransaction for using in test files
func (inRTx *InterceptedRewardTransaction) Hasher() hashing.Hasher {
	return inRTx.hasher
}
