package state

type AccountsHandler interface {
	RetrieveCode(state *AccountState) error
	RetrieveData(state *AccountState) error
	PutCode(state *AccountState, code []byte) error
	HasAccount(address Address) (bool, error)
	SaveAccountState(state *AccountState) error
	GetOrCreateAccount(address Address) (*AccountState, error)
	Undo() error
	Commit() error
}
