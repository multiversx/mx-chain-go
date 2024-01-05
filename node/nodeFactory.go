package node

type nodeFactory struct {
}

// NewNodeFactory creates a new node factory instance
func NewNodeFactory() *nodeFactory {
	return &nodeFactory{}
}

// CreateNewNode will create a new NodeHandler
func (nf *nodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	n, err := NewNode(opts...)

	if err != nil {
		return nil, err
	}

	return n, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (nf *nodeFactory) IsInterfaceNil() bool {
	return nf == nil
}
