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

	// OkReturnCode the const for the ok return code
	OkReturnCode = "ok"
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

// GetSysAccBytesAddress will return the system account bytes address
func GetSysAccBytesAddress(t *testing.T, nodeHandler process.NodeHandler) []byte {
	addressBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	require.Nil(t, err)

	return addressBytes
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
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	if txResult.Logs != nil && len(txResult.Logs.Events) > 0 {
		if txResult.Logs.Events[0].Identifier == signalError {
			require.Fail(t, string(txResult.Logs.Events[0].Topics[1]))
		}
	}
}

// RequireSignalError require that the transaction has specific signal error
func RequireSignalError(t *testing.T, txResult *transaction.ApiTransactionResult, error string) {
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	if txResult.Logs != nil && len(txResult.Logs.Events) > 0 {
		if txResult.Logs.Events[0].Identifier != signalError {
			require.Fail(t, "signal error event not found")
		}
		require.Equal(t, error, string(txResult.Logs.Events[0].Topics[1]))
	}
}

// RequireAccountHasToken checks if the account has the amount of tokens (can also be zero)
func RequireAccountHasToken(
	t *testing.T,
	cs ChainSimulator,
	token string,
	address string,
	value *big.Int,
) {
	nodeHandler := cs.GetNodeHandler(0)

	pubKey, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(address)
	require.Nil(t, err)

	addressShardID := nodeHandler.GetShardCoordinator().ComputeId(pubKey)
	tokens, _, err := cs.GetNodeHandler(addressShardID).GetFacadeHandler().GetAllESDTTokens(address, dataApi.AccountQueryOptions{})
	require.Nil(t, err)

	tokenData, found := tokens[token]

	if value.Cmp(big.NewInt(0)) == 0 {
		require.False(t, found)
		return
	}
	require.True(t, found)
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
) {
	esdtTransferArgs := core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte(token)) +
		"@" + hex.EncodeToString(amount.Bytes())
	txResult := SendTransaction(t, cs, sender, nonce, receiver, ZeroValue, esdtTransferArgs, uint64(5000000))
	RequireSuccessfulTransaction(t, txResult)
}

// IssueFungible will issue a fungible token
func IssueFungible(
	t *testing.T,
	cs ChainSimulator,
	nodeHandler process.NodeHandler,
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

	return getEsdtIdentifier(t, nodeHandler, tokenTicker, core.FungibleESDT)
}

func getEsdtIdentifier(t *testing.T, nodeHandler process.NodeHandler, ticker string, tokenType string) string {
	issuedTokens, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(tokenType)
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, ticker) {
			return issuedToken
		}
	}

	require.Fail(t, "could not issue semi fungible")
	return ""
}

// GetEsdtInWallet will add token key in wallet storage without adding key in system account
func GetEsdtInWallet(
	t *testing.T,
	cs ChainSimulator,
	wallet dtos.WalletAddress,
	token string,
	tokenData esdt.ESDigitalToken,
) {
	marshalledTokenData, err := cs.GetNodeHandler(0).GetCoreComponents().InternalMarshalizer().Marshal(&tokenData)
	require.NoError(t, err)

	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + token))
	tokenValue := hex.EncodeToString(marshalledTokenData)
	keyValueMap := map[string]string{
		tokenKey: tokenValue,
	}
	err = cs.SetKeyValueForAddress(wallet.Bech32, keyValueMap)
	require.NoError(t, err)
}
