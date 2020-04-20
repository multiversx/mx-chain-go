package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

func processDebugMissingData(handler dataRetriever.ResolverDebugHandler, topic string, hash []byte, err error) {
	if !handler.Enabled() {
		return
	}

	handler.FailedToResolveData(topic, hash, err)
}
