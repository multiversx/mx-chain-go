package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// ArgRequestingOnlyTrieNodeResolver is the argument structure used to create new RequestingOnlyTrieNodeResolver instance
type ArgRequestingOnlyTrieNodeResolver struct {
	ArgBaseResolver
}

// requestingOnlyTrieNodeResolver is a wrapper over Resolver that is specialized in requesting trie node
type requestingOnlyTrieNodeResolver struct {
	*baseResolver
	marshaller marshal.Marshalizer
}

// NewRequestingOnlyTrieNodeResolver creates a new requesting only trie node resolver
func NewRequestingOnlyTrieNodeResolver(args ArgRequestingOnlyTrieNodeResolver) (*requestingOnlyTrieNodeResolver, error) {
	err := checkArgBase(args.ArgBaseResolver)
	if err != nil {
		return nil, err
	}

	return &requestingOnlyTrieNodeResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: args.SenderResolver,
		},
		marshaller: args.Marshaller,
	}, nil
}

// RequestDataFromHash requests trie nodes from other peers having input a trie node hash
func (res *requestingOnlyTrieNodeResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
		},
		[][]byte{hash},
	)
}

// RequestDataFromHashArray requests trie nodes from other peers having input multiple trie node hashes
func (res *requestingOnlyTrieNodeResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := res.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
		},
		hashes,
	)
}

// RequestDataFromReferenceAndChunk requests a trie node's chunk by specifying the reference and the chunk index
func (res *requestingOnlyTrieNodeResolver) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:       dataRetriever.HashType,
			Value:      hash,
			ChunkIndex: chunkIndex,
		},
		[][]byte{hash},
	)
}

// ProcessReceivedMessage returns nil
func (res *requestingOnlyTrieNodeResolver) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *requestingOnlyTrieNodeResolver) IsInterfaceNil() bool {
	return res == nil
}
