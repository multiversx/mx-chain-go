package chainSimulator

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataApi "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	vmTypeHex                               = "0500"
	codeMetadata                            = "0500"
	minGasPrice                             = 1000000000
	txVersion                               = 1
	mockTxSignature                         = "sig"
	maxNumOfBlocksToGenerateWhenExecutingTx = 10
	signalError                             = "signalError"
	internalVMError                         = "internalVMErrors"

	// OkReturnCode the const for the ok return code
	OkReturnCode = "ok"
	// ESDTSystemAccount the bech32 address for esdt system account
	ESDTSystemAccount = "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t"
)

var (
	// ZeroValue the variable for the zero big int
	ZeroValue = big.NewInt(0)
	// OneEGLD the variable for one egld value
	OneEGLD = big.NewInt(1000000000000000000)
	// MinimumStakeValue the variable for the minimum stake value
	MinimumStakeValue = big.NewInt(0).Mul(OneEGLD, big.NewInt(2500))
	// InitialAmount the variable for initial minting amount in account
	InitialAmount = big.NewInt(0).Mul(OneEGLD, big.NewInt(100))
)

// ArgsDepositToken holds the arguments for a token
type ArgsDepositToken struct {
	Identifier string
	Nonce      uint64
	Amount     *big.Int
	Type       core.ESDTType
}

// Account holds the arguments for a user account
type Account struct {
	Wallet dtos.WalletAddress
	Nonce  uint64
}

// GetSysContactDeployAddressBytes will return the system contract deploy address
func GetSysContactDeployAddressBytes(t *testing.T, nodeHandler process.NodeHandler) []byte {
	addressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	require.Nil(t, err)

	return addressBytes
}

// GetShardForAddress will return the shard of the address
func GetShardForAddress(cs ChainSimulator, address string) uint32 {
	nodeHandler := cs.GetNodeHandler(0)
	pubKey, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(address)
	return nodeHandler.GetShardCoordinator().ComputeId(pubKey)
}

// DeployContract will deploy a smart contract and return its address
func DeployContract(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	receiver []byte,
	data string,
	wasmPath string,
) []byte {
	data = wasm.GetSCCode(wasmPath) + "@" + vmTypeHex + "@" + codeMetadata + data

	tx := GenerateTransaction(sender, *nonce, receiver, ZeroValue, data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	*nonce++

	require.Nil(t, err)
	RequireSuccessfulTransaction(t, txResult)

	address := txResult.Logs.Events[0].Topics[0]
	require.NotNil(t, address)
	return address
}

// GenerateTransaction will generate a transaction object
func GenerateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}

// SendTransactionWithSuccess will send a transaction and expect successful execution and return the result
func SendTransactionWithSuccess(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	receiver []byte,
	value *big.Int,
	data string,
	gasLimit uint64,
) *transaction.ApiTransactionResult {
	txResult := SendTransaction(t, cs, sender, nonce, receiver, value, data, gasLimit)
	RequireSuccessfulTransaction(t, txResult)
	return txResult
}

// SendTransaction will send a transaction and return the result
func SendTransaction(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	receiver []byte,
	value *big.Int,
	data string,
	gasLimit uint64,
) *transaction.ApiTransactionResult {
	tx := GenerateTransaction(sender, *nonce, receiver, value, data, gasLimit)
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	*nonce++
	require.Nil(t, err)

	return txResult
}

// RequireSuccessfulTransaction require that the transaction doesn't have signal error event
func RequireSuccessfulTransaction(t *testing.T, txResult *transaction.ApiTransactionResult) {
	require.NotNil(t, txResult)
	event := getEvent(txResult.Logs, signalError)
	if event != nil {
		require.Fail(t, string(event.Topics[1]))
	}
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
}

// RequireSignalError require that the transaction has specific signal error
func RequireSignalError(t *testing.T, txResult *transaction.ApiTransactionResult, error string) {
	require.NotNil(t, txResult)
	event := getEvent(txResult.Logs, signalError)
	if event == nil {
		require.Fail(t, "%s event not found", signalError)
		return
	}
	require.Equal(t, error, string(event.Topics[1]))
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
}

// RequireInternalVMError require that the transaction has specific invernal vm error
func RequireInternalVMError(t *testing.T, txResult *transaction.ApiTransactionResult, error string) {
	require.NotNil(t, txResult)
	event := getEvent(txResult.Logs, internalVMError)
	if event == nil {
		require.Fail(t, "%s event not found", internalVMError)
		return
	}
	require.Contains(t, string(event.Data), error)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
}

func getEvent(logs *transaction.ApiLogs, eventID string) *transaction.Events {
	if logs == nil || len(logs.Events) == 0 {
		return nil
	}

	for _, event := range logs.Events {
		if event.Identifier == eventID {
			return event
		}
	}
	return nil
}

// RequireAccountHasToken checks if the account has the amount of tokens (can also be zero)
func RequireAccountHasToken(
	t *testing.T,
	cs ChainSimulator,
	token string,
	address string,
	value *big.Int,
) {
	addressShardID := GetShardForAddress(cs, address)
	tokens, _, err := cs.GetNodeHandler(addressShardID).GetFacadeHandler().GetAllESDTTokens(address, dataApi.AccountQueryOptions{})
	require.Nil(t, err)

	tokenData, found := tokens[token]

	if value.Cmp(big.NewInt(0)) == 0 {
		require.False(t, found)
		return
	}
	require.True(t, found, fmt.Sprintf("%s token not found", token))
	require.Equal(t, tokenData.Value, value)
}

// TransferESDT will transfer the amount of esdt token to an address
func TransferESDT(
	t *testing.T,
	cs ChainSimulator,
	sender, receiver []byte,
	nonce *uint64,
	token string,
	amount *big.Int,
	args ...[]byte,
) {
	esdtTransferArgs := core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" + hex.EncodeToString(amount.Bytes())
	for _, arg := range args {
		esdtTransferArgs = esdtTransferArgs +
			"@" + hex.EncodeToString(arg)
	}
	txResult := SendTransaction(t, cs, sender, nonce, receiver, ZeroValue, esdtTransferArgs, uint64(5000000))
	RequireSuccessfulTransaction(t, txResult)
}

// TransferESDTNFT will transfer the amount of NFT/SFT token to an address
func TransferESDTNFT(
	t *testing.T,
	cs ChainSimulator,
	sender, receiver []byte,
	nonce *uint64,
	token string,
	tokenNonce uint64,
	amount *big.Int,
	args ...[]byte,
) {
	esdtNftTransferArgs :=
		core.BuiltInFunctionESDTNFTTransfer +
			"@" + hex.EncodeToString([]byte(token)) +
			"@" + hex.EncodeToString(big.NewInt(int64(tokenNonce)).Bytes()) +
			"@" + hex.EncodeToString(amount.Bytes()) +
			"@" + hex.EncodeToString(receiver)
	for _, arg := range args {
		esdtNftTransferArgs = esdtNftTransferArgs +
			"@" + hex.EncodeToString(arg)
	}
	txResult := SendTransaction(t, cs, sender, nonce, sender, ZeroValue, esdtNftTransferArgs, uint64(5000000))
	RequireSuccessfulTransaction(t, txResult)
}

// IssueFungible will issue a fungible token
func IssueFungible(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	issueCost *big.Int,
	tokenName string,
	tokenTicker string,
	numDecimals int,
	supply *big.Int,
) string {
	issueArgs := "issue" +
		"@" + hex.EncodeToString([]byte(tokenName)) +
		"@" + hex.EncodeToString([]byte(tokenTicker)) +
		"@" + hex.EncodeToString(supply.Bytes()) +
		"@" + fmt.Sprintf("%X", numDecimals) +
		"@" + hex.EncodeToString([]byte("canAddSpecialRoles")) +
		"@" + hex.EncodeToString([]byte("true"))
	txResult := SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))
	RequireSuccessfulTransaction(t, txResult)

	return GetIssuedEsdtIdentifier(t, cs, tokenTicker, core.FungibleESDT)
}

// GetIssuedEsdtIdentifier will return the token identifier for and issued token
func GetIssuedEsdtIdentifier(t *testing.T, cs ChainSimulator, ticker string, tokenType string) string {
	esdtScAddressShardId := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(vm.ESDTSCAddress)
	issuedTokens, err := cs.GetNodeHandler(esdtScAddressShardId).GetFacadeHandler().GetAllIssuedESDTs(tokenType)
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1, "no issued tokens found of type %s", tokenType)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, ticker) {
			return issuedToken
		}
	}

	require.Fail(t, "could not find the issued token")
	return ""
}

// InitAddressesAndSysAccState will initialize system account state and other addresses if provided
func InitAddressesAndSysAccState(
	t *testing.T,
	cs ChainSimulator,
	initialAddresses ...string,
) {
	addressesState := []*dtos.AddressState{
		{
			Address: ESDTSystemAccount,
		},
	}
	for _, address := range initialAddresses {
		addressesState = append(addressesState,
			&dtos.AddressState{
				Address: address,
				Balance: "10000000000000000000000",
			},
		)
	}
	err := cs.SetStateMultiple(addressesState)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

// SetEsdtInWallet will add token key in wallet storage without adding key in system account
func SetEsdtInWallet(
	t *testing.T,
	cs ChainSimulator,
	wallet dtos.WalletAddress,
	token string,
	tokenNonce uint64,
	tokenData esdt.ESDigitalToken,
) {
	marshalledTokenData, err := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer().Marshal(&tokenData)
	require.NoError(t, err)

	nonce := ""
	if tokenNonce != 0 {
		nonce = hex.EncodeToString(big.NewInt(0).SetUint64(tokenNonce).Bytes())
	}
	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix+core.ESDTKeyIdentifier+token)) + nonce
	tokenValue := hex.EncodeToString(marshalledTokenData)
	keyValueMap := map[string]string{
		tokenKey: tokenValue,
	}
	err = cs.SetKeyValueForAddress(wallet.Bech32, keyValueMap)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

// RegisterAndSetAllRoles will issue an esdt collection with all roles enabled
func RegisterAndSetAllRoles(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	issueCost *big.Int,
	esdtName string,
	esdtTicker string,
	tokenType string,
	numDecimals int,
) string {
	esdtType := getTokenRegisterType(tokenType)
	registerArgs := "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte(esdtName)) +
		"@" + hex.EncodeToString([]byte(esdtTicker)) +
		"@" + hex.EncodeToString([]byte(esdtType)) +
		"@" + fmt.Sprintf("%02X", numDecimals)
	SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, registerArgs, uint64(60000000))

	return GetIssuedEsdtIdentifier(t, cs, esdtTicker, tokenType)
}

// RegisterAndSetAllRolesDynamic will issue a dynamic esdt collection with all roles enabled
func RegisterAndSetAllRolesDynamic(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	issueCost *big.Int,
	esdtName string,
	esdtTicker string,
	tokenType string,
	numDecimals int,
) string {
	esdtType := getTokenRegisterType(tokenType)
	registerArgs := "registerAndSetAllRolesDynamic" +
		"@" + hex.EncodeToString([]byte(esdtName)) +
		"@" + hex.EncodeToString([]byte(esdtTicker)) +
		"@" + hex.EncodeToString([]byte(esdtType))
	if numDecimals != 0 {
		registerArgs += "@" + fmt.Sprintf("%02X", numDecimals)
	}
	SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, registerArgs, uint64(60000000))

	return GetIssuedEsdtIdentifier(t, cs, esdtTicker, tokenType)
}

func getTokenRegisterType(tokenType string) string {
	switch tokenType {
	case core.FungibleESDT:
		return "FNG"
	case core.NonFungibleESDT, core.NonFungibleESDTv2, core.DynamicNFTESDT:
		return "NFT"
	case core.SemiFungibleESDT, core.DynamicSFTESDT:
		return "SFT"
	case core.MetaESDT, core.DynamicMetaESDT:
		return "META"
	}
	return ""
}
