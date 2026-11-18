package smoke

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/require"
)

func TestSupernovaSmokeSignedKeyValue(t *testing.T) {
	f := newFixture(t)
	owner, outsider := f.wallet(0), f.wallet(0)
	node := f.cs.GetNodeHandler(0)
	storage := func() map[string]string {
		values, _, err := node.GetFacadeHandler().GetKeyValuePairs(f.address(owner.Address).Bech32, apiData.AccountQueryOptions{})
		require.NoError(t, err)
		return values
	}
	execute := func(sender *integrationTests.TestWalletAccount, tx *transaction.Transaction, success bool) *transaction.ApiTransactionResult {
		before := f.account(sender.Address)
		result := f.execute(tx)
		if success {
			requireExecutionSuccess(t, result)
		} else {
			requireExecutionFailure(t, result)
		}
		require.Equal(t, before.Nonce+1, f.account(sender.Address).Nonce)
		require.Equal(t, new(big.Int).Sub(amount(t, before.Balance), amount(t, result.Fee)).String(), f.account(sender.Address).Balance)
		return result
	}
	key := hexArg([]byte("smoke:key"))
	initial := storage()
	for _, value := range []string{"01", "0203", ""} {
		execute(owner, f.tx(owner, owner.Address, 0, core.BuiltInFunctionSaveKeyValue+"@"+key+"@"+value, 10_000_000), true)
		values := storage()
		if value == "" {
			require.NotContains(t, values, key, "empty value deletes the key")
		} else {
			require.Equal(t, value, values[key])
		}
	}
	require.Equal(t, initial, storage())
	beforeOwner, beforeStorage := f.account(owner.Address), storage()
	failed := execute(outsider, f.tx(outsider, owner.Address, 0, core.BuiltInFunctionSaveKeyValue+"@"+key+"@ff", 10_000_000), false)
	require.Contains(t, executionMessages(failed), "not the owner of the account")
	require.Equal(t, beforeOwner, f.account(owner.Address))
	require.Equal(t, beforeStorage, storage())

	// Budget exactly one fresh pair using the active schedule. A successful
	// single-pair control proves the batch reaches storage before running out.
	schedule, err := node.GetFacadeHandler().GetGasConfigs()
	require.NoError(t, err)
	base := schedule["BuiltInCost"]["SaveKeyValue"]
	persist, store := schedule["BaseOperationCost"]["PersistPerByte"], schedule["BaseOperationCost"]["StorePerByte"]
	require.Positive(t, base)
	require.Positive(t, persist)
	require.Positive(t, store)
	const keyBytes, valueBytes = uint64(4), uint64(4)
	onePair := (keyBytes+valueBytes)*persist + valueBytes*store
	firstPair := "@01000000@02000000"
	batch := firstPair + "@03000000@04000000@05000000@06000000"
	budgeted := func(arguments string, pairs uint64) *transaction.Transaction {
		tx := f.tx(owner, owner.Address, 0, core.BuiltInFunctionSaveKeyValue+arguments, 10_000_000)
		tx.GasLimit = node.GetCoreComponents().EconomicsData().ComputeGasLimit(tx) + base + pairs*onePair
		f.sign(tx, owner, nil, nil)
		return tx
	}
	execute(owner, budgeted(firstPair, 1), true)
	require.Equal(t, "02000000", storage()["01000000"])
	execute(owner, f.tx(owner, owner.Address, 0, core.BuiltInFunctionSaveKeyValue+"@01000000@", 10_000_000), true)
	require.Equal(t, initial, storage())
	beforeOwner, beforeStorage = f.account(owner.Address), storage()
	failed = execute(owner, budgeted(batch, 1), false)
	require.Contains(t, executionMessages(failed), "not enough gas")
	require.Equal(t, beforeOwner.RootHash, f.account(owner.Address).RootHash)
	require.Equal(t, beforeStorage, storage(), "a later pair failure must roll back earlier writes")
	// The same batch with sufficient gas writes all three pairs.
	execute(owner, budgeted(batch, 3), true)
	require.Equal(t, map[string]string{"01000000": "02000000", "03000000": "04000000", "05000000": "06000000"}, storage())
}
