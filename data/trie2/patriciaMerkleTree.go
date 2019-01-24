package trie2

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

var hasher = keccak.Keccak{}

type PatriciaMerkleTree struct {
	root   Node
	hasher hashing.Hasher
}

func NewTrie(hsh hashing.Hasher) *PatriciaMerkleTree {

	if hsh == nil {
		hsh = hasher
	}

	trie := &PatriciaMerkleTree{
		hasher: hsh,
	}

	return trie
}

func (t *PatriciaMerkleTree) Update(key, value []byte) error {
	k := KeyBytesToHex(key)
	if len(value) != 0 {
		_, n, err := t.insert(t.root, nil, k, valueNode(value))
		if err != nil {
			return err
		}
		t.root = n
	}
	//else {
	//	_, n, err := t.delete(t.root, nil, k)
	//	if err != nil {
	//		return err
	//	}
	//	t.root = n
	//}
	return nil
}

func (t *PatriciaMerkleTree) insert(n Node, prefix, key []byte, value Node) (bool, Node, error) {
	if len(key) == 0 {
		if v, ok := n.(valueNode); ok {
			return !bytes.Equal(v, value.(valueNode)), value, nil
		}
		return true, value, nil
	}
	switch n := n.(type) {
	case *shortNode:
		matchlen := PrefixLen(key, n.Key)
		// If the whole key matches, keep this short node as is
		// and only update the value.
		if matchlen == len(n.Key) {
			dirty, nn, err := t.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &shortNode{n.Key, nn}, nil
		}
		// Otherwise branch out at the index where they differ.
		branch := &fullNode{}
		var err error
		_, branch.Children[n.Key[matchlen]], err = t.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
		if err != nil {
			return false, nil, err
		}
		_, branch.Children[key[matchlen]], err = t.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if matchlen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return true, &shortNode{key[:matchlen], branch}, nil

	case *fullNode:
		dirty, nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.Children[key[0]] = nn
		return true, n, nil

	case nil:
		return true, &shortNode{key, value}, nil

	default:
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}
