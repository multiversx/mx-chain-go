package bridge

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	issuePaymentCost         = "50000000000000000" // esdt-safe contract without header-verifier checks
	esdtSafeWasmPath         = "testdata/mvx-esdt-safe.wasm"
	enshrineEsdtSafeWasmPath = "testdata/enshrine-esdt-safe.wasm"
	//enshrine esdt-safe contract without checks for prefix or issue cost paid for new tokens
	simpleEnshrineEsdtSafeWasmPath = "testdata/simple-enshrine-esdt-safe.wasm"
	feeMarketWasmPath              = "testdata/fee-market.wasm"
)

// ArgsEsdtSafe holds the arguments for esdt safe contract argument
type ArgsEsdtSafe struct {
	ChainPrefix       string
	IssuePaymentToken string
}

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
	OwnerAccount     chainSim.Account
}

type transferData struct {
	GasLimit uint64
	Function []byte
	Args     [][]byte
}

func initOwnerAndSysAccState(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress string,
	argsEsdtSafe ArgsEsdtSafe,
) {
	chainSim.InitAddressesAndSysAccState(t, cs, ownerAddress)

	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + argsEsdtSafe.IssuePaymentToken))
	err := cs.SetKeyValueForAddress(chainSim.ESDTSystemAccount,
		map[string]string{
			tokenKey: "0400",
		},
	)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

type esdtSafeDeployFunc func(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress []byte,
	nonce *uint64,
	systemContractDeploy []byte,
	argsEsdtSafe ArgsEsdtSafe,
	contractWasmPath string,
) []byte

// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func deployBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress string,
	argsEsdtSafe ArgsEsdtSafe,
	deployEsdtSafeContract esdtSafeDeployFunc,
	contractWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(0)

	systemContractDeploy := chainSim.GetSysContactDeployAddressBytes(t, nodeHandler)
	ownerAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(ownerAddress)
	require.Nil(t, err)
	nonce := uint64(0)

	esdtSafeAddress := deployEsdtSafeContract(t, cs, ownerAddrBytes, &nonce, systemContractDeploy, argsEsdtSafe, contractWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@00" // no fee
	feeMarketAddress := chainSim.DeployContract(t, cs, ownerAddrBytes, &nonce, systemContractDeploy, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransactionWithSuccess(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
		OwnerAccount: chainSim.Account{
			Wallet: dtos.WalletAddress{Bech32: ownerAddress, Bytes: ownerAddrBytes},
			Nonce:  nonce,
		},
	}
}

func esdtSafeContract(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress []byte,
	nonce *uint64,
	systemContractDeploy []byte,
	_ ArgsEsdtSafe,
	contractWasmPath string,
) []byte {
	esdtSafeArgs := "@00000000000000000500f3c3c6f64f8a20ced95a45a7de596ab08f9082d0f8ef" // dummy header_verifier_address
	return chainSim.DeployContract(t, cs, ownerAddress, nonce, systemContractDeploy, esdtSafeArgs, contractWasmPath)
}

func enshrineEsdtSafeContract(
	t *testing.T,
	cs chainSim.ChainSimulator,
	ownerAddress []byte,
	nonce *uint64,
	systemContractDeploy []byte,
	argsEsdtSafe ArgsEsdtSafe,
	contractWasmPath string,
) []byte {
	esdtSafeArgs := "@" + // is_sovereign_chain
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.IssuePaymentToken)) + hex.EncodeToString([]byte(argsEsdtSafe.IssuePaymentToken)) + // token identifier
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.ChainPrefix)) + hex.EncodeToString([]byte(argsEsdtSafe.ChainPrefix)) // prefix
	return chainSim.DeployContract(t, cs, ownerAddress, nonce, systemContractDeploy, esdtSafeArgs, contractWasmPath)
}

// deposit will deposit tokens in the bridge sc safe contract
func deposit(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	tokens []chainSim.ArgsDepositToken,
	receiver []byte,
) *transaction.ApiTransactionResult {
	require.True(t, len(tokens) > 0)

	depositArgs := core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(contract) +
		"@" + fmt.Sprintf("%02X", len(tokens))

	for _, token := range tokens {
		depositArgs = depositArgs +
			"@" + hex.EncodeToString([]byte(token.Identifier)) +
			"@" + getTokenNonce(token.Nonce) +
			"@" + hex.EncodeToString(token.Amount.Bytes())
	}

	depositArgs = depositArgs +
		"@" + hex.EncodeToString([]byte("deposit")) +
		"@" + hex.EncodeToString(receiver)

	return chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

func registerSovereignNewTokens(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	issuePaymentToken string,
	tokens []string,
) {
	lenSovTokens := big.NewInt(int64(len(tokens)))
	issueCostOneToken, _ := big.NewInt(0).SetString(issuePaymentCost, 10)
	issueTotalCost := big.NewInt(0).Mul(issueCostOneToken, lenSovTokens)

	registerData := core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte(issuePaymentToken)) +
		"@" + hex.EncodeToString(issueTotalCost.Bytes()) +
		"@" + hex.EncodeToString([]byte("registerNewTokenID"))
	for _, token := range tokens {
		registerData = registerData +
			"@" + hex.EncodeToString([]byte(token))
	}
	txResult := chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, registerData, uint64(10000000))
	chainSim.RequireSuccessfulTransaction(t, txResult)
}

func executeOperation(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	receiver []byte,
	nonce *uint64,
	esdtSafeAddress []byte,
	bridgedInTokens []chainSim.ArgsDepositToken,
	originalSender []byte,
	transferData *transferData,
) *transaction.ApiTransactionResult {
	executeBridgeOpsData := "executeBridgeOps" +
		"@" + generateRandomHash() + //dummy hash
		"@" + // operation
		hex.EncodeToString(receiver) + // receiver address
		lengthOn4Bytes(len(bridgedInTokens)) + // nr of tokens
		getTokenDataArgs(wallet.Bytes, bridgedInTokens) + // tokens encoded arg
		getUint64Bytes(0) + // event nonce
		hex.EncodeToString(originalSender) + // sender address from other chain
		getTransferDataArgs(transferData)
	return chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(100000000))
}

func generateRandomHash() string {
	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func getTokenDataArgs(creator []byte, tokens []chainSim.ArgsDepositToken) string {
	var arg string
	for _, token := range tokens {
		arg = arg +
			lengthOn4Bytes(len(token.Identifier)) + // length of token identifier
			hex.EncodeToString([]byte(token.Identifier)) + //token identifier
			getUint64Bytes(token.Nonce) + // nonce
			fmt.Sprintf("%02x", uint32(token.Type)) + // type
			lengthOn4Bytes(len(token.Amount.Bytes())) + // length of amount
			hex.EncodeToString(token.Amount.Bytes()) + // amount
			"00" + // not frozen
			lengthOn4Bytes(0) + // length of hash
			lengthOn4Bytes(4) + // length of name
			hex.EncodeToString([]byte("ESDT")) + // name
			lengthOn4Bytes(0) + // length of attributes
			hex.EncodeToString(creator) + // creator
			lengthOn4Bytes(0) + //length of royalties
			lengthOn4Bytes(0) // length of uris
	}
	return arg
}

func getTransferDataArgs(transferData *transferData) string {
	if transferData == nil {
		return "00"
	}

	transferDataArgs := "01" +
		getUint64Bytes(transferData.GasLimit) +
		lengthOn4Bytes(len(transferData.Function)) +
		hex.EncodeToString(transferData.Function) +
		lengthOn4Bytes(len(transferData.Args))
	for _, arg := range transferData.Args {
		transferDataArgs = transferDataArgs +
			lengthOn4Bytes(len(arg)) +
			hex.EncodeToString(arg)
	}
	return transferDataArgs
}

func getTokenIdentifier(token chainSim.ArgsDepositToken) string {
	if token.Nonce == 0 {
		return token.Identifier
	}
	return token.Identifier + "-" + fmt.Sprintf("%02x", token.Nonce)
}

func getTokenNonce(nonce uint64) string {
	if nonce == 0 {
		return ""
	}

	hexStr := fmt.Sprintf("%X", nonce)
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	return hexStr
}

func getUint64Bytes(number uint64) string {
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, number)
	return hex.EncodeToString(nonceBytes)
}

func groupTokens(tokens []chainSim.ArgsDepositToken) []chainSim.ArgsDepositToken {
	groupMap := make(map[string]*chainSim.ArgsDepositToken)

	for _, token := range tokens {
		key := fmt.Sprintf("%s:%d", token.Identifier, token.Nonce)
		if existingToken, found := groupMap[key]; found {
			existingToken.Amount.Add(existingToken.Amount, token.Amount)
		} else {
			newAmount := new(big.Int).Set(token.Amount)
			groupMap[key] = &chainSim.ArgsDepositToken{
				Identifier: token.Identifier,
				Nonce:      token.Nonce,
				Amount:     newAmount,
				Type:       token.Type,
			}
		}
	}

	result := make([]chainSim.ArgsDepositToken, 0, len(groupMap))
	for _, token := range groupMap {
		result = append(result, *token)
	}

	return result
}

func lengthOn4Bytes(number int) string {
	numberBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(numberBytes, uint32(number))
	return hex.EncodeToString(numberBytes)
}
