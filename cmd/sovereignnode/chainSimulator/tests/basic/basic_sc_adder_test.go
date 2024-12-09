package basic

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
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

func TestSovereignChainSimulator_SmartContract_Adder(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: true,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	time.Sleep(time.Second) // wait for VM to be ready for processing queries

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, chainSim.InitialAmount)
	require.Nil(t, err)
	nonce := uint64(0)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	initSum := big.NewInt(123)
	initArgs := "@" + hex.EncodeToString(initSum.Bytes())
	deployedContractAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, initArgs, adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum, sum)

	addedSum := big.NewInt(10)
	addTxData := "add@" + hex.EncodeToString(addedSum.Bytes())
	chainSim.SendTransactionWithSuccess(t, cs, wallet.Bytes, &nonce, deployedContractAddress, chainSim.ZeroValue, addTxData, uint64(10000000))

	res, _, err = nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	require.Equal(t, chainSim.OkReturnCode, res.ReturnCode)
	sum = big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum.Add(initSum, addedSum), sum)
}
