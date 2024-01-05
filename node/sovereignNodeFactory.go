package node

type sovereignNodeFactory struct {
}

// NewSovereignNodeFactory creates a new sovereign node factory instance
func NewSovereignNodeFactory() *sovereignNodeFactory {
	return &sovereignNodeFactory{}
}

// CreateNewNode will create a new NodeHandler
func (snf *sovereignNodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	nd, err := NewNode(opts...)
	if err != nil {
		return nil, err
	}

	snd, err := NewSovereignNode(nd)
	if err != nil {
		return nil, err
	}

	return snd, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (snf *sovereignNodeFactory) IsInterfaceNil() bool {
	return snf == nil
}
