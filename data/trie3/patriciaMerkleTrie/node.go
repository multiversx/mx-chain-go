package patriciaMerkleTrie

const nrOfChildren = 17

type node interface {
	getHash() []byte
}

type branchNode struct {
	Children      [nrOfChildren][]byte
	childrenNodes [nrOfChildren]node
	hash          []byte
	dirty         bool
}

func (bn *branchNode) ChildrenNodes() [nrOfChildren]node {
	return bn.childrenNodes
}

func (bn *branchNode) copy() *branchNode {
	cpy := *bn
	return &cpy
}

type extensionNode struct {
	Key      []byte
	Next     []byte
	nextNode node
	hash     []byte
	dirty    bool
}

func (en *extensionNode) NextNode() node {
	return en.nextNode
}

func (en *extensionNode) copy() *extensionNode {
	cpy := *en
	return &cpy
}

type leafNode struct {
	Key   []byte
	Value []byte
	hash  []byte
	dirty bool
}

func (ln *leafNode) copy() *leafNode {
	cpy := *ln
	return &cpy
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
