package txsimulator

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/realcomponents/txsimulator")

func TestTransactionSimulationComponentConstruction(t *testing.T) {
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
