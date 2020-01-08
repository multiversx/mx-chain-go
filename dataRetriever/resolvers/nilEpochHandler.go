package resolvers

type nilEpochHandler struct {
}

// Epoch returns the current epoch
func (n *nilEpochHandler) Epoch() uint32 {
	return 0
}

// IsInterfaceNil returns if underlying struct is nil
func (n *nilEpochHandler) IsInterfaceNil() bool {
	return n == nil
}
