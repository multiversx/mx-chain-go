package state

// AccountsDbIdentifier is the type of accounts db
type AccountsDbIdentifier byte

const (
	// UserAccountsState is the user accounts
	UserAccountsState AccountsDbIdentifier = 0
	// PeerAccountsState is the peer accounts
	PeerAccountsState AccountsDbIdentifier = 1
)

// TriePruningIdentifier is the type for trie pruning identifiers
type TriePruningIdentifier byte

const (
	// OldRoot is appended to the key when oldHashes are added to the evictionWaitingList
	OldRoot TriePruningIdentifier = 0
	// NewRoot is appended to the key when newHashes are added to the evictionWaitingList
	NewRoot TriePruningIdentifier = 1
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32
