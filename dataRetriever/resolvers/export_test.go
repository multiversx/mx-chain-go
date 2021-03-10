package resolvers

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

// MaxBuffToSendTrieNodes -
var MaxBuffToSendTrieNodes = maxBuffToSendTrieNodes

func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}
