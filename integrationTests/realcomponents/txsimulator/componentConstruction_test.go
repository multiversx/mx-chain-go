package txsimulator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/realcomponents/txsimulator")

func TestTransactionSimulationComponentConstructionOnMetachain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cfg := testscommon.CreateTestConfigs(t, "../../../cmd/node/config")
	cfg.EpochConfig.EnableEpochs.ESDTEnableEpoch = 0
	cfg.EpochConfig.EnableEpochs.BuiltInFunctionsEnableEpoch = 0
	cfg.PreferencesConfig.Preferences.DestinationShardAsObserver = "metachain" // the problem was only on the metachain

	pr := realcomponents.NewProcessorRunner(t, *cfg)
	defer pr.Close(t)

	senderShardID := uint32(0) // doesn't matter
	alice := pr.GenerateAddress(senderShardID)
	log.Info("generated address",
		"alice", pr.CoreComponents.AddressPubKeyConverter().SilentEncode(alice, log),
		"shard", senderShardID,
	)

	rootHash, err := pr.StateComponents.AccountsAdapter().Commit()
	require.Nil(t, err)

	err = pr.DataComponents.Blockchain().SetCurrentBlockHeaderAndRootHash(
		&block.MetaBlock{
			Nonce:    1,
			RootHash: rootHash,
		}, rootHash)
	require.Nil(t, err)

	issueCost, _ := big.NewInt(0).SetString(pr.Config.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost, 10)

	txForSimulation := &transaction.Transaction{
		Nonce:    pr.GetUserAccount(t, alice).GetNonce(),
		Value:    issueCost,
		RcvAddr:  vm.ESDTSCAddress,
		SndAddr:  alice,
		GasPrice: pr.CoreComponents.EconomicsData().MinGasPrice(),
		GasLimit: 60_000_000,
		Data:     []byte(fmt.Sprintf("issue@%x@%x@0100@02", "token", "tkn")),
		ChainID:  []byte(pr.CoreComponents.ChainID()),
		Version:  1,
	}

	_, err = pr.ProcessComponents.TransactionSimulatorProcessor().ProcessTx(txForSimulation)
	assert.Nil(t, err)
	assert.Equal(t, 0, pr.StateComponents.AccountsAdapter().JournalLen()) // state for processing should not be dirtied
}

func TestTransactionSimulationComponentConstructionOnShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cfg := testscommon.CreateTestConfigs(t, "../../../cmd/node/config")
	cfg.EpochConfig.EnableEpochs.SCDeployEnableEpoch = 0
	cfg.PreferencesConfig.Preferences.DestinationShardAsObserver = "0"
	cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{
		{
			StartEpoch: 0,
			Version:    "v1.4",
		},
	}

	pr := realcomponents.NewProcessorRunner(t, *cfg)
	defer pr.Close(t)

	senderShardID := pr.ProcessComponents.ShardCoordinator().SelfId()
	alice := pr.GenerateAddress(senderShardID)
	log.Info("generated address",
		"alice", pr.CoreComponents.AddressPubKeyConverter().SilentEncode(alice, log),
		"shard", senderShardID,
	)

	// mint some tokens for alice
	mintValue, _ := big.NewInt(0).SetString("1000000000000000000", 10) // 1 EGLD
	pr.AddBalanceToAccount(t, alice, mintValue)

	// deploy the contract
	txDeploy, hash := pr.CreateDeploySCTx(t, alice, "../testdata/adder/adder.wasm", 3000000, []string{"01"})
	err := pr.ExecuteTransactionAsScheduled(t, txDeploy)
	require.Nil(t, err)

	// get the contract address from logs
	logs, found := pr.ProcessComponents.TxLogsProcessor().GetLogFromCache(hash)
	require.True(t, found)
	events := logs.GetLogEvents()
	require.Equal(t, 1, len(events))
	require.Equal(t, "SCDeploy", string(events[0].GetIdentifier()))
	contractAddress := events[0].GetAddress()

	rootHash, err := pr.StateComponents.AccountsAdapter().Commit()
	require.Nil(t, err)

	err = pr.DataComponents.Blockchain().SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    1,
			RootHash: rootHash,
		}, rootHash)
	require.Nil(t, err)

	txForSimulation := &transaction.Transaction{
		Nonce:    pr.GetUserAccount(t, alice).GetNonce(),
		Value:    big.NewInt(0),
		RcvAddr:  contractAddress,
		SndAddr:  alice,
		GasPrice: pr.CoreComponents.EconomicsData().MinGasPrice(),
		GasLimit: 3_000_000,
		Data:     []byte("add@06"),
		ChainID:  []byte(pr.CoreComponents.ChainID()),
		Version:  1,
	}

	_, err = pr.ProcessComponents.TransactionSimulatorProcessor().ProcessTx(txForSimulation)
	assert.Nil(t, err)
	assert.Equal(t, 0, pr.StateComponents.AccountsAdapter().JournalLen()) // state for processing should not be dirtied
}
