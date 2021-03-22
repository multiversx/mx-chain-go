package disabled

type epochHandler struct {
}

// NewEpochHandler returns a disabled implementation of epoch provider that will return epoch 0 for each request
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
