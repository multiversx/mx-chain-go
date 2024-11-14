package trie

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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

func (bn *branchNode) setVersionForChild(version core.TrieNodeVersion, childPos byte) {
	sliceNotInitialized := len(bn.ChildrenVersion) == 0

	if version == core.NotSpecified && sliceNotInitialized {
		return
	}

	if sliceNotInitialized {
		bn.ChildrenVersion = make([]byte, nrOfChildren)
	}

	bn.ChildrenVersion[int(childPos)] = byte(version)

	if version == core.NotSpecified {
		bn.revertChildrenVersionSliceIfNeeded()
	}
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
			if !hasValidHash(bn.children[i]) {
				return nil, ErrNodeHashIsNotSet
			}
			collapsed.EncodedChildren[i] = bn.children[i].getHash()
			collapsed.children[i] = nil
		}
	}
	return collapsed, nil
}

func (bn *branchNode) setHash(goRoutinesManager common.TrieGoroutinesManager) {
	if len(bn.hash) != 0 {
		return
	}

	waitGroup := sync.WaitGroup{}

	for i := 0; i < nrOfChildren; i++ {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		if !bn.shouldSetHashForChild(i) {
			continue
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			bn.children[i].setHash(goRoutinesManager)
			encChild, err := encodeNodeAndGetHash(bn.children[i])
			if err != nil {
				goRoutinesManager.SetError(err)
				return
			}

			bn.childrenMutexes[i].Lock()
			bn.EncodedChildren[i] = encChild
			bn.childrenMutexes[i].Unlock()
			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			bn.children[childPos].setHash(goRoutinesManager)
			encChild, err := encodeNodeAndGetHash(bn.children[childPos])
			if err != nil {
				goRoutinesManager.SetError(err)
				return
			}
			bn.childrenMutexes[childPos].Lock()
			bn.EncodedChildren[childPos] = encChild
			bn.childrenMutexes[childPos].Unlock()
			waitGroup.Done()
		}(i)
	}

	waitGroup.Wait()

	hash, err := encodeNodeAndGetHash(bn)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}
	bn.hash = hash
}

func (bn *branchNode) shouldSetHashForChild(childPos int) bool {
	bn.childrenMutexes[childPos].RLock()
	defer bn.childrenMutexes[childPos].RUnlock()

	if bn.children[childPos] != nil && bn.EncodedChildren[childPos] == nil {
		return true
	}

	return false
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

func (bn *branchNode) commitDirty(level byte, maxTrieLevelInMemory uint, originDb common.TrieStorageInteractor, targetDb common.BaseStorer) error {
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

func (bn *branchNode) commitSnapshot(
	db common.TrieStorageInteractor,
	leavesChan chan core.KeyValueHolder,
	missingNodesChan chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider) {
		return core.ErrContextClosing
	}

	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit snapshot error %w", err)
	}

	for i := range bn.children {
		_, err = bn.resolveIfCollapsed(byte(i), db)
		childIsMissing, err := treatCommitSnapshotError(err, bn.EncodedChildren[i], missingNodesChan)
		if err != nil {
			return err
		}
		if childIsMissing {
			continue
		}

		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].commitSnapshot(db, leavesChan, missingNodesChan, ctx, stats, idleProvider, depthLevel+1)
		if err != nil {
			return err
		}
	}

	return bn.saveToStorage(db, stats, depthLevel)
}

func (bn *branchNode) saveToStorage(targetDb common.BaseStorer, stats common.TrieStatisticsHandler, depthLevel int) error {
	nodeSize, err := encodeNodeAndCommitToDB(bn, targetDb)
	if err != nil {
		return err
	}

	stats.AddBranchNode(depthLevel, uint64(nodeSize))

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

func (bn *branchNode) isCollapsed() bool {
	for i := range bn.children {
		if bn.children[i] != nil {
			return false
		}
	}
	return true
}

func (bn *branchNode) resolveIfCollapsed(pos byte, db common.TrieStorageInteractor) (node, error) {
	if childPosOutOfRange(pos) {
		return nil, ErrChildPosOutOfRange
	}

	bn.childrenMutexes[pos].Lock()
	defer bn.childrenMutexes[pos].Unlock()

	isPosCollapsed := bn.children[pos] == nil && len(bn.EncodedChildren[pos]) != 0
	if !isPosCollapsed {
		handleStorageInteractorStats(db)
		return bn.children[pos], nil
	}

	child, err := getNodeFromDBAndDecode(bn.EncodedChildren[pos], db, bn.marsh, bn.hasher)
	if err != nil {
		return nil, err
	}
	bn.children[pos] = child
	return child, nil
}

func (bn *branchNode) tryGet(key []byte, currentDepth uint32, db common.TrieStorageInteractor) (value []byte, maxDepth uint32, err error) {
	if len(key) == 0 {
		return nil, currentDepth, nil
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, currentDepth, ErrChildPosOutOfRange
	}
	key = key[1:]

	child, err := bn.resolveIfCollapsed(childPos, db)
	if err != nil {
		return nil, currentDepth, err
	}
	if check.IfNil(child) {
		return nil, currentDepth, nil
	}

	return child.tryGet(key, currentDepth+1, db)
}

func (bn *branchNode) getNext(key []byte, db common.TrieStorageInteractor) (node, []byte, error) {
	if len(key) == 0 {
		return nil, nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	_, err := bn.resolveIfCollapsed(childPos, db)
	if err != nil {
		return nil, nil, err
	}

	if bn.children[childPos] == nil {
		return nil, nil, ErrNodeNotFound
	}
	return bn.children[childPos], key, nil
}

func (bn *branchNode) insert(
	newData []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	db common.TrieStorageInteractor,
) (node, [][]byte) {
	dataForInsertion, err := splitDataForChildren(newData)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil, [][]byte{}
	}

	hashesMutex := &sync.Mutex{}
	modifiedHashes := make([][]byte, 0)
	bnHasBeenModified := &atomic.Flag{}
	newChildrenMutex := &sync.Mutex{}
	bnNewChildren := make([]node, nrOfChildren)

	waitGroup := sync.WaitGroup{}

	for childPos := range dataForInsertion {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return nil, [][]byte{}
		}

		if len(dataForInsertion[childPos]) == 0 {
			continue
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			newChild, newModifiedHashes := bn.insertOnChild(dataForInsertion[childPos], childPos, goRoutinesManager, db)
			if !check.IfNil(newChild) {
				newChildrenMutex.Lock()
				bnNewChildren[childPos] = newChild
				newChildrenMutex.Unlock()
				bnHasBeenModified.SetValue(true)
			}

			if len(newModifiedHashes) != 0 {
				hashesMutex.Lock()
				modifiedHashes = append(modifiedHashes, newModifiedHashes...)
				hashesMutex.Unlock()
			}

			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			newChild, newModifiedHashes := bn.insertOnChild(dataForInsertion[childPos], childPos, goRoutinesManager, db)
			if !check.IfNil(newChild) {
				newChildrenMutex.Lock()
				bnNewChildren[childPos] = newChild
				newChildrenMutex.Unlock()
				bnHasBeenModified.SetValue(true)
			}
			if len(newModifiedHashes) != 0 {
				hashesMutex.Lock()
				modifiedHashes = append(modifiedHashes, newModifiedHashes...)
				hashesMutex.Unlock()
			}
			waitGroup.Done()
		}(childPos)
	}

	waitGroup.Wait()

	if !bnHasBeenModified.IsSet() {
		return nil, [][]byte{}
	}

	modifiedHashes, err = bn.modifyNodeAfterInsert(modifiedHashes, bnNewChildren)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil, [][]byte{}
	}

	return bn, modifiedHashes
}

func (bn *branchNode) insertOnChild(
	dataForInsertion []core.TrieData,
	childPos int,
	goRoutinesManager common.TrieGoroutinesManager,
	db common.TrieStorageInteractor,
) (node, [][]byte) {
	child, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil, nil
	}

	if check.IfNil(child) {
		return bn.insertOnNilChild(dataForInsertion, goRoutinesManager, db)
	}

	newNode, modifiedHashes := child.insert(dataForInsertion, goRoutinesManager, db)
	if check.IfNil(newNode) {
		return nil, [][]byte{}
	}

	return newNode, modifiedHashes
}

// the prerequisite for this to work is that the data is already sorted
func splitDataForChildren(newData []core.TrieData) ([][]core.TrieData, error) {
	if len(newData) == 0 {
		return nil, ErrValueTooShort
	}
	childrenData := make([][]core.TrieData, nrOfChildren)

	startIndex := 0
	childPos := byte(0)
	prevChildPos := byte(0)
	for i := range newData {
		if len(newData[i].Key) == 0 {
			return nil, ErrValueTooShort
		}
		childPos = newData[i].Key[firstByte]
		if childPosOutOfRange(childPos) {
			return nil, ErrChildPosOutOfRange
		}
		newData[i].Key = newData[i].Key[1:]

		if i == 0 {
			prevChildPos = childPos
			continue
		}

		if childPos == prevChildPos {
			continue
		}

		childrenData[prevChildPos] = newData[startIndex:i]
		startIndex = i
		prevChildPos = childPos
	}

	childrenData[childPos] = newData[startIndex:]
	return childrenData, nil
}

func (bn *branchNode) insertOnNilChild(
	newData []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	db common.TrieStorageInteractor,
) (node, [][]byte) {
	if len(newData) == 0 {
		goRoutinesManager.SetError(ErrValueTooShort)
		return nil, [][]byte{}
	}

	var newNode node
	modifiedHashes := make([][]byte, 0)

	newNode, err := newLeafNode(newData[0], bn.marsh, bn.hasher)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil, [][]byte{}
	}

	if len(newData) > 1 {
		newNode, modifiedHashes = newNode.insert(newData[1:], goRoutinesManager, db)
		if check.IfNil(newNode) {
			return nil, [][]byte{}
		}
	}

	return newNode, modifiedHashes
}

func (bn *branchNode) modifyNodeAfterInsert(
	modifiedHashes [][]byte,
	newBnChildren []node,
) ([][]byte, error) {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	if !bn.dirty {
		modifiedHashes = append(modifiedHashes, bn.hash)
	}

	for i := range newBnChildren {
		if check.IfNil(newBnChildren[i]) {
			continue
		}

		childVersion, err := newBnChildren[i].getVersion()
		if err != nil {
			return nil, err
		}

		bn.children[i] = newBnChildren[i]
		bn.EncodedChildren[i] = nil
		bn.setVersionForChild(childVersion, byte(i))
	}

	bn.dirty = true
	bn.hash = nil

	return modifiedHashes, nil
}

func (bn *branchNode) delete(
	data []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	db common.TrieStorageInteractor,
) (bool, node, [][]byte) {
	dataForRemoval, err := splitDataForChildren(data)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil, [][]byte{}
	}

	hashesMutex := &sync.Mutex{}
	modifiedHashes := make([][]byte, 0)
	hasBeenModified := &atomic.Flag{}
	newChildren := NewModifiedChildren()
	waitGroup := sync.WaitGroup{}

	for childPos := range dataForRemoval {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return false, nil, [][]byte{}
		}

		if len(dataForRemoval[childPos]) == 0 {
			continue
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			newChild, newModifiedHashes, hasNewChild := bn.deleteChild(dataForRemoval[childPos], childPos, goRoutinesManager, db)
			if !hasNewChild {
				continue
			}

			newChildren.AddChild(childPos, newChild)
			hasBeenModified.SetValue(true)

			if len(newModifiedHashes) != 0 {
				hashesMutex.Lock()
				modifiedHashes = append(modifiedHashes, newModifiedHashes...)
				hashesMutex.Unlock()
			}

			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			newChild, newModifiedHashes, hasNewChild := bn.deleteChild(dataForRemoval[childPos], childPos, goRoutinesManager, db)
			if !hasNewChild {
				waitGroup.Done()
				return
			}

			newChildren.AddChild(childPos, newChild)
			hasBeenModified.SetValue(true)

			if len(newModifiedHashes) != 0 {
				hashesMutex.Lock()
				modifiedHashes = append(modifiedHashes, newModifiedHashes...)
				hashesMutex.Unlock()
			}
			waitGroup.Done()
		}(childPos)
	}

	waitGroup.Wait()

	if !hasBeenModified.IsSet() {
		return false, bn, [][]byte{}
	}

	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	modifiedHashes, err = bn.setNewChildren(modifiedHashes, newChildren)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil, [][]byte{}
	}

	numChildren, pos := getChildPosition(bn)
	if numChildren == 0 {
		return true, nil, modifiedHashes
	}
	if numChildren == 1 {
		_, err = bn.resolveIfCollapsed(byte(pos), db)
		if err != nil {
			goRoutinesManager.SetError(err)
			return false, nil, [][]byte{}
		}

		var newChildHash bool
		newNode, newChildHash, err := bn.children[pos].reduceNode(pos, db)
		if err != nil {
			goRoutinesManager.SetError(err)
			return false, nil, [][]byte{}
		}

		if newChildHash && !bn.children[pos].isDirty() {
			modifiedHashes = append(modifiedHashes, bn.children[pos].getHash())
		}

		return true, newNode, modifiedHashes
	}

	return true, bn, modifiedHashes
}

func (bn *branchNode) deleteChild(
	dataForRemoval []core.TrieData,
	childPos int,
	goRoutinesManager common.TrieGoroutinesManager,
	db common.TrieStorageInteractor,
) (node, [][]byte, bool) {
	child, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil, [][]byte{}, false
	}

	if check.IfNil(child) {
		return nil, [][]byte{}, false
	}

	dirty, newNode, oldHashes := child.delete(dataForRemoval, goRoutinesManager, db)
	if !goRoutinesManager.ShouldContinueProcessing() || !dirty {
		return nil, oldHashes, false
	}

	return newNode, oldHashes, true
}

func (bn *branchNode) setNewChildren(
	modifiedHashes [][]byte,
	newChildrenMap *modifiedChildren,
) ([][]byte, error) {
	if !bn.dirty {
		modifiedHashes = append(modifiedHashes, bn.hash)
	}

	bn.hash = nil
	bn.dirty = true

	var rangeError error

	newChildrenMap.Range(func(childPos int, newChild node) {
		bn.children[childPos] = newChild
		bn.EncodedChildren[childPos] = nil
		if check.IfNil(newChild) {
			bn.setVersionForChild(core.NotSpecified, byte(childPos))

			return
		}

		childVersion, err := newChild.getVersion()
		if err != nil {
			rangeError = err
			return
		}
		bn.setVersionForChild(childVersion, byte(childPos))
	})

	return modifiedHashes, rangeError
}

func (bn *branchNode) revertChildrenVersionSliceIfNeeded() {
	notSpecifiedVersion := byte(core.NotSpecified)
	for i := range bn.ChildrenVersion {
		if bn.ChildrenVersion[i] != notSpecifiedVersion {
			return
		}
	}

	bn.ChildrenVersion = []byte(nil)
}

func (bn *branchNode) reduceNode(pos int, _ common.TrieStorageInteractor) (node, bool, error) {
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

func (bn *branchNode) print(writer io.Writer, index int, db common.TrieStorageInteractor) {
	if bn == nil {
		return
	}

	str := fmt.Sprintf("B: %v - %v", hex.EncodeToString(bn.hash), bn.dirty)
	_, _ = fmt.Fprintln(writer, str)
	for i := 0; i < len(bn.children); i++ {
		_, err := bn.resolveIfCollapsed(byte(i), db)
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

func (bn *branchNode) getDirtyHashes(hashes common.ModifiedHashes) error {
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

func (bn *branchNode) getChildren(db common.TrieStorageInteractor) ([]node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getChildren error %w", err)
	}

	nextNodes := make([]node, 0)

	for i := range bn.children {
		childNode, err := bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			return nil, err
		}

		if check.IfNil(childNode) {
			continue
		}

		nextNodes = append(nextNodes, childNode)
	}

	return nextNodes, nil
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
	keyBuilder common.KeyBuilder,
	trieLeafParser common.TrieLeafParser,
	db common.TrieStorageInteractor,
	marshalizer marshal.Marshalizer,
	chanClose chan struct{},
	ctx context.Context,
) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	for i := range bn.children {
		select {
		case <-chanClose:
			log.Trace("branchNode.getAllLeavesOnChannel interrupted")
			return nil
		case <-ctx.Done():
			log.Trace("branchNode.getAllLeavesOnChannel context done")
			return nil
		default:
			_, err = bn.resolveIfCollapsed(byte(i), db)
			if err != nil {
				return err
			}

			if bn.children[i] == nil {
				continue
			}

			clonedKeyBuilder := keyBuilder.Clone()
			clonedKeyBuilder.BuildKey([]byte{byte(i)})
			err = bn.children[i].getAllLeavesOnChannel(leavesChannel, clonedKeyBuilder, trieLeafParser, db, marshalizer, chanClose, ctx)
			if err != nil {
				return err
			}

			bn.children[i] = nil
		}
	}

	return nil
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

func (bn *branchNode) getValue() []byte {
	return []byte{}
}

func (bn *branchNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, db common.TrieStorageInteractor) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("collectStats error %w", err)
	}

	for i := range bn.children {
		_, err = bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			return err
		}

		if bn.children[i] == nil {
			continue
		}

		err = bn.children[i].collectStats(ts, depthLevel+1, db)
		if err != nil {
			return err
		}
	}

	val, err := collapseAndEncodeNode(bn)
	if err != nil {
		return err
	}

	ts.AddBranchNode(depthLevel, uint64(len(val)))
	return nil
}

func (bn *branchNode) getVersion() (core.TrieNodeVersion, error) {
	if len(bn.ChildrenVersion) == 0 {
		return core.NotSpecified, nil
	}

	index := 0
	var nodeVersion byte
	for i := range bn.children {
		index++
		if bn.children[i] == nil && len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		nodeVersion = bn.ChildrenVersion[i]
		break
	}

	for i := index; i < len(bn.children); i++ {
		if bn.children[i] == nil && len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		if bn.ChildrenVersion[i] != nodeVersion {
			return core.NotSpecified, nil
		}
	}

	return core.TrieNodeVersion(nodeVersion), nil
}

func (bn *branchNode) getVersionForChild(childIndex byte) core.TrieNodeVersion {
	if len(bn.ChildrenVersion) == 0 {
		return core.NotSpecified
	}

	return core.TrieNodeVersion(bn.ChildrenVersion[childIndex])
}

func (bn *branchNode) collectLeavesForMigration(
	migrationArgs vmcommon.ArgsMigrateDataTrieLeaves,
	db common.TrieStorageInteractor,
	keyBuilder common.KeyBuilder,
) (bool, error) {
	shouldContinue := migrationArgs.TrieMigrator.ConsumeStorageLoadGas()
	if !shouldContinue {
		return false, nil
	}

	shouldMigrateNode, err := shouldMigrateCurrentNode(bn, migrationArgs)
	if err != nil {
		return false, err
	}
	if !shouldMigrateNode {
		return true, nil
	}

	for i := range bn.children {
		if bn.children[i] == nil && len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		if bn.getVersionForChild(byte(i)) != migrationArgs.OldVersion {
			continue
		}

		childNode, err := bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			return false, err
		}

		clonedKeyBuilder := keyBuilder.Clone()
		clonedKeyBuilder.BuildKey([]byte{byte(i)})
		shouldContinueMigrating, err := childNode.collectLeavesForMigration(migrationArgs, db, clonedKeyBuilder)
		if err != nil {
			return false, err
		}

		if !shouldContinueMigrating {
			return false, nil
		}
	}

	return true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bn *branchNode) IsInterfaceNil() bool {
	return bn == nil
}
