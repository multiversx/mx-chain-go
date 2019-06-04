package trie2

import (
	"bytes"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/capnp"
	protobuf "github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/proto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	capn "github.com/glycerine/go-capnproto"
)

// Save saves the serialized data of a leaf node into a stream through Capnp protocol
func (ln *leafNode) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	leafNodeGoToCapn(seg, ln)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a leaf node object through Capnp protocol
func (ln *leafNode) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootLeafNodeCapn(capMsg)
	leafNodeCapnToGo(z, ln)
	return nil
}

func leafNodeGoToCapn(seg *capn.Segment, src *leafNode) capnp.LeafNodeCapn {
	dest := capnp.AutoNewLeafNodeCapn(seg)

	dest.SetKey(src.Key)
	dest.SetValue(src.Value)

	return dest
}

func leafNodeCapnToGo(src capnp.LeafNodeCapn, dest *leafNode) *leafNode {
	if dest == nil {
		dest = &leafNode{}
	}

	dest.Value = src.Value()
	dest.Key = src.Key()

	return dest
}

func newLeafNode(key, value []byte) *leafNode {
	return &leafNode{
		CollapsedLn: protobuf.CollapsedLn{
			Key:   key,
			Value: value,
		},
		hash:  nil,
		dirty: true,
	}
}

func (ln *leafNode) getHash() []byte {
	return ln.hash
}

func (ln *leafNode) isDirty() bool {
	return ln.dirty
}

func (ln *leafNode) getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	return ln, nil
}

func (ln *leafNode) setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return err
	}
	if ln.getHash() != nil {
		return nil
	}
	hash, err := hashChildrenAndNode(ln, marshalizer, hasher)
	if err != nil {
		return err
	}
	ln.hash = hash
	return nil
}

func (ln *leafNode) setHashConcurrent(marshalizer marshal.Marshalizer, hasher hashing.Hasher, wg *sync.WaitGroup, c chan error) {
	err := ln.setHash(marshalizer, hasher)
	if err != nil {
		c <- err
	}
	wg.Done()
}

func (ln *leafNode) setRootHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	return ln.setHash(marshalizer, hasher)
}

func (ln *leafNode) hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	return nil
}

func (ln *leafNode) hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	return encodeNodeAndGetHash(ln, marshalizer, hasher)
}

func (ln *leafNode) commit(level byte, db DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return err
	}
	if !ln.dirty {
		return nil
	}
	ln.dirty = false
	return encodeNodeAndCommitToDB(ln, db, marshalizer, hasher)
}

func (ln *leafNode) getEncodedNode(marshalizer marshal.Marshalizer) ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := marshalizer.Marshal(ln)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, leaf)
	return marshaledNode, nil
}

func (ln *leafNode) resolveCollapsed(pos byte, db DBWriteCacher, marshalizer marshal.Marshalizer) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(pos int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, nil
	}
	return nil, ErrNodeNotFound
}

func (ln *leafNode) getNext(key []byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}

func (ln *leafNode) insert(n *leafNode, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if bytes.Equal(n.Key, ln.Key) {
		ln.Value = n.Value
		ln.dirty = true
		ln.hash = nil
		return true, ln, nil
	}

	keyMatchLen := prefixLen(n.Key, ln.Key)
	branch := newBranchNode()
	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, ErrChildPosOutOfRange
	}

	branch.children[oldChildPos] = newLeafNode(ln.Key[keyMatchLen+1:], ln.Value)
	branch.children[newChildPos] = newLeafNode(n.Key[keyMatchLen+1:], n.Value)

	if keyMatchLen == 0 {
		return true, branch, nil
	}
	return true, newExtensionNode(ln.Key[:keyMatchLen], branch), nil
}

func (ln *leafNode) delete(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	keyMatchLen := prefixLen(key, ln.Key)
	if keyMatchLen == len(key) {
		return true, nil, nil
	}
	return false, ln, nil
}

func (ln *leafNode) reduceNode(pos int) node {
	k := append([]byte{byte(pos)}, ln.Key...)
	return newLeafNode(k, ln.Value)
}

func (ln *leafNode) isEmptyOrNil() error {
	if ln == nil {
		return ErrNilNode
	}
	if ln.Value == nil {
		return ErrEmptyNode
	}
	return nil
}
