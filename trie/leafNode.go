package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
)

var _ = node(&leafNode{})

func newLeafNode(key, value []byte, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*leafNode, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &leafNode{
		CollapsedLn: CollapsedLn{
			Key:   key,
			Value: value,
		},
		baseNode: &baseNode{
			dirty:  true,
			marsh:  marshalizer,
			hasher: hasher,
		},
	}, nil
}

func (ln *leafNode) getHash() []byte {
	return ln.hash
}

func (ln *leafNode) setGivenHash(hash []byte) {
	ln.hash = hash
}

func (ln *leafNode) isDirty() bool {
	return ln.dirty
}

func (ln *leafNode) getMarshalizer() marshal.Marshalizer {
	return ln.marsh
}

func (ln *leafNode) setMarshalizer(marshalizer marshal.Marshalizer) {
	ln.marsh = marshalizer
}

func (ln *leafNode) getHasher() hashing.Hasher {
	return ln.hasher
}

func (ln *leafNode) setHasher(hasher hashing.Hasher) {
	ln.hasher = hasher
}

func (ln *leafNode) getCollapsed() (node, error) {
	return ln, nil
}

func (ln *leafNode) setHash() error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("setHash error %w", err)
	}
	if ln.getHash() != nil {
		return nil
	}
	hash, err := hashChildrenAndNode(ln)
	if err != nil {
		return err
	}
	ln.hash = hash
	return nil
}

func (ln *leafNode) setHashConcurrent(wg *sync.WaitGroup, c chan error) {
	err := ln.setHash()
	if err != nil {
		c <- err
	}
	wg.Done()
}

func (ln *leafNode) setRootHash() error {
	return ln.setHash()
}

func (ln *leafNode) hashChildren() error {
	return nil
}

func (ln *leafNode) hashNode() ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("hashNode error %w", err)
	}
	return encodeNodeAndGetHash(ln)
}

func (ln *leafNode) commitDirty(_ byte, _ uint, _ common.TrieStorageInteractor, targetDb common.BaseStorer) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit error %w", err)
	}

	if !ln.dirty {
		return nil
	}

	ln.dirty = false
	_, err = encodeNodeAndCommitToDB(ln, targetDb)

	return err
}

func (ln *leafNode) commitCheckpoint(
	_ common.TrieStorageInteractor,
	targetDb common.BaseStorer,
	checkpointHashes CheckpointHashesHolder,
	leavesChan chan core.KeyValueHolder,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider) {
		return core.ErrContextClosing
	}

	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit checkpoint error %w", err)
	}

	hash, err := computeAndSetNodeHash(ln)
	if err != nil {
		return err
	}

	shouldCommit := checkpointHashes.ShouldCommit(hash)
	if !shouldCommit {
		return nil
	}

	err = writeNodeOnChannel(ln, leavesChan)
	if err != nil {
		return err
	}

	checkpointHashes.Remove(hash)

	nodeSize, err := encodeNodeAndCommitToDB(ln, targetDb)
	if err != nil {
		return err
	}

	stats.AddLeafNode(depthLevel, uint64(nodeSize))

	return nil
}

func (ln *leafNode) commitSnapshot(
	db common.TrieStorageInteractor,
	leavesChan chan core.KeyValueHolder,
	_ chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider) {
		return core.ErrContextClosing
	}

	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit snapshot error %w", err)
	}

	err = writeNodeOnChannel(ln, leavesChan)
	if err != nil {
		return err
	}

	nodeSize, err := encodeNodeAndCommitToDB(ln, db)
	if err != nil {
		return err
	}

	stats.AddLeafNode(depthLevel, uint64(nodeSize))

	return nil
}

func writeNodeOnChannel(ln *leafNode, leavesChan chan core.KeyValueHolder) error {
	if leavesChan == nil {
		return nil
	}

	leafHash, err := computeAndSetNodeHash(ln)
	if err != nil {
		return err
	}

	trieLeaf := keyValStorage.NewKeyValStorage(leafHash, ln.Value)
	leavesChan <- trieLeaf

	return nil
}

func (ln *leafNode) getEncodedNode() ([]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getEncodedNode error %w", err)
	}
	marshaledNode, err := ln.marsh.Marshal(ln)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, leaf)
	return marshaledNode, nil
}

func (ln *leafNode) resolveCollapsed(_ byte, _ common.TrieStorageInteractor) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(_ int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte, currentDepth uint32, _ common.TrieStorageInteractor) (value []byte, maxDepth uint32, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, currentDepth, fmt.Errorf("tryGet error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, currentDepth, nil
	}

	return nil, currentDepth, nil
}

func (ln *leafNode) getNext(key []byte, _ common.TrieStorageInteractor) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("getNext error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}
func (ln *leafNode) insert(n *leafNode, _ common.TrieStorageInteractor) (node, [][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, [][]byte{}, fmt.Errorf("insert error %w", err)
	}

	oldHash := make([][]byte, 0)
	if !ln.dirty {
		oldHash = append(oldHash, ln.hash)
	}

	insertedKey := n.Key
	nodeKey := ln.Key

	if bytes.Equal(insertedKey, nodeKey) {
		return ln.insertInSameLn(n, oldHash)
	}

	keyMatchLen := prefixLen(insertedKey, nodeKey)
	bn, err := ln.insertInNewBn(n, keyMatchLen)
	if err != nil {
		return nil, [][]byte{}, err
	}

	if keyMatchLen == 0 {
		return bn, oldHash, nil
	}

	newEn, err := newExtensionNode(nodeKey[:keyMatchLen], bn, ln.marsh, ln.hasher)
	if err != nil {
		return nil, [][]byte{}, err
	}

	return newEn, oldHash, nil
}

func (ln *leafNode) insertInSameLn(n *leafNode, oldHashes [][]byte) (node, [][]byte, error) {
	if bytes.Equal(ln.Value, n.Value) {
		return nil, [][]byte{}, nil
	}

	ln.Value = n.Value
	ln.dirty = true
	ln.hash = nil
	return ln, oldHashes, nil
}

func (ln *leafNode) insertInNewBn(n *leafNode, keyMatchLen int) (node, error) {
	bn, err := newBranchNode(ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}

	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return nil, ErrChildPosOutOfRange
	}

	newLnOldChildPos, err := newLeafNode(ln.Key[keyMatchLen+1:], ln.Value, ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}
	bn.children[oldChildPos] = newLnOldChildPos

	newLnNewChildPos, err := newLeafNode(n.Key[keyMatchLen+1:], n.Value, ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}
	bn.children[newChildPos] = newLnNewChildPos

	return bn, nil
}

func (ln *leafNode) delete(key []byte, _ common.TrieStorageInteractor) (bool, node, [][]byte, error) {
	if bytes.Equal(key, ln.Key) {
		oldHash := make([][]byte, 0)
		if !ln.dirty {
			oldHash = append(oldHash, ln.hash)
		}

		return true, nil, oldHash, nil
	}
	return false, ln, [][]byte{}, nil
}

func (ln *leafNode) reduceNode(pos int) (node, bool, error) {
	k := append([]byte{byte(pos)}, ln.Key...)

	newLn, err := newLeafNode(k, ln.Value, ln.marsh, ln.hasher)
	if err != nil {
		return nil, false, err
	}

	return newLn, true, nil
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

func (ln *leafNode) print(writer io.Writer, _ int, _ common.TrieStorageInteractor) {
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

	_, _ = fmt.Fprintf(writer, "L: key= %v, (%v) - %v\n", ln.Key, hex.EncodeToString(ln.hash), ln.dirty)
}

func (ln *leafNode) getDirtyHashes(hashes common.ModifiedHashes) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getDirtyHashes error %w", err)
	}

	if !ln.isDirty() {
		return nil
	}

	hashes[string(ln.getHash())] = struct{}{}
	return nil
}

func (ln *leafNode) getChildren(_ common.TrieStorageInteractor) ([]node, error) {
	return nil, nil
}

func (ln *leafNode) isValid() bool {
	return len(ln.Value) > 0
}

func (ln *leafNode) setDirty(dirty bool) {
	ln.dirty = dirty
}

func (ln *leafNode) loadChildren(_ func([]byte) (node, error)) ([][]byte, []node, error) {
	return nil, nil, nil
}

func (ln *leafNode) getAllLeavesOnChannel(
	leavesChannel chan core.KeyValueHolder,
	keyBuilder common.KeyBuilder,
	_ common.TrieStorageInteractor,
	_ marshal.Marshalizer,
	chanClose chan struct{},
	ctx context.Context,
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

	trieLeaf := keyValStorage.NewKeyValStorage(nodeKey, ln.Value)
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

func (ln *leafNode) getAllHashes(_ common.TrieStorageInteractor) ([][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getAllHashes error: %w", err)
	}

	return [][]byte{ln.hash}, nil
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

	// hasher + marshalizer  + dirty flag = numNodeInnerPointers * pointerSizeInBytes + 1
	nodeSize := len(ln.hash) + len(ln.Key) + len(ln.Value) + numNodeInnerPointers*pointerSizeInBytes + 1

	return nodeSize
}

func (ln *leafNode) getValue() []byte {
	return ln.Value
}

func (ln *leafNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, _ common.TrieStorageInteractor) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("collectStats error %w", err)
	}

	val, err := collapseAndEncodeNode(ln)
	if err != nil {
		return err
	}

	ts.AddLeafNode(depthLevel, uint64(len(val)))
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ln *leafNode) IsInterfaceNil() bool {
	return ln == nil
}
