package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// ESDTRoleHandlerStub -
type ESDTRoleHandlerStub struct {
	CheckAllowedToExecuteCalled func(account state.UserAccountHandler, tokenID []byte, action []byte) error
}

// CheckAllowedToExecute -
func (e *ESDTRoleHandlerStub) CheckAllowedToExecute(account state.UserAccountHandler, tokenID []byte, action []byte) error {
	if e.CheckAllowedToExecuteCalled != nil {
		return e.CheckAllowedToExecuteCalled(account, tokenID, action)
	}

	return nil
}

// IsInterfaceNil -
func (e *ESDTRoleHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
