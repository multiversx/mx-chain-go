package trie2

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/encoding"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
)

var hasher = keccak.Keccak{}
var marshalizer = marshal.JsonMarshalizer{}

type PatriciaMerkleTree struct {
	root Node
}

func NewTrie() *PatriciaMerkleTree {
	trie := &PatriciaMerkleTree{}
	return trie
}

func (tr *PatriciaMerkleTree) Root() []byte {
	h, _ := SetHash(tr.root)
	return h.GetHash()
}

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
// The value bytes must not be modified by the caller.
// If a node was not found in the database, a MissingNodeError is returned.
func (tr *PatriciaMerkleTree) Get(key []byte) ([]byte, error) {
	key = encoding.KeyBytesToHex(key)
	value, newroot, didResolve, err := tr.tryGet(tr.root, key, 0)
	if err == nil && didResolve {
		tr.root = newroot
	}
	return value, err
}

// Copy returns a copy of Trie.
func (tr *PatriciaMerkleTree) Copy() Trie {
	cpy := *tr
	return &cpy
}

func (tr *PatriciaMerkleTree) tryGet(origNode Node, key []byte, pos int) (value []byte, newnode Node, didResolve bool, err error) {
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

func (tr *PatriciaMerkleTree) insert(n Node, prefix, key []byte, value Node) (bool, Node, error) {
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
			return true, &shortNode{n.Key, nn, nil}, nil
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
		return true, &shortNode{key[:matchlen], branch, nil}, nil

	case *fullNode:
		dirty, nn, err := tr.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.Children[key[0]] = nn
		return true, n, nil

	case nil:
		return true, &shortNode{key, value, nil}, nil

	default:
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

func (tr *PatriciaMerkleTree) delete(n Node, prefix, key []byte) (bool, Node, error) {
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
			return true, &shortNode{concat(n.Key, child.Key...), child.Val, nil}, nil
		default:
			return true, &shortNode{n.Key, child, nil}, nil
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
					return true, &shortNode{k, cnode.Val, nil}, nil
				}
			}
			// Otherwise, n is replaced by a one-nibble short node
			// containing the child.
			return true, &shortNode{[]byte{byte(pos)}, n.Children[pos], nil}, nil
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

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}
