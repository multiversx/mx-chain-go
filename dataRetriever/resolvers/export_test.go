package resolvers

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// EpochHandler -
func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}

// ResolveMultipleHashes -
func (tnRes *TrieNodeResolver) ResolveMultipleHashes(hashesBuff []byte, message p2p.MessageP2P) error {
	return tnRes.resolveMultipleHashes(hashesBuff, message)
}
