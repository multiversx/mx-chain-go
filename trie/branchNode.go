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
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/trieNodeData"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ = node(&branchNode{})

func newBranchNode() *branchNode {
	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: CollapsedBn{
			ChildrenHashes: encChildren,
		},
		children: children,
		dirty:    true,
	}
}

func (bn *branchNode) isDirty() bool {
	return bn.dirty
}

func (bn *branchNode) setDirty(dirty bool) {
	bn.dirty = dirty
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
	pathKey common.KeyBuilder,
	maxTrieLevelInMemory uint,
	goRoutinesManager common.TrieGoroutinesManager,
	hashesCollector common.TrieHashesCollector,
	trieCtx common.TrieContext,
) {
	level++

	if !bn.dirty {
		return
	}

	waitGroup := sync.WaitGroup{}
	nodeMutexKey := getMutexKeyFromPath(pathKey)

	for i := 0; i < nrOfChildren; i++ {
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		trieCtx.RLock(nodeMutexKey)
		child := bn.children[i]
		childHash := bn.ChildrenHashes[i]
		trieCtx.RUnlock(nodeMutexKey)

		if child == nil {
			continue
		}

		if len(childHash) == 0 {
			goRoutinesManager.SetError(ErrInvalidNodeState)
			return
		}

		childPathKey := getChildPathKey(pathKey, []byte{byte(i)})
		if !goRoutinesManager.CanStartGoRoutine() {
			child.commitDirty(level, childPathKey, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, trieCtx)
			if !goRoutinesManager.ShouldContinueProcessing() {
				return
			}
			continue
		}

		waitGroup.Add(1)
		go func(childPos int, childKey common.KeyBuilder) {
			defer func() {
				goRoutinesManager.EndGoRoutineProcessing()
				waitGroup.Done()
			}()

			child.commitDirty(level, childKey, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, trieCtx)
			if !goRoutinesManager.ShouldContinueProcessing() {
				return
			}
		}(i, childPathKey)
	}

	waitGroup.Wait()

	ok := saveDirtyNodeToStorage(bn, goRoutinesManager, hashesCollector, trieCtx)
	if !ok {
		return
	}

	if uint(level) == maxTrieLevelInMemory {
		log.Trace("collapse branch node on commit")

		for i := 0; i < nrOfChildren; i++ {
			trieCtx.Lock(nodeMutexKey)
			bn.children[i] = nil
			trieCtx.Unlock(nodeMutexKey)
		}
	}
}

func (bn *branchNode) commitSnapshot(
	trieCtx common.TrieContext,
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

		child, childBytes, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], trieCtx)
		childIsMissing, err := treatCommitSnapshotError(err, bn.ChildrenHashes[i], missingNodesChan)
		if err != nil {
			return err
		}
		if childIsMissing {
			continue
		}

		err = child.commitSnapshot(trieCtx, leavesChan, missingNodesChan, ctx, stats, idleProvider, childBytes, depthLevel+1)
		if err != nil {
			return err
		}

		err = trieCtx.Put(bn.ChildrenHashes[i], childBytes)
		if err != nil {
			return err
		}
	}

	stats.AddBranchNode(depthLevel, uint64(len(nodeBytes)))
	return nil
}

func (bn *branchNode) getEncodedNode(trieCtx common.TrieContext) ([]byte, error) {
	marshaledNode, err := trieCtx.Marshal(bn)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, branch)
	return marshaledNode, nil
}

func (bn *branchNode) resolveIfCollapsed(pos byte, trieCtx common.TrieContext) (node, []byte, error) {
	if childPosOutOfRange(pos) {
		return nil, nil, ErrChildPosOutOfRange
	}

	isPosCollapsed := check.IfNil(bn.children[pos]) && len(bn.ChildrenHashes[pos]) != 0
	if isPosCollapsed {
		child, _, err := getNodeFromDBAndDecode(bn.ChildrenHashes[pos], trieCtx)
		if err != nil {
			return nil, nil, err
		}
		bn.children[pos] = child
		return child, bn.ChildrenHashes[pos], nil
	}

	if !check.IfNil(bn.children[pos]) && len(bn.ChildrenHashes[pos]) == 0 {
		return nil, nil, ErrInvalidNodeState
	}

	handleStorageInteractorStats(trieCtx.GetStorage())
	return bn.children[pos], bn.ChildrenHashes[pos], nil
}

func (bn *branchNode) tryGet(keyData *keyData, currentDepth uint32, trieCtx common.TrieContext) (value []byte, maxDepth uint32, err error) {
	if len(keyData.keyRemainder) == 0 {
		return nil, currentDepth, nil
	}
	childPos := keyData.keyRemainder[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, currentDepth, ErrChildPosOutOfRange
	}
	keyData.keyRemainder = keyData.keyRemainder[1:]

	mutexKey := getMutexKeyFromBytes(keyData.pathKey)
	trieCtx.Lock(mutexKey)
	child, _, err := bn.resolveIfCollapsed(childPos, trieCtx)
	trieCtx.Unlock(mutexKey)
	if err != nil {
		return nil, currentDepth, err
	}
	if check.IfNil(child) {
		return nil, currentDepth, nil
	}

	keyData.pathKey = append(keyData.pathKey, childPos)
	return child.tryGet(keyData, currentDepth+1, trieCtx)
}

func (bn *branchNode) getNext(key []byte, trieCtx common.TrieContext) (*nodeData, error) {
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
	childNode, encodedNode, err := getNodeFromDBAndDecode(bn.ChildrenHashes[childPos], trieCtx)
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
	pathKey common.KeyBuilder,
	newData []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	trieCtx common.TrieContext,
) node {
	dataForInsertion, err := splitDataForChildren(newData)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	bnHasBeenModified := &atomic.Flag{}
	bn.updateNode(pathKey, dataForInsertion, goRoutinesManager, modifiedHashes, bnHasBeenModified, trieCtx, bn.insertOnChild)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return nil
	}

	if !bnHasBeenModified.IsSet() {
		return nil
	}

	return bn
}

func (bn *branchNode) updateNode(
	pathKey common.KeyBuilder,
	data [][]core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	hasBeenModified *atomic.Flag,
	trieCtx common.TrieContext,
	updateFunc func([]core.TrieData, int, common.KeyBuilder, common.TrieGoroutinesManager, common.AtomicBytesSlice, *atomic.Flag, common.TrieContext),
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
			updateFunc(data[childPos], childPos, pathKey, goRoutinesManager, modifiedHashes, hasBeenModified, trieCtx)

			continue
		}

		waitGroup.Add(1)
		go func(childPos int) {
			updateFunc(data[childPos], childPos, pathKey, goRoutinesManager, modifiedHashes, hasBeenModified, trieCtx)

			goRoutinesManager.EndGoRoutineProcessing()
			waitGroup.Done()
		}(childPos)
	}

	waitGroup.Wait()
}

func (bn *branchNode) insertOnChild(
	dataForInsertion []core.TrieData,
	childPos int,
	pathKey common.KeyBuilder,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	bnHasBeenModified *atomic.Flag,
	trieCtx common.TrieContext,
) {
	mutexKey := getMutexKeyFromPath(pathKey)
	trieCtx.Lock(mutexKey)
	child, childHash, err := bn.resolveIfCollapsed(byte(childPos), trieCtx)
	trieCtx.Unlock(mutexKey)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	if child == nil {
		bn.insertOnNilChild(pathKey, dataForInsertion, byte(childPos), goRoutinesManager, modifiedHashes, bnHasBeenModified, trieCtx)
		return
	}

	var originalChildHash []byte
	if !child.isDirty() {
		originalChildHash = childHash
	}

	childKeyPath := getChildPathKey(pathKey, []byte{byte(childPos)})
	newNode := child.insert(childKeyPath, dataForInsertion, goRoutinesManager, modifiedHashes, trieCtx)
	if check.IfNil(newNode) {
		return
	}

	if len(originalChildHash) != 0 {
		modifiedHashes.Append([][]byte{originalChildHash})
	}

	err = bn.modifyNodeAfterInsert(mutexKey, byte(childPos), newNode, trieCtx)
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
	pathKey common.KeyBuilder,
	newData []core.TrieData,
	childPos byte,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	bnHasBeenModified *atomic.Flag,
	trieCtx common.TrieContext,
) {
	if len(newData) == 0 {
		goRoutinesManager.SetError(ErrValueTooShort)
		return
	}

	var newNode node
	newNode = newLeafNode(newData[0])

	if len(newData) > 1 {
		childPathKey := getChildPathKey(pathKey, []byte{childPos})
		newNode = newNode.insert(childPathKey, newData[1:], goRoutinesManager, modifiedHashes, trieCtx)
		if check.IfNil(newNode) {
			return
		}
	}

	err := bn.modifyNodeAfterInsert(getMutexKeyFromPath(pathKey), childPos, newNode, trieCtx)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}

	bnHasBeenModified.SetValue(true)
}

func (bn *branchNode) modifyNodeAfterInsert(
	mutexKey string,
	childPos byte,
	newNode node,
	trieCtx common.TrieContext,
) error {
	trieCtx.Lock(mutexKey)
	defer trieCtx.Unlock(mutexKey)
	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}

	newHashForChild, err := encodeNodeAndGetHash(newNode, trieCtx)
	if err != nil {
		return err
	}

	bn.ChildrenHashes[childPos] = newHashForChild
	bn.children[childPos] = newNode

	bn.setVersionForChild(childVersion, childPos)
	bn.dirty = true
	return nil
}

func (bn *branchNode) delete(
	pathKey common.KeyBuilder,
	data []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	trieCtx common.TrieContext,
) (bool, node) {
	dataForRemoval, err := splitDataForChildren(data)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	hasBeenModified := &atomic.Flag{}

	bn.updateNode(pathKey, dataForRemoval, goRoutinesManager, modifiedHashes, hasBeenModified, trieCtx, bn.deleteChild)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return false, nil
	}

	if !hasBeenModified.IsSet() {
		return false, bn
	}

	return bn.reduceNodeIfNeeded(pathKey, goRoutinesManager, modifiedHashes, trieCtx)
}

func (bn *branchNode) reduceNodeIfNeeded(
	pathKey common.KeyBuilder,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	trieCtx common.TrieContext,
) (bool, node) {
	mutexKey := getMutexKeyFromPath(pathKey)
	trieCtx.Lock(mutexKey)
	defer trieCtx.Unlock(mutexKey)

	numChildren, pos := getChildPosition(bn)
	if numChildren == 0 {
		return true, nil
	}

	if numChildren != 1 {
		return true, bn
	}

	child, originalChildHash, err := bn.resolveIfCollapsed(byte(pos), trieCtx)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	childPathKey := getChildPathKey(pathKey, []byte{byte(pos)})
	var newChildHash bool
	newNode, newChildHash, err := child.reduceNode(pos, getMutexKeyFromPath(childPathKey), trieCtx)
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
	pathKey common.KeyBuilder,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	hasBeenModified *atomic.Flag,
	trieCtx common.TrieContext,
) {
	mutexKey := getMutexKeyFromPath(pathKey)
	trieCtx.Lock(mutexKey)
	child, childHash, err := bn.resolveIfCollapsed(byte(childPos), trieCtx)
	trieCtx.Unlock(mutexKey)
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

	childPathKey := getChildPathKey(pathKey, []byte{byte(childPos)})
	dirty, newNode := child.delete(childPathKey, dataForRemoval, goRoutinesManager, modifiedHashes, trieCtx)
	if !goRoutinesManager.ShouldContinueProcessing() || !dirty {
		return
	}

	err = bn.setNewChild(mutexKey, byte(childPos), newNode, trieCtx)
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
	mutexKey string,
	childPos byte,
	newNode node,
	trieCtx common.TrieContext,
) error {
	trieCtx.Lock(mutexKey)
	defer trieCtx.Unlock(mutexKey)

	bn.dirty = true

	if check.IfNil(newNode) {
		bn.children[childPos] = nil
		bn.ChildrenHashes[childPos] = nil
		bn.setVersionForChild(core.NotSpecified, childPos)
		return nil
	}
	childVersion, err := newNode.getVersion()
	if err != nil {
		return err
	}

	newChildHash, err := encodeNodeAndGetHash(newNode, trieCtx)
	if err != nil {
		return err
	}

	bn.children[childPos] = newNode
	bn.ChildrenHashes[childPos] = newChildHash
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

func (bn *branchNode) reduceNode(pos int, _ string, trieCtx common.TrieContext) (node, bool, error) {
	newEn, err := newExtensionNode([]byte{byte(pos)}, bn)
	if err != nil {
		return nil, false, err
	}
	hash, err := encodeNodeAndGetHash(bn, trieCtx)
	if err != nil {
		return nil, false, err
	}
	newEn.ChildHash = hash

	return newEn, false, nil
}

func getChildPosition(n *branchNode) (numChildren int, childPos int) {
	for i := 0; i < nrOfChildren; i++ {
		if n.children[i] != nil || len(n.ChildrenHashes[i]) != 0 {
			numChildren++
			childPos = i
		}
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

func (bn *branchNode) print(writer io.Writer, index int, trieCtx common.TrieContext) {
	if bn == nil {
		return
	}

	str := fmt.Sprintf("B: %v", bn.dirty)
	_, _ = fmt.Fprintln(writer, str)
	for i := 0; i < len(bn.children); i++ {
		_, _, err := bn.resolveIfCollapsed(byte(i), trieCtx)
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
		str2 := fmt.Sprintf("+ %d: - hash: %v ", i, hex.EncodeToString(bn.ChildrenHashes[i]))
		_, _ = fmt.Fprint(writer, str2)
		childIndex := index + len(str) - 1 + len(str2)
		child.print(writer, childIndex, trieCtx)
	}
}

func (bn *branchNode) getChildren(trieCtx common.TrieContext) ([]nodeWithHash, error) {
	nextNodes := make([]nodeWithHash, 0)

	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		childNode, _, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], trieCtx)
		if err != nil {
			return nil, err
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
	chanClose chan struct{},
	ctx context.Context,
	trieCtx common.TrieContext,
) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	for i := range bn.ChildrenHashes {
		select {
		case <-chanClose:
			log.Trace("branchNode.getAllLeavesOnChannel interrupted")
			return nil
		case <-ctx.Done():
			log.Trace("branchNode.getAllLeavesOnChannel context done")
			return nil
		default:
			if len(bn.ChildrenHashes[i]) == 0 {
				continue
			}

			child, _, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], trieCtx)
			if err != nil {
				return err
			}

			clonedKeyBuilder := keyBuilder.ShallowClone()
			clonedKeyBuilder.BuildKey([]byte{byte(i)})
			err = child.getAllLeavesOnChannel(
				leavesChannel,
				clonedKeyBuilder,
				trieLeafParser,
				chanClose,
				ctx,
				trieCtx,
			)
			if err != nil {
				return err
			}
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

	nodeSize := dirtyFlagSizeInBytes
	for _, collapsed := range bn.ChildrenHashes {
		nodeSize += len(collapsed)
	}
	nodeSize += len(bn.children) * pointerSizeInBytes
	nodeSize += len(bn.ChildrenVersion)

	return nodeSize
}

func (bn *branchNode) getValue() []byte {
	return []byte{}
}

func (bn *branchNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, trieCtx common.TrieContext) error {
	for i := range bn.ChildrenHashes {
		if len(bn.ChildrenHashes[i]) == 0 {
			continue
		}

		child, childBytes, err := getNodeFromDBAndDecode(bn.ChildrenHashes[i], trieCtx)
		if err != nil {
			return err
		}

		err = child.collectStats(ts, depthLevel+1, uint64(len(childBytes)), trieCtx)
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
	keyBuilder common.KeyBuilder,
	trieCtx common.TrieContext,
) (bool, error) {
	shouldContinue := migrationArgs.TrieMigrator.ConsumeStorageLoadGas()
	if !shouldContinue {
		return false, nil
	}

	mutexKey := getMutexKeyFromPath(keyBuilder)
	trieCtx.Lock(mutexKey)
	defer trieCtx.Unlock(mutexKey)

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

		childNode, _, err := bn.resolveIfCollapsed(byte(i), trieCtx)
		if err != nil {
			return false, err
		}

		clonedKeyBuilder := keyBuilder.ShallowClone()
		clonedKeyBuilder.BuildKey([]byte{byte(i)})
		shouldContinueMigrating, err := childNode.collectLeavesForMigration(migrationArgs, clonedKeyBuilder, trieCtx)
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
