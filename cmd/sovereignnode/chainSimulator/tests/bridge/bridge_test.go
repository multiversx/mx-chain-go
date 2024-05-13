package bridge

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

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
	multiSigWasmPath           = "../testdata/multisigverifier.wasm"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))
var issueCost = big.NewInt(5000000000000000000)

func TestBridge(t *testing.T) {
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
		ConsensusGroupSize:          1,
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
			ConsensusGroupSize:     2,
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
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setFeeMarketAddressData, uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	multiSigAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "", multiSigWasmPath)

	setEsdtSafeAddressData := "setEsdtSafeAddress" +
		"@" + hex.EncodeToString(esdtSafeAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, multiSigAddress, big.NewInt(0), setEsdtSafeAddressData, uint64(10000000))

	setMultiSigAddressData := "setMultisigAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setMultiSigAddressData, uint64(10000000))

	setSovereignBridgeAddressData := "setSovereignBridgeAddress" +
		"@" + hex.EncodeToString(multiSigAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setSovereignBridgeAddressData, uint64(10000000))

	// ---------------------------------------------------------------------------------------------------------------

	sovNodeHandler := scs.GetNodeHandler(core.SovereignChainShardId)
	sovSystemScAddress, _ := sovNodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	sovNonce := uint64(0)

	sovWallet, err := scs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeSovArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(sovWallet.Bytes) // initiator_address
	esdtSafeSovAddress := utils.DeployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, esdtSafeSovArgs, esdtSafeWasmPath)

	feeMarketSovArgs := "@" + hex.EncodeToString(esdtSafeSovAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketSovAddress := utils.DeployContract(t, scs, sovWallet.Bytes, &sovNonce, sovSystemScAddress, feeMarketSovArgs, feeMarketWasmPath)

	setSovFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketSovAddress)
	utils.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, big.NewInt(0), setSovFeeMarketAddressData, uint64(10000000))

	utils.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, feeMarketSovAddress, big.NewInt(0), "disableFee", uint64(10000000))

	utils.SendTransaction(t, scs, sovWallet.Bytes, &sovNonce, esdtSafeSovAddress, big.NewInt(0), "unpause", uint64(10000000))
}

func Test_ESDTRoleBurnForAll(t *testing.T) {
	epochConfig, economicsConfig, sovereignExtraConfig, err := sovereignChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
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
			ConsensusGroupSize:     2,
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
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	systemEsdtScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")
	nonce := uint64(0)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), setFeeMarketAddressData, uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, big.NewInt(0), "disableFee", uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, big.NewInt(0), "unpause", uint64(10000000))

	issueArgs := "issue@536f7665726569676e546b6e@53564e@5c003ec0dd34509ad40000@12@63616e4164645370656369616c526f6c6573@74727565"
	txResult := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtScAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]
	require.True(t, len(tokenIdentifier) > 7)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
	depositArgs := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(esdtSafeAddress) +
		"@01" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + //nonce
		"@06aaf7c8516d0c0000" + //amount
		"@6465706f736974" + //deposit func
		"@" + hex.EncodeToString(wallet.Bytes) //receiver from other side
	txResult = utils.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))
	require.False(t, string(txResult.Logs.Events[0].Topics[1]) == "action is not allowed")
}
