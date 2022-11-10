package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgTrieNodeRequestHandler is the argument structure used to create a new a trie node request handler instance
type ArgTrieNodeRequestHandler struct {
	ArgBaseRequestHandler
}

type trieNodeRequestHandler struct {
	*baseRequestHandler
}

// NewTrieNodeRequestHandler returns a new instance of trie node request handler
func NewTrieNodeRequestHandler(args ArgTrieNodeRequestHandler) (*trieNodeRequestHandler, error) {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return nil, err
	}

	return &trieNodeRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
	}, nil
}

// RequestDataFromHashArray requests trie nodes from other peers having input multiple trie node hashes
func (handler *trieNodeRequestHandler) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := handler.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
			Epoch: epoch,
		},
		hashes,
	)
}

// RequestDataFromReferenceAndChunk requests a trie node's chunk by specifying the reference and the chunk index
func (handler *trieNodeRequestHandler) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:       dataRetriever.HashType,
			Value:      hash,
			ChunkIndex: chunkIndex,
		},
		[][]byte{hash},
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *trieNodeRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
