package node

type sovereignNodeFactory struct {
}

func NewSovereignNodeFactory() *sovereignNodeFactory {
	return &sovereignNodeFactory{}
}

func (snf *sovereignNodeFactory) CreateNewNode(opts ...Option) (NodeHandler, error) {
	nd, err := NewNode(opts...)

	if err != nil {
		return nil, err
	}

	snd := NewSovereignNode(nd)
	return snd, nil
}

func (snf *sovereignNodeFactory) IsInterfaceNil() bool {
	return snf == nil
}
