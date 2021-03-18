package disabled

type epochHandler struct {
}

// NewEpochHandler returns a nil implementation of epoch provider that knows only the epoch 0
func NewEpochHandler() *epochHandler {
	return &epochHandler{}
}

// MetaEpoch returns the current meta epoch
func (eh *epochHandler) MetaEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns if underlying struct is nil
func (eh *epochHandler) IsInterfaceNil() bool {
	return eh == nil
}
