package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func getDefaultInterceptedTrieNodeParameters() ([]byte, data.DBWriteCacher, marshal.Marshalizer, hashing.Hasher) {
	tr := initTrie()
	nodes, _ := getEncodedTrieNodesAndHashes(tr)

	return nodes[0], memorydb.New(), &mock.ProtobufMarshalizerMock{}, &mock.KeccakMock{}
}

func getEncodedTrieNodesAndHashes(tr data.Trie) ([][]byte, [][]byte) {
	it, _ := trie.NewIterator(tr)
	encNode, _ := it.GetMarshalizedNode()

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)

	hashes := make([][]byte, 0)
	hash, _ := it.GetHash()
	hashes = append(hashes, hash)

	for it.HasNext() {
		_ = it.Next()
		encNode, _ = it.GetMarshalizedNode()

		nodes = append(nodes, encNode)
		hash, _ = it.GetHash()
		hashes = append(hashes, hash)
	}

	return nodes, hashes
}

func TestNewInterceptedTrieNode_EmptyBufferShouldFail(t *testing.T) {
	t.Parallel()

	_, db, marsh, hasher := getDefaultInterceptedTrieNodeParameters()
	interceptedNode, err := trie.NewInterceptedTrieNode([]byte{}, db, marsh, hasher)
	assert.Nil(t, interceptedNode)
	assert.Equal(t, trie.ErrValueTooShort, err)
}

func TestNewInterceptedTrieNode_NilDbShouldFail(t *testing.T) {
	t.Parallel()

	buff, _, marsh, hasher := getDefaultInterceptedTrieNodeParameters()
	interceptedNode, err := trie.NewInterceptedTrieNode(buff, nil, marsh, hasher)
	assert.Nil(t, interceptedNode)
	assert.Equal(t, trie.ErrNilDatabase, err)
}

func TestNewInterceptedTrieNode_NilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	buff, db, _, hasher := getDefaultInterceptedTrieNodeParameters()
	interceptedNode, err := trie.NewInterceptedTrieNode(buff, db, nil, hasher)
	assert.Nil(t, interceptedNode)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewInterceptedTrieNode_NilHasherShouldFail(t *testing.T) {
	t.Parallel()

	buff, db, marsh, _ := getDefaultInterceptedTrieNodeParameters()
	interceptedNode, err := trie.NewInterceptedTrieNode(buff, db, marsh, nil)
	assert.Nil(t, interceptedNode)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewInterceptedTrieNode_OkParametersShouldWork(t *testing.T) {
	t.Parallel()

	interceptedNode, err := trie.NewInterceptedTrieNode(getDefaultInterceptedTrieNodeParameters())
	assert.NotNil(t, interceptedNode)
	assert.Nil(t, err)
}

func TestInterceptedTrieNode_CheckValidity(t *testing.T) {
	t.Parallel()

	interceptedNode, _ := trie.NewInterceptedTrieNode(getDefaultInterceptedTrieNodeParameters())

	err := interceptedNode.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedTrieNode_Hash(t *testing.T) {
	t.Parallel()

	interceptedNode, _ := trie.NewInterceptedTrieNode(getDefaultInterceptedTrieNodeParameters())
	tr := initTrie()
	_, hashes := getEncodedTrieNodesAndHashes(tr)

	hash := interceptedNode.Hash()
	assert.Equal(t, hashes[0], hash)
}

func TestInterceptedTrieNode_EncodedNode(t *testing.T) {
	t.Parallel()

	interceptedNode, _ := trie.NewInterceptedTrieNode(getDefaultInterceptedTrieNodeParameters())
	tr := initTrie()
	nodes, _ := getEncodedTrieNodesAndHashes(tr)

	encNode := interceptedNode.EncodedNode()
	assert.Equal(t, nodes[0], encNode)
}
