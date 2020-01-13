package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

//TODO: implement a real epoch provider by nonce

// simpleEpochProviderByNonce will use the epoch handler to return the current epoch
type simpleEpochProviderByNonce struct {
	epochHandler dataRetriever.EpochHandler
}

// NewSimpleEpochProviderByNonce will return a new instance of simpleEpochProviderByNonce
func NewSimpleEpochProviderByNonce(epochHandler dataRetriever.EpochHandler) *simpleEpochProviderByNonce {
	return &simpleEpochProviderByNonce{
		epochHandler: epochHandler,
	}
}

// EpochForNonce will return the current epoch from the epoch handler
func (sepbn *simpleEpochProviderByNonce) EpochForNonce(nonce uint64) (uint32, error) {
	return sepbn.epochHandler.Epoch(), nil
}

// IsInterfaceNil returns true if there is not value under the interface
func (sepbn *simpleEpochProviderByNonce) IsInterfaceNil() bool {
	return sepbn == nil
}
