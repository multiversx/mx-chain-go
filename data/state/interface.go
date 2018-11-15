package state

//RetrieveData(state *AccountState) error
//Commit() ([]byte, error)

// AccountsHandler is used for the structure that manages the accounts
type AccountsHandler interface {
	RetrieveCode(state *AccountState) error
	PutCode(state *AccountState, code []byte) error
	RemoveCode(codeHash []byte) error
	HasAccount(address Address) (bool, error)
	SaveAccountState(state *AccountState) error
	RemoveAccount(address Address) error
	GetOrCreateAccount(address Address) (*AccountState, error)
	Jurnal() *Jurnal
}

// JurnalEntry will be used to implement different state changes to be able to easily revert them
type JurnalEntry interface {
	Revert(accounts AccountsHandler) error
	DirtyAddress() *Address
}
