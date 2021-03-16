package builtInFunctions

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
)

func TestEsdtNFTCreateRoleTransfer_Constructor(t *testing.T) {
	t.Parallel()

	e, err := NewESDTNFTCreateRoleTransfer(nil, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilMarshalizer)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, nil, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilAccountsAdapter)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, nil)
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilShardCoordinator)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, err)
	assert.NotNil(t, e)
	assert.False(t, e.IsInterfaceNil())

	e.SetNewGasConfig(&process.GasCost{})
}

func TestESDTNFTCreateRoleTransfer_ProcessWithErrors(t *testing.T) {
	t.Parallel()

	e, err := NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, err)
	assert.NotNil(t, e)

	vmOutput, err := e.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{})
	assert.Equal(t, err, process.ErrNilValue)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(10)}})
	assert.Equal(t, err, process.ErrBuiltInFunctionCalledWithValue)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)

	vmInput := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}}
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmOutput, err = e.ProcessBuiltinFunction(&mock.UserAccountStub{}, nil, vmInput)
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Equal(t, err, process.ErrNilUserAccount)
	assert.Nil(t, vmOutput)

	vmInput.CallerAddr = vm.ESDTSCAddress
	vmInput.Arguments = [][]byte{{1}, {2}, {3}}
	vmOutput, err = e.ProcessBuiltinFunction(nil, &mock.UserAccountStub{}, vmInput)
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)

	vmInput.Arguments = [][]byte{{1}, {2}}
	vmOutput, err = e.ProcessBuiltinFunction(nil, &mock.UserAccountStub{}, vmInput)
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)
}

func createESDTNFTCreateRoleTransferComponent(t *testing.T) *esdtNFTCreateRoleTransfer {
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(1, 0)
	trieStoreManager := createTrieStorageManager(createMemUnit(), marshalizer, hasher)
	tr, _ := trie.NewTrie(trieStoreManager, marshalizer, hasher, 6)
	accounts, _ := state.NewAccountsDB(tr, hasher, marshalizer, factory.NewAccountCreator())

	e, err := NewESDTNFTCreateRoleTransfer(marshalizer, accounts, shardCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, e)
	return e
}

func TestESDTNFTCreateRoleTransfer_ProcessAtCurrentShard(t *testing.T) {
	t.Parallel()

	e := createESDTNFTCreateRoleTransferComponent(t)

	tokenID := []byte("NFT")
	currentOwner := bytes.Repeat([]byte{1}, 32)
	destinationAddr := bytes.Repeat([]byte{2}, 32)
	vmInput := &vmcommon.ContractCallInput{}
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = vm.ESDTSCAddress
	vmInput.Arguments = [][]byte{tokenID, destinationAddr}

	destAcc, _ := e.accounts.LoadAccount(currentOwner)
	userAcc := destAcc.(state.UserAccountHandler)
	vmOutput, err := e.ProcessBuiltinFunction(nil, userAcc, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, err, state.ErrNilTrie)

	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	err = saveRolesToAccount(userAcc, esdtTokenRoleKey, &esdt.ESDTRoles{Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate), []byte(core.ESDTRoleNFTAddQuantity)}}, e.marshalizer)
	_ = saveLatestNonce(userAcc, tokenID, 100)
	_ = e.accounts.SaveAccount(userAcc)
	_, _ = e.accounts.Commit()
	destAcc, _ = e.accounts.LoadAccount(currentOwner)
	userAcc = destAcc.(state.UserAccountHandler)

	vmOutput, err = e.ProcessBuiltinFunction(nil, userAcc, vmInput)
	assert.Nil(t, err)
	assert.Equal(t, len(vmOutput.OutputAccounts), 1)

	_ = e.accounts.SaveAccount(userAcc)
	_, _ = e.accounts.Commit()
	checkLatestNonce(t, e, currentOwner, tokenID, 0)
	checkNFTCreateRoleExists(t, e, currentOwner, tokenID, -1)

	checkLatestNonce(t, e, destinationAddr, tokenID, 100)
	checkNFTCreateRoleExists(t, e, destinationAddr, tokenID, 0)
}

func TestESDTNFTCreateRoleTransfer_ProcessCrossShard(t *testing.T) {
	t.Parallel()

	e := createESDTNFTCreateRoleTransferComponent(t)

	tokenID := []byte("NFT")
	currentOwner := bytes.Repeat([]byte{1}, 32)
	destinationAddr := bytes.Repeat([]byte{2}, 32)
	vmInput := &vmcommon.ContractCallInput{}
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = currentOwner
	nonce := uint64(100)
	vmInput.Arguments = [][]byte{tokenID, big.NewInt(0).SetUint64(nonce).Bytes()}

	destAcc, _ := e.accounts.LoadAccount(destinationAddr)
	userAcc := destAcc.(state.UserAccountHandler)
	vmOutput, err := e.ProcessBuiltinFunction(nil, userAcc, vmInput)
	assert.Nil(t, err)
	assert.Equal(t, len(vmOutput.OutputAccounts), 0)

	_ = e.accounts.SaveAccount(userAcc)
	_, _ = e.accounts.Commit()
	checkLatestNonce(t, e, destinationAddr, tokenID, 100)
	checkNFTCreateRoleExists(t, e, destinationAddr, tokenID, 0)

	destAcc, _ = e.accounts.LoadAccount(destinationAddr)
	userAcc = destAcc.(state.UserAccountHandler)
	vmOutput, err = e.ProcessBuiltinFunction(nil, userAcc, vmInput)
	assert.Nil(t, err)
	assert.Equal(t, len(vmOutput.OutputAccounts), 0)

	_ = e.accounts.SaveAccount(userAcc)
	_, _ = e.accounts.Commit()
	checkLatestNonce(t, e, destinationAddr, tokenID, 100)
	checkNFTCreateRoleExists(t, e, destinationAddr, tokenID, 0)

	vmInput.Arguments = append(vmInput.Arguments, []byte{100})
	vmOutput, err = e.ProcessBuiltinFunction(nil, userAcc, vmInput)
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)
}

func checkLatestNonce(t *testing.T, e *esdtNFTCreateRoleTransfer, addr []byte, tokenID []byte, expectedNonce uint64) {
	destAcc, _ := e.accounts.LoadAccount(addr)
	userAcc := destAcc.(state.UserAccountHandler)
	nonce, _ := getLatestNonce(userAcc, tokenID)
	assert.Equal(t, nonce, expectedNonce)
}

func checkNFTCreateRoleExists(t *testing.T, e *esdtNFTCreateRoleTransfer, addr []byte, tokenID []byte, expectedIndex int) {
	destAcc, _ := e.accounts.LoadAccount(addr)
	userAcc := destAcc.(state.UserAccountHandler)
	esdtTokenRoleKey := append(roleKeyPrefix, tokenID...)
	roles, _, _ := getESDTRolesForAcnt(e.marshalizer, userAcc, esdtTokenRoleKey)
	assert.Equal(t, 1, len(roles.Roles))
	index, _ := doesRoleExist(roles, []byte(core.ESDTRoleNFTCreate))
	assert.Equal(t, expectedIndex, index)
}
