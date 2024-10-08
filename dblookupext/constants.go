package dblookupext

// HashPrefix identifies the prefix of a hash
type HashPrefix uint8

const (
	// RelayedTxHash defines the prefix of a relayed transaction hash
	RelayedTxHash HashPrefix = iota
)
