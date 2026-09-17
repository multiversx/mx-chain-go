package smoke

import (
	"math/big"
	"strings"
	"testing"

	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

// Read the executed callback itself: the originating transaction's API result
// does not include all downstream logs. Bound the search to blocks this call produced.
func (f *fixture) callbackResult(firstBlock uint64, hash string, from, to []byte) *transaction.ApiTransactionResult {
	f.t.Helper()
	node := f.cs.GetNodeHandler(0)
	lastBlock := node.GetChainHandler().GetLastExecutionResult().GetHeaderNonce()
	// The next proposal persists the latest execution result for block API reads.
	require.NoError(f.t, f.cs.GenerateBlocks(1))
	matches := make(map[string]*transaction.ApiTransactionResult)
	for nonce := firstBlock; nonce <= lastBlock; nonce++ {
		block, err := node.GetFacadeHandler().GetBlockByNonce(nonce, apiData.BlockQueryOptions{WithTransactions: true, WithLogs: true})
		require.NoError(f.t, err, "block %d, range %d..%d", nonce, firstBlock, lastBlock)
		for _, mb := range block.MiniBlocks {
			for _, tx := range mb.Transactions {
				if tx.OriginalTransactionHash == hash && tx.Sender == f.address(from).Bech32 && tx.Receiver == f.address(to).Bech32 {
					matches[tx.Hash] = tx
				}
			}
		}
	}
	require.Len(f.t, matches, 1, "expected exactly one executed callback")
	for _, tx := range matches {
		return tx
	}
	return nil
}

func executionMessages(result *transaction.ApiTransactionResult) string {
	messages := result.ReturnMessage
	if result.Logs != nil {
		for _, event := range result.Logs.Events {
			for _, topic := range event.Topics {
				messages += " " + string(topic)
			}
		}
	}
	return messages
}

func TestSupernovaSmokeCrossShardPromises(t *testing.T) {
	f := newFixture(t)
	sender, remoteOwner := f.wallet(0), f.wallet(1)
	forwarder, _ := f.deployArtifact(sender, "testdata/promise-probe.wasm", "0506", "")
	vault, _ := f.deployArtifact(remoteOwner, "../../vm/txsFee/testdata/forwarderQueue/vault-promises.wasm", "0506", "")
	for _, scenario := range []struct {
		name, endpoint, remoteError, callbackError string
		remoteGas                                  uint64
		callbackMode                               byte
	}{
		{"destination-failure", "missing_endpoint", "invalid function", "", 10_000_000, 0},
		{"destination-out-of-gas", "accept_funds", "not enough gas", "", 1_000, 0},
		{"callback-failure", "accept_funds", "", "smoke callback failure", 10_000_000, 1},
		{"callback-out-of-gas", "accept_funds", "", "gas", 10_000_000, 2},
		{"success", "accept_funds", "", "", 10_000_000, 0},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			f := &fixture{t: t, cs: f.cs}
			firstBlock, _, _ := f.cs.GetNodeHandler(0).GetChainHandler().GetLastExecutedBlockInfo()
			firstBlock++
			beforeSender, beforeForwarder, beforeVault := f.account(sender.Address), f.account(forwarder), f.account(vault)
			callbacksBefore := f.query(forwarder, "getCallbackCount")
			data := "start@" + hexArg(vault) + "@" + hexArg([]byte(scenario.endpoint)) + "@" + hexArg(new(big.Int).SetUint64(scenario.remoteGas).Bytes()) + "@" + hexArg([]byte{scenario.callbackMode})
			tx := f.tx(sender, forwarder, 100, data, 50_000_000)
			initial := f.success(tx)
			result, err := f.cs.GetNodeHandler(0).GetFacadeHandler().GetTransaction(initial.Hash, true)
			require.NoError(t, err)
			var remote *transaction.ApiTransactionResult
			for _, scr := range result.SmartContractResults {
				if scr.RcvAddr == f.address(vault).Bech32 {
					require.Nil(t, remote, "exactly one remote call")
					remote, err = f.cs.GetNodeHandler(1).GetFacadeHandler().GetTransaction(scr.Hash, true)
					require.NoError(t, err)
				}
			}
			require.NotNil(t, remote)
			callback := f.callbackResult(firstBlock, result.Hash, vault, forwarder)
			if scenario.remoteError != "" {
				requireExecutionFailure(t, remote)
				require.Contains(t, executionMessages(remote), scenario.remoteError)
				require.Equal(t, beforeVault.Balance, f.account(vault).Balance)
				require.Equal(t, beforeVault.RootHash, f.account(vault).RootHash)
				// A failed remote call refunds its immediate caller contract.
				require.Equal(t, new(big.Int).Add(amount(t, beforeForwarder.Balance), big.NewInt(100)).String(), f.account(forwarder).Balance)
				require.Equal(t, "100", callback.Value)
			} else {
				requireExecutionSuccess(t, remote)
				require.Equal(t, new(big.Int).Add(amount(t, beforeVault.Balance), big.NewInt(100)).String(), f.account(vault).Balance)
				require.Equal(t, beforeForwarder.Balance, f.account(forwarder).Balance)
				require.Equal(t, "0", callback.Value)
			}
			if scenario.callbackError != "" {
				requireExecutionFailure(t, callback)
				require.Contains(t, strings.ToLower(executionMessages(callback)), scenario.callbackError)
				require.Equal(t, callbacksBefore, f.query(forwarder, "getCallbackCount"), "write before failing callback must roll back")
			} else {
				requireExecutionSuccess(t, callback)
				require.Equal(t, new(big.Int).Add(callbacksBefore, big.NewInt(1)), f.query(forwarder, "getCallbackCount"))
			}
			// All attached value is held exactly once by the vault or refunded forwarder;
			// the original user pays that value plus the final reported transaction fee.
			require.Equal(t, new(big.Int).Sub(amount(t, beforeSender.Balance), new(big.Int).Add(amount(t, result.Fee), big.NewInt(100))).String(), f.account(sender.Address).Balance)
			require.Equal(t, beforeSender.Nonce+1, f.account(sender.Address).Nonce)
			countAfter := f.query(forwarder, "getCallbackCount")
			f.rejected(tx, sender.Address, forwarder, vault)
			require.Equal(t, countAfter, f.query(forwarder, "getCallbackCount"))
		})
	}
}
