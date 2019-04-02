package trie2

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func newExtensionNode(key []byte, child node) *extensionNode {
	return &extensionNode{
		Key:          key,
		EncodedChild: nil,
		child:        child,
		hash:         nil,
		dirty:        true,
	}
}

func (en *extensionNode) getHash() []byte {
	return en.hash
}

func (en *extensionNode) isDirty() bool {
	return en.dirty
}

func (en *extensionNode) getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	collapsed := en.clone()
	if en.isCollapsed() {
		return collapsed, nil
	}
	ok, err := hasValidHash(en.child)
	if err != nil {
		return nil, err
	}
	if !ok {
		err := en.child.setHash(marshalizer, hasher)
		if err != nil {
			return nil, err
		}
	}
	collapsed.EncodedChild = en.child.getHash()
	collapsed.child = nil
	return collapsed, nil
}

func (en *extensionNode) setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.isCollapsed() {
		hash, err := encodeNodeAndGetHash(en, marshalizer, hasher)
		if err != nil {
			return err
		}
		en.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(en, marshalizer, hasher)
	if err != nil {
		return err
	}
	en.hash = hash
	return nil
}

func (en *extensionNode) hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.child != nil {
		err = en.child.setHash(marshalizer, hasher)
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *extensionNode) hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if en.child != nil {
		encChild, err := encodeNodeAndGetHash(en.child, marshalizer, hasher)
		if err != nil {
			return nil, err
		}
		en.EncodedChild = encChild
	}
	return encodeNodeAndGetHash(en, marshalizer, hasher)
}

func (en *extensionNode) commit(db DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	if en.child != nil {
		err = en.child.commit(db, marshalizer, hasher)
		if err != nil {
			return err
		}
	}
	if !en.dirty {
		return nil
	}
	en.dirty = false
	return encodeNodeAndCommitToDB(en, db, marshalizer, hasher)
}

func (en *extensionNode) getEncodedNode(marshalizer marshal.Marshalizer) ([]byte, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := marshalizer.Marshal(en)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, extension)
	return marshaledNode, nil
}

func (en *extensionNode) resolveCollapsed(pos byte, db DBWriteCacher, marshalizer marshal.Marshalizer) error {
	err := en.isEmptyOrNil()
	if err != nil {
		return err
	}
	child, err := getNodeFromDBAndDecode(en.EncodedChild, db, marshalizer)
	if err != nil {
		return err
	}
	en.child = child
	return nil
}

func (en *extensionNode) isCollapsed() bool {
	if en.child == nil && en.EncodedChild != nil {
		return true
	}
	return false
}

func (en *extensionNode) tryGet(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = en.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	keyTooShort := len(key) < len(en.Key)
	if keyTooShort {
		return nil, ErrNodeNotFound
	}
	keysDontMatch := !bytes.Equal(en.Key, key[:len(en.Key)])
	if keysDontMatch {
		return nil, ErrNodeNotFound
	}
	key = key[len(en.Key):]
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return nil, err
	}
	value, err = en.child.tryGet(key, db, marshalizer)
	return value, err
}

func (en *extensionNode) insert(n *leafNode, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return false, nil, err
	}
	keyMatchLen := prefixLen(n.Key, en.Key)

	// If the whole key matches, keep this extension node as is
	// and only update the value.
	if keyMatchLen == len(en.Key) {
		n.Key = n.Key[keyMatchLen:]
		dirty, newNode, err := en.child.insert(n, db, marshalizer)
		if !dirty || err != nil {
			return false, nil, err
		}
		return true, newExtensionNode(en.Key, newNode), nil
	}
	// Otherwise branch out at the index where they differ.
	branch := &branchNode{}
	branch.dirty = true
	oldChildPos := en.Key[keyMatchLen]
	newChildPos := n.Key[keyMatchLen]
	if childPosOutOfRange(oldChildPos) || childPosOutOfRange(newChildPos) {
		return false, nil, ErrChildPosOutOfRange
	}
	branch.children[oldChildPos] = newExtensionNode(en.Key[keyMatchLen+1:], en.child)
	n.Key = n.Key[keyMatchLen+1:]
	branch.children[newChildPos] = n

	if keyMatchLen == 0 {
		return true, branch, nil
	}
	return true, newExtensionNode(en.Key[:keyMatchLen], branch), nil
}

func (en *extensionNode) delete(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := en.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if len(key) == 0 {
		return false, nil, ErrValueTooShort
	}
	keyMatchLen := prefixLen(key, en.Key)
	if keyMatchLen < len(en.Key) {
		return false, en, nil // don't replace n on mismatch
	}
	err = resolveIfCollapsed(en, 0, db, marshalizer)
	if err != nil {
		return false, nil, err
	}

	dirty, newNode, err := en.child.delete(key[len(en.Key):], db, marshalizer)
	if !dirty || err != nil {
		return false, en, err
	}

	switch newNode := newNode.(type) {
	case *leafNode:
		return true, newLeafNode(concat(en.Key, newNode.Key...), newNode.Value), nil
	case *extensionNode:
		return true, newExtensionNode(concat(en.Key, newNode.Key...), newNode.child), nil
	default:
		return true, newExtensionNode(en.Key, newNode), nil
	}
}

func (en *extensionNode) reduceNode(pos int) node {
	k := append([]byte{byte(pos)}, en.Key...)
	return newExtensionNode(k, en.child)
}

func (en *extensionNode) nextChild(previousState *nodeIteratorState, path []byte) (newState *nodeIteratorState, newPath []byte, ok bool) {
	if previousState.index < 0 {
		hash := en.child.getHash()
		state := newIteratorState(hash, en.child, previousState.hash, len(path))
		newPath := append(path, en.Key...)
		return state, newPath, true
	}
	return previousState, path, false
}

func (en *extensionNode) clone() *extensionNode {
	nodeClone := *en
	return &nodeClone
}

func (en *extensionNode) isEmptyOrNil() error {
	if en == nil {
		return ErrNilNode
	}
	if en.child == nil && en.EncodedChild == nil {
		return ErrEmptyNode
	}
	return nil
}
