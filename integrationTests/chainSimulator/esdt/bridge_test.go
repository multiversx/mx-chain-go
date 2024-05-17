package esdt

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))

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

	whiteListedAddresses := make([]string, 0)
	whiteListedAddresses = append(whiteListedAddresses, "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5")

	numOfShards := uint32(1)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
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
			cfg.SystemSCConfig.ESDTSystemSCConfig.WhiteListedCrossChainMintAddresses = whiteListedAddresses
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.MetachainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

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
	esdtSafeCode := getSCCode("testdata/esdt-safe.wasm")
	esdtSafeArgs := "@0500@0500" +
		"@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	deployEsdtSafeData := esdtSafeCode + esdtSafeArgs
	deployTx := generateTransaction(wallet.Bytes, 0, systemScAddress, big.NewInt(0), deployEsdtSafeData, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	esdtSafeAddress := txResult.Logs.Events[0].Topics[0]

	feeMarketCode := getSCCode("testdata/fee-market.wasm")
	feeMarketArgs := "@0500@0500" +
		"@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	deployFeeMarketData := feeMarketCode + feeMarketArgs
	deployTx = generateTransaction(wallet.Bytes, 1, systemScAddress, big.NewInt(0), deployFeeMarketData, uint64(200000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	feeMarketAddress := txResult.Logs.Events[0].Topics[0]

	deployTx = generateTransaction(wallet.Bytes, 2, esdtSafeAddress, big.NewInt(0), "setFeeMarketAddress@"+hex.EncodeToString(feeMarketAddress), uint64(10000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	deployTx = generateTransaction(wallet.Bytes, 3, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	deployTx = generateTransaction(wallet.Bytes, 4, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	executeBridgeOpsData := "executeBridgeOps" +
		"@de96b8d3842668aad676f915f545403b3e706f8f724cefb0c15b728e83864ce7" + //dummy hash
		"@" + // operation
		hex.EncodeToString(wallet.Bytes) + // receiver address
		"00000001" + // nr of tokens
		"00000010" + // length of token identifier
		"736f76312d534f56542d356438663536" + //token identifier
		"0000000000000000000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" + // nonce + token data
		"0000000000000000" + // event nonce
		hex.EncodeToString(wallet.Bytes) + // sender address from other chain
		"00" // no transfer data
	tx := generateTransaction(wallet.Bytes, 5, esdtSafeAddress, big.NewInt(0), executeBridgeOpsData, uint64(50000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	expectedMintValue, _ := big.NewInt(0).SetString("123000000000000000000", 10)
	esdts, _, err := cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Contains(t, esdts, "sov1-SOVT-5d8f56")
	require.Equal(t, esdts["sov1-SOVT-5d8f56"].Value, expectedMintValue)

	amountToBurn, _ := big.NewInt(0).SetString("120000000000000000000", 10)
	depositTxData := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(esdtSafeAddress) +
		"@01" +
		"@736f76312d534f56542d356438663536" + // token
		"@" + // nonce
		"@" + hex.EncodeToString(amountToBurn.Bytes()) + // amount
		"@6465706f736974" + //deposit
		"@" + hex.EncodeToString(wallet.Bytes) // receiver on the other chain
	tx = generateTransaction(wallet.Bytes, 6, wallet.Bytes, big.NewInt(0), depositTxData, uint64(50000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	expectedRemainingValue := expectedMintValue.Sub(expectedMintValue, amountToBurn)
	esdts, _, err = cs.GetNodeHandler(0).GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Contains(t, esdts, "sov1-SOVT-5d8f56")
	require.Equal(t, esdts["sov1-SOVT-5d8f56"].Value, expectedRemainingValue)
}

func getSCCode(fileName string) string {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}

func generateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	minGasPrice := uint64(1000000000)
	txVersion := uint32(1)
	mockTxSignature := "sig"

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
