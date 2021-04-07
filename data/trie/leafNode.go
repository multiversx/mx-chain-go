package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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

func (ln *leafNode) commit(force bool, _ byte, _ uint, _ data.DBWriteCacher, targetDb data.DBWriteCacher) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("commit error %w", err)
	}

	shouldNotCommit := !ln.dirty && !force
	if shouldNotCommit {
		return nil
	}

	ln.dirty = false
	return encodeNodeAndCommitToDB(ln, targetDb)
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

func (ln *leafNode) resolveCollapsed(_ byte, _ data.DBWriteCacher) error {
	return nil
}

func (ln *leafNode) isCollapsed() bool {
	return false
}

func (ln *leafNode) isPosCollapsed(_ int) bool {
	return false
}

func (ln *leafNode) tryGet(key []byte, _ data.DBWriteCacher) (value []byte, err error) {
	err = ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("tryGet error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return ln.Value, nil
	}

	return nil, nil
}

func (ln *leafNode) getNext(key []byte, _ data.DBWriteCacher) (node, []byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, nil, fmt.Errorf("getNext error %w", err)
	}
	if bytes.Equal(key, ln.Key) {
		return nil, nil, nil
	}
	return nil, nil, ErrNodeNotFound
}

func (ln *leafNode) insert(n *leafNode, _ data.DBWriteCacher) (bool, node, [][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return false, nil, [][]byte{}, fmt.Errorf("insert error %w", err)
	}

	oldHash := make([][]byte, 0)
	if !ln.dirty {
		oldHash = append(oldHash, ln.hash)
	}

	if bytes.Equal(n.Key, ln.Key) {
		if bytes.Equal(ln.Value, n.Value) {
			return false, ln, [][]byte{}, nil
		}

		ln.Value = n.Value
		ln.dirty = true
		ln.hash = nil
		return true, ln, oldHash, nil
	}

	keyMatchLen := prefixLen(n.Key, ln.Key)
	bn, err := newBranchNode(ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	oldChildPos := ln.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, [][]byte{}, ErrChildPosOutOfRange
	}

	newLnOldChildPos, err := newLeafNode(ln.Key[keyMatchLen+1:], ln.Value, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	bn.children[oldChildPos] = newLnOldChildPos

	newLnNewChildPos, err := newLeafNode(n.Key[keyMatchLen+1:], n.Value, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}
	bn.children[newChildPos] = newLnNewChildPos

	if keyMatchLen == 0 {
		return true, bn, oldHash, nil
	}

	newEn, err := newExtensionNode(ln.Key[:keyMatchLen], bn, ln.marsh, ln.hasher)
	if err != nil {
		return false, nil, [][]byte{}, err
	}

	return true, newEn, oldHash, nil
}

func (ln *leafNode) delete(key []byte, _ data.DBWriteCacher) (bool, node, [][]byte, error) {
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

func (ln *leafNode) print(writer io.Writer, _ int, _ data.DBWriteCacher) {
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

func (ln *leafNode) deepClone() node {
	if ln == nil {
		return nil
	}

	clonedNode := &leafNode{baseNode: &baseNode{}}

	if ln.Key != nil {
		clonedNode.Key = make([]byte, len(ln.Key))
		copy(clonedNode.Key, ln.Key)
	}

	if ln.Value != nil {
		clonedNode.Value = make([]byte, len(ln.Value))
		copy(clonedNode.Value, ln.Value)
	}

	if ln.hash != nil {
		clonedNode.hash = make([]byte, len(ln.hash))
		copy(clonedNode.hash, ln.hash)
	}

	clonedNode.dirty = ln.dirty
	clonedNode.marsh = ln.marsh
	clonedNode.hasher = ln.hasher

	return clonedNode
}

func (ln *leafNode) getDirtyHashes(hashes data.ModifiedHashes) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getDirtyHashes error %w", err)
	}

	if !ln.isDirty() {
		return nil
	}

	hashes[hex.EncodeToString(ln.getHash())] = struct{}{}
	return nil
}

func (ln *leafNode) getChildren(_ data.DBWriteCacher) ([]node, error) {
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
	key []byte,
	_ data.DBWriteCacher,
	_ marshal.Marshalizer,
	_ context.Context,
) error {
	err := ln.isEmptyOrNil()
	if err != nil {
		return fmt.Errorf("getAllLeavesOnChannel error: %w", err)
	}

	nodeKey := append(key, ln.Key...)
	nodeKey, err = hexToKeyBytes(nodeKey)
	if err != nil {
		return err
	}

	trieLeaf := keyValStorage.NewKeyValStorage(nodeKey, ln.Value)
	leavesChannel <- trieLeaf

	return nil
}

func (ln *leafNode) getAllHashes(_ data.DBWriteCacher) ([][]byte, error) {
	err := ln.isEmptyOrNil()
	if err != nil {
		return nil, fmt.Errorf("getAllHashes error: %w", err)
	}

	return [][]byte{ln.hash}, nil
}

func (ln *leafNode) getNextHashAndKey(key []byte) (bool, []byte, []byte) {
	if ln.isInterfaceNil() {
		return false, nil, nil
	}

	if bytes.Equal(key, ln.Key) {
		return true, nil, nil
	}

	return false, nil, nil
}

//TODO(iulian) add tests
func (ln *leafNode) sizeInBytes() int {
	if ln == nil {
		return 0
	}

	// hasher + marshalizer  + dirty flag = 2 * pointerSizeInBytes + 1
	nodeSize := len(ln.hash) + len(ln.Key) + len(ln.Value) + 2*pointerSizeInBytes + 1

	return nodeSize
}

func (ln *leafNode) isInterfaceNil() bool {
	return ln == nil
}
