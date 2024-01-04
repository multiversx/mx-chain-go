package node

type nodeFactory struct {
}

func NewNodeFactory() *nodeFactory {
	return &nodeFactory{}
}

func (nf *nodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	n, err := NewNode(opts...)

	if err != nil {
		return nil, err
	}

	return n, nil
}

func (nf *nodeFactory) IsInterfaceNil() bool {
	return nf == nil
}
