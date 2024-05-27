package basic

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	adderWasmPath              = "../testdata/adder.wasm"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))

func TestSmartContract_Adder(t *testing.T) {
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

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	initSum := big.NewInt(123)
	initArgs := "@" + hex.EncodeToString(initSum.Bytes())
	deployedContractAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, initArgs, adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum, sum)

	addedSum := big.NewInt(10)
	addTxData := "add@" + hex.EncodeToString(addedSum.Bytes())
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), addTxData, uint64(10000000))

	res, _, err = nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum = big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum.Add(initSum, addedSum), sum)
}
