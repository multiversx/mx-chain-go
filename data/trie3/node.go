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
