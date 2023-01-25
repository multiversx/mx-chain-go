package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
)

var _ = node(&extensionNode{})

func newExtensionNode(key []byte, child node, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*extensionNode, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:          key,
			EncodedChild: nil,
		},
		child: child,
		baseNode: &baseNode{
			dirty:  true,
			marsh:  marshalizer,
			hasher: hasher,
		},
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
	collapsed := en.clone()
	ok, err := hasValidHash(en.child)
	if err != nil {
		return nil, err
	}
	if !ok {
		err = en.child.setHash()
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
		return fmt.Errorf("setHash error %w", err)
	}
	if en.getHash() != nil {
		return nil
	}
	if en.isCollapsed() {
		var hash []byte
		hash, err = encodeNodeAndGetHash(en)
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
		return fmt.Errorf("hashChildren error %w", err)
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

func (en *extensionNode) commitDirty(level byte, maxTrieLevelInMemory uint, originDb common.DBWriteCacher, targetDb common.DBWriteCacher) error {
	level++
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit error %w", err)
	}

	if !en.dirty {
		return nil
	}

	if en.child != nil {
		err = en.child.commitDirty(level, maxTrieLevelInMemory, originDb, targetDb)
		if err != nil {
			return err
		}
	}

	en.dirty = false
	_, err = encodeNodeAndCommitToDB(en, targetDb)
	if err != nil {
		return err
	}
	if uint(level) == maxTrieLevelInMemory {
		log.Trace("collapse extension node on commit")

		var collapsedEn *extensionNode
		collapsedEn, err = en.getCollapsedEn()
		if err != nil {
			return err
		}

		*en = *collapsedEn
	}
	return nil
}

func (en *extensionNode) commitCheckpoint(
	originDb common.DBWriteCacher,
	targetDb common.DBWriteCacher,
	checkpointHashes CheckpointHashesHolder,
	leavesChan chan core.KeyValueHolder,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDone(ctx, idleProvider) {
		return errors.ErrContextClosing
	}

	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit checkpoint error %w", err)
	}

	err = resolveIfCollapsed(en, 0, originDb)
	if err != nil {
		return err
	}

	hash, err := computeAndSetNodeHash(en)
	if err != nil {
		return err
	}

	shouldCommit := checkpointHashes.ShouldCommit(hash)
	if !shouldCommit {
		return nil
	}

	err = en.child.commitCheckpoint(originDb, targetDb, checkpointHashes, leavesChan, ctx, stats, idleProvider, depthLevel+1)
	if err != nil {
		return err
	}

	checkpointHashes.Remove(hash)
	return en.saveToStorage(targetDb, stats, depthLevel)
}

func (en *extensionNode) commitSnapshot(
	db common.DBWriteCacher,
	leavesChan chan core.KeyValueHolder,
	missingNodesChan chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDone(ctx, idleProvider) {
		return errors.ErrContextClosing
	}

	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit snapshot error %w", err)
	}

	err = resolveIfCollapsed(en, 0, db)
	isMissingNodeErr := false
	if err != nil {
		isMissingNodeErr = strings.Contains(err.Error(), common.GetNodeFromDBErrorString)
		if !isMissingNodeErr {
			return err
		}
	}

	if isMissingNodeErr {
		treatCommitSnapshotError(err, en.EncodedChild, missingNodesChan)
	} else {
		err = en.child.commitSnapshot(db, leavesChan, missingNodesChan, ctx, stats, idleProvider, depthLevel+1)
		if err != nil {
			return err
		}
	}

	return en.saveToStorage(db, stats, depthLevel)
}

func (en *extensionNode) saveToStorage(targetDb common.DBWriteCacher, stats common.TrieStatisticsHandler, depthLevel int) error {
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

func (en *extensionNode) resolveCollapsed(_ byte, db common.DBWriteCacher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("resolveCollapsed error %w", err)
	}
	child, err := getNodeFromDBAndDecode(en.EncodedChild, db, en.marsh, en.hasher)
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

func (en *extensionNode) isPosCollapsed(_ int) bool {
	return en.isCollapsed()
}

func (en *extensionNode) tryGet(key []byte, currentDepth uint32, db common.DBWriteCacher) (value []byte, maxDepth uint32, err error) {
	err = en.isEmptyOrNil()
	if err != nil {
		return nil, currentDepth, fmt.Errorf("tryGet error %w", err)
	}
	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, currentDepth, nil
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, currentDepth, nil
	}
	key = key[len(en.Key):]
	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return nil, currentDepth, err
	}

	return en.child.tryGet(key, currentDepth+1, db)
}

func (en *extensionNode) getNext(key []byte, db common.DBWriteCacher) (node, []byte, error) {
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
	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return nil, nil, err
	}

	key = key[len(en.Key):]
	return en.child, key, nil
}

func (en *extensionNode) insert(n *leafNode, db common.DBWriteCacher) (node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, emptyHashes, fmt.Errorf("insert error %w", err)
	}
	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return nil, emptyHashes, err
	}

	keyMatchLen := prefixLen(n.Key, en.Key)

	// If the whole key matches, keep this extension node as is
	// and only update the value.
	if keyMatchLen == len(en.Key) {
		return en.insertInSameEn(n, keyMatchLen, db)
	}

	// Otherwise branch out at the index where they differ.
	return en.insertInNewBn(n, keyMatchLen)
}

func (en *extensionNode) insertInSameEn(n *leafNode, keyMatchLen int, db common.DBWriteCacher) (node, [][]byte, error) {
	n.Key = n.Key[keyMatchLen:]
	newNode, oldHashes, err := en.child.insert(n, db)
	if check.IfNil(newNode) || err != nil {
		return nil, [][]byte{}, err
	}

	if !en.dirty {
		oldHashes = append(oldHashes, en.hash)
	}

	newEn, err := newExtensionNode(en.Key, newNode, en.marsh, en.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	return newEn, oldHashes, nil
}

func (en *extensionNode) insertInNewBn(n *leafNode, keyMatchLen int) (node, [][]byte, error) {
	oldHash := make([][]byte, 0)
	if !en.dirty {
		oldHash = append(oldHash, en.hash)
	}

	bn, err := newBranchNode(en.marsh, en.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	oldChildPos := en.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return nil, [][]byte{}, ErrChildPosOutOfRange
	}

	followingExtensionNode, err := newExtensionNode(en.Key[keyMatchLen+1:], en.child, en.marsh, en.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	if len(followingExtensionNode.Key) < 1 {
		bn.children[oldChildPos] = en.child
	} else {
		bn.children[oldChildPos] = followingExtensionNode
	}
	n.Key = n.Key[keyMatchLen+1:]
	bn.children[newChildPos] = n

	if keyMatchLen == 0 {
		return bn, oldHash, nil
	}

	newEn, err := newExtensionNode(en.Key[:keyMatchLen], bn, en.marsh, en.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	return newEn, oldHash, nil
}

func (en *extensionNode) delete(key []byte, db common.DBWriteCacher) (bool, node, [][]byte, error) {
	emptyHashes := make([][]byte, 0)
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, emptyHashes, fmt.Errorf("delete error %w", err)
	}
	if len(key) == 0 {
		return false, nil, emptyHashes, ErrValueTooShort
	}
	keyMatchLen := prefixLen(key, en.Key)
	if keyMatchLen < len(en.Key) {
		return false, en, emptyHashes, nil
	}
	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return false, nil, emptyHashes, err
	}

	dirty, newNode, oldHashes, err := en.child.delete(key[len(en.Key):], db)
	if !dirty || err != nil {
		return false, en, emptyHashes, err
	}

	if !en.dirty {
		oldHashes = append(oldHashes, en.hash)
	}

	var n node
	switch newNode := newNode.(type) {
	case *leafNode:
		n, err = newLeafNode(concat(en.Key, newNode.Key...), newNode.Value, en.marsh, en.hasher)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		return true, n, oldHashes, nil
	case *extensionNode:
		n, err = newExtensionNode(concat(en.Key, newNode.Key...), newNode.child, en.marsh, en.hasher)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		return true, n, oldHashes, nil
	default:
		n, err = newExtensionNode(en.Key, newNode, en.marsh, en.hasher)
		if err != nil {
			return false, nil, emptyHashes, err
		}

		return true, n, oldHashes, nil
	}
}

func (en *extensionNode) reduceNode(pos int) (node, bool, error) {
	k := append([]byte{byte(pos)}, en.Key...)

	newEn, err := newExtensionNode(k, en.child, en.marsh, en.hasher)
	if err != nil {
		return nil, false, err
	}

	return newEn, true, nil
}

func (en *extensionNode) clone() *extensionNode {
	nodeClone := *en
	return &nodeClone
}

func (en *extensionNode) isEmptyOrNil() error {
	if en == nil {
		return ErrNilExtensionNode
	}
	if en.child == nil && len(en.EncodedChild) == 0 {
		return ErrEmptyExtensionNode
	}
	return nil
}

func (en *extensionNode) print(writer io.Writer, index int, db common.DBWriteCacher) {
	if en == nil {
		return
	}

	err := resolveIfCollapsed(en, 0, db)
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

func (en *extensionNode) getDirtyHashes(hashes common.ModifiedHashes) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getDirtyHashes error %w", err)
	}

	if !en.isDirty() {
		return nil
	}

	if en.child == nil {
		return nil
	}

	err = en.child.getDirtyHashes(hashes)
	if err != nil {
		return err
	}
	hashes[string(en.getHash())] = struct{}{}

	return nil
}

func (en *extensionNode) getChildren(db common.DBWriteCacher) ([]node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getChildren error %w", err)
	}

	nextNodes := make([]node, 0)

	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return nil, err
	}

	nextNodes = append(nextNodes, en.child)

	return nextNodes, nil
}

func (en *extensionNode) getNumNodes() common.NumNodesDTO {
	if check.IfNil(en) {
		return common.NumNodesDTO{}
	}

	childNumNodes := en.child.getNumNodes()
	childNumNodes.Extensions++
	childNumNodes.MaxLevel++

	return childNumNodes
}

func (en *extensionNode) isValid() bool {
	if len(en.EncodedChild) == 0 && en.child == nil {
		return false
	}

	if len(en.Key) == 0 {
		return false
	}

	return true
}

func (en *extensionNode) setDirty(dirty bool) {
	en.dirty = dirty
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
	db common.DBWriteCacher,
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
		err = resolveIfCollapsed(en, 0, db)
		if err != nil {
			return err
		}

		keyBuilder.BuildKey(en.Key)
		err = en.child.getAllLeavesOnChannel(leavesChannel, keyBuilder.Clone(), db, marshalizer, chanClose, ctx)
		if err != nil {
			return err
		}

		en.child = nil
	}

	return nil
}

func (en *extensionNode) getAllHashes(db common.DBWriteCacher) ([][]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getAllHashes error: %w", err)
	}

	err = resolveIfCollapsed(en, 0, db)
	if err != nil {
		return nil, err
	}

	hashes, err := en.child.getAllHashes(db)
	if err != nil {
		return nil, err
	}

	hashes = append(hashes, en.hash)

	return hashes, nil
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

func (en *extensionNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, db common.DBWriteCacher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("collectStats error %w", err)
	}

	err = resolveIfCollapsed(en, 0, db)
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

// IsInterfaceNil returns true if there is no value under the interface
func (en *extensionNode) IsInterfaceNil() bool {
	return en == nil
}
