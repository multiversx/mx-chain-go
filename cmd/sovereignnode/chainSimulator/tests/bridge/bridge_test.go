package bridge

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

func TestBridge(t *testing.T) {
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
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	//nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	//systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")
	//
	//oneEgld := big.NewInt(1000000000000000000)
	//initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(10))
	//wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	//require.Nil(t, err)
	//walletAddress := hex.EncodeToString(wallet.Bytes)

	//esdtSafeData := getSCCode("../testdata/esdt-safe.wasm") + "@0500@0500@@@" + walletAddress
	//tx0 := chainSimulator.GenerateTransaction(wallet.Bytes, 0, systemScAddress, big.NewInt(0), esdtSafeData, uint64(200000000))
	//txRes, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx0, 1)
	//require.Nil(t, err)
	//require.NotNil(t, txRes)
	//esdtSafeAddress := txRes.Logs.Events[0].Topics[0]
	//
	//
	//issueCost := big.NewInt(50000000000000000)
	//tx1 := chainSimulator.GenerateTransaction(wallet.Bytes, 1, deployedContractAddress, issueCost, "issue", uint64(60000000))
	//txRes, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx1, 100)
	//require.Nil(t, err)
	//require.NotNil(t, txRes)
	//require.False(t, string(txRes.Logs.Events[0].Topics[1]) == "sending value to non payable contract")
}

func getSCCode(fileName string) string {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}
