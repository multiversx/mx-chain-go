package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

// EpochHandler -
func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}
