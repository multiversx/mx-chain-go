package node

type nodeFactory struct {
}

// NewNodeFactory creates a new node factory instance for regular/main chain
func NewNodeFactory() *nodeFactory {
	return &nodeFactory{}
}

// CreateNewNode will create a new NodeHandler
func (nf *nodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	return NewNode(opts...)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (nf *nodeFactory) IsInterfaceNil() bool {
	return nf == nil
}
