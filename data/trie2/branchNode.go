package trie2

import (
	"bytes"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/trie2/capnp"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie2/proto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	capn "github.com/glycerine/go-capnproto"
)

// Save saves the serialized data of a branch node into a stream through Capnp protocol
func (bn *branchNode) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	branchNodeGoToCapn(seg, bn)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a branch node object through Capnp protocol
func (bn *branchNode) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootBranchNodeCapn(capMsg)
	branchNodeCapnToGo(z, bn)
	return nil
}

func branchNodeGoToCapn(seg *capn.Segment, src *branchNode) capnp.BranchNodeCapn {
	dest := capnp.AutoNewBranchNodeCapn(seg)

	children := seg.NewDataList(len(src.EncodedChildren))
	for i := range src.EncodedChildren {
		children.Set(i, src.EncodedChildren[i])
	}
	dest.SetEncodedChildren(children)

	return dest
}

func branchNodeCapnToGo(src capnp.BranchNodeCapn, dest *branchNode) *branchNode {
	if dest == nil {
		dest = newBranchNode()
	}

	for i := 0; i < nrOfChildren; i++ {
		child := src.EncodedChildren().At(i)
		if bytes.Equal(child, []byte{}) {
			dest.EncodedChildren[i] = nil
		} else {
			dest.EncodedChildren[i] = child
		}

	}
	return dest
}

func newBranchNode() *branchNode {
	var children [nrOfChildren]node
	EncChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: protobuf.CollapsedBn{
			EncodedChildren: EncChildren,
		},
		children: children,
		hash:     nil,
		dirty:    true,
	}
}

func (bn *branchNode) getHash() []byte {
	return bn.hash
}

func (bn *branchNode) isDirty() bool {
	return bn.dirty
}

func (bn *branchNode) getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if bn.isCollapsed() {
		return bn, nil
	}
	collapsed := bn.clone()
	for i := range bn.children {
		if bn.children[i] != nil {
			ok, err := hasValidHash(bn.children[i])
			if err != nil {
				return nil, err
			}
			if !ok {
				err := bn.children[i].setHash(marshalizer, hasher)
				if err != nil {
					return nil, err
				}
			}
			collapsed.EncodedChildren[i] = bn.children[i].getHash()
			collapsed.children[i] = nil
		}
	}
	return collapsed, nil
}

func (bn *branchNode) setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if bn.getHash() != nil {
		return nil
	}
	if bn.isCollapsed() {
		hash, err := encodeNodeAndGetHash(bn, marshalizer, hasher)
		if err != nil {
			return err
		}
		bn.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(bn, marshalizer, hasher)
	if err != nil {
		return err
	}
	bn.hash = hash
	return nil
}

func (bn *branchNode) setRootHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if bn.getHash() != nil {
		return nil
	}
	if bn.isCollapsed() {
		hash, err := encodeNodeAndGetHash(bn, marshalizer, hasher)
		if err != nil {
			return err
		}
		bn.hash = hash
		return nil
	}

	var wg sync.WaitGroup
	errc := make(chan error, nrOfChildren)

	for i := 0; i < nrOfChildren; i++ {
		if bn.children[i] != nil {
			wg.Add(1)
			go bn.children[i].setHashConcurrent(marshalizer, hasher, &wg, errc)
		}
	}
	wg.Wait()
	if len(errc) != 0 {
		for err := range errc {
			return err
		}
	}

	hashed, err := bn.hashNode(marshalizer, hasher)
	if err != nil {
		return err
	}

	bn.hash = hashed
	return nil
}

func (bn *branchNode) setHashConcurrent(marshalizer marshal.Marshalizer, hasher hashing.Hasher, wg *sync.WaitGroup, c chan error) {
	defer wg.Done()
	err := bn.isEmptyOrNil()
	if err != nil {
		c <- err
		return
	}
	if bn.getHash() != nil {
		return
	}
	if bn.isCollapsed() {
		hash, err := encodeNodeAndGetHash(bn, marshalizer, hasher)
		if err != nil {
			c <- err
			return
		}
		bn.hash = hash
		return
	}
	hash, err := hashChildrenAndNode(bn, marshalizer, hasher)
	if err != nil {
		c <- err
		return
	}
	bn.hash = hash
	return
}

func (bn *branchNode) hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	for i := 0; i < nrOfChildren; i++ {
		if bn.children[i] != nil {
			err := bn.children[i].setHash(marshalizer, hasher)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bn *branchNode) hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	for i := range bn.EncodedChildren {
		if bn.children[i] != nil {
			encChild, err := encodeNodeAndGetHash(bn.children[i], marshalizer, hasher)
			if err != nil {
				return nil, err
			}
			bn.EncodedChildren[i] = encChild
		}
	}
	return encodeNodeAndGetHash(bn, marshalizer, hasher)
}

func (bn *branchNode) commit(level byte, db DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	level++
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if !bn.dirty {
		return nil
	}
	for i := range bn.children {
		if bn.children[i] != nil {
			err := bn.children[i].commit(level, db, marshalizer, hasher)
			if err != nil {
				return err
			}
		}
	}
	bn.dirty = false
	err = encodeNodeAndCommitToDB(bn, db, marshalizer, hasher)
	if err != nil {
		return err
	}
	if level == maxTrieLevelAfterCommit {
		collapsed, err := bn.getCollapsed(marshalizer, hasher)
		if err != nil {
			return err
		}
		if n, ok := collapsed.(*branchNode); ok {
			*bn = *n
		}
	}
	return nil
}

func (bn *branchNode) getEncodedNode(marshalizer marshal.Marshalizer) ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := marshalizer.Marshal(bn)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, branch)
	return marshaledNode, nil
}

func (bn *branchNode) resolveCollapsed(pos byte, db DBWriteCacher, marshalizer marshal.Marshalizer) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if childPosOutOfRange(pos) {
		return ErrChildPosOutOfRange
	}
	if len(bn.EncodedChildren[pos]) != 0 {
		child, err := getNodeFromDBAndDecode(bn.EncodedChildren[pos], db, marshalizer)
		if err != nil {
			return err
		}
		bn.children[pos] = child
	}
	return nil
}

func (bn *branchNode) isCollapsed() bool {
	for i := range bn.children {
		if bn.children[i] != nil {
			return false
		}
	}
	return true
}

func (bn *branchNode) isPosCollapsed(pos int) bool {
	return bn.children[pos] == nil && len(bn.EncodedChildren[pos]) != 0
}

func (bn *branchNode) tryGet(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return nil, err
	}
	if bn.children[childPos] == nil {
		return nil, ErrNodeNotFound
	}
	value, err = bn.children[childPos].tryGet(key, db, marshalizer)
	return value, err
}

func (bn *branchNode) getNext(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (node, []byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, nil, err
	}
	if len(key) == 0 {
		return nil, nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return nil, nil, err
	}

	if bn.children[childPos] == nil {
		return nil, nil, ErrNodeNotFound
	}
	return bn.children[childPos], key, nil
}

func (bn *branchNode) insert(n *leafNode, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if len(n.Key) == 0 {
		return false, nil, ErrValueTooShort
	}
	childPos := n.Key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, ErrChildPosOutOfRange
	}
	n.Key = n.Key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return false, nil, err
	}

	if bn.children[childPos] != nil {
		dirty, newNode, err := bn.children[childPos].insert(n, db, marshalizer)
		if !dirty || err != nil {
			return false, bn, err
		}
		bn.children[childPos] = newNode
		bn.dirty = dirty
		if dirty {
			bn.hash = nil
		}
		return true, bn, nil
	}
	bn.children[childPos] = newLeafNode(n.Key, n.Value)
	bn.dirty = true
	bn.hash = nil
	return true, bn, nil
}

func (bn *branchNode) delete(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if len(key) == 0 {
		return false, nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return false, nil, err
	}

	dirty, newNode, err := bn.children[childPos].delete(key, db, marshalizer)
	if !dirty || err != nil {
		return false, nil, err
	}

	bn.hash = nil
	bn.children[childPos] = newNode
	if newNode == nil {
		bn.EncodedChildren[childPos] = nil
	}

	nrOfChildren, pos := getChildPosition(bn)

	if nrOfChildren == 1 {
		err = bn.resolveCollapsed(byte(pos), db, marshalizer)
		if err != nil {
			return false, nil, err
		}
		if childPos != 16 {
			newNode := bn.reduceNode(pos)
			return true, newNode, nil
		}
		child := bn.children[pos]
		if child, ok := child.(*leafNode); ok {
			return true, newLeafNode([]byte{byte(pos)}, child.Value), nil
		}

	}

	bn.dirty = dirty
	return true, bn, nil
}

func (bn *branchNode) reduceNode(pos int) node {
	if child, ok := bn.children[pos].(*leafNode); ok {
		return newLeafNode(concat([]byte{byte(pos)}, child.Key...), child.Value)
	}
	return newExtensionNode([]byte{byte(pos)}, bn.children[pos])
}

func getChildPosition(n *branchNode) (nrOfChildren int, childPos int) {
	for i := range n.children {
		if n.children[i] != nil || len(n.EncodedChildren[i]) != 0 {
			nrOfChildren++
			childPos = i
		}
	}
	return
}

func (bn *branchNode) clone() *branchNode {
	nodeClone := *bn
	return &nodeClone
}

func (bn *branchNode) isEmptyOrNil() error {
	if bn == nil {
		return ErrNilNode
	}
	for i := range bn.children {
		if bn.children[i] != nil || len(bn.EncodedChildren[i]) != 0 {
			return nil
		}
	}
	return ErrEmptyNode
}
