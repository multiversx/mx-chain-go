package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var _ = node(&leafNode{})

func newLeafNode(
	newData *dataForInsertion,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (*leafNode, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &leafNode{
		CollapsedLn: CollapsedLn{
			Key:     newData.key,
			Value:   newData.value,
			Version: uint32(newData.version),
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

func (ln *leafNode) commitDirty(_ byte, _ uint, _ common.DBWriteCacher, targetDb common.DBWriteCacher) error {
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
	_ common.DBWriteCacher,
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
	db common.DBWriteCacher,
	leavesChan chan core.KeyValueHolder,
	_ chan []byte,
	ctx context.Context,
	stats common.TrieStatisticsHandler,
	idleProvider IdleNodeProvider,
	depthLevel int,
) error {
	if shouldStopIfContextDone(ctx, idleProvider) {
		return errors.ErrContextClosing
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

func (ln *leafNode) resolveCollapsed(_ byte, _ common.DBWriteCacher) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(_ int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte, currentDepth uint32, _ common.DBWriteCacher) (value []byte, maxDepth uint32, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, currentDepth, fmt.Errorf("tryGet error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, currentDepth, nil
	}

	return nil, currentDepth, nil
}

func (ln *leafNode) getNext(key []byte, _ common.DBWriteCacher) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("getNext error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}
func (ln *leafNode) insert(newData *dataForInsertion, _ common.DBWriteCacher) (node, [][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, [][]byte{}, fmt.Errorf("insert error %w", err)
	}

	oldHash := make([][]byte, 0)
	if !ln.dirty {
		oldHash = append(oldHash, ln.hash)
	}

	nodeKey := ln.Key

	if bytes.Equal(newData.key, nodeKey) {
		return ln.insertInSameLn(newData, oldHash)
	}

	keyMatchLen := prefixLen(newData.key, nodeKey)
	bn, err := ln.insertInNewBn(newData, keyMatchLen)
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

func (ln *leafNode) insertInSameLn(newData *dataForInsertion, oldHashes [][]byte) (node, [][]byte, error) {
	if bytes.Equal(ln.Value, newData.value) {
		return nil, [][]byte{}, nil
	}

	ln.Value = newData.value
	ln.dirty = true
	ln.hash = nil
	return ln, oldHashes, nil
}

func (ln *leafNode) insertInNewBn(newData *dataForInsertion, keyMatchLen int) (node, error) {
	bn, err := newBranchNode(ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}

	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := newData.key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return nil, ErrChildPosOutOfRange
	}

	oldLnVersion, err := ln.getVersion()
	if err != nil {
		return nil, err
	}

	oldLnData := &dataForInsertion{
		key:     ln.Key[keyMatchLen+1:],
		value:   ln.Value,
		version: oldLnVersion,
	}
	newLnOldChildPos, err := newLeafNode(oldLnData, ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}
	bn.children[oldChildPos] = newLnOldChildPos

	newData.key = newData.key[keyMatchLen+1:]
	newLnNewChildPos, err := newLeafNode(newData, ln.marsh, ln.hasher)
	if err != nil {
		return nil, err
	}
	bn.children[newChildPos] = newLnNewChildPos

	return bn, nil
}

func (ln *leafNode) delete(key []byte, _ common.DBWriteCacher) (bool, node, [][]byte, error) {
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

	oldLnVersion, err := ln.getVersion()
	if err != nil {
		return nil, false, err
	}

	oldLnData := &dataForInsertion{
		key:     k,
		value:   ln.Value,
		version: oldLnVersion,
	}

	newLn, err := newLeafNode(oldLnData, ln.marsh, ln.hasher)
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

func (ln *leafNode) print(writer io.Writer, _ int, _ common.DBWriteCacher) {
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

func (ln *leafNode) getChildren(_ common.DBWriteCacher) ([]node, error) {
	return nil, nil
}

func (ln *leafNode) getNumNodes() common.NumNodesDTO {
	return common.NumNodesDTO{
		Leaves:   1,
		MaxLevel: 1,
	}
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
	trieLeafParser common.TrieLeafParser,
	_ common.DBWriteCacher,
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

	trieLeaf, err := trieLeafParser.ParseLeaf(nodeKey, ln.Value)
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

func (ln *leafNode) getAllHashes(_ common.DBWriteCacher) ([][]byte, error) {
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

func (ln *leafNode) collectStats(ts common.TrieStatisticsHandler, depthLevel int, _ common.DBWriteCacher) error {
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

func (ln *leafNode) getVersion() (common.TrieNodeVersion, error) {
	// TODO modify to use appropriate flags in order to know the returned val
	return common.NotSpecified, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ln *leafNode) IsInterfaceNil() bool {
	return ln == nil
}
