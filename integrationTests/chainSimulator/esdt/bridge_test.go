package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSimulator2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

// TODO: Delete this in feat/chain-go-sdk when merged and use existing functions
// ArgsBridgeSetup holds the arguments for bridge setup
type BridgeSetupAddresses struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

func TestChainSimulator_ExecuteBridgeOpsWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	whiteListedAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.VirtualMachine.Execution.TransferAndExecuteByUserAddresses = []string{whiteListedAddress}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	initialAddress := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	initialAddrBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddress)
	require.Nil(t, err)
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: initialAddress,
			Balance: "10000000000000000000000",
		},
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t", // init sys account
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	wallet := dtos.WalletAddress{Bech32: initialAddress, Bytes: initialAddrBytes}
	bridgeSetup := deployMainChainBridgeContracts(t, cs, wallet, "testdata/esdt-safe.wasm", "testdata/fee-market.wasm")

	esdtSafeEncoded, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(bridgeSetup.ESDTSafeAddress)
	require.Equal(t, esdtSafeEncoded, whiteListedAddress)

	// We will deposit a prefixed token from a sovereign chain to the main chain,
	// expecting these tokens to be minted by the whitelisted ESDT safe sc and transferred to our address.
	depositToken := "sov1-SOVT-5d8f56"
	expectedMintValue, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		"00000001" + // nr of tokens
		"00000010" + // length of token identifier
		hex.EncodeToString([]byte(depositToken)) + //token identifier
		"0000000000000000000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" + // nonce + token data
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	_ = sendTx(t, cs, wallet.Bytes, 5, bridgeSetup.ESDTSafeAddress, big.NewInt(0), executeBridgeOpsData, uint64(50000000))
	requireAccountHasToken(t, cs, depositToken, wallet.Bech32, expectedMintValue)

	// Now we will transfer a part of the previously mentioned tokens from main chain to a sovereign chain,
	// expecting these tokens to be burned by the whitelisted ESDT safe sc.
	amountToBurn, _ := big.NewInt(0).SetString("120000000000000000000", 10)
	remainingValueAfterBridge := expectedMintValue.Sub(expectedMintValue, amountToBurn)
	depositTxData := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(bridgeSetup.ESDTSafeAddress) +
		"@01" +
		fmt.Sprintf("@%x", depositToken) + // token
		"@" + // nonce
		"@" + hex.EncodeToString(amountToBurn.Bytes()) + // amount
		"@6465706f736974" + //deposit
		"@" + hex.EncodeToString(wallet.Bytes) // receiver on the other chain

	_ = sendTx(t, cs, wallet.Bytes, 6, wallet.Bytes, big.NewInt(0), depositTxData, uint64(50000000))
	requireAccountHasToken(t, cs, depositToken, wallet.Bech32, remainingValueAfterBridge)

	// Send some of the bridged prefixed tokens to another address
	receiver := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	receiverBytes, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(receiver)
	require.Nil(t, err)

	receivedTokens := big.NewInt(11)
	transferESDT(t, cs, wallet.Bytes, receiverBytes, 7, receivedTokens, depositToken)
	requireAccountHasToken(t, cs, depositToken, receiver, receivedTokens)
	requireAccountHasToken(t, cs, depositToken, initialAddress, big.NewInt(0).Sub(remainingValueAfterBridge, receivedTokens))
}

// TODO: Delete this in feat/chain-go-sdk when merged and use existing functions
func deployMainChainBridgeContracts(
	t *testing.T,
	cs chainSimulator2.ChainSimulator,
	wallet dtos.WalletAddress,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) BridgeSetupAddresses {
	nodeHandler := cs.GetNodeHandler(core.MetachainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	esdtSafeCode := wasm.GetSCCode(esdtSafeWasmPath)
	esdtSafeArgs := "@0500@0500" +
		"@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	deployEsdtSafeData := esdtSafeCode + esdtSafeArgs
	txResult := sendTx(t, cs, wallet.Bytes, 0, systemScAddress, big.NewInt(0), deployEsdtSafeData, uint64(200000000))
	esdtSafeAddress := txResult.Logs.Events[0].Topics[0]

	feeMarketCode := wasm.GetSCCode(feeMarketWasmPath)
	feeMarketArgs := "@0500@0500" +
		"@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	deployFeeMarketData := feeMarketCode + feeMarketArgs
	txResult = sendTx(t, cs, wallet.Bytes, 1, systemScAddress, big.NewInt(0), deployFeeMarketData, uint64(200000000))
	feeMarketAddress := txResult.Logs.Events[0].Topics[0]

	_ = sendTx(t, cs, wallet.Bytes, 2, esdtSafeAddress, big.NewInt(0), "setFeeMarketAddress@"+hex.EncodeToString(feeMarketAddress), uint64(10000000))
	_ = sendTx(t, cs, wallet.Bytes, 3, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))
	_ = sendTx(t, cs, wallet.Bytes, 4, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	return BridgeSetupAddresses{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
	}
}

func sendTx(t *testing.T, cs chainSimulator2.ChainSimulator, sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.ApiTransactionResult {
	tx := chainSimulator2.GenerateTransaction(sender, nonce, receiver, value, data, gasLimit)
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	return txResult
}
