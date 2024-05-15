package esdt

import (
	"encoding/hex"
	"math/big"
	"strings"
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

func TestEsdt_IssueSft(t *testing.T) {
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
	//cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
	//	BypassTxSignatureCheck: false,
	//	TempDir:                t.TempDir(),
	//	PathToInitialConfig:    defaultPathToInitialConfig,
	//	NumOfShards:            3,
	//	GenesisTimestamp:       time.Now().Unix(),
	//	RoundDurationInMillis:  uint64(6000),
	//	RoundsPerEpoch: core.OptionalUint64{
	//		HasValue: true,
	//		Value:    30,
	//	},
	//	InitialEpoch:                3,
	//	ApiInterface:                api.NewNoApiInterface(),
	//	MinNodesPerShard:            1,
	//	MetaChainMinNodes:           1,
	//	ConsensusGroupSize:          1,
	//	MetaChainConsensusGroupSize: 1,
	//})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	data := "registerAndSetAllRoles@4f4354534654@4f4354534654@534654@"
	txResult := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, data, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	//esdts, _, err := nodeHandler.GetFacadeHandler().GetESDTsWithRole(wallet.Bech32, "ESDTRoleNFTCreate", coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.Equal(t, string(tokenIdentifier), esdts[0])
	//
	//tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.NotNil(t, tokens)
	//
	//keys, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(wallet.Bech32, coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.NotNil(t, keys)

	data = "ESDTNFTCreate@" + hex.EncodeToString(tokenIdentifier) + "@04d2@4f4354534654202331@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	tx := utils.GenerateTransaction(wallet.Bytes, 1, wallet.Bytes, big.NewInt(0), data, uint64(60000000))
	txRes, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txRes)
	require.False(t, "action is not allowed" == string(txRes.Logs.Events[0].Topics[1]))
	require.False(t, strings.Contains(string(txRes.Logs.Events[0].Topics[1]), "trie was not found for hash"))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)
}

func TestEsdt_RegisterSft(t *testing.T) {
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

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	data := "registerAndSetAllRoles@4f4354534654@4f4354534654@534654@"
	txResult := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, data, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	data = "ESDTNFTCreate@" + hex.EncodeToString(tokenIdentifier) + "@04d2@4f4354534654202331@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), data, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data2 := "registerAndSetAllRoles@5346544e32@5346544e32@534654@"
	txResult2 := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, data2, uint64(60000000))
	tokenIdentifier2 := txResult2.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data2 = "ESDTNFTCreate@" + hex.EncodeToString(tokenIdentifier2) + "@04d2@5346544e32202332@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), data2, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
}

func TestEsdt_IssueTwoSfts(t *testing.T) {
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

	err = cs.GenerateBlocks(3)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemEsdtAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	data := "issueSemiFungible@4f4354534654@4f4354534654"
	txResult := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, data, uint64(60000000))
	tokenIdentifier := txResult.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data = "setSpecialRole@" + hex.EncodeToString(tokenIdentifier) + "@" + hex.EncodeToString(wallet.Bytes) + "@45534454526f6c654e4654437265617465@45534454526f6c654e46544164645175616e74697479"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, big.NewInt(0), data, uint64(60000000))

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data = "ESDTNFTCreate@" + hex.EncodeToString(tokenIdentifier) + "@04d2@4f4354534654202331@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), data, uint64(60000000))

	tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 2)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data = "issueSemiFungible@5346544e32@5346544e32"
	txResult = utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, issueCost, data, uint64(60000000))
	tokenIdentifier2 := txResult.Logs.Events[0].Topics[0]

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data = "setSpecialRole@" + hex.EncodeToString(tokenIdentifier2) + "@" + hex.EncodeToString(wallet.Bytes) + "@45534454526f6c654e4654437265617465@45534454526f6c654e46544164645175616e74697479"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, systemEsdtAddress, big.NewInt(0), data, uint64(60000000))

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	data = "ESDTNFTCreate@" + hex.EncodeToString(tokenIdentifier2) + "@04d2@5346544e32202332@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	utils.SendTransaction(t, cs, wallet.Bytes, &nonce, wallet.Bytes, big.NewInt(0), data, uint64(60000000))

	tokens, _, err = nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, tokens)
	require.True(t, len(tokens) == 3)
}
