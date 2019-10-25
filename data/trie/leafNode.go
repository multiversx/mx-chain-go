package trie

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie/capnp"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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

func (ln *leafNode) setGivenHash(hash []byte) {
	ln.hash = hash
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

func (ln *leafNode) commit(force bool, level byte, originDb data.DBWriteCacher, targetDb data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return err
	}

	shouldNotCommit := !ln.dirty && !force
	if shouldNotCommit {
		return nil
	}

	ln.dirty = false
	return encodeNodeAndCommitToDB(ln, targetDb, marshalizer, hasher)
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

func (ln *leafNode) resolveCollapsed(pos byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(pos int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, nil
	}

	return nil, nil
}

func (ln *leafNode) getNext(key []byte, dbw data.DBWriteCacher, marshalizer marshal.Marshalizer) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}

func (ln *leafNode) insert(n *leafNode, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, [][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	oldHash := make([][]byte, 0)
	if !ln.dirty {
		oldHash = append(oldHash, ln.hash)
	}

	if bytes.Equal(n.Key, ln.Key) {
		ln.Value = n.Value
		ln.dirty = true
		ln.hash = nil
		return true, ln, oldHash, nil
	}

	keyMatchLen := prefixLen(n.Key, ln.Key)
	branch := newBranchNode()
	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, [][]byte{}, ErrChildPosOutOfRange
	}

	branch.children[oldChildPos] = newLeafNode(ln.Key[keyMatchLen+1:], ln.Value)
	branch.children[newChildPos] = newLeafNode(n.Key[keyMatchLen+1:], n.Value)

	if keyMatchLen == 0 {
		return true, branch, oldHash, nil
	}
	return true, newExtensionNode(ln.Key[:keyMatchLen], branch), oldHash, nil
}

func (ln *leafNode) delete(key []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, [][]byte, error) {
	keyMatchLen := prefixLen(key, ln.Key)
	if keyMatchLen == len(key) {
		oldHash := make([][]byte, 0)
		if !ln.dirty {
			oldHash = append(oldHash, ln.hash)
		}

		return true, nil, oldHash, nil
	}
	return false, ln, [][]byte{}, nil
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

func (ln *leafNode) print(writer io.Writer, index int) {
	if ln == nil {
		return
	}

	key := ""
	for _, k := range ln.Key {
		key += fmt.Sprintf("%d", k)
	}

	val := ""
	for _, v := range ln.Value {
		val += fmt.Sprintf("%d", v)
	}

	_, _ = fmt.Fprintf(writer, "L:(%s - %s)\n", key, val)
}

func (ln *leafNode) deepClone() node {
	if ln == nil {
		return nil
	}

	clonedNode := &leafNode{}

	if ln.Key != nil {
		clonedNode.Key = make([]byte, len(ln.Key))
		copy(clonedNode.Key, ln.Key)
	}

	if ln.Value != nil {
		clonedNode.Value = make([]byte, len(ln.Value))
		copy(clonedNode.Value, ln.Value)
	}

	if ln.hash != nil {
		clonedNode.hash = make([]byte, len(ln.hash))
		copy(clonedNode.hash, ln.hash)
	}

	clonedNode.dirty = ln.dirty

	return clonedNode
}

func (ln *leafNode) getDirtyHashes() ([][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make([][]byte, 0)

	if !ln.isDirty() {
		return dirtyHashes, nil
	}

	dirtyHashes = append(dirtyHashes, ln.getHash())
	return dirtyHashes, nil
}
