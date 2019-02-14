package patriciaMerkleTrie

const (
	nrOfChildren = 17
	typeNode     = false
	typeVal      = true
)

type node interface {
	getHash() []byte
}

type (
	fullNode struct {
		Children [nrOfChildren]node
		hash     hashNode
	}
	shortNode struct {
		Key     []byte
		Val     node
		hash    hashNode
		ValType bool // needed for decoding, to know if the Val holds a valueNode
	}
	hashNode  []byte
	valueNode []byte
)

func (n *fullNode) copy() *fullNode   { cpy := *n; return &cpy }
func (n *shortNode) copy() *shortNode { cpy := *n; return &cpy }

func (n *fullNode) getHash() []byte  { return n.hash }
func (n *shortNode) getHash() []byte { return n.hash }
func (n valueNode) getHash() []byte  { return nil }
func (n hashNode) getHash() []byte   { return nil }
