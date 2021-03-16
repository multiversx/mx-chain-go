package builtInFunctions

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ process.BuiltinFunction = (*esdtNFTCreateRoleTransfer)(nil)

type esdtNFTCreateRoleTransfer struct {
	keyPrefix        []byte
	marshalizer      marshal.Marshalizer
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
}

// NewESDTNFTCreateRoleTransfer returns the esdt NFT create role transfer built-in function component
func NewESDTNFTCreateRoleTransfer(
	marshalizer marshal.Marshalizer,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
) (*esdtNFTCreateRoleTransfer, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	e := &esdtNFTCreateRoleTransfer{
		keyPrefix:        []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		marshalizer:      marshalizer,
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtNFTCreateRoleTransfer) SetNewGasConfig(_ *process.GasCost) {
}

// ProcessBuiltinFunction resolves ESDT create role transfer function call
func (e *esdtNFTCreateRoleTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {

	err := checkBasicESDTArguments(vmInput)
	if err != nil {
		return nil, err
	}
	if !check.IfNil(acntSnd) {
		return nil, process.ErrInvalidArguments
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilUserAccount
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}
	if bytes.Equal(vmInput.CallerAddr, vm.ESDTSCAddress) {
		outAcc, errExec := e.executeTransferNFTCreateChangeAtCurrentOwner(acntDst, vmInput)
		if errExec != nil {
			return nil, errExec
		}
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
		vmOutput.OutputAccounts[string(outAcc.Address)] = outAcc
	} else {
		err = e.executeTransferNFTCreateChangeAtNextOwner(acntDst, vmInput)
		if err != nil {
			return nil, err
		}
	}

	return vmOutput, nil
}

func (e *esdtNFTCreateRoleTransfer) executeTransferNFTCreateChangeAtCurrentOwner(
	acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.OutputAccount, error) {
	if len(vmInput.Arguments) != 2 {
		return nil, process.ErrInvalidArguments
	}
	if len(vmInput.Arguments[1]) != len(vmInput.CallerAddr) {
		return nil, process.ErrInvalidArguments
	}

	tokenID := vmInput.Arguments[0]
	nonce, err := getLatestNonce(acntDst, tokenID)
	if err != nil {
		return nil, err
	}

	err = saveLatestNonce(acntDst, tokenID, 0)
	if err != nil {
		return nil, err
	}

	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	err = e.deleteCreateRoleFromAccount(acntDst, esdtTokenRoleKey)
	if err != nil {
		return nil, err
	}

	destAddress := vmInput.Arguments[1]
	if e.shardCoordinator.ComputeId(destAddress) == e.shardCoordinator.SelfId() {
		newDestinationAcc, errLoad := e.accounts.LoadAccount(destAddress)
		if errLoad != nil {
			return nil, errLoad
		}
		newDestUserAcc, ok := newDestinationAcc.(state.UserAccountHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		err = saveLatestNonce(newDestUserAcc, tokenID, nonce)
		if err != nil {
			return nil, err
		}

		err = e.addCreateRoleToAccount(newDestUserAcc, esdtTokenRoleKey)
		if err != nil {
			return nil, err
		}

		err = e.accounts.SaveAccount(newDestUserAcc)
		if err != nil {
			return nil, err
		}
	}

	outAcc := &vmcommon.OutputAccount{
		Address:         destAddress,
		Balance:         big.NewInt(0),
		BalanceDelta:    big.NewInt(0),
		OutputTransfers: make([]vmcommon.OutputTransfer, 0),
	}
	outTransfer := vmcommon.OutputTransfer{
		Value: big.NewInt(0),
		Data:  []byte(core.BuiltInFunctionESDTNFTCreateRoleTransfer + "@" + hex.EncodeToString(big.NewInt(0).SetUint64(nonce).Bytes())),
	}
	outAcc.OutputTransfers = append(outAcc.OutputTransfers, outTransfer)

	return outAcc, nil
}

func (e *esdtNFTCreateRoleTransfer) deleteCreateRoleFromAccount(
	acntDst state.UserAccountHandler,
	esdtTokenRoleKey []byte,
) error {
	roles, _, err := getESDTRolesForAcnt(e.marshalizer, acntDst, esdtTokenRoleKey)
	if err != nil {
		return err
	}

	deleteRoles(roles, [][]byte{[]byte(core.ESDTRoleNFTCreate)})
	return saveRolesToAccount(acntDst, esdtTokenRoleKey, roles, e.marshalizer)
}

func (e *esdtNFTCreateRoleTransfer) addCreateRoleToAccount(
	acntDst state.UserAccountHandler,
	esdtTokenRoleKey []byte,
) error {
	roles, _, err := getESDTRolesForAcnt(e.marshalizer, acntDst, esdtTokenRoleKey)
	if err != nil {
		return err
	}

	for _, role := range roles.Roles {
		if bytes.Equal(role, []byte(core.ESDTRoleNFTCreate)) {
			log.Error("addCreateRoleToAccount which already has create role")
			return nil
		}
	}

	roles.Roles = append(roles.Roles, []byte(core.ESDTRoleNFTCreate))
	return saveRolesToAccount(acntDst, esdtTokenRoleKey, roles, e.marshalizer)
}

func saveRolesToAccount(
	acntDst state.UserAccountHandler,
	esdtTokenRoleKey []byte,
	roles *esdt.ESDTRoles,
	marshalizer marshal.Marshalizer,
) error {
	marshaledData, err := marshalizer.Marshal(roles)
	if err != nil {
		return err
	}
	err = acntDst.DataTrieTracker().SaveKeyValue(esdtTokenRoleKey, marshaledData)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdtNFTCreateRoleTransfer) executeTransferNFTCreateChangeAtNextOwner(
	acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) error {
	if len(vmInput.Arguments) != 2 {
		return process.ErrInvalidArguments
	}

	tokenID := vmInput.Arguments[0]
	nonce := big.NewInt(0).SetBytes(vmInput.Arguments[1]).Uint64()

	err := saveLatestNonce(acntDst, tokenID, nonce)
	if err != nil {
		return err
	}

	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	err = e.addCreateRoleToAccount(acntDst, esdtTokenRoleKey)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtNFTCreateRoleTransfer) IsInterfaceNil() bool {
	return e == nil
}
