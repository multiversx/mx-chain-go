package disabled

type hasher struct {
}

// NewDisabledHasher -
func NewDisabledHasher() *hasher {
	return &hasher{}
}

// Compute - return nothing
func (h *hasher) Compute(_ string) []byte {
	return make([]byte, 0)
}

// Size - returns size
func (h *hasher) Size() int {
	return 0
}

// IsInterfaceNil - returns true if there is no value under the interface
func (h *hasher) IsInterfaceNil() bool {
	return h == nil
}
