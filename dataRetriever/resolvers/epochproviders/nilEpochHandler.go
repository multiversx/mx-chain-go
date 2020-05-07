package epochproviders

type nilEpochHandler struct {
}

// NewNilEpochHandler returns a nil implementation of epoch provider that knows only the epoch 0
func NewNilEpochHandler() *nilEpochHandler {
	return &nilEpochHandler{}
}

// MetaEpoch returns the current meta epoch
func (n *nilEpochHandler) MetaEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns if underlying struct is nil
func (n *nilEpochHandler) IsInterfaceNil() bool {
	return n == nil
}
