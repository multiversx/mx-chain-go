package vm

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/utils"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	adderWasmPath              = "../testdata/adder.wasm"
)

func TestSmartContract_IssueToken(t *testing.T) {
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

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "@10000000", adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 268435456, int(sum))

	issueCost := big.NewInt(5000000000000000000)
	tx1 := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))
	require.False(t, string(tx1.Logs.Events[0].Topics[1]) == "sending value to non payable contract")

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	esdts, err := nodeHandler.GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)
}

func TestSmartContract_IssueToken_MainChain(t *testing.T) {
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 1,
		GenesisTimestamp:            time.Now().Unix(),
		RoundDurationInMillis:       uint64(6000),
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            2,
		MetaChainMinNodes:           2,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t",
		},
	})
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(5)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	deployedContractAddress := utils.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, "@10000000", adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 268435456, int(sum))

	issueCost := big.NewInt(5000000000000000000)
	tx1 := utils.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, issueCost, "issue", uint64(60000000))
	require.False(t, string(tx1.Logs.Events[0].Topics[1]) == "sending value to non payable contract")

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	deployedAddrBech32, err := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Encode(deployedContractAddress)
	require.Nil(t, err)

	_ = deployedAddrBech32

	esdts, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().GetAllIssuedESDTs("FungibleESDT")
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) == 1)
}
