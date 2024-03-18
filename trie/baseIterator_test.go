package trie_test

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseIterator(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	it, err := trie.NewBaseIterator(tr)
	assert.Nil(t, err)
	assert.NotNil(t, it)
}

func TestNewBaseIteratorNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	var tr common.Trie

	it, err := trie.NewBaseIterator(tr)
	assert.Nil(t, it)
	assert.Equal(t, trie.ErrNilTrie, err)
}

func TestBaseIterator_HasNext(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	_ = tr.Update([]byte("dog"), []byte("dog"))
	trie.ExecuteUpdatesFromBatch(tr)
	it, _ := trie.NewBaseIterator(tr)
	assert.False(t, it.HasNext())

	_ = tr.Update([]byte("doe"), []byte("doe"))
	trie.ExecuteUpdatesFromBatch(tr)
	it, _ = trie.NewBaseIterator(tr)
	assert.True(t, it.HasNext())
}

func TestBaseIterator_GetMarshalizedNode(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	it, _ := trie.NewBaseIterator(tr)

	encNode, err := it.MarshalizedNode()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(encNode))

	branchNodeIdentifier := uint8(2)
	lastByte := len(encNode) - 1
	assert.Equal(t, branchNodeIdentifier, encNode[lastByte])
}

func TestBaseIterator_GetHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.RootHash()
	it, _ := trie.NewBaseIterator(tr)

	hash, err := it.GetHash()
	assert.Nil(t, err)
	assert.Equal(t, rootHash, hash)
}

func TestIterator_Search(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Update([]byte("ddoge"), []byte("foo"))
	trie.ExecuteUpdatesFromBatch(tr)

	expectedHashes := []string{
		"ecc2304769996585131ad6276c1422265813a2b79d60392130c4baa19a9b4e06",
		"4461526b45b5c399db2792b4bcbaef10e94e5164cd5bc73cf92907b6e792a717",
		"bb50a879071cef1050f1d10f3d9ea4ddbdb48ccb150341b8a1c79db44d5011d3",
		"b06c8aa3fabebd819d0d24d682edb3a055759a33327aee6afd754cc34c6f61fc",
		"828790d06af95af1d8620b060b65c1166dc164235e737f73e776949b0802cd62",
		"3ce38be296cd56a9caec202554c0d94c2df1367941be37ecbe5e089c29188f59",
		"fdab2ad38702ee7b55b8b1d4956f9e9efdec6a201a04ebd8a7709ded07435492",
		"5d153a4f4cbf552ec3b5640ca429a33ab0f0955ae7bfdc03d8b5820522108270",
		"7121fb99a2c85ff3370e8ebf73a23a9f389a767891fece807c506f03abb1bc67",
	}

	t.Run("dfs iterator search", func(t *testing.T) {
		t.Parallel()

		expectedHashesDFSOrder := []string{
			expectedHashes[0],
			expectedHashes[1],
			expectedHashes[3],
			expectedHashes[5],
			expectedHashes[6],
			expectedHashes[2],
			expectedHashes[4],
			expectedHashes[7],
			expectedHashes[8],
		}

		it, _ := trie.NewDFSIterator(tr)
		nodeHash, err := it.GetHash()
		require.Nil(t, err)

		nodesHashes := make([]string, 0)
		nodesHashes = append(nodesHashes, hex.EncodeToString(nodeHash))

		for it.HasNext() {
			err := it.Next()
			require.Nil(t, err)

			nodeHash, err := it.GetHash()
			require.Nil(t, err)

			nodesHashes = append(nodesHashes, hex.EncodeToString(nodeHash))
		}

		require.Equal(t, expectedHashesDFSOrder, nodesHashes)
	})
}
