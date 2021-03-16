package builtInFunctions

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createNftCreateWithStubArguments() *esdtNFTCreate {
	nftCreate, _ := NewESDTNFTCreateFunc(
		1,
		process.BaseOperationCost{},
		&mock.MarshalizerStub{},
		&mock.PauseHandlerStub{},
		&mock.ESDTRoleHandlerStub{},
	)

	return nftCreate
}

func createNftCreateWithMockArguments(pauseHandler process.ESDTPauseHandler) (*esdtNFTCreate, state.AccountsAdapter) {
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	trieStoreManager := createTrieStorageManager(createMemUnit(), marshalizer, hasher)
	tr, _ := trie.NewTrie(trieStoreManager, marshalizer, hasher, 6)
	accounts, _ := state.NewAccountsDB(tr, hasher, marshalizer, factory.NewAccountCreator())

	nftCreate, _ := NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		marshalizer,
		pauseHandler,
		&mock.ESDTRoleHandlerStub{},
	)

	return nftCreate, accounts
}

func TestNewESDTNFTCreateFunc_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	nftCreate, err := NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		nil,
		&mock.PauseHandlerStub{},
		&mock.ESDTRoleHandlerStub{},
	)
	assert.True(t, check.IfNil(nftCreate))
	assert.Equal(t, process.ErrNilMarshalizer, err)

	nftCreate, err = NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		&mock.MarshalizerStub{},
		nil,
		&mock.ESDTRoleHandlerStub{},
	)
	assert.True(t, check.IfNil(nftCreate))
	assert.Equal(t, process.ErrNilPauseHandler, err)

	nftCreate, err = NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		&mock.MarshalizerStub{},
		&mock.PauseHandlerStub{},
		nil,
	)
	assert.True(t, check.IfNil(nftCreate))
	assert.Equal(t, process.ErrNilRolesHandler, err)
}

func TestNewESDTNFTCreateFunc(t *testing.T) {
	t.Parallel()

	nftCreate, err := NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		&mock.MarshalizerStub{},
		&mock.PauseHandlerStub{},
		&mock.ESDTRoleHandlerStub{},
	)
	assert.False(t, check.IfNil(nftCreate))
	assert.Nil(t, err)
}

func TestEsdtNFTCreate_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	nftCreate := createNftCreateWithStubArguments()
	nftCreate.SetNewGasConfig(nil)
	assert.Equal(t, uint64(1), nftCreate.funcGasCost)
	assert.Equal(t, process.BaseOperationCost{}, nftCreate.gasConfig)

	gasCost := createMockGasCost()
	nftCreate.SetNewGasConfig(&gasCost)
	assert.Equal(t, gasCost.BuiltInCost.ESDTNFTCreate, nftCreate.funcGasCost)
	assert.Equal(t, gasCost.BaseOperationCost, nftCreate.gasConfig)
}

func TestEsdtNFTCreate_ProcessBuiltinFunctionInvalidArguments(t *testing.T) {
	t.Parallel()

	nftCreate := createNftCreateWithStubArguments()
	sender := mock.NewAccountWrapMock([]byte("address"))
	vmOutput, err := nftCreate.ProcessBuiltinFunction(sender, nil, nil)
	assert.Nil(t, vmOutput)
	assert.Equal(t, process.ErrNilVmInput, err)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: []byte("caller"),
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), []byte("arg2")},
		},
		RecipientAddr: []byte("recipient"),
	}
	vmOutput, err = nftCreate.ProcessBuiltinFunction(sender, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, process.ErrInvalidRcvAddr, err)

	vmInput = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: sender.AddressBytes(),
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), []byte("arg2")},
		},
		RecipientAddr: sender.AddressBytes(),
	}
	vmOutput, err = nftCreate.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, process.ErrNilUserAccount, err)

	vmInput = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: sender.AddressBytes(),
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), []byte("arg2")},
		},
		RecipientAddr: sender.AddressBytes(),
	}
	vmOutput, err = nftCreate.ProcessBuiltinFunction(sender, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, process.ErrNotEnoughGas, err)

	vmInput = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender.AddressBytes(),
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), []byte("arg2")},
			GasProvided: 1,
		},
		RecipientAddr: sender.AddressBytes(),
	}
	vmOutput, err = nftCreate.ProcessBuiltinFunction(sender, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.True(t, errors.Is(err, process.ErrInvalidArguments))
}

func TestEsdtNFTCreate_ProcessBuiltinFunctionNotAllowedToExecute(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	nftCreate, _ := NewESDTNFTCreateFunc(
		0,
		process.BaseOperationCost{},
		&mock.MarshalizerStub{},
		&mock.PauseHandlerStub{},
		&mock.ESDTRoleHandlerStub{
			CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
				return expectedErr
			},
		},
	)
	sender := mock.NewAccountWrapMock([]byte("address"))
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: sender.AddressBytes(),
			CallValue:  big.NewInt(0),
			Arguments:  make([][]byte, 7),
		},
		RecipientAddr: sender.AddressBytes(),
	}
	vmOutput, err := nftCreate.ProcessBuiltinFunction(sender, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, expectedErr, err)
}

func TestEsdtNFTCreate_ProcessBuiltinFunctionShouldWork(t *testing.T) {
	t.Parallel()

	nftCreate, accounts := createNftCreateWithMockArguments(&mock.PauseHandlerStub{})
	address := bytes.Repeat([]byte{1}, 32)
	sender, _ := accounts.LoadAccount(address)
	//add some data in the trie, otherwise the creation will fail (it won't happen in real case usage as the create NFT
	//will be called after the creation permission was set in the account's data)
	_ = sender.(state.UserAccountHandler).DataTrieTracker().SaveKeyValue([]byte("key"), []byte("value"))
	_ = accounts.SaveAccount(sender)
	_, _ = accounts.Commit()

	sender, _ = accounts.LoadAccount(address)

	token := "token"
	quantity := big.NewInt(2)
	name := "name"
	royalties := 100 //1%
	hash := []byte("12345678901234567890123456789012")
	attibutes := []byte("attributes")
	uris := [][]byte{[]byte("uri1"), []byte("uri2")}
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: sender.AddressBytes(),
			CallValue:  big.NewInt(0),
			Arguments: [][]byte{
				[]byte(token),
				quantity.Bytes(),
				[]byte(name),
				big.NewInt(int64(royalties)).Bytes(),
				hash,
				attibutes,
				uris[0],
				uris[1],
			},
		},
		RecipientAddr: sender.AddressBytes(),
	}
	vmOutput, err := nftCreate.ProcessBuiltinFunction(sender.(state.UserAccountHandler), nil, vmInput)
	assert.Nil(t, err)
	require.NotNil(t, vmOutput)

	_ = accounts.SaveAccount(sender)
	_, _ = accounts.Commit()

	createdEsdt, latestNonce := readNFTData(t, accounts, nftCreate.marshalizer, []byte(token), 1, address)
	assert.Equal(t, uint64(1), latestNonce)
	expectedEsdt := &esdt.ESDigitalToken{
		Type:       uint32(core.NonFungible),
		Value:      quantity,
		Properties: nil,
		TokenMetaData: &esdt.MetaData{
			Nonce:      1,
			Name:       []byte(name),
			Creator:    address,
			Royalties:  uint32(royalties),
			Hash:       hash,
			URIs:       uris,
			Attributes: attibutes,
		},
	}
	assert.Equal(t, expectedEsdt, createdEsdt)
}

func readNFTData(t *testing.T, accounts state.AccountsAdapter, marshalizer marshal.Marshalizer, tokenID []byte, nonce uint64, address []byte) (*esdt.ESDigitalToken, uint64) {
	account, err := accounts.LoadAccount(address)
	require.Nil(t, err)

	nonceKey := getNonceKey(tokenID)
	latestNonceBytes, err := account.(state.UserAccountHandler).DataTrieTracker().RetrieveValue(nonceKey)
	latestNonce := big.NewInt(0).SetBytes(latestNonceBytes).Uint64()

	createdTokenID := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier)
	createdTokenID = append(createdTokenID, tokenID...)
	tokenKey := computeESDTNFTTokenKey(createdTokenID, nonce)
	data, err := account.(state.UserAccountHandler).DataTrieTracker().RetrieveValue(tokenKey)
	require.Nil(t, err)

	esdtData := &esdt.ESDigitalToken{}
	err = marshalizer.Unmarshal(esdtData, data)
	require.Nil(t, err)

	return esdtData, latestNonce
}
