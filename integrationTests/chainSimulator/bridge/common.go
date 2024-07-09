package bridge

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	issuePaymentCost = "50000000000000000"
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
	esdtSafeWasmPath string,
	argsEsdtSafe ArgsEsdtSafe,
	feeMarketWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(0)

	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)
	ownerAddrBytes, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode(ownerAddress)
	require.Nil(t, err)
	nonce := uint64(0)

	tokenKey := hex.EncodeToString([]byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + argsEsdtSafe.IssuePaymentToken))
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: ownerAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t", // init sys account
			Pairs: map[string]string{
				tokenKey: "0400",
			},
		},
	})
	require.Nil(t, err)

	esdtSafeArgs := "@" + // is_sovereign_chain
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.IssuePaymentToken)) + hex.EncodeToString([]byte(argsEsdtSafe.IssuePaymentToken)) + // token identifier
		"@01" + lengthOn4Bytes(len(argsEsdtSafe.ChainPrefix)) + hex.EncodeToString([]byte(argsEsdtSafe.ChainPrefix)) // prefix
	esdtSafeAddress := chainSim.DeployContract(t, cs, ownerAddrBytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSim.DeployContract(t, cs, ownerAddrBytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransaction(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, ownerAddrBytes, &nonce, feeMarketAddress, chainSim.ZeroValue, "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, ownerAddrBytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
		OwnerAccount: chainSim.Account{
			Wallet: dtos.WalletAddress{Bech32: ownerAddress, Bytes: ownerAddrBytes},
			Nonce:  nonce,
		},
	}
}

// This function will deposit tokens in the bridge sc safe contract
func deposit(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	tokens []chainSim.ArgsDepositToken,
	receiver []byte,
) {
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

	chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
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
	chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, registerData, uint64(10000000))
}

func executeMintOperation(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	nonce *uint64,
	esdtSafeAddress []byte,
	bridgedInTokens []chainSim.ArgsDepositToken,
) {
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		lengthOn4Bytes(len(bridgedInTokens)) + // nr of tokens
		getTokenDataArgs(wallet.Bytes, bridgedInTokens) + // tokens encoded arg
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	chainSim.SendTransaction(t, cs, wallet.Bytes, nonce, esdtSafeAddress, chainSim.ZeroValue, executeBridgeOpsData, uint64(50000000))
	for _, token := range groupTokens(bridgedInTokens) {
		chainSim.RequireAccountHasToken(t, cs, getTokenIdentifier(token), wallet.Bech32, token.Amount)
	}
}

func getTokenDataArgs(creator []byte, tokens []chainSim.ArgsDepositToken) string {
	var arg string
	for _, token := range tokens {
		arg = arg +
			lengthOn4Bytes(len(token.Identifier)) + // length of token identifier
			hex.EncodeToString([]byte(token.Identifier)) + //token identifier
			getNonceHex(token.Nonce) + // nonce
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
	return fmt.Sprintf("%02X", nonce)
}

func getNonceHex(nonce uint64) string {
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
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
