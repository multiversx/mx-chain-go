package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// EpochHandler -
func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}

// ResolveMultipleHashes -
func (tnRes *TrieNodeResolver) ResolveMultipleHashes(hashesBuff []byte, message core.MessageP2P) error {
	return tnRes.resolveMultipleHashes(hashesBuff, message)
}
