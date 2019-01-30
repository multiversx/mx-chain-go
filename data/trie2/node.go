package trie2

const nrOfChildrenNodes = 17

type (
	fullNode struct {
		Children [nrOfChildrenNodes]Node
		hash     hashNode
	}
	shortNode struct {
		Key  []byte
		Val  Node
		hash hashNode
	}
	valueNode []byte
	hashNode  []byte
)

func (n *fullNode) copy() *fullNode   { cpy := *n; return &cpy }
func (n *shortNode) copy() *shortNode { cpy := *n; return &cpy }

func (n *fullNode) Hash() hashNode  { return n.hash }
func (n *shortNode) Hash() hashNode { return n.hash }
func (n hashNode) Hash() hashNode   { return nil }
func (n valueNode) Hash() hashNode  { return nil }

func Hash(n Node) (Node, error) {
	node, err := getChildren(n)
	if err != nil {
		return hashNode{}, err
	}

	hashed, err := hash(node)
	if err != nil {
		return hashNode{}, err
	}

	hn, _ := hashed.(hashNode)

	switch n := node.(type) {
	case *shortNode:
		n.hash = hn

	case *fullNode:
		n.hash = hn

	}

	return node, nil

}

func getChildren(original Node) (Node, error) {
	var err error

	switch n := original.(type) {
	case *shortNode:
		node := n.copy()
		if _, ok := n.Val.(valueNode); !ok {
			node.Val, err = Hash(n.Val)
			if err != nil {
				return original, err
			}
		}
		return node, nil

	case *fullNode:
		node := n.copy()

		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				node.Children[i], err = Hash(n.Children[i])
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

func hash(n Node) (Node, error) {

	if _, isHash := n.(hashNode); n == nil || isHash {
		return n, nil
	}

	tmp, err := marshalizer.Marshal(n)
	if err != nil {
		return n, err
	}

	hash := n.Hash()
	if hash == nil {
		hash = hashNode(hasher.Compute(string(tmp)))
	}

	return hash, nil
}
