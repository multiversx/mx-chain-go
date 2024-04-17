package bridge

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/utils"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	esdtSafeWasmPath           = "../../../../node/config/genesisContracts/esdt-safe.wasm"
	feeMarketWasmPath          = "../../../../node/config/genesisContracts/fee-market.wasm"
	multiSigWasmPath           = "../testdata/multisig-verifier.wasm"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(10))
var issueCost = big.NewInt(50000000000000000)

func TestBridge(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              core.OptionalUint64{},
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	epochConfig, economicsConfig, sovereignExtraConfig, err := sovereignChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	scs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignExtraConfig: *sovereignExtraConfig,
		ChainSimulatorArgs: chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			MetaChainMinNodes:      0,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EconomicsConfig = economicsConfig
				cfg.EpochConfig = epochConfig
				cfg.GeneralConfig.SovereignConfig = *sovereignExtraConfig
				cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
				cfg.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, scs)

	defer scs.Close()

	nodeHandler := cs.GetNodeHandler(core.MetachainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	nonce := uint64(0)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeArgs := "@" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) + // initiator_address
		"@" // signers
	esdtSafeAddress := deployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketAddress := deployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	sendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, setFeeMarketAddressData, uint64(10000000))

	sendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, "disableFee", uint64(10000000))

	sendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, "unpause", uint64(10000000))

	multiSigAddress := deployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", multiSigWasmPath)

	setEsdtSafeAddressData := "setEsdtSafeAddress" +
		"@" + hex.EncodeToString(esdtSafeAddress)
	sendTransaction(t, cs, wallet.Bytes, &nonce, multiSigAddress, setEsdtSafeAddressData, uint64(10000000))

	setMultiSigAddressData := "setMultisigAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	sendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, setMultiSigAddressData, uint64(10000000))

	setSovereignBridgeAddressData := "setSovereignBridgeAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	sendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, setSovereignBridgeAddressData, uint64(10000000))

	// ---------------------------------------------------------------------------------------------------------------

	sovNodeHandler := scs.GetNodeHandler(core.SovereignChainShardId)
	sovSystemScAddress, _ := sovNodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	sovNonce := uint64(0)

	sovWallet, err := scs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeSovArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(sovWallet.Bytes) + // initiator_address
		"@" // signers
	esdtSafeSovAddress := deployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, esdtSafeSovArgs, esdtSafeWasmPath)

	feeMarketSovArgs := "@" + hex.EncodeToString(esdtSafeSovAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketSovAddress := deployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, feeMarketSovArgs, feeMarketWasmPath)

	setSovFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketSovAddress)
	sendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, setSovFeeMarketAddressData, uint64(10000000))

	sendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, "disableFee", uint64(10000000))

	sendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, "unpause", uint64(10000000))
}

func deployContract(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string, wasmPath string) []byte {
	data = utils.GetSCCode(wasmPath) + "@0500@0500" + data

	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
	*nonce++

	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	address := txResult.Logs.Events[0].Topics[0]
	require.NotNil(t, address)
	return address
}

func sendTransaction(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string, gasLimit uint64) {
	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, gasLimit)
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
	*nonce++
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
}

//
//func deployFeeMarketContract(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string) (address []byte) {
//	data = utils.GetSCCode(feeMarketWasmPath) + "@0500@0500" + data
//
//	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(200000000))
//	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
//	*nonce++
//
//	require.Nil(t, err)
//	require.NotNil(t, txResult)
//	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
//
//	feeMarketAddress := txResult.Logs.Events[0].Topics[0]
//	require.NotNil(t, feeMarketAddress)
//	return feeMarketAddress
//}
//
//func deployMultiSigContract(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string) (address []byte) {
//	data = utils.GetSCCode(multiSigWasmPath) + "@0500@0500" + data
//
//	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(200000000))
//	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
//	*nonce++
//
//	require.Nil(t, err)
//	require.NotNil(t, txResult)
//	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
//
//	multiSigAddress := txResult.Logs.Events[0].Topics[0]
//	require.NotNil(t, multiSigAddress)
//	return multiSigAddress
//}

//func setFeeMarketAddress(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string) {
//	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(10000000))
//	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
//	*nonce++
//	require.Nil(t, err)
//	require.NotNil(t, txResult)
//	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
//}
//
//func disableFeeMarketContract(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string) {
//	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(10000000))
//	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
//	*nonce++
//	require.Nil(t, err)
//	require.NotNil(t, txResult)
//	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
//}
//
//func unpauseEsdtSafeContract(t *testing.T, cs *chainSimulator.Simulator, sender []byte, nonce *uint64, receiver []byte, data string) {
//	tx := chainSimulator.GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(10000000))
//	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 1)
//	*nonce++
//	require.Nil(t, err)
//	require.NotNil(t, txResult)
//	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
//}
