package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ = node(&extensionNode{})

func newExtensionNode(key []byte, child node, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*extensionNode, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(child) {
		return nil, ErrNilNode
	}

	childVersion, err := child.getVersion()
	if err != nil {
		return nil, err
	}

	return &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:          key,
			EncodedChild: nil,
			ChildVersion: uint32(childVersion),
		},
		child: child,
		baseNode: &baseNode{
			dirty:  true,
			marsh:  marshalizer,
			hasher: hasher,
		},
	}, nil
}

func (en *extensionNode) getCollapsed() (node, error) {
	return en.getCollapsedEn()
}

func (en *extensionNode) getCollapsedEn() (*extensionNode, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getCollapsed error %w", err)
	}
	if en.isCollapsed() {
		return en, nil
	}
	if !hasValidHash(en.child) {
		return nil, ErrNodeHashIsNotSet
	}
	en.EncodedChild = en.child.getHash()
	en.child = nil
	return en, nil
}

func (en *extensionNode) setHash(goRoutinesManager common.TrieGoroutinesManager) {
	if len(en.hash) != 0 {
		return
	}

	if !goRoutinesManager.ShouldContinueProcessing() {
		return
	}

	if en.shouldSetHashForChild() {
		en.child.setHash(goRoutinesManager)
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		encChild, err := encodeNodeAndGetHash(en.child)
		if err != nil {
			goRoutinesManager.SetError(err)
			return
		}
		en.EncodedChild = encChild
	}

	hash, err := encodeNodeAndGetHash(en)
	if err != nil {
		goRoutinesManager.SetError(err)
		return
	}
	en.hash = hash
}

func (en *extensionNode) shouldSetHashForChild() bool {
	en.childMutex.RLock()
	defer en.childMutex.RUnlock()

	if en.child != nil && en.EncodedChild == nil {
		return true
	}

	return false
}

func (en *extensionNode) hashNode() ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("hashNode error %w", err)
	}
	if en.child != nil {
		var encChild []byte
		encChild, err = encodeNodeAndGetHash(en.child)
		if err != nil {
			return nil, err
		}
		en.EncodedChild = encChild
	}
	return encodeNodeAndGetHash(en)
}

func (en *extensionNode) commitDirty(
	level byte,
	maxTrieLevelInMemory uint,
	goRoutinesManager common.TrieGoroutinesManager,
	hashesCollector common.TrieHashesCollector,
	originDb common.TrieStorageInteractor,
	targetDb common.BaseStorer,
) {
	level++

	if !en.dirty {
		return
	}

	if !goRoutinesManager.ShouldContinueProcessing() {
		return
	}

	en.childMutex.RLock()
	child := en.child
	en.childMutex.RUnlock()

	if child != nil {
		child.commitDirty(level, maxTrieLevelInMemory, goRoutinesManager, hashesCollector, originDb, targetDb)
		if !goRoutinesManager.ShouldContinueProcessing() {
			return
		}

		en.EncodedChild = child.getHash()
	}

	ok := saveDirtyNodeToStorage(en, goRoutinesManager, hashesCollector, targetDb, en.hasher)
	if !ok {
		return
	}

	if uint(level) == maxTrieLevelInMemory {
		log.Trace("collapse extension node on commit")

		en.childMutex.Lock()
		en.child = nil
		en.childMutex.Unlock()
	}
}

func (en *extensionNode) commitSnapshot(
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

	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit snapshot error %w", err)
	}

	_, err = en.resolveIfCollapsed(db)
	childIsMissing, err := treatCommitSnapshotError(err, en.EncodedChild, missingNodesChan)
	if err != nil {
		return err
	}

	if !childIsMissing {
		err = en.child.commitSnapshot(db, leavesChan, missingNodesChan, ctx, stats, idleProvider, depthLevel+1)
		if err != nil {
			return err
		}
	}

	return en.saveToStorage(db, stats, depthLevel)
}

func (en *extensionNode) saveToStorage(targetDb common.BaseStorer, stats common.TrieStatisticsHandler, depthLevel int) error {
	nodeSize, err := encodeNodeAndCommitToDB(en, targetDb)
	if err != nil {
		return err
	}

	stats.AddExtensionNode(depthLevel, uint64(nodeSize))

	en.child = nil
	return nil
}

func (en *extensionNode) getEncodedNode() ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getEncodedNode error %w", err)
	}
	marshaledNode, err := en.marsh.Marshal(en)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, extension)
	return marshaledNode, nil
}

func (en *extensionNode) isCollapsed() bool {
	return en.child == nil && len(en.EncodedChild) != 0
}

func (en *extensionNode) resolveIfCollapsed(db common.TrieStorageInteractor) (node, error) {
	en.childMutex.Lock()
	defer en.childMutex.Unlock()

	if !en.isCollapsed() {
		handleStorageInteractorStats(db)
		return en.child, nil
	}

	child, err := getNodeFromDBAndDecode(en.EncodedChild, db, en.marsh, en.hasher)
	if err != nil {
		return nil, err
	}
	en.child = child
	return child, nil
}

func (en *extensionNode) getChild(key []byte, db common.TrieStorageInteractor) (node, []byte, error) {
	en.mutex.RLock()
	defer en.mutex.RUnlock()

	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, nil, nil
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, nil, nil
	}
	key = key[len(en.Key):]
	child, err := en.resolveIfCollapsed(db)
	if err != nil {
		return nil, nil, err
	}

	return child, key, nil
}

func (en *extensionNode) tryGet(key []byte, currentDepth uint32, db common.TrieStorageInteractor) (value []byte, maxDepth uint32, err error) {
	child, key, err := en.getChild(key, db)
	if err != nil {
		return nil, currentDepth, err
	}
	if check.IfNil(child) {
		return nil, currentDepth, nil
	}

	return child.tryGet(key, currentDepth+1, db)
}

func (en *extensionNode) getNext(key []byte, db common.TrieStorageInteractor) (node, []byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("getNext error %w", err)
	}
	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, nil, ErrNodeNotFound
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, nil, ErrNodeNotFound
	}
	childNode, err := en.resolveIfCollapsed(db)
	if err != nil {
		return nil, nil, err
	}

	key = key[len(en.Key):]
	return childNode, key, nil
}

func (en *extensionNode) insert(
	newData []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) node {
	childNode, err := en.resolveIfCollapsed(db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	keyMatchLen, index := getMinKeyMatchLen(newData, en.Key)

	// If the whole key matches, keep this extension node as is
	// and only update the value.
	if keyMatchLen == len(en.Key) {
		return en.insertAtSameKey(newData, childNode, keyMatchLen, goRoutinesManager, modifiedHashes, db)
	}

	// Otherwise branch out at the index where they differ.
	return en.insertInNewBn(newData, childNode, goRoutinesManager, modifiedHashes, db, keyMatchLen, index)
}

func getMinKeyMatchLen(newData []core.TrieData, enKey []byte) (int, int) {
	minKeyMatchLen := len(enKey)
	index := 0
	for i, data := range newData {
		if minKeyMatchLen == 0 {
			return 0, index
		}
		matchLen := prefixLen(data.Key, enKey)
		if matchLen < minKeyMatchLen {
			minKeyMatchLen = matchLen
			index = i
		}
	}

	return minKeyMatchLen, index
}

func removeCommonPrefix(newData []core.TrieData, prefixLen int) error {
	for i := range newData {
		if len(newData[i].Key) < prefixLen {
			return ErrValueTooShort
		}
		newData[i].Key = newData[i].Key[prefixLen:]
	}

	return nil
}

func (en *extensionNode) insertAtSameKey(
	newData []core.TrieData,
	child node,
	keyMatchLen int,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) node {
	for i := range newData {
		newData[i].Key = newData[i].Key[keyMatchLen:]
	}
	newNode := child.insert(newData, goRoutinesManager, modifiedHashes, db)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return newNode
	}

	if check.IfNil(newNode) {
		return nil
	}

	if !en.dirty {
		modifiedHashes.Append([][]byte{en.hash})
	}

	newEn, err := newExtensionNode(en.Key, newNode, en.marsh, en.hasher)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	return newEn
}

func (en *extensionNode) insertInNewBn(
	newData []core.TrieData,
	childNode node,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
	keyMatchLen int,
	index int,
) node {
	if !en.dirty {
		modifiedHashes.Append([][]byte{en.hash})
	}

	bn, err := newBranchNode(en.marsh, en.hasher)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	oldChildPos := en.Key[keyMatchLen]
	newChildPos := newData[index].Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		goRoutinesManager.SetError(ErrChildPosOutOfRange)
		return nil
	}

	err = en.insertOldChildInBn(bn, childNode, oldChildPos, keyMatchLen)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	newChild := newData[index]
	newData = append(newData[:index], newData[index+1:]...)

	err = en.insertNewChildInBn(bn, newChild, newChildPos, keyMatchLen)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	err = removeCommonPrefix(newData, keyMatchLen)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	var newNode node
	newNode = bn
	if len(newData) != 0 {
		newNode = bn.insert(newData, goRoutinesManager, modifiedHashes, db)
		if !goRoutinesManager.ShouldContinueProcessing() {
			return nil
		}
	}

	if keyMatchLen == 0 {
		return newNode
	}

	newEn, err := newExtensionNode(en.Key[:keyMatchLen], newNode, en.marsh, en.hasher)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	return newEn
}

func (en *extensionNode) insertOldChildInBn(bn *branchNode, childNode node, oldChildPos byte, keyMatchLen int) error {
	keyReminder := en.Key[keyMatchLen+1:]
	childVersion, err := childNode.getVersion()
	if err != nil {
		return err
	}
	bn.setVersionForChild(childVersion, oldChildPos)

	if len(keyReminder) < 1 {
		bn.children[oldChildPos] = en.child
		return nil
	}

	followingExtensionNode, err := newExtensionNode(en.Key[keyMatchLen+1:], childNode, en.marsh, en.hasher)
	if err != nil {
		return err
	}

	bn.children[oldChildPos] = followingExtensionNode
	return nil
}

func (en *extensionNode) insertNewChildInBn(bn *branchNode, newChild core.TrieData, newChildPos byte, keyMatchLen int) error {
	newChild.Key = newChild.Key[keyMatchLen+1:]

	newLeaf, err := newLeafNode(newChild, en.marsh, en.hasher)
	if err != nil {
		return err
	}

	bn.children[newChildPos] = newLeaf
	bn.setVersionForChild(newChild.Version, newChildPos)
	return nil
}

func (en *extensionNode) getDataWithMatchingPrefix(data []core.TrieData) []core.TrieData {
	dataWithMatchingKey := make([]core.TrieData, 0)
	for _, d := range data {
		if len(en.Key) == prefixLen(d.Key, en.Key) {
			d.Key = d.Key[len(en.Key):]
			dataWithMatchingKey = append(dataWithMatchingKey, d)
		}
	}

	return dataWithMatchingKey
}

func (en *extensionNode) delete(
	data []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	db common.TrieStorageInteractor,
) (bool, node) {
	dataWithMatchingKey := en.getDataWithMatchingPrefix(data)
	if len(dataWithMatchingKey) == 0 {
		return false, en
	}
	childNode, err := en.resolveIfCollapsed(db)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false, nil
	}

	dirty, newNode := childNode.delete(dataWithMatchingKey, goRoutinesManager, modifiedHashes, db)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return false, nil
	}
	if !dirty {
		return false, en
	}

	if !en.dirty {
		modifiedHashes.Append([][]byte{en.hash})
	}

	switch newNode := newNode.(type) {
	case *leafNode:
		newLeafData := core.TrieData{
			Key:     concat(en.Key, newNode.Key...),
			Value:   newNode.Value,
			Version: core.TrieNodeVersion(newNode.Version),
		}
		n, err := newLeafNode(newLeafData, en.marsh, en.hasher)
		if err != nil {
			goRoutinesManager.SetError(err)
			return false, nil
		}

		return true, n
	case *extensionNode:
		n, err := newExtensionNode(concat(en.Key, newNode.Key...), newNode.child, en.marsh, en.hasher)
		if err != nil {
			goRoutinesManager.SetError(err)
			return false, nil
		}

		return true, n
	case *branchNode:
		n, err := newExtensionNode(en.Key, newNode, en.marsh, en.hasher)
		if err != nil {
			goRoutinesManager.SetError(err)
			return false, nil
		}

		return true, n
	case nil:
		return true, nil
	default:
		goRoutinesManager.SetError(ErrInvalidNode)
		return false, nil
	}
}

func (en *extensionNode) reduceNode(pos int, db common.TrieStorageInteractor) (node, bool, error) {
	k := append([]byte{byte(pos)}, en.Key...)

	child, err := en.resolveIfCollapsed(db)
	if err != nil {
		return nil, false, err
	}

	newEn, err := newExtensionNode(k, child, en.marsh, en.hasher)
	if err != nil {
		return nil, false, err
	}

	return newEn, true, nil
}

func (en *extensionNode) isEmptyOrNil() error {
	if en == nil {
		return ErrNilExtensionNode
	}

	en.childMutex.RLock()
	defer en.childMutex.RUnlock()
	if en.child == nil && len(en.EncodedChild) == 0 {
		return ErrEmptyExtensionNode
	}
	return nil
}

func (en *extensionNode) print(writer io.Writer, index int, db common.TrieStorageInteractor) {
	if en == nil {
		return
	}

	_, err := en.resolveIfCollapsed(db)
	if err != nil {
		log.Debug("extension node: print trie err", "error", err, "hash", en.EncodedChild)
	}

	key := ""
	for _, k := range en.Key {
		key += fmt.Sprintf("%d", k)
	}

	str := fmt.Sprintf("E: key= %v, (%v) - %v", en.Key, hex.EncodeToString(en.hash), en.dirty)
	_, _ = fmt.Fprint(writer, str)

	if en.child == nil {
		return
	}
	en.child.print(writer, index+len(str), db)
}

func (en *extensionNode) getChildren(db common.TrieStorageInteractor) ([]node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getChildren error %w", err)
	}

	nextNodes := make([]node, 0)

	childNode, err := en.resolveIfCollapsed(db)
	if err != nil {
		return nil, err
	}

	nextNodes = append(nextNodes, childNode)

	return nextNodes, nil
}

func (en *extensionNode) loadChildren(getNode func([]byte) (node, error)) ([][]byte, []node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("loadChildren error %w", err)
	}

	if en.EncodedChild == nil {
		return nil, nil, ErrNilExtensionNode
	}

	child, err := getNode(en.EncodedChild)
	if err != nil {
		return [][]byte{en.EncodedChild}, nil, nil
	}
	log.Trace("load extension node child", "child hash", en.EncodedChild)
	en.child = child

	return nil, []node{child}, nil
}

func (en *extensionNode) getAllLeavesOnChannel(
	leavesChannel chan core.KeyValueHolder,
	keyBuilder common.KeyBuilder,
	trieLeafParser common.TrieLeafParser,
	db common.TrieStorageInteractor,
	marshalizer marshal.Marshalizer,
	chanClose chan struct{},
	ctx context.Context,
) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	select {
	case <-chanClose:
		log.Trace("extensionNode.getAllLeavesOnChannel interrupted")
		return nil
	case <-ctx.Done():
		log.Trace("extensionNode.getAllLeavesOnChannel: context done")
		return nil
	default:
		_, err = en.resolveIfCollapsed(db)
		if err != nil {
			return err
		}

		keyBuilder.BuildKey(en.Key)
		err = en.child.getAllLeavesOnChannel(leavesChannel, keyBuilder.Clone(), trieLeafParser, db, marshalizer, chanClose, ctx)
		if err != nil {
			return err
		}

		en.child = nil
	}

	return nil
}

func (en *extensionNode) getNextHashAndKey(key []byte) (bool, []byte, []byte) {
	if len(key) == 0 || check.IfNil(en) {
		return false, nil, nil
	}

	nextKey := key[len(en.Key):]
	wantHash := en.EncodedChild

	return false, wantHash, nextKey
}

func (en *extensionNode) sizeInBytes() int {
	if en == nil {
		return 0
	}

	// hasher + marshalizer + child + dirty flag = 3 * pointerSizeInBytes + 1
	nodeSize := len(en.hash) + len(en.Key) + (numNodeInnerPointers+1)*pointerSizeInBytes + 1
	nodeSize += len(en.EncodedChild)

	return nodeSize
}

func (en *extensionNode) getValue() []byte {
	return []byte{}
}

func (en *extensionNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, db common.TrieStorageInteractor) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("collectStats error %w", err)
	}

	_, err = en.resolveIfCollapsed(db)
	if err != nil {
		return err
	}

	err = en.child.collectStats(ts, depthLevel+1, db)
	if err != nil {
		return err
	}

	val, err := collapseAndEncodeNode(en)
	if err != nil {
		return err
	}

	ts.AddExtensionNode(depthLevel, uint64(len(val)))
	return nil
}

func (en *extensionNode) getVersion() (core.TrieNodeVersion, error) {
	if en.ChildVersion > math.MaxUint8 {
		log.Warn("invalid trie node version for extension node", "child version", en.ChildVersion, "max version", math.MaxUint8)
		return core.NotSpecified, ErrInvalidNodeVersion
	}

	return core.TrieNodeVersion(en.ChildVersion), nil
}

func (en *extensionNode) collectLeavesForMigration(
	migrationArgs vmcommon.ArgsMigrateDataTrieLeaves,
	db common.TrieStorageInteractor,
	keyBuilder common.KeyBuilder,
) (bool, error) {
	hasEnoughGasToContinueMigration := migrationArgs.TrieMigrator.ConsumeStorageLoadGas()
	if !hasEnoughGasToContinueMigration {
		return false, nil
	}

	shouldMigrateNode, err := shouldMigrateCurrentNode(en, migrationArgs)
	if err != nil {
		return false, err
	}
	if !shouldMigrateNode {
		return true, nil
	}

	childNode, err := en.resolveIfCollapsed(db)
	if err != nil {
		return false, err
	}

	keyBuilder.BuildKey(en.Key)
	return childNode.collectLeavesForMigration(migrationArgs, db, keyBuilder.Clone())
}

// IsInterfaceNil returns true if there is no value under the interface
func (en *extensionNode) IsInterfaceNil() bool {
	return en == nil
}
