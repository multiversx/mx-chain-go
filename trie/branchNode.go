package trie

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/trieNodeData"
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
			ChildrenHashes: encChildren,
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

		if len(bn.ChildrenHashes[i]) == 0 {
			goRoutinesManager.SetError(ErrInvalidNodeState)
			return
		}

		if !goRoutinesManager.CanStartGoRoutine() {
			child.commitDirty(level, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, originDb, targetDb)
			if !goRoutinesManager.ShouldContinueProcessing() {
				return
			}
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

	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		child, childBytes, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], db, bn.marsh, bn.hasher)
		childIsMissing, err := treatCommitSnapshotError(err, bn.ChildrenHashes[i], missingNodesChan)
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

		err = db.Put(bn.ChildrenHashes[i], childBytes)
		if err != nil {
			return err
		}
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

func (bn *branchNode) resolveIfCollapsed(pos byte, db common.TrieStorageInteractor) (node, []byte, error) {
	if childPosOutOfRange(pos) {
		return nil, nil, ErrChildPosOutOfRange
	}

	bn.childrenMutexes[pos].Lock()
	defer bn.childrenMutexes[pos].Unlock()

	isPosCollapsed := check.IfNil(bn.children[pos]) && len(bn.ChildrenHashes[pos]) != 0
	if isPosCollapsed {
		child, _, err := getNodeFromDBAndDecode(bn.ChildrenHashes[pos], db, bn.marsh, bn.hasher)
		if err != nil {
			return nil, nil, err
		}
		bn.children[pos] = child
		return child, bn.ChildrenHashes[pos], nil
	}

	if !check.IfNil(bn.children[pos]) && len(bn.ChildrenHashes[pos]) == 0 {
		return nil, nil, ErrInvalidNodeState
	}

	handleStorageInteractorStats(db)
	return bn.children[pos], bn.ChildrenHashes[pos], nil
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

	child, _, err := bn.resolveIfCollapsed(childPos, db)
	if err != nil {
		return nil, currentDepth, err
	}
	if check.IfNil(child) {
		return nil, currentDepth, nil
	}

	return child.tryGet(key, currentDepth+1, db)
}

func (bn *branchNode) getNext(key []byte, db common.TrieStorageInteractor) (*nodeData, error) {
	if len(key) == 0 {
		return nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	if len(bn.ChildrenHashes[childPos]) == 0 {
		return nil, ErrNodeNotFound
	}
	childNode, encodedNode, err := getNodeFromDBAndDecode(bn.ChildrenHashes[childPos], db, bn.marsh, bn.hasher)
	if err != nil {
		return nil, err
	}

	return &nodeData{
		currentNode: childNode,
		encodedNode: encodedNode,
		hexKey:      key,
	}, nil
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

	if !bnHasBeenModified.IsSet() {
		return nil
	}

	return bn
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
	child, childHash, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if child == nil {
		bn.insertOnNilChild(dataForInsertion, byte(childPos), goRoutinesManager, modifiedHashes, bnHasBeenModified, db)
		return
	}

	var originalChildHash []byte
	if !child.isDirty() {
		originalChildHash = childHash
	}

	newNode := child.insert(dataForInsertion, goRoutinesManager, modifiedHashes, db)
	if check.IfNil(newNode) {
		return
	}

	if len(originalChildHash) != 0 {
		modifiedHashes.Append([][]byte{originalChildHash})
	}

	err = bn.modifyNodeAfterInsert(byte(childPos), newNode)
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

	err = bn.modifyNodeAfterInsert(childPos, newNode)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	bnHasBeenModified.SetValue(true)
}

func (bn *branchNode) modifyNodeAfterInsert(
	childPos byte,
	newNode node,
) error {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}

	newHashForChild, err := encodeNodeAndGetHash(newNode)
	if err != nil {
		return err
	}

	bn.childrenMutexes[childPos].Lock()
	bn.ChildrenHashes[childPos] = newHashForChild
	bn.children[childPos] = newNode
	bn.childrenMutexes[childPos].Unlock()

	bn.setVersionForChild(childVersion, childPos)
	bn.dirty = true

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

	child, originalChildHash, err := bn.resolveIfCollapsed(byte(pos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	var newChildHash bool
	newNode, newChildHash, err := child.reduceNode(pos, originalChildHash, db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	if newChildHash && !child.isDirty() {
		modifiedHashes.Append([][]byte{originalChildHash})
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
	child, childHash, err := bn.resolveIfCollapsed(byte(childPos), db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if check.IfNil(child) {
		return
	}

	var originalChildHash []byte
	if !child.isDirty() {
		originalChildHash = childHash
	}

	dirty, newNode := child.delete(dataForRemoval, goRoutinesManager, modifiedHashes, db)
	if !goRoutinesManager.ShouldContinueProcessing() || !dirty {
		return
	}

	err = bn.setNewChild(byte(childPos), newNode)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if len(originalChildHash) != 0 {
		modifiedHashes.Append([][]byte{originalChildHash})
	}

	hasBeenModified.SetValue(true)
}

func (bn *branchNode) setNewChild(
	childPos byte,
	newNode node,
) error {
	bn.mutex.Lock()
	defer bn.mutex.Unlock()

	bn.dirty = true

	if check.IfNil(newNode) {
		bn.childrenMutexes[childPos].Lock()
		bn.children[childPos] = nil
		bn.ChildrenHashes[childPos] = nil
		bn.setVersionForChild(core.NotSpecified, childPos)
		bn.childrenMutexes[childPos].Unlock()
		return nil
	}
	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}

	newChildHash, err := encodeNodeAndGetHash(newNode)
	if err != nil {
		return err
	}

	bn.childrenMutexes[childPos].Lock()
	bn.children[childPos] = newNode
	bn.ChildrenHashes[childPos] = newChildHash
	bn.setVersionForChild(childVersion, childPos)
	bn.childrenMutexes[childPos].Unlock()

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

func (bn *branchNode) reduceNode(pos int, hash []byte, _ common.TrieStorageInteractor) (node, bool, error) {
	newEn, err := newExtensionNode([]byte{byte(pos)}, bn, bn.marsh, bn.hasher)
	if err != nil {
		return nil, false, err
	}
	newEn.ChildHash = hash

	return newEn, false, nil
}

func getChildPosition(n *branchNode) (numChildren int, childPos int) {
	for i := 0; i < nrOfChildren; i++ {
		n.childrenMutexes[i].RLock()
		if n.children[i] != nil || len(n.ChildrenHashes[i]) != 0 {
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
		if bn.children[i] != nil || len(bn.ChildrenHashes[i]) != 0 {
			return nil
		}
	}
	return ErrEmptyBranchNode
}

func (bn *branchNode) print(writer io.Writer, index int, db common.TrieStorageInteractor) {
	if bn == nil {
		return
	}

	str := fmt.Sprintf("B: %v", bn.dirty)
	_, _ = fmt.Fprintln(writer, str)
	for i := 0; i < len(bn.children); i++ {
		_, _, err := bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			log.Debug("branch node: print trie err", "error", err, "hash", bn.ChildrenHashes[i])
		}

		if bn.children[i] == nil {
			continue
		}

		child := bn.children[i]
		for j := 0; j < index+len(str)-1; j++ {
			_, _ = fmt.Fprint(writer, " ")
		}
		str2 := fmt.Sprintf("+ %d: - hash: %v ", i, bn.ChildrenHashes[i])
		_, _ = fmt.Fprint(writer, str2)
		childIndex := index + len(str) - 1 + len(str2)
		child.print(writer, childIndex, db)
	}
}

func (bn *branchNode) getChildren(db common.TrieStorageInteractor) ([]nodeWithHash, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getChildren error %w", err)
	}

	nextNodes := make([]nodeWithHash, 0)

	for i := range bn.children {
		childNode, _, err := bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			return nil, err
		}

		if check.IfNil(childNode) {
			continue
		}

		nextNodes = append(nextNodes, nodeWithHash{
			node: childNode,
			hash: bn.ChildrenHashes[i],
		})
	}

	return nextNodes, nil
}

func (bn *branchNode) loadChildren(getNode func([]byte) (node, error)) ([][]byte, []nodeWithHash, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("loadChildren error %w", err)
	}

	existingChildren := make([]nodeWithHash, 0)
	missingChildren := make([][]byte, 0)
	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		var child node
		child, err = getNode(bn.ChildrenHashes[i])
		if err != nil {
			missingChildren = append(missingChildren, bn.ChildrenHashes[i])
			continue
		}

		existingChildren = append(existingChildren, nodeWithHash{
			node: child,
			hash: bn.ChildrenHashes[i],
		})
		log.Trace("load branch node child", "child hash", bn.ChildrenHashes[i])
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
			_, _, err = bn.resolveIfCollapsed(byte(i), db)
			if err != nil {
				return err
			}

			if bn.children[i] == nil {
				continue
			}

			clonedKeyBuilder := keyBuilder.ShallowClone()
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

	wantHash := bn.ChildrenHashes[key[0]]
	nextKey := key[1:]

	return false, wantHash, nextKey
}

func (bn *branchNode) sizeInBytes() int {
	if bn == nil {
		return 0
	}

	// hasher + marshalizer + dirty flag = numNodeInnerPointers * pointerSizeInBytes + 1
	nodeSize := +numNodeInnerPointers*pointerSizeInBytes + 1
	for _, collapsed := range bn.ChildrenHashes {
		nodeSize += len(collapsed)
	}
	nodeSize += len(bn.children) * pointerSizeInBytes

	return nodeSize
}

func (bn *branchNode) getValue() []byte {
	return []byte{}
}

func (bn *branchNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, db common.TrieStorageInteractor) error {
	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		child, childBytes, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], db, bn.marsh, bn.hasher)
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
		if bn.children[i] == nil && len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		nodeVersion = bn.ChildrenVersion[i]
		break
	}

	for i := index; i < len(bn.children); i++ {
		if bn.children[i] == nil && len(bn.ChildrenHashes[i]) == 0 {
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
		if bn.children[i] == nil && len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		if bn.getVersionForChild(byte(i)) != migrationArgs.OldVersion {
			continue
		}

		childNode, _, err := bn.resolveIfCollapsed(byte(i), db)
		if err != nil {
			return false, err
		}

		clonedKeyBuilder := keyBuilder.ShallowClone()
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

func (bn *branchNode) getNodeData(keyBuilder common.KeyBuilder) ([]common.TrieNodeData, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getNodeData error %w", err)
	}

	data := make([]common.TrieNodeData, 0)
	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		clonedKeyBuilder := keyBuilder.DeepClone()
		clonedKeyBuilder.BuildKey([]byte{byte(i)})
		childData, err := trieNodeData.NewIntermediaryNodeData(clonedKeyBuilder, bn.ChildrenHashes[i])
		if err != nil {
			return nil, err
		}
		data = append(data, childData)
	}

	return data, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bn *branchNode) IsInterfaceNil() bool {
	return bn == nil
}
