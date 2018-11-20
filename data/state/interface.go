package state

// AccountsHandler is used for the structure that manages the accounts
type AccountsHandler interface {
	RetrieveCode(state *AccountState) error
	PutCode(state *AccountState, code []byte) error
	RemoveCode(codeHash []byte) error

	RetrieveDataTrie(state *AccountState) error

	HasAccount(address Address) (bool, error)
	SaveAccountState(state *AccountState) error
	RemoveAccount(address Address) error
	GetOrCreateAccount(address Address) (*AccountState, error)
	Journal() *Journal
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert(accounts AccountsHandler) error
	DirtyAddress() *Address
}
