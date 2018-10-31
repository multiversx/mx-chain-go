package state

// AccountsHandler is used for the structure that manages the accounts
type AccountsHandler interface {
	RetrieveCode(state *AccountState) error
	RetrieveData(state *AccountState) error
	PutCode(state *AccountState, code []byte) error
	HasAccount(address Address) (bool, error)
	SaveAccountState(state *AccountState) error
	GetOrCreateAccount(address Address) (*AccountState, error)
	Commit() ([]byte, error)
}
