package chainSimulator

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	vmTypeHex                               = "0500"
	codeMetadata                            = "0500"
	minGasPrice                             = 1000000000
	txVersion                               = 1
	mockTxSignature                         = "sig"
	deposit                                 = "deposit"
	multiEsdtTransfer                       = "MultiESDTNFTTransfer"
	maxNumOfBlocksToGenerateWhenExecutingTx = 1
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
	Identifier []byte
	Nonce      uint64
	Amount     *big.Int
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

	tx := GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	*nonce++

	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

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
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	if txResult.Logs != nil && txResult.Logs.Events != nil && len(txResult.Logs.Events) > 0 {
		require.NotEqual(t, signalError, txResult.Logs.Events[0].Identifier)
	}

	return txResult
}

// RegisterAndSetAllRoles will issue an esdt token with all roles enabled
func RegisterAndSetAllRoles(
	t *testing.T,
	cs ChainSimulator,
	nodeHandler process.NodeHandler,
	sender []byte,
	nonce *uint64,
	issueCost *big.Int,
	esdtName string,
	esdtTicker string,
	esdtType string,
	numDecimals int,
) string {
	registerArgs := "registerAndSetAllRoles" +
		"@" + hex.EncodeToString([]byte(esdtName)) +
		"@" + hex.EncodeToString([]byte(esdtTicker)) +
		"@" + hex.EncodeToString([]byte(esdtType)) +
		"@" + fmt.Sprintf("%02X", numDecimals)
	SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, registerArgs, uint64(60000000))

	issuedTokens, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.SemiFungibleESDT)
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, esdtTicker) {
			return issuedToken
		}
	}

	require.Fail(t, "could not issue semi fungible")
	return ""
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
	SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))

	issuedTokens, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.FungibleESDT)
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, tokenTicker) {
			return issuedToken
		}
	}

	require.Fail(t, "could not issue fungible")
	return ""
}

// IssueSemiFungible will issue a semi fungible token
func IssueSemiFungible(
	t *testing.T,
	cs ChainSimulator,
	nodeHandler process.NodeHandler,
	sender []byte,
	nonce *uint64,
	issueCost *big.Int,
	tokenName string,
	tokenTicker string,
) string {
	issueArgs := "issueSemiFungible" +
		"@" + hex.EncodeToString([]byte(tokenName)) +
		"@" + hex.EncodeToString([]byte(tokenTicker))
	SendTransaction(t, cs, sender, nonce, vm.ESDTSCAddress, issueCost, issueArgs, uint64(60000000))

	issuedTokens, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs(core.SemiFungibleESDT)
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(issuedTokens), 1)

	for _, issuedToken := range issuedTokens {
		if strings.Contains(issuedToken, tokenTicker) {
			return issuedToken
		}
	}

	require.Fail(t, "could not issue semi fungible")
	return ""
}

// Deposit will deposit tokens in a contract
func Deposit(
	t *testing.T,
	cs ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	tokens []ArgsDepositToken,
	receiver []byte,
) {
	require.True(t, len(tokens) > 0)

	depositArgs := multiEsdtTransfer +
		"@" + hex.EncodeToString(contract) +
		"@" + fmt.Sprintf("%02X", len(tokens))

	for _, token := range tokens {
		depositArgs = depositArgs +
			"@" + hex.EncodeToString(token.Identifier) +
			"@" + getTokenNonce(token.Nonce) +
			"@" + hex.EncodeToString(token.Amount.Bytes())
	}

	depositArgs = depositArgs +
		"@" + hex.EncodeToString([]byte(deposit)) +
		"@" + hex.EncodeToString(receiver)

	SendTransaction(t, cs, sender, nonce, sender, ZeroValue, depositArgs, uint64(20000000))
}

func getTokenNonce(nonce uint64) string {
	if nonce == 0 {
		return ""
	}
	return fmt.Sprintf("%02X", nonce)
}
