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
			bn.setHashForChild(i, goRoutinesManager)
			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			bn.setHashForChild(childPos, goRoutinesManager)
			goRoutinesManager.EndGoRoutineProcessing()
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

func (bn *branchNode) setHashForChild(childPos int, goRoutinesManager common.TrieGoroutinesManager) {
	bn.children[childPos].setHash(goRoutinesManager)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return
	}

	encChild, err := encodeNodeAndGetHash(bn.children[childPos])
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	bn.childrenMutexes[childPos].Lock()
	bn.EncodedChildren[childPos] = encChild
	bn.childrenMutexes[childPos].Unlock()
}

func (bn *branchNode) shouldSetHashForChild(childPos int) bool {
	bn.childrenMutexes[childPos].RLock()
	defer bn.childrenMutexes[childPos].RUnlock()

	if bn.children[childPos] != nil && bn.EncodedChildren[childPos] == nil {
		return true
	}

	return false
}

func (bn *branchNode) commitDirty(
	level byte,
	maxTrieLevelInMemory uint,
	goRoutinesManager common.TrieGoroutinesManager,
	hashesCollector common.TrieHashesCollector,
	originDb common.TrieStorageInteractor,
	targetDb common.BaseStorer,
) {
	level++

	if !bn.dirty {
		return
	}

	waitGroup := sync.WaitGroup{}

	for i := 0; i < nrOfChildren; i++ {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		bn.childrenMutexes[i].RLock()
		child := bn.children[i]
		bn.childrenMutexes[i].RUnlock()

		if child == nil {
			continue
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			child.commitDirty(level, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, originDb, targetDb)
			if !goRoutinesManager.ShouldContinueProcessing() {
				return
			}

			bn.childrenMutexes[i].Lock()
			bn.EncodedChildren[i] = child.getHash()
			bn.childrenMutexes[i].Unlock()

			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			defer func() {
				goRoutinesManager.EndGoRoutineProcessing()
				waitGroup.Done()
			}()

			child.commitDirty(level, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, originDb, targetDb)
			if !goRoutinesManager.ShouldContinueProcessing() {
				return
			}

			bn.childrenMutexes[childPos].Lock()
			bn.EncodedChildren[childPos] = child.getHash()
			bn.childrenMutexes[childPos].Unlock()
		}(i)

	}

	waitGroup.Wait()

	ok := saveDirtyNodeToStorage(bn, goRoutinesManager, hashesCollector, targetDb, bn.hasher)
	if !ok {
		return
	}

	if uint(level) == maxTrieLevelInMemory {
		log.Trace("collapse branch node on commit")

		for i := range bn.children {
			bn.childrenMutexes[i].Lock()
			bn.children[i] = nil
			bn.childrenMutexes[i].Unlock()
		}
	}
}

func (bn *branchNode) commitSnapshot(
	db common.TrieStorageInteractor,
	leavesChan chan core.KeyValueHolder,
	missingNodesChan chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	nodeBytes []byte,
	depthLevel int,
) error {
	if shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider) {
		return core.ErrContextClosing
	}

	for i := range bn.EncodedChildren {
		if len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		child, childBytes, err := getNodeFromDBAndDecode(bn.EncodedChildren[i], db, bn.marsh, bn.hasher)
		childIsMissing, err := treatCommitSnapshotError(err, bn.EncodedChildren[i], missingNodesChan)
		if err != nil {
			return err
		}
		if childIsMissing {
			continue
		}

		err = child.commitSnapshot(db, leavesChan, missingNodesChan, ctx, stats, idleProvider, childBytes, depthLevel+1)
		if err != nil {
			return err
		}
	}

	return bn.saveToStorage(db, stats, nodeBytes, depthLevel)
}

func (bn *branchNode) saveToStorage(targetDb common.BaseStorer, stats common.TrieStatisticsHandler, nodeBytes []byte, depthLevel int) error {
	err := targetDb.Put(bn.hash, nodeBytes)
	if err != nil {
		return err
	}

	stats.AddBranchNode(depthLevel, uint64(len(nodeBytes)))
	return nil
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

	child, _, err := getNodeFromDBAndDecode(bn.EncodedChildren[pos], db, bn.marsh, bn.hasher)
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
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) node {
	dataForInsertion, err := splitDataForChildren(newData)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	bnHasBeenModified := &atomic.Flag{}
	bn.updateNode(dataForInsertion, goRoutinesManager, modifiedHashes, bnHasBeenModified, db, bn.insertOnChild)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return nil
	}

	if bnHasBeenModified.IsSet() {
		return bn
	}

	return nil
}

func (bn *branchNode) updateNode(
	data [][]core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	hasBeenModified *atomic.Flag,
	db common.TrieStorageInteractor,
	updateFunc func([]core.TrieData, int, common.TrieGoroutinesManager, common.AtomicBytesSlice, *atomic.Flag, common.TrieStorageInteractor),
) {
	waitGroup := sync.WaitGroup{}

	for childPos := range data {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		if len(data[childPos]) == 0 {
			continue
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			updateFunc(data[childPos], childPos, goRoutinesManager, modifiedHashes, hasBeenModified, db)

			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			updateFunc(data[childPos], childPos, goRoutinesManager, modifiedHashes, hasBeenModified, db)

			goRoutinesManager.EndGoRoutineProcessing()
			waitGroup.Done()
		}(childPos)
	}

	waitGroup.Wait()
}

func (bn *branchNode) insertOnChild(
	dataForInsertion []core.TrieData,
	childPos int,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	bnHasBeenModified *atomic.Flag,
	db common.TrieStorageInteractor,
) {
	child, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if child == nil {
		bn.insertOnNilChild(dataForInsertion, byte(childPos), goRoutinesManager, modifiedHashes, bnHasBeenModified, db)
		return
	}

	newNode := child.insert(dataForInsertion, goRoutinesManager, modifiedHashes, db)
	if check.IfNil(newNode) {
		return
	}

	err = bn.modifyNodeAfterInsert(modifiedHashes, byte(childPos), newNode)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	bnHasBeenModified.SetValue(true)
}

// the prerequisite for this to work is that the data is already sorted
func splitDataForChildren(newSortedData []core.TrieData) ([][]core.TrieData, error) {
	if len(newSortedData) == 0 {
		return nil, ErrValueTooShort
	}
	childrenData := make([][]core.TrieData, nrOfChildren)

	startIndex := 0
	childPos := byte(0)
	prevChildPos := byte(0)
	for i := range newSortedData {
		if len(newSortedData[i].Key) == 0 {
			return nil, ErrValueTooShort
		}
		childPos = newSortedData[i].Key[firstByte]
		if childPosOutOfRange(childPos) {
			return nil, ErrChildPosOutOfRange
		}
		newSortedData[i].Key = newSortedData[i].Key[1:]

		if i == 0 {
			prevChildPos = childPos
			continue
		}

		if childPos == prevChildPos {
			continue
		}

		childrenData[prevChildPos] = newSortedData[startIndex:i]
		startIndex = i
		prevChildPos = childPos
	}

	childrenData[childPos] = newSortedData[startIndex:]
	return childrenData, nil
}

func (bn *branchNode) insertOnNilChild(
	newData []core.TrieData,
	childPos byte,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	bnHasBeenModified *atomic.Flag,
	db common.TrieStorageInteractor,
) {
	if len(newData) == 0 {
		goRoutinesManager.SetError(ErrValueTooShort)
		return
	}

	var newNode node

	newNode, err := newLeafNode(newData[0], bn.marsh, bn.hasher)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if len(newData) > 1 {
		newNode = newNode.insert(newData[1:], goRoutinesManager, modifiedHashes, db)
		if check.IfNil(newNode) {
			return
		}
	}

	err = bn.modifyNodeAfterInsert(modifiedHashes, childPos, newNode)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	bnHasBeenModified.SetValue(true)
}

func (bn *branchNode) modifyNodeAfterInsert(modifiedHashes common.AtomicBytesSlice, childPos byte, newNode node) error {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	if !bn.dirty {
		modifiedHashes.Append([][]byte{bn.hash})
	}

	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}

	bn.childrenMutexes[childPos].Lock()
	bn.EncodedChildren[childPos] = nil
	bn.children[childPos] = newNode
	bn.childrenMutexes[childPos].Unlock()

	bn.setVersionForChild(childVersion, childPos)
	bn.dirty = true
	bn.hash = nil

	return nil
}

func (bn *branchNode) delete(
	data []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) (bool, node) {
	dataForRemoval, err := splitDataForChildren(data)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	hasBeenModified := &atomic.Flag{}

	bn.updateNode(dataForRemoval, goRoutinesManager, modifiedHashes, hasBeenModified, db, bn.deleteChild)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return false, nil
	}

	if !hasBeenModified.IsSet() {
		return false, bn
	}

	return bn.reduceNodeIfNeeded(goRoutinesManager, modifiedHashes, db)
}

func (bn *branchNode) reduceNodeIfNeeded(
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) (bool, node) {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	numChildren, pos := getChildPosition(bn)
	if numChildren == 0 {
		return true, nil
	}

	if numChildren != 1 {
		return true, bn
	}

	child, err := bn.resolveIfCollapsed(byte(pos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	var newChildHash bool
	newNode, newChildHash, err := child.reduceNode(pos, db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	if newChildHash && !child.isDirty() {
		modifiedHashes.Append([][]byte{child.getHash()})
	}

	return true, newNode
}

func (bn *branchNode) deleteChild(
	dataForRemoval []core.TrieData,
	childPos int,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	hasBeenModified *atomic.Flag,
	db common.TrieStorageInteractor,
) {
	child, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if check.IfNil(child) {
		return
	}

	dirty, newNode := child.delete(dataForRemoval, goRoutinesManager, modifiedHashes, db)
	if !goRoutinesManager.ShouldContinueProcessing() || !dirty {
		return
	}

	err = bn.setNewChild(byte(childPos), newNode, modifiedHashes)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	hasBeenModified.SetValue(true)
}

func (bn *branchNode) setNewChild(childPos byte, newNode node, modifiedHashes common.AtomicBytesSlice) error {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	if !bn.dirty && len(bn.hash) != 0 {
		modifiedHashes.Append([][]byte{bn.hash})
	}

	bn.hash = nil
	bn.dirty = true

	bn.childrenMutexes[childPos].Lock()
	bn.children[childPos] = newNode
	bn.EncodedChildren[childPos] = nil
	bn.childrenMutexes[childPos].Unlock()

	if check.IfNil(newNode) {
		bn.setVersionForChild(core.NotSpecified, childPos)
		return nil
	}
	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}
	bn.setVersionForChild(childVersion, childPos)
	return nil
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

func getChildPosition(n *branchNode) (numChildren int, childPos int) {
	for i := 0; i < nrOfChildren; i++ {
		n.childrenMutexes[i].RLock()
		if n.children[i] != nil || len(n.EncodedChildren[i]) != 0 {
			numChildren++
			childPos = i
		}
		n.childrenMutexes[i].RUnlock()
	}
	return
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

func (bn *branchNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, db common.TrieStorageInteractor) error {
	for i := range bn.EncodedChildren {
		if len(bn.EncodedChildren[i]) == 0 {
			continue
		}

		child, childBytes, err := getNodeFromDBAndDecode(bn.EncodedChildren[i], db, bn.marsh, bn.hasher)
		if err != nil {
			return err
		}

		err = child.collectStats(ts, depthLevel+1, uint64(len(childBytes)), db)
		if err != nil {
			return err
		}
	}

	ts.AddBranchNode(depthLevel, nodeSize)
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
