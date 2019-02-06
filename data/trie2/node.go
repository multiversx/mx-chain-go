package trie2

const nrOfChildrenNodes = 17

type (
	fullNode struct {
		Children [nrOfChildrenNodes]Node
		hash     []byte
	}
	shortNode struct {
		Key  []byte
		Val  Node
		hash []byte
	}
	valueNode []byte
)

func (n *fullNode) copy() *fullNode   { cpy := *n; return &cpy }
func (n *shortNode) copy() *shortNode { cpy := *n; return &cpy }

func (n *fullNode) GetHash() []byte  { return n.hash }
func (n *shortNode) GetHash() []byte { return n.hash }
func (n valueNode) GetHash() []byte  { return nil }

func SetHash(n Node) (Node, error) {
	node, err := nextChildren(n)
	if err != nil {
		return nil, err
	}

	hashed, err := hash(node)
	if err != nil {
		return nil, err
	}

	switch n := node.(type) {
	case *shortNode:
		n.hash = hashed
	case *fullNode:
		n.hash = hashed
	}

	return node, nil
}

func nextChildren(original Node) (Node, error) {
	var err error

	switch n := original.(type) {
	case *shortNode:
		node := n.copy()
		if _, ok := n.Val.(valueNode); !ok {
			node.Val, err = SetHash(n.Val)
			if err != nil {
				return original, err
			}
		}
		return node, nil

	case *fullNode:
		node := n.copy()

		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				node.Children[i], err = SetHash(n.Children[i])
				if err != nil {
					return original, err
				}
			}
		}
		node.Children[16] = n.Children[16]
		return node, nil

	default:
		return original, nil
	}

}

func hash(n Node) ([]byte, error) {
	tmp, err := marshalizer.Marshal(n)
	if err != nil {
		return nil, err
	}

	hash := n.GetHash()
	if hash == nil {
		hash = hasher.Compute(string(tmp))
	}

	return hash, nil
}
