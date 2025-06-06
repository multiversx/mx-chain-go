package trie

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/trieNodeData"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ = node(&leafNode{})

func newLeafNode(
	newData core.TrieData,
) *leafNode {
	return &leafNode{
		CollapsedLn: CollapsedLn{
			Key:     newData.Key,
			Value:   newData.Value,
			Version: uint32(newData.Version),
		},
		baseNode: &baseNode{
			dirty: true,
		},
	}
}

func (ln *leafNode) commitDirty(
	_ byte,
	_ uint,
	goRoutinesManager common.TrieGoroutinesManager,
	hashesCollector common.TrieHashesCollector,
	trieCtx common.TrieContext,
) {
	if !ln.dirty {
		return
	}

	saveDirtyNodeToStorage(ln, goRoutinesManager, hashesCollector, trieCtx)
}

func (ln *leafNode) commitSnapshot(
	trieCtx common.TrieContext,
	leavesChan chan core.KeyValueHolder,
	_ chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	nodeBytes []byte,
	depthLevel int,
) error {
	if shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider) {
		return core.ErrContextClosing
	}

	err := writeNodeOnChannel(ln, leavesChan, trieCtx)
	if err != nil {
		return err
	}

	version, err := ln.getVersion()
	if err != nil {
		return err
	}

	stats.AddLeafNode(depthLevel, uint64(len(nodeBytes)), version)
	return nil
}

func writeNodeOnChannel(ln *leafNode, leavesChan chan core.KeyValueHolder, trieCtx common.TrieContext) error {
	if leavesChan == nil {
		return nil
	}

	hash, err := encodeNodeAndGetHash(ln, trieCtx)
	if err != nil {
		return err
	}

	trieLeaf := keyValStorage.NewKeyValStorage(hash, ln.Value)
	leavesChan <- trieLeaf

	return nil
}

func (ln *leafNode) getEncodedNode(trieCtx common.TrieContext) ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getEncodedNode error %w", err)
	}
	marshaledNode, err := trieCtx.Marshal(ln)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, leaf)
	return marshaledNode, nil
}

func (ln *leafNode) tryGet(key []byte, currentDepth uint32, _ common.TrieContext) (value []byte, maxDepth uint32, err error) {
	ln.mutex.RLock()
	defer ln.mutex.RUnlock()

	if bytes.Equal(key, ln.Key) {
		return ln.Value, currentDepth, nil
	}

	return nil, currentDepth, nil
}

func (ln *leafNode) getNext(key []byte, _ common.TrieContext) (*nodeData, error) {
	if bytes.Equal(key, ln.Key) {
		return nil, nil
	}
	return nil, ErrNodeNotFound
}

func (ln *leafNode) insert(
	newData []core.TrieData,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	trieCtx common.TrieContext,
) node {
	if len(newData) == 1 && bytes.Equal(newData[0].Key, ln.Key) {
		return ln.insertInSameLn(newData[0])
	}

	keyMatchLen, _ := getMinKeyMatchLen(newData, ln.Key)
	bn := ln.insertInNewBn(newData, keyMatchLen, goRoutinesManager, modifiedHashes, trieCtx)
	if !goRoutinesManager.ShouldContinueProcessing() {
		return nil
	}

	if keyMatchLen == 0 {
		return bn
	}

	newEn, err := newExtensionNode(ln.Key[:keyMatchLen], bn)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}
	childHash, err := encodeNodeAndGetHash(bn, trieCtx)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}
	newEn.ChildHash = childHash

	return newEn
}

func (ln *leafNode) insertInSameLn(newData core.TrieData) node {
	if bytes.Equal(ln.Value, newData.Value) {
		return nil
	}

	ln.mutex.Lock()
	defer ln.mutex.Unlock()

	ln.Value = newData.Value
	ln.Version = uint32(newData.Version)
	ln.dirty = true
	return ln
}

func trimKeys(data []core.TrieData, keyMatchLen int) {
	for i := range data {
		data[i].Key = data[i].Key[keyMatchLen:]
	}
}

func (ln *leafNode) insertInNewBn(
	newData []core.TrieData,
	keyMatchLen int,
	goRoutinesManager common.TrieGoroutinesManager,
	modifiedHashes common.AtomicBytesSlice,
	trieCtx common.TrieContext,
) node {
	bn := newBranchNode()

	lnVersion, err := ln.getVersion()
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}

	var newKeyForOldLn []byte
	posForOldLn := byte(keyBuilder.HexTerminator)
	if len(ln.Key) > keyMatchLen {
		newKeyForOldLn = ln.Key[keyMatchLen+1:]
		posForOldLn = ln.Key[keyMatchLen]
	}

	lnData := core.TrieData{
		Key:     newKeyForOldLn,
		Value:   ln.Value,
		Version: lnVersion,
	}

	oldLn := newLeafNode(lnData)
	oldLnHash, err := encodeNodeAndGetHash(oldLn, trieCtx)
	if err != nil {
		goRoutinesManager.SetError(err)
		return nil
	}
	bn.children[posForOldLn] = oldLn
	bn.ChildrenHashes[posForOldLn] = oldLnHash
	bn.setVersionForChild(lnVersion, posForOldLn)

	trimKeys(newData, keyMatchLen)
	return bn.insert(newData, goRoutinesManager, modifiedHashes, trieCtx)
}

func (ln *leafNode) delete(
	data []core.TrieData,
	_ common.TrieGoroutinesManager,
	_ common.AtomicBytesSlice,
	_ common.TrieContext,
) (bool, node) {
	ln.mutex.RLock()
	defer ln.mutex.RUnlock()

	for _, d := range data {
		if bytes.Equal(d.Key, ln.Key) {
			return true, nil
		}
	}
	return false, ln
}

func (ln *leafNode) reduceNode(pos int, _ []byte, _ common.TrieContext) (node, bool, error) {
	k := append([]byte{byte(pos)}, ln.Key...)

	oldLnVersion, err := ln.getVersion()
	if err != nil {
		return nil, false, err
	}

	oldLnData := core.TrieData{
		Key:     k,
		Value:   ln.Value,
		Version: oldLnVersion,
	}

	return newLeafNode(oldLnData), true, nil
}

func (ln *leafNode) isEmptyOrNil() error {
	if ln == nil {
		return ErrNilLeafNode
	}
	if ln.Value == nil {
		return ErrEmptyLeafNode
	}
	return nil
}

func (ln *leafNode) print(writer io.Writer, _ int, _ common.TrieContext) {
	if ln == nil {
		return
	}

	key := ""
	for _, k := range ln.Key {
		key += fmt.Sprintf("%d", k)
	}

	val := ""
	for _, v := range ln.Value {
		val += fmt.Sprintf("%d", v)
	}

	_, _ = fmt.Fprintf(writer, "L: key= %v, %v\n", ln.Key, ln.dirty)
}

func (ln *leafNode) getChildren(_ common.TrieContext) ([]nodeWithHash, error) {
	return nil, nil
}

func (ln *leafNode) loadChildren(_ func([]byte) (node, error)) ([][]byte, []nodeWithHash, error) {
	return nil, nil, nil
}

func (ln *leafNode) getAllLeavesOnChannel(
	leavesChannel chan core.KeyValueHolder,
	keyBuilder common.KeyBuilder,
	trieLeafParser common.TrieLeafParser,
	chanClose chan struct{},
	ctx context.Context,
	_ common.TrieContext,
) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	keyBuilder.BuildKey(ln.Key)
	nodeKey, err := keyBuilder.GetKey()
	if err != nil {
		return err
	}

	version, err := ln.getVersion()
	if err != nil {
		return err
	}

	trieLeaf, err := trieLeafParser.ParseLeaf(nodeKey, ln.Value, version)
	if err != nil {
		return err
	}

	for {
		select {
		case <-chanClose:
			log.Trace("leafNode.getAllLeavesOnChannel interrupted")
			return nil
		case <-ctx.Done():
			log.Trace("leafNode.getAllLeavesOnChannel: context done")
			return nil
		case leavesChannel <- trieLeaf:
			return nil
		}
	}
}

func (ln *leafNode) getNextHashAndKey(key []byte) (bool, []byte, []byte) {
	if check.IfNil(ln) {
		return false, nil, nil
	}

	if bytes.Equal(key, ln.Key) {
		return true, nil, nil
	}

	return false, nil, nil
}

func (ln *leafNode) sizeInBytes() int {
	if ln == nil {
		return 0
	}

	nodeSize := len(ln.Key) + len(ln.Value) + versionSizeInBytes + dirtyFlagSizeInBytes + mutexSizeInBytes

	return nodeSize
}

func (ln *leafNode) getValue() []byte {
	return ln.Value
}

func (ln *leafNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, _ common.TrieContext) error {
	version, err := ln.getVersion()
	if err != nil {
		return err
	}

	ts.AddLeafNode(depthLevel, nodeSize, version)
	return nil
}

func (ln *leafNode) getVersion() (core.TrieNodeVersion, error) {
	if ln.Version > math.MaxUint8 {
		log.Warn("invalid trie node version", "version", ln.Version, "max version", math.MaxUint8)
		return core.NotSpecified, ErrInvalidNodeVersion
	}

	return core.TrieNodeVersion(ln.Version), nil
}

func (ln *leafNode) collectLeavesForMigration(
	migrationArgs vmcommon.ArgsMigrateDataTrieLeaves,
	keyBuilder common.KeyBuilder,
	_ common.TrieContext,
) (bool, error) {
	shouldContinue := migrationArgs.TrieMigrator.ConsumeStorageLoadGas()
	if !shouldContinue {
		return false, nil
	}

	shouldMigrateNode, err := shouldMigrateCurrentNode(ln, migrationArgs)
	if err != nil {
		return false, err
	}
	if !shouldMigrateNode {
		return true, nil
	}

	keyBuilder.BuildKey(ln.Key)
	key, err := keyBuilder.GetKey()
	if err != nil {
		return false, err
	}

	version, err := ln.getVersion()
	if err != nil {
		return false, err
	}

	leafData := core.TrieData{
		Key:     key,
		Value:   ln.Value,
		Version: version,
	}

	return migrationArgs.TrieMigrator.AddLeafToMigrationQueue(leafData, migrationArgs.NewVersion)
}

func (ln *leafNode) getNodeData(keyBuilder common.KeyBuilder) ([]common.TrieNodeData, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getNodeData error %w", err)
	}

	version, err := ln.getVersion()
	if err != nil {
		return nil, err
	}

	data := make([]common.TrieNodeData, 1)
	clonedKeyBuilder := keyBuilder.DeepClone()
	clonedKeyBuilder.BuildKey(ln.Key)
	nodeData, err := trieNodeData.NewLeafNodeData(clonedKeyBuilder, ln.Value, version)
	if err != nil {
		return nil, err
	}
	data[0] = nodeData

	return data, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ln *leafNode) IsInterfaceNil() bool {
	return ln == nil
}
