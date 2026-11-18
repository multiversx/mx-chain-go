package trie

type disabledTrieGoRoutinesThrottler struct {
}

// NewDisabledTrieGoRoutinesThrottler returns a new instance of a disabledTrieGoRoutinesThrottler
func NewDisabledTrieGoRoutinesThrottler() *disabledTrieGoRoutinesThrottler {
	return &disabledTrieGoRoutinesThrottler{}
}

// CanProcess will always return false
func (d *disabledTrieGoRoutinesThrottler) CanProcess() bool {
	return false
}

// StartProcessing won't do anything
func (d *disabledTrieGoRoutinesThrottler) StartProcessing() {
}

// EndProcessing won't do anything
func (d *disabledTrieGoRoutinesThrottler) EndProcessing() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledTrieGoRoutinesThrottler) IsInterfaceNil() bool {
	return d == nil
}
