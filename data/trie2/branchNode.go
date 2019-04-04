package trie2

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func (bn *branchNode) getHash() []byte {
	return bn.hash
}

func (bn *branchNode) isDirty() bool {
	return bn.dirty
}

func (bn *branchNode) getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	collapsed := bn.clone()
	if bn.isCollapsed() {
		return collapsed, nil
	}
	for i := range bn.children {
		if bn.children[i] != nil {
			ok, err := hasValidHash(bn.children[i])
			if err != nil {
				return nil, err
			}
			if !ok {
				err := bn.children[i].setHash(marshalizer, hasher)
				if err != nil {
					return nil, err
				}
			}
			collapsed.EncodedChildren[i] = bn.children[i].getHash()
			collapsed.children[i] = nil
		}
	}
	return collapsed, nil
}

func (bn *branchNode) setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if bn.isCollapsed() {
		hash, err := encodeNodeAndGetHash(bn, marshalizer, hasher)
		if err != nil {
			return err
		}
		bn.hash = hash
		return nil
	}
	hash, err := hashChildrenAndNode(bn, marshalizer, hasher)
	if err != nil {
		return err
	}
	bn.hash = hash
	return nil
}

func (bn *branchNode) hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	for i := 0; i < nrOfChildren; i++ {
		if bn.children[i] != nil {
			err := bn.children[i].setHash(marshalizer, hasher)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bn *branchNode) hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	for i := range bn.EncodedChildren {
		if bn.children[i] != nil {
			encChild, err := encodeNodeAndGetHash(bn.children[i], marshalizer, hasher)
			if err != nil {
				return nil, err
			}
			bn.EncodedChildren[i] = encChild
		}
	}
	return encodeNodeAndGetHash(bn, marshalizer, hasher)
}

func (bn *branchNode) commit(db DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	for i := range bn.children {
		if bn.children[i] != nil {
			err := bn.children[i].commit(db, marshalizer, hasher)
			if err != nil {
				return err
			}
		}
	}
	if !bn.dirty {
		return nil
	}
	bn.dirty = false
	return encodeNodeAndCommitToDB(bn, db, marshalizer, hasher)
}

func (bn *branchNode) getEncodedNode(marshalizer marshal.Marshalizer) ([]byte, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	marshaledNode, err := marshalizer.Marshal(bn)
	if err != nil {
		return nil, err
	}
	marshaledNode = append(marshaledNode, branch)
	return marshaledNode, nil
}

func (bn *branchNode) resolveCollapsed(pos byte, db DBWriteCacher, marshalizer marshal.Marshalizer) error {
	err := bn.isEmptyOrNil()
	if err != nil {
		return err
	}
	if childPosOutOfRange(pos) {
		return ErrChildPosOutOfRange
	}
	if bn.EncodedChildren[pos] != nil {
		child, err := getNodeFromDBAndDecode(bn.EncodedChildren[pos], db, marshalizer)
		if err != nil {
			return err
		}
		bn.children[pos] = child
	}
	return nil
}

func (bn *branchNode) isCollapsed() bool {
	for i := range bn.children {
		if bn.children[i] != nil {
			return false
		}
	}
	return true
}

func (bn *branchNode) tryGet(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (value []byte, err error) {
	err = bn.isEmptyOrNil()
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return nil, err
	}
	if bn.children[childPos] == nil {
		return nil, ErrNodeNotFound
	}
	value, err = bn.children[childPos].tryGet(key, db, marshalizer)
	return value, err
}

func (bn *branchNode) insert(n *leafNode, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if len(n.Key) == 0 {
		return false, nil, ErrValueTooShort
	}
	childPos := n.Key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, ErrChildPosOutOfRange
	}
	n.Key = n.Key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return false, nil, err
	}

	if bn.children[childPos] != nil {
		dirty, newNode, err := bn.children[childPos].insert(n, db, marshalizer)
		if !dirty || err != nil {
			return false, bn, err
		}
		bn.children[childPos] = newNode
		bn.dirty = dirty
		return true, bn, nil
	}
	bn.children[childPos] = newLeafNode(n.Key, n.Value)
	return true, bn, nil
}

func (bn *branchNode) delete(key []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error) {
	err := bn.isEmptyOrNil()
	if err != nil {
		return false, nil, err
	}
	if len(key) == 0 {
		return false, nil, ErrValueTooShort
	}
	childPos := key[firstByte]
	if childPosOutOfRange(childPos) {
		return false, nil, ErrChildPosOutOfRange
	}
	key = key[1:]
	err = resolveIfCollapsed(bn, childPos, db, marshalizer)
	if err != nil {
		return false, nil, err
	}
	dirty, newNode, err := bn.children[childPos].delete(key, db, marshalizer)
	if !dirty || err != nil {
		return false, nil, err
	}

	bn.hash = nil
	bn.children[childPos] = newNode
	if newNode == nil {
		bn.EncodedChildren[childPos] = nil
	}

	nrOfChildren, pos := getChildPosition(bn)

	if nrOfChildren == 1 {
		err = bn.resolveCollapsed(byte(pos), db, marshalizer)
		if err != nil {
			return false, nil, err
		}
		if childPos != 16 {
			newNode := bn.children[pos].reduceNode(pos)
			return true, newNode, nil
		}
		child := bn.children[pos]
		if child, ok := child.(*leafNode); ok {
			return true, newLeafNode([]byte{byte(pos)}, child.Value), nil
		}

	}

	bn.dirty = dirty
	return true, bn, nil
}

func (bn *branchNode) reduceNode(pos int) node {
	return newExtensionNode([]byte{byte(pos)}, bn.children[pos])
}

func (bn *branchNode) nextChild(previousState *nodeIteratorState, path []byte) (newState *nodeIteratorState, newPath []byte, ok bool) {
	for i := previousState.index + 1; i < len(bn.children); i++ {
		child := bn.children[i]
		if child != nil {
			hash := child.getHash()
			state := newIteratorState(hash, child, previousState.hash, len(path))
			newPath := append(path, byte(i))
			if child, ok := child.(*leafNode); ok {
				newPath = append(newPath, child.Key...)
			}
			previousState.index = i - 1
			return state, newPath, true
		}
	}
	return previousState, path, false
}

func getChildPosition(n *branchNode) (nrOfChildren int, childPos int) {
	for i := range &n.children {
		if n.children[i] != nil || n.EncodedChildren[i] != nil {
			nrOfChildren++
			childPos = i
		}
	}
	return
}

func (bn *branchNode) clone() *branchNode {
	nodeClone := *bn
	return &nodeClone
}

func (bn *branchNode) isEmptyOrNil() error {
	if bn == nil {
		return ErrNilNode
	}
	for i := range bn.children {
		if bn.children[i] != nil || bn.EncodedChildren[i] != nil {
			return nil
		}
	}
	return ErrEmptyNode
}
