package bridge

import (
	"encoding/hex"
	"math/big"
	"testing"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

// DeploySovereignBridgeSetup will deploy all bridge contracts
func DeploySovereignBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	systemScAddress, err := chainSim.GetSysAccBytesAddress(nodeHandler)
	require.Nil(t, err)

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
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
	}
}
