package patriciaMerkleTrie

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie/encoding"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// PatriciaMerkleTree is an implementation of Trie interface
type PatriciaMerkleTree struct {
	root        node
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// New returns a new PatriciaMerkleTree with the given hasher and marshalizer. If hasher and marshalizer
// are nil, New uses keccak hasher and json marshalizer
func New(hsh hashing.Hasher, msh marshal.Marshalizer) *PatriciaMerkleTree {

	if hsh == nil {
		hsh = keccak.Keccak{}
	}

	if msh == nil {
		msh = marshal.JsonMarshalizer{}
	}

	return &PatriciaMerkleTree{hasher: hsh, marshalizer: msh}
}

// NodeIterator returns a new node iterator for the current trie
func (tr *PatriciaMerkleTree) NodeIterator() trie2.NodeIterator {
	return newNodeIterator(tr)
}

// Root returns the hash of the root node
func (tr *PatriciaMerkleTree) Root() []byte {
	_, h, _ := tr.setHash(tr.root)
	return h.getHash()
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *PatriciaMerkleTree) Update(key, value []byte) error {
	k := encoding.KeyBytesToHex(key)
	if len(value) != 0 {
		_, n, err := tr.insert(tr.root, nil, k, valueNode(value))
		if err != nil {
			return err
		}
		tr.root = n
	} else {
		_, n, err := tr.delete(tr.root, nil, k)
		if err != nil {
			return err
		}
		tr.root = n
	}
	return nil
}

func (tr *PatriciaMerkleTree) insert(n node, prefix, key []byte, value node) (bool, node, error) {
	if len(key) == 0 {
		if v, ok := n.(valueNode); ok {
			return !bytes.Equal(v, value.(valueNode)), value, nil
		}
		return true, value, nil
	}
	switch n := n.(type) {
	case *shortNode:
		matchlen := encoding.PrefixLen(key, n.Key)
		// If the whole key matches, keep this short node as is
		// and only update the value.
		if matchlen == len(n.Key) {
			dirty, nn, err := tr.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &shortNode{n.Key, nn, nil, typeNode}, nil
		}
		// Otherwise branch out at the index where they differ.
		branch := &fullNode{}
		var err error
		_, branch.Children[n.Key[matchlen]], err = tr.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
		if err != nil {
			return false, nil, err
		}
		_, branch.Children[key[matchlen]], err = tr.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if matchlen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return true, &shortNode{key[:matchlen], branch, nil, typeNode}, nil

	case *fullNode:
		dirty, nn, err := tr.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.Children[key[0]] = nn
		return true, n, nil

	case nil:
		return true, &shortNode{key, value, nil, typeVal}, nil

	default:
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

func (tr *PatriciaMerkleTree) delete(n node, prefix, key []byte) (bool, node, error) {
	switch n := n.(type) {
	case *shortNode:
		matchlen := encoding.PrefixLen(key, n.Key)
		if matchlen < len(n.Key) {
			return false, n, nil // don't replace n on mismatch
		}
		if matchlen == len(key) {
			return true, nil, nil // remove n entirely for whole matches
		}
		// The key is longer than n.Key. Remove the remaining suffix
		// from the subtrie. Child can never be nil here since the
		// subtrie must contain at least two other values with keys
		// longer than n.Key.
		dirty, child, err := tr.delete(n.Val, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if !dirty || err != nil {
			return false, n, err
		}
		switch child := child.(type) {
		case *shortNode:
			// Deleting from the subtrie reduced it to another
			// short node. Merge the nodes to avoid creating a
			// shortNode{..., shortNode{...}}. Use concat (which
			// always creates a new slice) instead of append to
			// avoid modifying n.Key since it might be shared with
			// other nodes.
			return true, &shortNode{concat(n.Key, child.Key...), child.Val, nil, typeVal}, nil
		default:
			return true, &shortNode{n.Key, child, nil, typeNode}, nil
		}

	case *fullNode:
		dirty, nn, err := tr.delete(n.Children[key[0]], append(prefix, key[0]), key[1:])
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.hash = nil
		n.Children[key[0]] = nn

		// Check how many non-nil entries are left after deleting and
		// reduce the full node to a short node if only one entry is
		// left. Since n must've contained at least two children
		// before deletion (otherwise it would not be a full node) n
		// can never be reduced to nil.
		//
		// When the loop is done, pos contains the index of the single
		// value that is left in n or -2 if n contains at least two
		// values.
		pos := -1
		for i, cld := range &n.Children {
			if cld != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				// If the remaining entry is a short node, it replaces
				// n and its key gets the missing nibble tacked to the
				// front. This avoids creating an invalid
				// shortNode{..., shortNode{...}}.
				cnode := n.Children[pos]
				if err != nil {
					return false, nil, err
				}
				if cnode, ok := cnode.(*shortNode); ok {
					k := append([]byte{byte(pos)}, cnode.Key...)
					return true, &shortNode{k, cnode.Val, nil, typeVal}, nil
				}
			}
			// Otherwise, n is replaced by a one-nibble short node
			// containing the child.
			return true, &shortNode{[]byte{byte(pos)}, n.Children[pos], nil, typeVal}, nil
		}
		// n still contains at least two values and cannot be reduced.
		return true, n, nil

	case valueNode:
		return true, nil, nil

	case nil:
		return false, nil, nil

	default:
		panic(fmt.Sprintf("%T: invalid node: %v (%v)", n, n, key))
	}
}

// Delete removes any existing value for key from the trie.
func (tr *PatriciaMerkleTree) Delete(key []byte) error {
	k := encoding.KeyBytesToHex(key)
	_, n, err := tr.delete(tr.root, nil, k)
	if err != nil {
		return err
	}
	tr.root = n
	return nil
}

// Get returns the value for key stored in the trie.
func (tr *PatriciaMerkleTree) Get(key []byte) ([]byte, error) {
	key = encoding.KeyBytesToHex(key)
	value, newroot, didResolve, err := tr.tryGet(tr.root, key, 0)
	if err == nil && didResolve {
		tr.root = newroot
	}
	return value, err
}

func (tr *PatriciaMerkleTree) tryGet(origNode node, key []byte, pos int) (value []byte, newnode node, didResolve bool, err error) {
	switch n := (origNode).(type) {
	case nil:
		return nil, nil, false, nil
	case valueNode:
		return n, n, false, nil
	case *shortNode:
		if len(key)-pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos+len(n.Key)]) {
			// key not found in trie
			return nil, n, false, nil
		}
		value, newnode, didResolve, err = tr.tryGet(n.Val, key, pos+len(n.Key))
		if err == nil && didResolve {
			n = n.copy()
			n.Val = newnode
		}
		return value, n, didResolve, err
	case *fullNode:
		value, newnode, didResolve, err = tr.tryGet(n.Children[key[pos]], key, pos+1)
		if err == nil && didResolve {
			n = n.copy()
			n.Children[key[pos]] = newnode
		}
		return value, n, didResolve, err
	default:
		panic(fmt.Sprintf("%T: invalid node: %v", origNode, origNode))
	}
}

// Copy returns a copy of Trie.
func (tr *PatriciaMerkleTree) Copy() trie2.Trie {
	cpy := *tr
	return &cpy
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}

// setHash collapses a node down into a hash node, also returning a copy of the
// original node initialized with the computed hash to replace the original one.
func (tr *PatriciaMerkleTree) setHash(n node) (node, node, error) {
	if hash := n.getHash(); hash != nil {
		return hashNode(hash), n, nil
	}
	collapsed, node, err := tr.hashChildren(n)
	if err != nil {
		return hashNode{}, n, err
	}

	hashed, err := tr.hash(collapsed)
	if err != nil {
		return hashNode{}, n, err
	}

	cachedHash, _ := hashed.(hashNode)
	switch n := node.(type) {
	case *shortNode:
		n.hash = cachedHash
	case *fullNode:
		n.hash = cachedHash
	}

	return hashed, node, nil
}

// hashChildren replaces the children of a node with their hashes, returning the collapsed node as well
// as a replacement for the original node with the child hashes cached in.
func (tr *PatriciaMerkleTree) hashChildren(original node) (node, node, error) {
	var err error

	switch n := original.(type) {
	case *shortNode:
		collapsed, node := n.copy(), n.copy()

		if _, ok := n.Val.(valueNode); !ok {
			collapsed.Val, node.Val, err = tr.setHash(n.Val)
			if err != nil {
				return original, original, err
			}
		}
		return collapsed, node, nil

	case *fullNode:
		// Hash the full node's children, caching the newly hashed subtrees
		collapsed, node := n.copy(), n.copy()

		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				collapsed.Children[i], node.Children[i], err = tr.setHash(n.Children[i])
				if err != nil {
					return original, original, err
				}
			}
		}
		node.Children[16] = n.Children[16]
		return collapsed, node, nil

	default:
		// Value and hash nodes don't have children so they're left as were
		return n, original, nil
	}

}

// hash collapses any node into a hash
func (tr *PatriciaMerkleTree) hash(n node) (node, error) {
	if _, isHash := n.(hashNode); n == nil || isHash {
		return n, nil
	}

	tmp, err := tr.marshalizer.Marshal(n)

	if err != nil {
		return nil, err
	}

	hash := n.getHash()
	if hash == nil {
		hash = tr.hasher.Compute(string(tmp))
	}

	return hashNode(hash), nil
}

// short is a struct that is used in decoding a short node
type short struct {
	Key, Val []byte
	ValType  bool
}

// full is a struct that is used in decoding a full node
type full struct {
	Children [nrOfChildren][]byte
}

// Prove returns the Merkle proof for the given key
func (tr *PatriciaMerkleTree) Prove(key []byte) [][]byte {
	it := newNodeIterator(tr)

	for it.Next() {
		if it.Leaf() && bytes.Equal(it.LeafKey(), key) {
			return it.LeafProof()
		}
	}

	return nil
}

// VerifyProof checks Merkle proofs. The given proof must contain the value for
// key in a trie with the given root hash.
func (tr *PatriciaMerkleTree) VerifyProof(rootHash []byte, proofs [][]byte, key []byte) bool {
	key = encoding.KeyBytesToHex(key)
	wantHash := rootHash
	for i := range proofs {
		buf := proofs[i]
		if buf == nil {
			return false
		}
		n := decodeNode(buf, tr.marshalizer)

		keyrest, cld := get(n, key)
		switch cld := cld.(type) {
		case nil:
			// The trie doesn't contain the key.
			return false
		case hashNode:
			key = keyrest
			copy(wantHash[:], cld)
		case valueNode:
			return true
		}
	}
	return false
}

func decodeNode(proof []byte, msh marshal.Marshalizer) node {
	decodeShort := &short{}
	decodeFull := &full{}

	err := msh.Unmarshal(decodeShort, proof)
	if err != nil || isEmpty(decodeShort) {
		err = msh.Unmarshal(decodeFull, proof)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		fullN := &fullNode{}
		for i := 0; i < nrOfChildren-1; i++ {
			if decodeFull.Children[i] == nil {
				fullN.Children[i] = nil
			} else {
				fullN.Children[i] = hashNode(decodeFull.Children[i])
			}
		}
		fullN.Children[16] = valueNode(decodeFull.Children[16])
		return fullN
	}

	if decodeShort.ValType == typeVal {
		return &shortNode{decodeShort.Key, valueNode(decodeShort.Val), nil, typeVal}
	}
	return &shortNode{decodeShort.Key, hashNode(decodeShort.Val), nil, typeNode}

}

func isEmpty(node *short) bool {
	emptyShort := &short{}
	return bytes.Equal(emptyShort.Key, node.Key)
}

func get(tn node, key []byte) ([]byte, node) {
	for {
		switch n := tn.(type) {
		case *shortNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				return nil, nil
			}
			tn = n.Val
			key = key[len(n.Key):]
		case *fullNode:
			tn = n.Children[key[0]]
			key = key[1:]
		case hashNode:
			return key, n
		case nil:
			return key, nil
		case valueNode:
			return nil, n
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}
}
