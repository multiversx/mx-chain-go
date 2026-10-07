package requesters

import (
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// ArgTrieNodeRequester is the argument structure used to create a new trie node requester instance
type ArgTrieNodeRequester struct {
	ArgBaseRequester
}

type trieNodeRequester struct {
	*baseRequester
}

// NewTrieNodeRequester returns a new instance of trie node requester
func NewTrieNodeRequester(args ArgTrieNodeRequester) (*trieNodeRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &trieNodeRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests trie nodes from other peers by having input multiple trie node hashes
func (requester *trieNodeRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

func (requester *trieNodeRequester) RequestDataFromHashArrayForRecovery(hashes [][]byte, epoch uint32) error {
	buff, err := requester.marshaller.Marshal(&batch.Batch{Data: hashes})
	if err != nil {
		return err
	}
	return requester.SendOnRequestTopicIncludingMainPeers(&dataRetriever.RequestData{
		Type: dataRetriever.HashArrayType, Value: buff, Epoch: epoch,
	}, hashes)
}

func (requester *trieNodeRequester) RequestDataFromReferenceAndChunkForRecovery(hash []byte, chunkIndex uint32) error {
	return requester.SendOnRequestTopicIncludingMainPeers(&dataRetriever.RequestData{
		Type: dataRetriever.HashType, Value: hash, ChunkIndex: chunkIndex,
	}, [][]byte{hash})
}

// RequestDataFromReferenceAndChunk requests a trie node's chunk by specifying the reference and the chunk index
func (requester *trieNodeRequester) RequestDataFromReferenceAndChunk(hash []byte, chunkIndex uint32) error {
	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:       dataRetriever.HashType,
			Value:      hash,
			ChunkIndex: chunkIndex,
		},
		[][]byte{hash},
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *trieNodeRequester) IsInterfaceNil() bool {
	return requester == nil
}
