package mock

import "github.com/multiversx/mx-chain-go/common"

// ESDTRoleHandlerStub -
type ESDTRoleHandlerStub struct {
	CheckAllowedToExecuteCalled func(account common.UserAccountHandler, tokenID []byte, action []byte) error
}

// CheckAllowedToExecute -
func (e *ESDTRoleHandlerStub) CheckAllowedToExecute(account common.UserAccountHandler, tokenID []byte, action []byte) error {
	if e.CheckAllowedToExecuteCalled != nil {
		return e.CheckAllowedToExecuteCalled(account, tokenID, action)
	}

	return nil
}

// IsInterfaceNil -
func (e *ESDTRoleHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
