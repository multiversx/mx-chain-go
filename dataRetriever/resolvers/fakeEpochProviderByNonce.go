package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

//TODO: implement a real epoch provider by nonce

// fakeEpochProviderByNonce will use the epoch handler to return the current epoch
type fakeEpochProviderByNonce struct {
	epochHandler dataRetriever.EpochHandler
}

// NewFakeEpochProviderByNonce will return a new instance of fakeEpochProviderByNonce
func NewFakeEpochProviderByNonce(epochHandler dataRetriever.EpochHandler) *fakeEpochProviderByNonce {
	return &fakeEpochProviderByNonce{
		epochHandler: epochHandler,
	}
}

// EpochForNonce will return the current epoch from the epoch handler
func (fepbn *fakeEpochProviderByNonce) EpochForNonce(nonce uint64) (uint32, error) {
	return fepbn.epochHandler.Epoch(), nil
}

// IsInterfaceNil returns true if there is not value under the interface
func (fepbn *fakeEpochProviderByNonce) IsInterfaceNil() bool {
	return fepbn == nil
}
