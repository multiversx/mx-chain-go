package bridge

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
)

const (
	deposit = "deposit"
)

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

// DeployBridgeSetup will deploy all bridge contracts
// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func DeployBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	address []byte,
	nonce *uint64,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(0)

	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)

	esdtSafeArgs := "@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(address) // initiator_address
	esdtSafeAddress := chainSim.DeployContract(t, cs, address, nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSim.DeployContract(t, cs, address, nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransaction(t, cs, address, nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, address, nonce, feeMarketAddress, chainSim.ZeroValue, "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, address, nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
	}
}

// Deposit will deposit tokens in the bridge sc safe contract
func Deposit(
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
		"@" + hex.EncodeToString([]byte(deposit)) +
		"@" + hex.EncodeToString(receiver)

	chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

func getTokenNonce(nonce uint64) string {
	if nonce == 0 {
		return ""
	}
	return fmt.Sprintf("%02X", nonce)
}
