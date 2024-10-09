package node

type sovereignNodeFactory struct {
	nativeESDT string
}

// NewSovereignNodeFactory creates a new sovereign node factory instance
func NewSovereignNodeFactory(nativeESDT string) *sovereignNodeFactory {
	return &sovereignNodeFactory{
		nativeESDT: nativeESDT,
	}
}

// CreateNewNode creates a new sovereign node
func (snf *sovereignNodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	nd, err := NewNode(opts...)
	if err != nil {
		return nil, err
	}

	return NewSovereignNode(nd, snf.nativeESDT)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (snf *sovereignNodeFactory) IsInterfaceNil() bool {
	return snf == nil
}
