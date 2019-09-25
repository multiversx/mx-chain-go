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

// Save saves the serialized data of an extension node into a stream through Capnp protocol
func (en *extensionNode) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	extensionNodeGoToCapn(seg, en)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into an extension node object through Capnp protocol
func (en *extensionNode) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootExtensionNodeCapn(capMsg)
	extensionNodeCapnToGo(z, en)
	return nil
}

func extensionNodeGoToCapn(seg *capn.Segment, src *extensionNode) capnp.ExtensionNodeCapn {
	dest := capnp.AutoNewExtensionNodeCapn(seg)

	dest.SetKey(src.Key)
	dest.SetEncodedChild(src.EncodedChild)

	return dest
}

func extensionNodeCapnToGo(src capnp.ExtensionNodeCapn, dest *extensionNode) *extensionNode {
	if dest == nil {
		dest = &extensionNode{}
	}

	dest.EncodedChild = src.EncodedChild()
	dest.Key = src.Key()

	return dest
}

func newExtensionNode(key []byte, child node) *extensionNode {
	return &extensionNode{
		CollapsedEn: protobuf.CollapsedEn{
			Key:          key,
			EncodedChild: nil,
		},
		child: child,
		hash:  nil,
		dirty: true,
	}
}

func (en *extensionNode) getHash() []byte {
	return en.hash
}

func (en *extensionNode) setGivenHash(hash []byte) {
	en.hash = hash
}

func (en *extensionNode) isDirty() bool {
	return en.dirty
}

func (en *extensionNode) getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if en.isCollapsed() {
		return en, nil
	}
	collapsed := en.clone()
	ok, err := hasValidHash(en.child)
	if err != nil {
		return nil, err
	}
	if !ok {
		err := en.child.setHash(marshalizer, hasher)
		if err != nil {
			return nil, err
		}
	}
	collapsed.EncodedChild = en.child.getHash()
	collapsed.child = nil
	return collapsed, nil
}

func (en *extensionNode) setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.getHash() != nil {
		return nil
	}
	if en.isCollapsed() {
		hash, err := encodeNodeAndGetHash(en, marshalizer, hasher)
		if err != nil {
			return err
		}
		en.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(en, marshalizer, hasher)
	if err != nil {
		return err
	}
	en.hash = hash
	return nil
}

func (en *extensionNode) setHashConcurrent(marshalizer marshal.Marshalizer, hasher hashing.Hasher, wg *sync.WaitGroup, c chan error) {
	err := en.setHash(marshalizer, hasher)
	if err != nil {
		c <- err
	}
	wg.Done()
}
func (en *extensionNode) setRootHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	return en.setHash(marshalizer, hasher)
}

func (en *extensionNode) hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.child != nil {
		err = en.child.setHash(marshalizer, hasher)
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *extensionNode) hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if en.child != nil {
		encChild, err := encodeNodeAndGetHash(en.child, marshalizer, hasher)
		if err != nil {
			return nil, err
		}
		en.EncodedChild = encChild
	}
	return encodeNodeAndGetHash(en, marshalizer, hasher)
}

func (en *extensionNode) commit(level byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	level++
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if !en.dirty {
		return nil
	}
	if en.child != nil {
		err = en.child.commit(level, db, marshalizer, hasher)
		if err != nil {
			return err
		}
	}

	en.dirty = false
	err = encodeNodeAndCommitToDB(en, db, marshalizer, hasher)
	if err != nil {
		return err
	}
	if level == maxTrieLevelAfterCommit {
		collapsed, err := en.getCollapsed(marshalizer, hasher)
		if err != nil {
			return err
		}
		if n, ok := collapsed.(*extensionNode); ok {
			*en = *n
		}
	}
	return nil
}

func (en *extensionNode) getEncodedNode(marshalizer marshal.Marshalizer) ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := marshalizer.Marshal(en)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, extension)
	return marshaledNode, nil
}

func (en *extensionNode) resolveCollapsed(pos byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	child, err := getNodeFromDBAndDecode(en.EncodedChild, db, marshalizer)
	if err != nil {
		return err
	}
	child.setGivenHash(en.EncodedChild)
	en.child = child
	return nil
}

func (en *extensionNode) isCollapsed() bool {
	return en.child == nil && len(en.EncodedChild) != 0
}

func (en *extensionNode) isPosCollapsed(pos int) bool {
	return en.isCollapsed()
}

func (en *extensionNode) tryGet(key []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, nil
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, nil
	}
	key = key[len(en.Key):]
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return nil, err
	}

	return en.child.tryGet(key, db, marshalizer)
}

func (en *extensionNode) getNext(key []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (node, []byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, nil, err
	}
	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, nil, ErrNodeNotFound
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, nil, ErrNodeNotFound
	}
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return nil, nil, err
	}

	key = key[len(en.Key):]
	return en.child, key, nil
}

func (en *extensionNode) insert(n *leafNode, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, [][]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	keyMatchLen := prefixLen(n.Key, en.Key)

	// If the whole key matches, keep this extension node as is
	// and only update the value.
	if keyMatchLen == len(en.Key) {
		n.Key = n.Key[keyMatchLen:]
		dirty, newNode, oldHashes, err := en.child.insert(n, db, marshalizer)
		if !dirty || err != nil {
			return false, nil, [][]byte{}, err
		}

		if !en.dirty {
			oldHashes = append(oldHashes, en.hash)
		}

		return true, newExtensionNode(en.Key, newNode), oldHashes, nil
	}

	oldHash := make([][]byte, 0)
	if !en.dirty {
		oldHash = append(oldHash, en.hash)
	}

	// Otherwise branch out at the index where they differ.
	branch := newBranchNode()
	oldChildPos := en.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, [][]byte{}, ErrChildPosOutOfRange
	}

	followingExtensionNode := newExtensionNode(en.Key[keyMatchLen+1:], en.child)
	if len(followingExtensionNode.Key) < 1 {
		branch.children[oldChildPos] = en.child
	} else {
		branch.children[oldChildPos] = followingExtensionNode
	}
	n.Key = n.Key[keyMatchLen+1:]
	branch.children[newChildPos] = n

	if keyMatchLen == 0 {
		return true, branch, oldHash, nil
	}

	return true, newExtensionNode(en.Key[:keyMatchLen], branch), oldHash, nil
}

func (en *extensionNode) delete(key []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, [][]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	if len(key) == 0 {
		return false, nil, [][]byte{}, ErrValueTooShort
	}
	keyMatchLen := prefixLen(key, en.Key)
	if keyMatchLen < len(en.Key) {
		return false, en, [][]byte{}, nil
	}
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	dirty, newNode, oldHashes, err := en.child.delete(key[len(en.Key):], db, marshalizer)
	if !dirty || err != nil {
		return false, en, [][]byte{}, err
	}

	if !en.dirty {
		oldHashes = append(oldHashes, en.hash)
	}

	switch newNode := newNode.(type) {
	case *leafNode:
		return true, newLeafNode(concat(en.Key, newNode.Key...), newNode.Value), oldHashes, nil
	case *extensionNode:
		return true, newExtensionNode(concat(en.Key, newNode.Key...), newNode.child), oldHashes, nil
	default:
		return true, newExtensionNode(en.Key, newNode), oldHashes, nil
	}
}

func (en *extensionNode) reduceNode(pos int) node {
	k := append([]byte{byte(pos)}, en.Key...)
	return newExtensionNode(k, en.child)
}

func (en *extensionNode) clone() *extensionNode {
	nodeClone := *en
	return &nodeClone
}

func (en *extensionNode) isEmptyOrNil() error {
	if en == nil {
		return ErrNilNode
	}
	if en.child == nil && len(en.EncodedChild) == 0 {
		return ErrEmptyNode
	}
	return nil
}

func (en *extensionNode) print(writer io.Writer, index int) {
	if en == nil {
		return
	}

	key := ""
	for _, k := range en.Key {
		key += fmt.Sprintf("%d", k)
	}

	str := fmt.Sprintf("E:(%s) - ", key)
	_, _ = fmt.Fprint(writer, str)

	if en.child == nil {
		return
	}
	en.child.print(writer, index+len(str))
}

func (en *extensionNode) deepClone() node {
	if en == nil {
		return nil
	}

	clonedNode := &extensionNode{}

	if en.Key != nil {
		clonedNode.Key = make([]byte, len(en.Key))
		copy(clonedNode.Key, en.Key)
	}

	if en.EncodedChild != nil {
		clonedNode.EncodedChild = make([]byte, len(en.EncodedChild))
		copy(clonedNode.EncodedChild, en.EncodedChild)
	}

	if en.hash != nil {
		clonedNode.hash = make([]byte, len(en.hash))
		copy(clonedNode.hash, en.hash)
	}

	clonedNode.dirty = en.dirty

	if en.child != nil {
		clonedNode.child = en.child.deepClone()
	}

	return clonedNode
}
