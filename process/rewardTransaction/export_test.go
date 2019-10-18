package rewardTransaction

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// Hasher will return the hasher of InterceptedRewardTransaction for using in test files
func (irt *InterceptedRewardTransaction) Hasher() hashing.Hasher {
	return irt.hasher
}

// Marshalizer will return the marshalizer of RewardTxInterceptor for using in test files
func (rti *RewardTxInterceptor) Marshalizer() marshal.Marshalizer {
	return rti.marshalizer
}

// BroadcastCallbackHandler will call the broadcast callback handler of RewardTxInterceptor for using in test files
func (rti *RewardTxInterceptor) BroadcastCallbackHandler(buffToSend []byte) {
	rti.broadcastCallbackHandler(buffToSend)
}
