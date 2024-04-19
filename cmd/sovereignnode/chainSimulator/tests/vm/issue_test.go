package vm

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	chainSim "github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/process"
	sovChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
)

func TestSmartContract_IssueToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	epochConfig, economicsConfig, sovereignExtraConfig, err := sovChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	cs, err := sovChainSimulator.NewSovereignChainSimulator(sovereignExtraConfig, chainSim.ArgsChainSimulator{
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

	data := getSCCode("../testdata/adder.wasm") + "@0500@0500@10000000"
	tx0 := generateTransaction(wallet.Bytes, 0, systemScAddress, big.NewInt(0), data, uint64(60000000))
	txRes, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx0, 100)
	require.Nil(t, err)
	require.NotNil(t, txRes)
	deployedContractAddress := txRes.Logs.Events[0].Topics[0]

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0]).Int64()
	require.Equal(t, 268435456, int(sum))

	issueCost := big.NewInt(50000000000000000)
	tx1 := generateTransaction(wallet.Bytes, 1, deployedContractAddress, issueCost, "issue", uint64(60000000))
	txRes, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx1, 100)
	require.Nil(t, err)
	require.NotNil(t, txRes)
	require.False(t, string(txRes.Logs.Events[0].Topics[1]) == "sending value to non payable contract")

	esdtSC, err := cs.GetNodeHandler(0).GetStateComponents().AccountsAdapter().LoadAccount(vm.ESDTSCAddress)
	require.Nil(t, err)
	esdtSCAcc, ok := esdtSC.(state.UserAccountHandler)
	require.True(t, ok)

	existingCode := vmcommon.CodeMetadataFromBytes(esdtSCAcc.GetCodeMetadata())
	require.True(t, existingCode.Readable)
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

	transferValue := big.NewInt(0).Set(value)
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     transferValue,
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
