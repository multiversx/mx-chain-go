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

func newExtensionNode(key []byte, child node, db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*extensionNode, error) {
	if db == nil || db.IsInterfaceNil() {
		return nil, ErrNilDatabase
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, ErrNilHasher
	}

	return &extensionNode{
		CollapsedEn: protobuf.CollapsedEn{
			Key:          key,
			EncodedChild: nil,
		},
		child:  child,
		dirty:  true,
		db:     db,
		marsh:  marshalizer,
		hasher: hasher,
	}, nil
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

func (en *extensionNode) getMarshalizer() marshal.Marshalizer {
	return en.marsh
}

func (en *extensionNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	en.marsh = marshalizer
}

func (en *extensionNode) getHasher() hashing.Hasher {
	return en.hasher
}

func (en *extensionNode) setHasher(hasher hashing.Hasher) {
	en.hasher = hasher
}

func (en *extensionNode) getDb() data.DBWriteCacher {
	return en.db
}

func (en *extensionNode) setDb(db data.DBWriteCacher) {
	en.db = db
}

func (en *extensionNode) getCollapsed() (node, error) {
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
		err := en.child.setHash()
		if err != nil {
			return nil, err
		}
	}
	collapsed.EncodedChild = en.child.getHash()
	collapsed.child = nil
	return collapsed, nil
}

func (en *extensionNode) setHash() error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.getHash() != nil {
		return nil
	}
	if en.isCollapsed() {
		hash, err := encodeNodeAndGetHash(en)
		if err != nil {
			return err
		}
		en.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(en)
	if err != nil {
		return err
	}
	en.hash = hash
	return nil
}

func (en *extensionNode) setHashConcurrent(wg *sync.WaitGroup, c chan error) {
	err := en.setHash()
	if err != nil {
		c <- err
	}
	wg.Done()
}
func (en *extensionNode) setRootHash() error {
	return en.setHash()
}

func (en *extensionNode) hashChildren() error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.child != nil {
		err = en.child.setHash()
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *extensionNode) hashNode() ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if en.child != nil {
		encChild, err := encodeNodeAndGetHash(en.child)
		if err != nil {
			return nil, err
		}
		en.EncodedChild = encChild
	}
	return encodeNodeAndGetHash(en)
}

func (en *extensionNode) commit(force bool, level byte) error {
	level++
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}

	shouldNotCommit := !en.dirty && !force
	if shouldNotCommit {
		return nil
	}

	if en.child != nil {
		err = en.child.commit(force, level)
		if err != nil {
			return err
		}
	}

	en.dirty = false
	err = encodeNodeAndCommitToDB(en)
	if err != nil {
		return err
	}
	if level == maxTrieLevelAfterCommit {
		collapsed, err := en.getCollapsed()
		if err != nil {
			return err
		}
		if n, ok := collapsed.(*extensionNode); ok {
			*en = *n
		}
	}
	return nil
}

func (en *extensionNode) getEncodedNode() ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := en.marsh.Marshal(en)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, extension)
	return marshaledNode, nil
}

func (en *extensionNode) resolveCollapsed(pos byte) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	child, err := getNodeFromDBAndDecode(en.EncodedChild, en.db, en.marsh, en.hasher)
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

func (en *extensionNode) tryGet(key []byte) (value []byte, err error) {
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
	err = resolveIfCollapsed(en, 0)
	if err != nil {
		return nil, err
	}

	return en.child.tryGet(key)
}

func (en *extensionNode) getNext(key []byte) (node, []byte, error) {
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
	err = resolveIfCollapsed(en, 0)
	if err != nil {
		return nil, nil, err
	}

	key = key[len(en.Key):]
	return en.child, key, nil
}

func (en *extensionNode) insert(n *leafNode) (bool, node, [][]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	err = resolveIfCollapsed(en, 0)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	keyMatchLen := prefixLen(n.Key, en.Key)

	// If the whole key matches, keep this extension node as is
	// and only update the value.
	if keyMatchLen == len(en.Key) {
		n.Key = n.Key[keyMatchLen:]
		dirty, newNode, oldHashes, err := en.child.insert(n)
		if !dirty || err != nil {
			return false, nil, [][]byte{}, err
		}

		if !en.dirty {
			oldHashes = append(oldHashes, en.hash)
		}

		newEn, err := newExtensionNode(en.Key, newNode, en.db, en.marsh, en.hasher)
		if err != nil {
			return false, nil, [][]byte{}, err
		}

		return true, newEn, oldHashes, nil
	}

	oldHash := make([][]byte, 0)
	if !en.dirty {
		oldHash = append(oldHash, en.hash)
	}

	// Otherwise branch out at the index where they differ.
	branch, err := newBranchNode(en.db, en.marsh, en.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	oldChildPos := en.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, [][]byte{}, ErrChildPosOutOfRange
	}

	followingExtensionNode, err := newExtensionNode(en.Key[keyMatchLen+1:], en.child, en.db, en.marsh, en.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

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

	newEn, err := newExtensionNode(en.Key[:keyMatchLen], branch, en.db, en.marsh, en.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	return true, newEn, oldHash, nil
}

func (en *extensionNode) delete(key []byte) (bool, node, [][]byte, error) {
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
	err = resolveIfCollapsed(en, 0)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	dirty, newNode, oldHashes, err := en.child.delete(key[len(en.Key):])
	if !dirty || err != nil {
		return false, en, [][]byte{}, err
	}

	if !en.dirty {
		oldHashes = append(oldHashes, en.hash)
	}

	switch newNode := newNode.(type) {
	case *leafNode:
		newLn, err := newLeafNode(concat(en.Key, newNode.Key...), newNode.Value, en.db, en.marsh, en.hasher)
		if err != nil {
			return false, nil, [][]byte{}, err
		}

		return true, newLn, oldHashes, nil
	case *extensionNode:
		newEn, err := newExtensionNode(concat(en.Key, newNode.Key...), newNode.child, en.db, en.marsh, en.hasher)
		if err != nil {
			return false, nil, [][]byte{}, err
		}

		return true, newEn, oldHashes, nil
	default:
		newEn, err := newExtensionNode(en.Key, newNode, en.db, en.marsh, en.hasher)
		if err != nil {
			return false, nil, [][]byte{}, err
		}

		return true, newEn, oldHashes, nil
	}
}

func (en *extensionNode) reduceNode(pos int) (node, error) {
	k := append([]byte{byte(pos)}, en.Key...)

	newEn, err := newExtensionNode(k, en.child, en.db, en.marsh, en.hasher)
	if err != nil {
		return nil, err
	}

	return newEn, nil
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

	clonedNode.db = en.db
	clonedNode.marsh = en.marsh
	clonedNode.hasher = en.hasher

	return clonedNode
}

func (en *extensionNode) getDirtyHashes() ([][]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make([][]byte, 0)

	if !en.isDirty() {
		return dirtyHashes, nil
	}

	hashes, err := en.child.getDirtyHashes()
	if err != nil {
		return nil, err
	}

	dirtyHashes = append(dirtyHashes, hashes...)
	dirtyHashes = append(dirtyHashes, en.hash)
	return dirtyHashes, nil
}
