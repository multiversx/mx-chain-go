package esdt

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
)

func TestEsdt_IssueSft(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)

	data := "registerAndSetAllRoles@4f4354534654@4f4354534654@534654@"
	tx := utils.GenerateTransaction(wallet.Bytes, 0, systemEsdtAddress, big.NewInt(5000000000000000000), data, uint64(60000000))
	txRes, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txRes)
	//tokenIdentifer := string(txRes.Logs.Events[0].Topics[0])
	tokenIdentifierHex := hex.EncodeToString(txRes.Logs.Events[0].Topics[0])

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	//esdts, _, err := nodeHandler.GetFacadeHandler().GetESDTsWithRole(wallet.Bech32, "ESDTRoleNFTCreate", coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.Equal(t, tokenIdentifer, esdts[0])
	//
	//tokens, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(wallet.Bech32, coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.NotNil(t, tokens)
	//
	//keys, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(wallet.Bech32, coreAPI.AccountQueryOptions{})
	//require.Nil(t, err)
	//require.NotNil(t, keys)

	data = "ESDTNFTCreate@" + tokenIdentifierHex + "@04d2@4f4354534654202331@09c4@@746167733a746167312c746167323b77686174657665724b65793a776861746576657256616c7565@68747470733a2f2f6d656469612e72656d61726b61626c652e746f6f6c732f4465764e65742f52656d61726b61626c65546f6f6c732e706e67"
	tx = utils.GenerateTransaction(wallet.Bytes, 1, wallet.Bytes, big.NewInt(0), data, uint64(60000000))
	txRes, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txRes)
	require.False(t, "action is not allowed" == string(txRes.Logs.Events[0].Topics[1]))
}
