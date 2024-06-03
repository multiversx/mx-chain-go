package bridge

import (
	"encoding/hex"
	"fmt"
	"testing"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-sdk-abi-go/abi"
	"github.com/stretchr/testify/require"
)

const (
	deposit = "deposit"
)

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

// DeploySovereignBridgeSetup will deploy all bridge contracts
// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func DeploySovereignBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	esdtSafeArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, chainSim.ZeroValue, "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

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
	transferData *sovereign.TransferData,
) {
	require.True(t, len(tokens) > 0)

	dtaCodec := cs.GetNodeHandler(core.SovereignChainShardId).GetRunTypeComponents().DataCodecHandler()

	depositArgs := core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(contract) +
		"@" + fmt.Sprintf("%02X", len(tokens))

	for _, token := range tokens {
		depositArgs = depositArgs +
			"@" + hex.EncodeToString([]byte(token.Identifier)) +
			"@" + getTokenNonce(t, dtaCodec, token.Nonce) +
			"@" + hex.EncodeToString(token.Amount.Bytes())
	}

	transferDataArg := getTransferDataArg(t, dtaCodec, transferData)
	depositArgs = depositArgs +
		"@" + hex.EncodeToString([]byte(deposit)) +
		"@" + hex.EncodeToString(receiver) +
		transferDataArg

	chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

func getTokenNonce(t *testing.T, dataCodec dataCodec.SovereignDataCodec, nonce uint64) string {
	if nonce == 0 {
		return ""
	}

	hexNonce, err := dataCodec.Serialize([]any{&abi.U64Value{Value: nonce}})
	require.Nil(t, err)
	return hexNonce
}

func getTransferDataArg(t *testing.T, dataCodec dataCodec.SovereignDataCodec, transferData *sovereign.TransferData) string {
	if transferData == nil {
		return ""
	}

	arguments := make([]abi.SingleValue, len(transferData.Args))
	for i, arg := range transferData.Args {
		arguments[i] = &abi.BytesValue{Value: arg}
	}
	transferDataValue := &abi.MultiValue{
		Items: []any{
			&abi.U64Value{Value: transferData.GasLimit},
			&abi.BytesValue{Value: transferData.Function},
			&abi.ListValue{Items: arguments},
		},
	}
	args, err := dataCodec.Serialize([]any{transferDataValue})
	require.Nil(t, err)

	return "@" + args
}
