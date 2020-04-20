package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

func ProcessDebugMissingData(handler dataRetriever.ResolverDebugHandler, topic string, hash []byte, err error) {
	processDebugMissingData(handler, topic, hash, err)
}

func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}
