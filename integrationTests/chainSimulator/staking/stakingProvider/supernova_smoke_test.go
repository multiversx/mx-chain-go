package stakingProvider

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	chainTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func newSupernovaProviderSimulator(t *testing.T) chainTests.ChainSimulator {
	t.Helper()
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true, BypassCreateBlockTimeCheck: true,
		TempDir: t.TempDir(), PathToInitialConfig: defaultPathToInitialConfig,
		NumOfShards: 3, MinNodesPerShard: 3, MetaChainMinNodes: 3,
		NumNodesWaitingListMeta: 3, NumNodesWaitingListShard: 3,
		RoundDurationInMillis: 6000, SupernovaRoundDurationInMillis: 600,
		RoundsPerEpoch:          core.OptionalUint64{HasValue: true, Value: 20},
		SupernovaRoundsPerEpoch: core.OptionalUint64{HasValue: true, Value: 20},
		ApiInterface:            api.NewNoApiInterface(),
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 2
			cfg.RoundConfig.RoundActivations["SupernovaEnableRound"] = config.ActivationRoundByName{Round: "50"}
			// Keep the genesis validator set stable: these scenarios qualify cap/merge
			// behavior, while the jail suite exercises validator-set churn.
			for i := range cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch {
				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[i].MaxNumNodes = 24
				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[i].NodesToShufflePerShard = 0
			}
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
		},
	})
	require.NoError(t, err)
	t.Cleanup(cs.Close)
	chainTests.RequireSupernova(t, cs, 3, 200)
	return cs
}

func providerAmount(t *testing.T, cs chainTests.ChainSimulator, contract []byte, method string, args ...[]byte) *big.Int {
	t.Helper()
	out, err := executeQuery(cs, core.MetachainShardId, contract, method, args)
	require.NoError(t, err)
	require.Equal(t, "ok", out.ReturnCode, out.ReturnMessage)
	require.Len(t, out.ReturnData, 1)
	return new(big.Int).SetBytes(out.ReturnData[0])
}

func providerSuccess(t *testing.T, result *transaction.ApiTransactionResult) {
	t.Helper()
	require.Equal(t, transaction.TxStatusSuccess, result.Status)
	if result.Logs != nil {
		for _, event := range result.Logs.Events {
			require.NotEqual(t, core.SignalErrorOperation, event.Identifier, "%s", event.Topics)
		}
	}
}

func checkUnauthorizedMerge(t *testing.T, cs chainTests.ChainSimulator, owner dtos.WalletAddress, contract, key []byte) {
	t.Helper()
	before := providerAmount(t, cs, contract, "getTotalActiveStake")
	tx := chainTests.GenerateTransaction(owner.Bytes, 1, vm.DelegationManagerSCAddress, big.NewInt(0), "mergeValidatorToDelegationWithWhitelist@"+hex.EncodeToString(contract), gasLimitForMergeOperation)
	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 40)
	require.NoError(t, err)
	require.NotNil(t, result.Logs)
	found := false
	for _, event := range result.Logs.Events {
		if event.Identifier == core.SignalErrorOperation {
			found = true
		}
	}
	require.True(t, found, "merge without whitelisting must fail")
	require.Equal(t, before, providerAmount(t, cs, contract, "getTotalActiveStake"))
	require.Equal(t, owner.Bytes, getBLSKeyOwner(t, cs.GetNodeHandler(core.MetachainShardId), key))
}

func checkMergedFunds(t *testing.T, cs chainTests.ChainSimulator, owner, merged dtos.WalletAddress, contract []byte, nonce uint64) {
	t.Helper()
	stake := new(big.Int).Mul(chainTests.OneEGLD, big.NewInt(2600))
	require.Equal(t, new(big.Int).Mul(stake, big.NewInt(2)), providerAmount(t, cs, contract, "getTotalActiveStake"))
	for _, address := range [][]byte{owner.Bytes, merged.Bytes} {
		require.Equal(t, stake, providerAmount(t, cs, contract, "getUserActiveStake", address))
	}
	out, err := executeQuery(cs, core.MetachainShardId, contract, "getContractConfig", nil)
	require.NoError(t, err)
	require.Equal(t, owner.Bytes, out.ReturnData[0])
	// The merged validator keeps the right to recover its surplus stake.
	value := new(big.Int).Mul(chainTests.OneEGLD, big.NewInt(100))
	tx := chainTests.GenerateTransaction(merged.Bytes, nonce, contract, big.NewInt(0), "unDelegate@"+hex.EncodeToString(value.Bytes()), gasLimitForUndelegateOperation)
	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 40)
	require.NoError(t, err)
	providerSuccess(t, result)
	require.Equal(t, value, providerAmount(t, cs, contract, "getUserUnStakedValue", merged.Bytes))
	for i := 0; i < 160 && providerAmount(t, cs, contract, "getUserUnBondable", merged.Bytes).Cmp(value) != 0; i++ {
		require.NoError(t, cs.GenerateBlocks(1))
	}
	require.Equal(t, value, providerAmount(t, cs, contract, "getUserUnBondable", merged.Bytes))
	before, err := cs.GetAccount(merged)
	require.NoError(t, err)
	tx = chainTests.GenerateTransaction(merged.Bytes, nonce+1, contract, big.NewInt(0), "withdraw", gasLimitForUndelegateOperation)
	result, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 40)
	require.NoError(t, err)
	providerSuccess(t, result)
	require.NoError(t, cs.GenerateBlocks(8))
	after, err := cs.GetAccount(merged)
	require.NoError(t, err)
	balance, ok := new(big.Int).SetString(before.Balance, 10)
	require.True(t, ok)
	fee, ok := new(big.Int).SetString(result.Fee, 10)
	require.True(t, ok)
	require.Equal(t, new(big.Int).Sub(new(big.Int).Add(balance, value), fee).String(), after.Balance)
	require.Equal(t, stake, providerAmount(t, cs, contract, "getUserActiveStake", owner.Bytes))
	require.Equal(t, new(big.Int).Sub(stake, value), providerAmount(t, cs, contract, "getUserActiveStake", merged.Bytes))
}
