package patriciaMerkleTrie_test

//
//import (
//	"testing"
//
//	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie3"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie"
//	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
//	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
//)
//
//func testTrie() trie3.Trie {
//	tr, _ := patriciaMerkleTrie.NewTrie(keccak.Keccak{}, marshal.JsonMarshalizer{}, nil)
//
//	tr.Update([]byte("doe"), []byte("reindeer"))
//	tr.Update([]byte("dog"), []byte("puppy"))
//	tr.Update([]byte("dogglesworth"), []byte("cat"))
//
//	return tr
//}
//
//func TestNodeIterator_newIterator(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	assert.NotNil(t, it)
//}
//
//func TestNodeIterator_Hash(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	assert.Equal(t, []byte{}, it.Hash())
//
//	it.Next()
//
//	assert.NotEqual(t, []byte{}, it.Hash())
//}
//
//func TestNodeIterator_Parent(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	rootParent := it.Hash()
//
//	it.Next()
//	parent := it.Hash()
//
//	it.Next()
//	actual := it.Parent()
//
//	assert.Equal(t, parent, actual)
//	assert.Equal(t, []byte{}, rootParent)
//}
//
//func TestNodeIterator_Path(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	assert.Nil(t, it.Path())
//
//	it.Next()
//	it.Next()
//
//	assert.Equal(t, []byte{6, 4, 6, 15, 6}, it.Path())
//}
//
//func TestNodeIterator_Leaf(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	it.Next()
//	it.Next()
//	assert.False(t, it.Leaf())
//
//	it.Next()
//	it.Next()
//	assert.True(t, it.Leaf())
//}
//
//func TestNodeIterator_LeafKey(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	searchedKey := []byte("doe")
//	var key []byte
//	var err error
//
//	ok, _ := it.Next()
//
//	for ok {
//		if it.Leaf() {
//			key, err = it.LeafKey()
//			break
//		}
//		ok, _ = it.Next()
//	}
//
//	assert.Equal(t, searchedKey, key)
//	assert.Nil(t, err)
//}
//
//func TestNodeIterator_LeafBlob(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	searchedVal := []byte("reindeer")
//	var val []byte
//	var err error
//
//	ok, _ := it.Next()
//
//	for ok {
//		if it.Leaf() {
//			val, err = it.LeafBlob()
//			break
//		}
//		ok, _ = it.Next()
//	}
//
//	assert.Equal(t, searchedVal, val)
//	assert.Nil(t, err)
//}
//
//func TestNodeIterator_LeafProof(t *testing.T) {
//	tr := testTrie()
//	it := tr.NodeIterator()
//
//	var proofs [][][]byte
//
//	ok, _ := it.Next()
//
//	for ok {
//		if it.Leaf() {
//			proof, err := it.LeafProof()
//			assert.Nil(t, err)
//			proofs = append(proofs, proof)
//		}
//		ok, _ = it.Next()
//	}
//
//	ok, err := tr.VerifyProof(proofs[0], []byte("doe"))
//	assert.Nil(t, err)
//	assert.True(t, ok)
//
//	ok, err = tr.VerifyProof(proofs[1], []byte("dogglesworth"))
//	assert.Nil(t, err)
//	assert.True(t, ok)
//
//	ok, err = tr.VerifyProof(proofs[2], []byte("dog"))
//	assert.Nil(t, err)
//	assert.True(t, ok)
//
//	ok, err = tr.VerifyProof(proofs[2], []byte("doge"))
//	assert.Nil(t, err)
//	assert.False(t, ok)
//
//}
