package disabled

type epochHandler struct {
}

// NewEpochHandler returns a nil implementation of epoch provider that knows only the epoch 0
func NewEpochHandler() *epochHandler {
	return &epochHandler{}
}

// MetaEpoch returns epoch 0 as this is a disabled component
func (eh *epochHandler) MetaEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns if underlying struct is nil
func (eh *epochHandler) IsInterfaceNil() bool {
	return eh == nil
}
