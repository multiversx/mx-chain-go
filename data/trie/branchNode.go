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
		dest = emptyDirtyBranchNode()
	}
	dest.dirty = false

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

func newBranchNode(db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*branchNode, error) {
	if db == nil || db.IsInterfaceNil() {
		return nil, ErrNilDatabase
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, ErrNilHasher
	}

	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: protobuf.CollapsedBn{
			EncodedChildren: encChildren,
		},
		children: children,
		baseNode: &baseNode{
			dirty:  true,
			db:     db,
			marsh:  marshalizer,
			hasher: hasher,
		},
	}, nil
}

func emptyDirtyBranchNode() *branchNode {
	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: protobuf.CollapsedBn{
			EncodedChildren: encChildren,
		},
		children: children,
		baseNode: &baseNode{
			dirty: true,
		},
	}
}

func (bn *branchNode) getHash() []byte {
	return bn.hash
}

func (bn *branchNode) setGivenHash(hash []byte) {
	bn.hash = hash
}

func (bn *branchNode) isDirty() bool {
	return bn.dirty
}

func (bn *branchNode) getMarshalizer() marshal.Marshalizer {
	return bn.marsh
}

func (bn *branchNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	bn.marsh = marshalizer
}

func (bn *branchNode) getHasher() hashing.Hasher {
	return bn.hasher
}

func (bn *branchNode) setHasher(hasher hashing.Hasher) {
	bn.hasher = hasher
}

func (bn *branchNode) getDb() data.DBWriteCacher {
	return bn.db
}

func (bn *branchNode) setDb(db data.DBWriteCacher) {
	bn.db = db
}

func (bn *branchNode) getCollapsed() (node, error) {
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
			var ok bool
			ok, err = hasValidHash(bn.children[i])
			if err != nil {
				return nil, err
			}
			if !ok {
				err = bn.children[i].setHash()
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

func (bn *branchNode) setHash() error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if bn.getHash() != nil {
		return nil
	}
	if bn.isCollapsed() {
		var hash []byte
		hash, err = encodeNodeAndGetHash(bn)
		if err != nil {
			return err
		}
		bn.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(bn)
	if err != nil {
		return err
	}
	bn.hash = hash
	return nil
}

func (bn *branchNode) setRootHash() error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if bn.getHash() != nil {
		return nil
	}
	if bn.isCollapsed() {
		var hash []byte
		hash, err = encodeNodeAndGetHash(bn)
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
			go bn.children[i].setHashConcurrent(&wg, errc)
		}
	}
	wg.Wait()
	if len(errc) != 0 {
		for err = range errc {
			return err
		}
	}

	hashed, err := bn.hashNode()
	if err != nil {
		return err
	}

	bn.hash = hashed
	return nil
}

func (bn *branchNode) setHashConcurrent(wg *sync.WaitGroup, c chan error) {
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
		var hash []byte
		hash, err = encodeNodeAndGetHash(bn)
		if err != nil {
			c <- err
			return
		}
		bn.hash = hash
		return
	}
	hash, err := hashChildrenAndNode(bn)
	if err != nil {
		c <- err
		return
	}
	bn.hash = hash
	return
}

func (bn *branchNode) hashChildren() error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	for i := 0; i < nrOfChildren; i++ {
		if bn.children[i] != nil {
			err = bn.children[i].setHash()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bn *branchNode) hashNode() ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	for i := range bn.EncodedChildren {
		if bn.children[i] != nil {
			var encChild []byte
			encChild, err = encodeNodeAndGetHash(bn.children[i])
			if err != nil {
				return nil, err
			}
			bn.EncodedChildren[i] = encChild
		}
	}
	return encodeNodeAndGetHash(bn)
}

func (bn *branchNode) commit(force bool, level byte, targetDb data.DBWriteCacher) error {
	level++
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}

	shouldNotCommit := !bn.dirty && !force
	if shouldNotCommit {
		return nil
	}

	for i := range bn.children {
		if force {
			err = resolveIfCollapsed(bn, byte(i))
			if err != nil {
				return err
			}
		}

		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].commit(force, level, targetDb)
		if err != nil {
			return err
		}
	}
	bn.dirty = false
	err = encodeNodeAndCommitToDB(bn, targetDb)
	if err != nil {
		return err
	}
	if level == maxTrieLevelAfterCommit {
		var collapsed node
		collapsed, err = bn.getCollapsed()
		if err != nil {
			return err
		}
		if n, ok := collapsed.(*branchNode); ok {
			*bn = *n
		}
	}
	return nil
}

func (bn *branchNode) getEncodedNode() ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := bn.marsh.Marshal(bn)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, branch)
	return marshaledNode, nil
}

func (bn *branchNode) resolveCollapsed(pos byte) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if childPosOutOfRange(pos) {
		return ErrChildPosOutOfRange
	}
	if len(bn.EncodedChildren[pos]) != 0 {
		var child node
		child, err = getNodeFromDBAndDecode(bn.EncodedChildren[pos], bn.db, bn.marsh, bn.hasher)
		if err != nil {
			return err
		}
		child.setGivenHash(bn.EncodedChildren[pos])
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

func (bn *branchNode) tryGet(key []byte) (value []byte, err error) {
	err = bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, nil
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos)
	if err != nil {
		return nil, err
	}
	if bn.children[childPos] == nil {
		return nil, nil
	}

	return bn.children[childPos].tryGet(key)
}

func (bn *branchNode) getNext(key []byte) (node, []byte, error) {
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
	err = resolveIfCollapsed(bn, childPos)
	if err != nil {
		return nil, nil, err
	}

	if bn.children[childPos] == nil {
		return nil, nil, ErrNodeNotFound
	}
	return bn.children[childPos], key, nil
}

func (bn *branchNode) insert(n *leafNode) (bool, node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, emptyHashes, err
	}
	if len(n.Key) == 0 {
		return false, nil, emptyHashes, ErrValueTooShort
	}
	childPos := n.Key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, emptyHashes, ErrChildPosOutOfRange
	}
	n.Key = n.Key[1:]
	err = resolveIfCollapsed(bn, childPos)
	if err != nil {
		return false, nil, emptyHashes, err
	}

	if bn.children[childPos] != nil {
		var dirty bool
		var newNode node
		var oldHashes [][]byte

		dirty, newNode, oldHashes, err = bn.children[childPos].insert(n)
		if !dirty || err != nil {
			return false, bn, emptyHashes, err
		}

		if !bn.dirty {
			oldHashes = append(oldHashes, bn.hash)
		}

		bn.children[childPos] = newNode
		bn.dirty = dirty
		if dirty {
			bn.hash = nil
		}
		return true, bn, oldHashes, nil
	}

	newLn, err := newLeafNode(n.Key, n.Value, bn.db, bn.marsh, bn.hasher)
	if err != nil {
		return false, nil, emptyHashes, err
	}
	bn.children[childPos] = newLn

	oldHash := make([][]byte, 0)
	if !bn.dirty {
		oldHash = append(oldHash, bn.hash)
	}

	bn.dirty = true
	bn.hash = nil
	return true, bn, oldHash, nil
}

func (bn *branchNode) delete(key []byte) (bool, node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, emptyHashes, err
	}
	if len(key) == 0 {
		return false, nil, emptyHashes, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, emptyHashes, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos)
	if err != nil {
		return false, nil, emptyHashes, err
	}

	dirty, newNode, oldHashes, err := bn.children[childPos].delete(key)
	if !dirty || err != nil {
		return false, nil, emptyHashes, err
	}

	if !bn.dirty {
		oldHashes = append(oldHashes, bn.hash)
	}

	bn.hash = nil
	bn.children[childPos] = newNode
	if newNode == nil {
		bn.EncodedChildren[childPos] = nil
	}

	numChildren, pos := getChildPosition(bn)

	if numChildren == 1 {
		err = resolveIfCollapsed(bn, byte(pos))
		if err != nil {
			return false, nil, emptyHashes, err
		}

		newNode, err = bn.children[pos].reduceNode(pos)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		return true, newNode, oldHashes, nil
	}

	bn.dirty = dirty

	return true, bn, oldHashes, nil
}

func (bn *branchNode) reduceNode(pos int) (node, error) {
	newEn, err := newExtensionNode([]byte{byte(pos)}, bn, bn.db, bn.marsh, bn.hasher)
	if err != nil {
		return nil, err
	}

	return newEn, nil
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

func (bn *branchNode) print(writer io.Writer, index int) {
	if bn == nil {
		return
	}

	str := fmt.Sprintf("B:")
	_, _ = fmt.Fprintln(writer, str)
	for i := 0; i < len(bn.children); i++ {
		if bn.children[i] == nil {
			continue
		}

		child := bn.children[i]
		for j := 0; j < index+len(str)-1; j++ {
			_, _ = fmt.Fprint(writer, " ")
		}
		str2 := fmt.Sprintf("+ %d: ", i)
		_, _ = fmt.Fprint(writer, str2)
		child.print(writer, index+len(str)-1+len(str2))
	}
}

func (bn *branchNode) deepClone() node {
	if bn == nil {
		return nil
	}

	clonedNode := &branchNode{baseNode: &baseNode{}}

	if bn.hash != nil {
		clonedNode.hash = make([]byte, len(bn.hash))
		copy(clonedNode.hash, bn.hash)
	}

	clonedNode.EncodedChildren = make([][]byte, len(bn.EncodedChildren))
	for idx, encChild := range bn.EncodedChildren {
		if encChild == nil {
			continue
		}

		clonedEncChild := make([]byte, len(encChild))
		copy(clonedEncChild, encChild)

		clonedNode.EncodedChildren[idx] = clonedEncChild
	}

	for idx, child := range bn.children {
		if child == nil {
			continue
		}

		clonedNode.children[idx] = child.deepClone()
	}

	clonedNode.dirty = bn.dirty
	clonedNode.db = bn.db
	clonedNode.marsh = bn.marsh
	clonedNode.hasher = bn.hasher

	return clonedNode
}

func (bn *branchNode) getDirtyHashes() ([][]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make([][]byte, 0)

	if !bn.isDirty() {
		return dirtyHashes, nil
	}

	for i := range bn.children {
		if bn.children[i] == nil {
			continue
		}

		var hashes [][]byte
		hashes, err = bn.children[i].getDirtyHashes()
		if err != nil {
			return nil, err
		}

		dirtyHashes = append(dirtyHashes, hashes...)
	}

	dirtyHashes = append(dirtyHashes, bn.getHash())
	return dirtyHashes, nil
}

func (bn *branchNode) getChildren() ([]node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}

	nextNodes := make([]node, 0)

	for i := range bn.children {
		err = resolveIfCollapsed(bn, byte(i))
		if err != nil {
			return nil, err
		}

		if bn.children[i] == nil {
			continue
		}

		nextNodes = append(nextNodes, bn.children[i])
	}

	return nextNodes, nil
}

func (bn *branchNode) isValid() bool {
	nrChildren := 0
	for i := range bn.EncodedChildren {
		if len(bn.EncodedChildren[i]) != 0 || bn.children[i] != nil {
			nrChildren++
		}
	}

	if nrChildren < 2 {
		return false
	}

	return true
}

func (bn *branchNode) setDirty(dirty bool) {
	bn.dirty = dirty
}

func (bn *branchNode) loadChildren(syncer *trieSyncer) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}

	for i := range bn.EncodedChildren {
		if len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		var child node
		child, err = syncer.getNode(bn.EncodedChildren[i])
		if err != nil {
			return err
		}

		bn.children[i] = child
	}

	syncer.interceptedNodes.Remove(bn.hash)

	return nil
}
