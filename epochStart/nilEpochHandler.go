package epochStart

// nilEpochHandler is used for null object pattern
type nilEpochHandler struct {
}

// NewNilEpochHandler will return a new instance of a nilEpochHandler
func NewNilEpochHandler() *nilEpochHandler {
	return &nilEpochHandler{}
}

// Epoch will return 0
func (neh *nilEpochHandler) Epoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (neh *nilEpochHandler) IsInterfaceNil() bool {
	return neh == nil
}
