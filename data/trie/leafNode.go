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

func newLeafNode(key, value []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*leafNode, error) {
	if db == nil || db.IsInterfaceNil() {
		return nil, ErrNilDatabase
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, ErrNilHasher
	}

	return &leafNode{
		CollapsedLn: protobuf.CollapsedLn{
			Key:   key,
			Value: value,
		},
		baseNode: &baseNode{
			dirty:  true,
			db:     db,
			marsh:  marshalizer,
			hasher: hasher,
		},
	}, nil
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

func (ln *leafNode) getMarshalizer() marshal.Marshalizer {
	return ln.marsh
}

func (ln *leafNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	ln.marsh = marshalizer
}

func (ln *leafNode) getHasher() hashing.Hasher {
	return ln.hasher
}

func (ln *leafNode) setHasher(hasher hashing.Hasher) {
	ln.hasher = hasher
}

func (ln *leafNode) getDb() data.DBWriteCacher {
	return ln.db
}

func (ln *leafNode) setDb(db data.DBWriteCacher) {
	ln.db = db
}

func (ln *leafNode) getCollapsed() (node, error) {
	return ln, nil
}

func (ln *leafNode) setHash() error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return err
	}
	if ln.getHash() != nil {
		return nil
	}
	hash, err := hashChildrenAndNode(ln)
	if err != nil {
		return err
	}
	ln.hash = hash
	return nil
}

func (ln *leafNode) setHashConcurrent(wg *sync.WaitGroup, c chan error) {
	err := ln.setHash()
	if err != nil {
		c <- err
	}
	wg.Done()
}

func (ln *leafNode) setRootHash() error {
	return ln.setHash()
}

func (ln *leafNode) hashChildren() error {
	return nil
}

func (ln *leafNode) hashNode() ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	return encodeNodeAndGetHash(ln)
}

func (ln *leafNode) commit(force bool, level byte, targetDb data.DBWriteCacher) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return err
	}

	shouldNotCommit := !ln.dirty && !force
	if shouldNotCommit {
		return nil
	}

	ln.dirty = false
	return encodeNodeAndCommitToDB(ln, targetDb)
}

func (ln *leafNode) getEncodedNode() ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := ln.marsh.Marshal(ln)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, leaf)
	return marshaledNode, nil
}

func (ln *leafNode) resolveCollapsed(pos byte) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(pos int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte) (value []byte, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, nil
	}

	return nil, nil
}

func (ln *leafNode) getNext(key []byte) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}

func (ln *leafNode) insert(n *leafNode) (bool, node, [][]byte, error) {
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
	branch, err := newBranchNode(ln.db, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, [][]byte{}, ErrChildPosOutOfRange
	}

	newLnOldChildPos, err := newLeafNode(ln.Key[keyMatchLen+1:], ln.Value, ln.db, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	branch.children[oldChildPos] = newLnOldChildPos

	newLnNewChildPos, err := newLeafNode(n.Key[keyMatchLen+1:], n.Value, ln.db, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	branch.children[newChildPos] = newLnNewChildPos

	if keyMatchLen == 0 {
		return true, branch, oldHash, nil
	}

	newEn, err := newExtensionNode(ln.Key[:keyMatchLen], branch, ln.db, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	return true, newEn, oldHash, nil
}

func (ln *leafNode) delete(key []byte) (bool, node, [][]byte, error) {
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

func (ln *leafNode) reduceNode(pos int) (node, error) {
	k := append([]byte{byte(pos)}, ln.Key...)

	newLn, err := newLeafNode(k, ln.Value, ln.db, ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}

	return newLn, nil
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

	clonedNode := &leafNode{baseNode: &baseNode{}}

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
	clonedNode.db = ln.db
	clonedNode.marsh = ln.marsh
	clonedNode.hasher = ln.hasher

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

func (ln *leafNode) getChildren() ([]node, error) {
	return nil, nil
}

func (ln *leafNode) isValid() bool {
	if len(ln.Value) == 0 {
		return false
	}
	return true
}

func (ln *leafNode) setDirty(dirty bool) {
	ln.dirty = dirty
}

func (ln *leafNode) loadChildren(syncer *trieSyncer) error {
	syncer.interceptedNodes.Remove(ln.hash)
	return nil
}
