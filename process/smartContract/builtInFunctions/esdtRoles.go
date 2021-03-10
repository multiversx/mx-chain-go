package builtInFunctions

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ process.BuiltinFunction = (*esdtRoles)(nil)
var roleKeyPrefix = []byte(core.ElrondProtectedKeyPrefix + core.ESDTRoleIdentifier + core.ESDTKeyIdentifier)

type esdtRoles struct {
	set         bool
	marshalizer marshal.Marshalizer
}

// NewESDTRolesFunc returns the esdt change roles built-in function component
func NewESDTRolesFunc(
	marshalizer marshal.Marshalizer,
	set bool,
) (*esdtRoles, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	e := &esdtRoles{
		set:         set,
		marshalizer: marshalizer,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtRoles) SetNewGasConfig(_ *process.GasCost) {
}

// ProcessBuiltinFunction resolves ESDT change roles function call
func (e *esdtRoles) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	err := checkBasicESDTArguments(vmInput)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(vmInput.CallerAddr, vm.ESDTSCAddress) {
		return nil, process.ErrAddressIsNotESDTSystemSC
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilUserAccount
	}

	esdtTokenRoleKey := append(roleKeyPrefix, vmInput.Arguments[0]...)
	log.Trace(vmInput.Function, "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "key", esdtTokenRoleKey)

	roles, _, err := getESDTRolesForAcnt(e.marshalizer, acntDst, esdtTokenRoleKey)
	if err != nil {
		return nil, err
	}

	if e.set {
		roles.Roles = append(roles.Roles, vmInput.Arguments[1:]...)
	} else {
		deleteRoles(roles, vmInput.Arguments[1:])
	}

	err = saveRolesToAccount(acntDst, esdtTokenRoleKey, roles, e.marshalizer)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}
	return vmOutput, nil
}

func deleteRoles(roles *esdt.ESDTRoles, deleteRoles [][]byte) {
	for _, arg := range deleteRoles {
		index, exist := doesRoleExist(roles, arg)
		if !exist {
			continue
		}

		copy(roles.Roles[index:], roles.Roles[index+1:])
		roles.Roles[len(roles.Roles)-1] = nil
		roles.Roles = roles.Roles[:len(roles.Roles)-1]
	}
}

func doesRoleExist(roles *esdt.ESDTRoles, role []byte) (int, bool) {
	for i, currentRole := range roles.Roles {
		if bytes.Equal(currentRole, role) {
			return i, true
		}
	}
	return -1, false
}

func getESDTRolesForAcnt(
	marshalizer marshal.Marshalizer,
	acnt state.UserAccountHandler,
	key []byte,
) (*esdt.ESDTRoles, bool, error) {
	marshaledData, err := acnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		return nil, false, err
	}

	roles := &esdt.ESDTRoles{
		Roles: make([][]byte, 0),
	}
	if len(marshaledData) == 0 {
		return roles, true, nil
	}

	err = marshalizer.Unmarshal(roles, marshaledData)
	if err != nil {
		return nil, false, err
	}

	return roles, false, nil
}

// CheckAllowedToExecute returns error if the account is not allowed to execute the given action
func (e *esdtRoles) CheckAllowedToExecute(account state.UserAccountHandler, tokenID []byte, action []byte) error {
	if check.IfNil(account) {
		return process.ErrNilUserAccount
	}

	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	roles, isNew, err := getESDTRolesForAcnt(e.marshalizer, account, esdtTokenRoleKey)
	if err != nil {
		return err
	}
	if isNew {
		return process.ErrActionNotAllowed
	}

	for _, role := range roles.Roles {
		if bytes.Equal(role, action) {
			return nil
		}
	}

	return process.ErrActionNotAllowed
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtRoles) IsInterfaceNil() bool {
	return e == nil
}
