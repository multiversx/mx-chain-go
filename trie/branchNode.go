package trie

import (
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

var _ = node(&branchNode{})

func newBranchNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*branchNode, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: CollapsedBn{
			EncodedChildren: encChildren,
		},
		children: children,
		baseNode: &baseNode{
			dirty:  true,
			marsh:  marshalizer,
			hasher: hasher,
		},
	}, nil
}

func emptyDirtyBranchNode() *branchNode {
	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: CollapsedBn{
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

func (bn *branchNode) getCollapsed() (node, error) {
	return bn.getCollapsedBn()
}

func (bn *branchNode) getCollapsedBn() (*branchNode, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getCollapsed error %w", err)
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
		return fmt.Errorf("setHash error %w", err)
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
		return fmt.Errorf("setRootHash error %w", err)
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
		c <- fmt.Errorf("setHashConcurrent error %w", err)
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
}

func (bn *branchNode) hashChildren() error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("hashChildren error %w", err)
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
		return nil, fmt.Errorf("hashNode error %w", err)
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

func (bn *branchNode) commitDirty(level byte, maxTrieLevelInMemory uint, originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher) error {
	level++
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit error %w", err)
	}

	if !bn.dirty {
		return nil
	}

	for i := range bn.children {
		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].commitDirty(level, maxTrieLevelInMemory, originDb, targetDb)
		if err != nil {
			return err
		}
	}
	bn.dirty = false
	_, err = encodeNodeAndCommitToDB(bn, targetDb)
	if err != nil {
		return err
	}
	if uint(level) == maxTrieLevelInMemory {
		log.Trace("collapse branch node on commit")

		var collapsedBn *branchNode
		collapsedBn, err = bn.getCollapsedBn()
		if err != nil {
			return err
		}

		*bn = *collapsedBn
	}
	return nil
}

func (bn *branchNode) commitCheckpoint(
	originDb temporary.DBWriteCacher,
	targetDb temporary.DBWriteCacher,
	checkpointHashes temporary.CheckpointHashesHolder,
	leavesChan chan core.KeyValueHolder,
) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit checkpoint error %w", err)
	}

	hash, err := computeAndSetNodeHash(bn)
	if err != nil {
		return err
	}

	shouldCommit := checkpointHashes.ShouldCommit(hash)
	if !shouldCommit {
		return nil
	}

	for i := range bn.children {
		err = resolveIfCollapsed(bn, byte(i), originDb)
		if err != nil {
			return err
		}

		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].commitCheckpoint(originDb, targetDb, checkpointHashes, leavesChan)
		if err != nil {
			return err
		}
	}

	checkpointHashes.Remove(hash)
	return bn.saveToStorage(targetDb)
}

func (bn *branchNode) commitSnapshot(
	originDb temporary.DBWriteCacher,
	targetDb temporary.DBWriteCacher,
	leavesChan chan core.KeyValueHolder,
) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit snapshot error %w", err)
	}

	for i := range bn.children {
		err = resolveIfCollapsed(bn, byte(i), originDb)
		if err != nil {
			return err
		}

		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].commitSnapshot(originDb, targetDb, leavesChan)
		if err != nil {
			return err
		}
	}

	return bn.saveToStorage(targetDb)
}

func (bn *branchNode) saveToStorage(targetDb temporary.DBWriteCacher) error {
	_, err := encodeNodeAndCommitToDB(bn, targetDb)
	if err != nil {
		return err
	}

	bn.removeChildrenPointers()
	return nil
}

func (bn *branchNode) removeChildrenPointers() {
	for i := range bn.children {
		bn.children[i] = nil
	}
}

func (bn *branchNode) getEncodedNode() ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getEncodedNode error %w", err)
	}
	marshaledNode, err := bn.marsh.Marshal(bn)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, branch)
	return marshaledNode, nil
}

func (bn *branchNode) resolveCollapsed(pos byte, db temporary.DBWriteCacher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("resolveCollapsed error %w", err)
	}
	if childPosOutOfRange(pos) {
		return ErrChildPosOutOfRange
	}
	if len(bn.EncodedChildren[pos]) != 0 {
		var child node
		child, err = getNodeFromDBAndDecode(bn.EncodedChildren[pos], db, bn.marsh, bn.hasher)
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

func (bn *branchNode) tryGet(key []byte, db temporary.DBWriteCacher) (value []byte, err error) {
	err = bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("tryGet error %w", err)
	}
	if len(key) == 0 {
		return nil, nil
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db)
	if err != nil {
		return nil, err
	}
	if bn.children[childPos] == nil {
		return nil, nil
	}

	return bn.children[childPos].tryGet(key, db)
}

func (bn *branchNode) getNext(key []byte, db temporary.DBWriteCacher) (node, []byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("getNext error %w", err)
	}
	if len(key) == 0 {
		return nil, nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db)
	if err != nil {
		return nil, nil, err
	}

	if bn.children[childPos] == nil {
		return nil, nil, ErrNodeNotFound
	}
	return bn.children[childPos], key, nil
}

func (bn *branchNode) insert(n *leafNode, db temporary.DBWriteCacher) (node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, emptyHashes, fmt.Errorf("insert error %w", err)
	}

	insertedKey := n.Key
	if len(insertedKey) == 0 {
		return nil, emptyHashes, ErrValueTooShort
	}
	childPos := insertedKey[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, emptyHashes, ErrChildPosOutOfRange
	}

	n.Key = insertedKey[1:]
	err = resolveIfCollapsed(bn, childPos, db)
	if err != nil {
		return nil, emptyHashes, err
	}

	if bn.children[childPos] == nil {
		return bn.insertOnNilChild(n, childPos)
	}

	return bn.insertOnExistingChild(n, childPos, db)
}

func (bn *branchNode) insertOnNilChild(n *leafNode, childPos byte) (node, [][]byte, error) {
	newLn, err := newLeafNode(n.Key, n.Value, bn.marsh, bn.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	modifiedHashes := make([][]byte, 0)
	modifiedHashes = bn.modifyNodeAfterInsert(modifiedHashes, childPos, newLn)

	return bn, modifiedHashes, nil
}

func (bn *branchNode) insertOnExistingChild(n *leafNode, childPos byte, db temporary.DBWriteCacher) (node, [][]byte, error) {
	newNode, modifiedHashes, err := bn.children[childPos].insert(n, db)
	if check.IfNil(newNode) || err != nil {
		return nil, [][]byte{}, err
	}

	modifiedHashes = bn.modifyNodeAfterInsert(modifiedHashes, childPos, newNode)

	return bn, modifiedHashes, nil
}

func (bn *branchNode) modifyNodeAfterInsert(modifiedHashes [][]byte, childPos byte, newNode node) [][]byte {
	if !bn.dirty {
		modifiedHashes = append(modifiedHashes, bn.hash)
	}

	bn.children[childPos] = newNode
	bn.dirty = true
	bn.hash = nil

	return modifiedHashes
}

func (bn *branchNode) delete(key []byte, db temporary.DBWriteCacher) (bool, node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, emptyHashes, fmt.Errorf("delete error %w", err)
	}
	if len(key) == 0 {
		return false, nil, emptyHashes, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, emptyHashes, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db)
	if err != nil {
		return false, nil, emptyHashes, err
	}

	if bn.children[childPos] == nil {
		return false, bn, emptyHashes, nil
	}

	dirty, newNode, oldHashes, err := bn.children[childPos].delete(key, db)
	if !dirty || err != nil {
		return false, bn, emptyHashes, err
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
		err = resolveIfCollapsed(bn, byte(pos), db)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		err = resolveIfCollapsed(bn.children[pos], byte(pos), db)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		var newChildHash bool
		newNode, newChildHash, err = bn.children[pos].reduceNode(pos)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		if newChildHash && !bn.children[pos].isDirty() {
			oldHashes = append(oldHashes, bn.children[pos].getHash())
		}

		return true, newNode, oldHashes, nil
	}

	bn.dirty = dirty

	return true, bn, oldHashes, nil
}

func (bn *branchNode) reduceNode(pos int) (node, bool, error) {
	newEn, err := newExtensionNode([]byte{byte(pos)}, bn, bn.marsh, bn.hasher)
	if err != nil {
		return nil, false, err
	}

	return newEn, false, nil
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
		return ErrNilBranchNode
	}
	for i := range bn.children {
		if bn.children[i] != nil || len(bn.EncodedChildren[i]) != 0 {
			return nil
		}
	}
	return ErrEmptyBranchNode
}

func (bn *branchNode) print(writer io.Writer, index int, db temporary.DBWriteCacher) {
	if bn == nil {
		return
	}

	str := fmt.Sprintf("B: %v - %v", hex.EncodeToString(bn.hash), bn.dirty)
	_, _ = fmt.Fprintln(writer, str)
	for i := 0; i < len(bn.children); i++ {
		err := resolveIfCollapsed(bn, byte(i), db)
		if err != nil {
			log.Debug("branch node: print trie err", "error", err, "hash", bn.EncodedChildren[i])
		}

		if bn.children[i] == nil {
			continue
		}

		child := bn.children[i]
		for j := 0; j < index+len(str)-1; j++ {
			_, _ = fmt.Fprint(writer, " ")
		}
		str2 := fmt.Sprintf("+ %d: ", i)
		_, _ = fmt.Fprint(writer, str2)
		childIndex := index + len(str) - 1 + len(str2)
		child.print(writer, childIndex, db)
	}
}

func (bn *branchNode) getDirtyHashes(hashes temporary.ModifiedHashes) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getDirtyHashes error %w", err)
	}

	if !bn.isDirty() {
		return nil
	}

	for i := range bn.children {
		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].getDirtyHashes(hashes)
		if err != nil {
			return err
		}
	}

	hashes[string(bn.getHash())] = struct{}{}
	return nil
}

func (bn *branchNode) getChildren(db temporary.DBWriteCacher) ([]node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getChildren error %w", err)
	}

	nextNodes := make([]node, 0)

	for i := range bn.children {
		err = resolveIfCollapsed(bn, byte(i), db)
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

	return nrChildren >= 2
}

func (bn *branchNode) setDirty(dirty bool) {
	bn.dirty = dirty
}

func (bn *branchNode) loadChildren(getNode func([]byte) (node, error)) ([][]byte, []node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("loadChildren error %w", err)
	}

	existingChildren := make([]node, 0)
	missingChildren := make([][]byte, 0)
	for i := range bn.EncodedChildren {
		if len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		var child node
		child, err = getNode(bn.EncodedChildren[i])
		if err != nil {
			missingChildren = append(missingChildren, bn.EncodedChildren[i])
			continue
		}

		existingChildren = append(existingChildren, child)
		log.Trace("load branch node child", "child hash", bn.EncodedChildren[i])
		bn.children[i] = child
	}

	return missingChildren, existingChildren, nil
}

func (bn *branchNode) getAllLeavesOnChannel(
	leavesChannel chan core.KeyValueHolder,
	key []byte, db temporary.DBWriteCacher,
	marshalizer marshal.Marshalizer,
	chanClose chan struct{},
) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	for i := range bn.children {
		select {
		case <-chanClose:
			log.Trace("getAllLeavesOnChannel interrupted")
			return nil
		default:
			err = resolveIfCollapsed(bn, byte(i), db)
			if err != nil {
				return err
			}

			if bn.children[i] == nil {
				continue
			}

			childKey := append(key, byte(i))
			err = bn.children[i].getAllLeavesOnChannel(leavesChannel, childKey, db, marshalizer, chanClose)
			if err != nil {
				return err
			}

			bn.children[i] = nil
		}
	}

	return nil
}

func (bn *branchNode) getAllHashes(db temporary.DBWriteCacher) ([][]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getAllHashes error: %w", err)
	}

	var childrenHashes [][]byte
	hashes := make([][]byte, 0)
	for i := range bn.children {
		err = resolveIfCollapsed(bn, byte(i), db)
		if err != nil {
			return nil, err
		}

		if bn.children[i] == nil {
			continue
		}

		childrenHashes, err = bn.children[i].getAllHashes(db)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, childrenHashes...)
	}

	hashes = append(hashes, bn.hash)

	return hashes, nil
}

func (bn *branchNode) getNumNodes() temporary.NumNodesDTO {
	if check.IfNil(bn) {
		return temporary.NumNodesDTO{}
	}

	currentNumNodes := temporary.NumNodesDTO{
		Branches: 1,
	}

	for _, n := range bn.children {
		if check.IfNil(n) {
			continue
		}

		childNumNodes := n.getNumNodes()
		currentNumNodes.Branches += childNumNodes.Branches
		currentNumNodes.Leaves += childNumNodes.Leaves
		currentNumNodes.Extensions += childNumNodes.Extensions
		if childNumNodes.MaxLevel > currentNumNodes.MaxLevel {
			currentNumNodes.MaxLevel = childNumNodes.MaxLevel
		}
	}

	currentNumNodes.MaxLevel++

	return currentNumNodes
}

func (bn *branchNode) getNextHashAndKey(key []byte) (bool, []byte, []byte) {
	if len(key) == 0 || check.IfNil(bn) {
		return false, nil, nil
	}

	wantHash := bn.EncodedChildren[key[0]]
	nextKey := key[1:]

	return false, wantHash, nextKey
}

func (bn *branchNode) sizeInBytes() int {
	if bn == nil {
		return 0
	}

	// hasher + marshalizer + dirty flag = numNodeInnerPointers * pointerSizeInBytes + 1
	nodeSize := len(bn.hash) + numNodeInnerPointers*pointerSizeInBytes + 1
	for _, collapsed := range bn.EncodedChildren {
		nodeSize += len(collapsed)
	}
	nodeSize += len(bn.children) * pointerSizeInBytes

	return nodeSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (bn *branchNode) IsInterfaceNil() bool {
	return bn == nil
}
