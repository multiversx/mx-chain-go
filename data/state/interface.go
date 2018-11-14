package state

// AccountsHandler is used for the structure that manages the accounts
type AccountsHandler interface {
	RetrieveCode(state *AccountState) error
	RetrieveData(state *AccountState) error
	PutCode(state *AccountState, code []byte) error
	HasAccount(address Address) (bool, error)
	SaveAccountState(state *AccountState) error
	RemoveAccount(address Address) error
	GetOrCreateAccount(address Address) (*AccountState, error)
	Commit() ([]byte, error)
}

// JurnalEntry will be used to implement different state changes to be able to easily revert them
type JurnalEntry interface {
	Revert(accounts AccountsHandler) error
	DirtyAddress() *Address
}
