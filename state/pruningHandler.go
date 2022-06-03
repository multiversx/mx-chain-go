package state

type pruningHandler struct {
	isPruningEnabled bool
}

// NewPruningHandler returns a new instance of pruningHandler with the given parameters
func NewPruningHandler(isPruningEnabled bool) *pruningHandler {
	return &pruningHandler{
		isPruningEnabled: isPruningEnabled,
	}
}

// IsPruningEnabled returns the value of the underlying isPruningEnabled flag
func (ph *pruningHandler) IsPruningEnabled() bool {
	return ph.isPruningEnabled
}
