package trie3

const nrOfChildren = 17

type branchNode struct {
	EncodedChildren [nrOfChildren][]byte
	children        [nrOfChildren]node
	hash            []byte
	dirty           bool
}

type extensionNode struct {
	Key          []byte
	EncodedChild []byte
	child        node
	hash         []byte
	dirty        bool
}

type leafNode struct {
	Key   []byte
	Value []byte
	hash  []byte
	dirty bool
}

func (bn *branchNode) getHash() []byte {
	return bn.hash
}

func (en *extensionNode) getHash() []byte {
	return en.hash
}

func (ln *leafNode) getHash() []byte {
	return ln.hash
}

func (bn *branchNode) isCollapsed(pos byte) bool {
	if bn.children[pos] == nil && bn.EncodedChildren[pos] != nil {
		return true
	}
	return false
}

func (en *extensionNode) isCollapsed() bool {
	if en.child == nil && en.EncodedChild != nil {
		return true
	}
	return false
}

func (ln *leafNode) isCollapsed(tr *patriciaMerkleTree) bool {
	_, err := tr.decodeNode(ln.Value)
	if err != nil {
		return false
	}
	return true
}

func (bn *branchNode) copy() *branchNode {
	cpy := *bn
	return &cpy
}

func (en *extensionNode) copy() *extensionNode {
	cpy := *en
	return &cpy
}

func (ln *leafNode) copy() *leafNode {
	cpy := *ln
	return &cpy
}
