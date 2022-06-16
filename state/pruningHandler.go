package state

// PruningHandlerOperation defines the options that can be activated on pruning
type PruningHandlerOperation bool

const (
	// EnableDataRemoval enables trie pruning
	EnableDataRemoval PruningHandlerOperation = true

	// DisableDataRemoval stops trie pruning
	DisableDataRemoval PruningHandlerOperation = false
)

type pruningHandler struct {
	isPruningEnabled PruningHandlerOperation
}

// NewPruningHandler returns a new instance of pruningHandler with the given parameters
func NewPruningHandler(isPruningEnabled PruningHandlerOperation) *pruningHandler {
	return &pruningHandler{
		isPruningEnabled: isPruningEnabled,
	}
}

// IsPruningEnabled returns the value of the underlying isPruningEnabled flag
func (ph *pruningHandler) IsPruningEnabled() bool {
	return ph.isPruningEnabled == EnableDataRemoval
}
