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
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"
)

func TestOutgoingOperations(t *testing.T) {
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
			ConsensusGroupSize:     1,
			InitialRound:           100,
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
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	walletNonce := uint64(0)
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	esdtSafeSovArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeSovAddress := utils.DeployContract(t, cs, wallet.Bytes, &walletNonce, systemScAddress, esdtSafeSovArgs, esdtSafeWasmPath)

	feeMarketSovArgs := "@" + hex.EncodeToString(esdtSafeSovAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@555344432d333530633465" + // usdc_token_id
		"@5745474c442d613238633539" // wegld_token_id
	feeMarketSovAddress := utils.DeployContract(t, cs, wallet.Bytes, &walletNonce, systemScAddress, feeMarketSovArgs, feeMarketWasmPath)

	setSovFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketSovAddress)
	utils.SendTransaction(t, cs, wallet.Bytes, &walletNonce, esdtSafeSovAddress, big.NewInt(0), setSovFeeMarketAddressData, uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &walletNonce, feeMarketSovAddress, big.NewInt(0), "disableFee", uint64(10000000))

	utils.SendTransaction(t, cs, wallet.Bytes, &walletNonce, esdtSafeSovAddress, big.NewInt(0), "unpause", uint64(10000000))

	userNonce := uint64(0)
	userWallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, big.NewInt(0))
	require.Nil(t, err)

	utils.SendTransaction(t, cs, wallet.Bytes, &walletNonce, userWallet.Bytes, big.NewInt(0).Mul(oneEgld, big.NewInt(10)), "", uint64(50000))

	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(userWallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)

	issueArgs := "issue@536f7665726569676e546b6e@53564e@5c003ec0dd34509ad40000@12@63616e4164645370656369616c526f6c6573@74727565"
	txResult := utils.SendTransaction(t, cs, userWallet.Bytes, &userNonce, systemEsdtAddress, issueCost, issueArgs, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(userWallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)

	amount := big.NewInt(500000000000000000)
	depositArgs := "MultiESDTNFTTransfer" +
		"@" + hex.EncodeToString(esdtSafeSovAddress) +
		"@01" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" +
		"@" + hex.EncodeToString(amount.Bytes()) +
		"@6465706f736974" + //deposit
		"@" + hex.EncodeToString(userWallet.Bytes)
	res := utils.SendTransaction(t, cs, userWallet.Bytes, &userNonce, userWallet.Bytes, big.NewInt(0), depositArgs, uint64(20000000))
	require.NotNil(t, res)
}
