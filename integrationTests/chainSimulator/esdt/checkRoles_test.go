package esdt

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	common "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))

func TestGuildSC_CheckTransferRole(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       2,
		MetaChainMinNodes:      2,
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.MetachainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	err = cs.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	wallet, err := cs.GenerateAndMintWalletAddress(0, initialMinting)
	require.Nil(t, err)

	contractCode := getSCCode("guild-sc.wasm")
	deployData := contractCode + "@0500@0500" +
		"@55544b54312d616237356565" +
		"@" + hex.EncodeToString(big.NewInt(100000000000000).Bytes()) +
		"@000000000000000005007f452bddaab558c882723217e392abb5feda8eed1679" +
		"@" + hex.EncodeToString(wallet.Bytes) +
		"@0000000000000582" +
		"@" + hex.EncodeToString(big.NewInt(200).Bytes()) +
		"@" + hex.EncodeToString(wallet.Bytes)
	deployTx := common.GenerateTransaction(wallet.Bytes, 0, systemScAddress, big.NewInt(0), deployData, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	contractAddress := txResult.Logs.Events[0].Topics[0]

	txData := "registerFarmToken@55544b4641524d@55544b4641524d@08"
	tx := common.GenerateTransaction(wallet.Bytes, 1, contractAddress, big.NewInt(5000000000000000000), txData, uint64(100000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData = "registerUnbondToken@55544b554e424f4e44@55544b554e424e44@08"
	tx = common.GenerateTransaction(wallet.Bytes, 2, contractAddress, big.NewInt(5000000000000000000), txData, uint64(100000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData = "setTransferRoleFarmToken"
	tx = common.GenerateTransaction(wallet.Bytes, 3, contractAddress, big.NewInt(0), txData, uint64(100860000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData = "setTransferRoleUnbondToken"
	tx = common.GenerateTransaction(wallet.Bytes, 4, contractAddress, big.NewInt(0), txData, uint64(100890000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	txData = "checkLocalRolesSet"
	tx = common.GenerateTransaction(wallet.Bytes, 5, contractAddress, big.NewInt(0), txData, uint64(100000000))
	txResult, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	require.False(t, string(txResult.Logs.Events[0].Topics[1]) == "Transfer role not set for farm token")
}

func getSCCode(fileName string) string {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}
