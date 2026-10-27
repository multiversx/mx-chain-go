package smoke

import (
	"math/big"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
)

func (f *fixture) deploy(owner *integrationTests.TestWalletAccount) ([]byte, string) {
	return f.deployArtifact(owner, "../relayedTx/testData/adder.wasm", "0102", "@00")
}

func (f *fixture) deployArtifact(owner *integrationTests.TestWalletAccount, path, metadata, arguments string) ([]byte, string) {
	f.t.Helper()
	code, err := os.ReadFile(path)
	require.NoError(f.t, err)
	result := f.success(f.tx(owner, make([]byte, 32), 0, hexArg(code)+"@0500@"+metadata+arguments, 300_000_000))
	require.NotNil(f.t, result.Logs)
	for _, event := range result.Logs.Events {
		if event.Identifier != "SCDeploy" {
			continue
		}
		address, err := f.cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(event.Address)
		require.NoError(f.t, err)
		return address, hexArg(code)
	}
	f.t.Fatal("missing SCDeploy event")
	return nil, ""
}

func TestSupernovaSmokeSimulationPreservesState(t *testing.T) {
	f := newFixture(t)
	owner, receiver := f.wallet(0), f.wallet(0)
	contract, _ := f.deploy(owner)
	for _, scenario := range []struct {
		name, call  string
		destination []byte
		value       int64
		success     bool
	}{
		{"transfer", "", receiver.Address, 100, true},
		{"storage-write", "add@01", contract, 0, true},
		{"failed-call", "missingEndpoint", contract, 0, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			f := &fixture{t: t, cs: f.cs}
			tx := f.tx(owner, scenario.destination, scenario.value, scenario.call, 10_000_000)
			facade := f.cs.GetNodeHandler(0).GetFacadeHandler()
			beforeOwner, beforeReceiver, beforeContract := f.account(owner.Address), f.account(receiver.Address), f.account(contract)
			storage, _, err := facade.GetKeyValuePairs(f.address(contract).Bech32, apiData.AccountQueryOptions{})
			require.NoError(t, err)
			roots := f.roots()
			require.NoError(t, facade.ValidateTransactionForSimulation(tx, true))
			result, err := facade.SimulateTransactionExecution(tx)
			require.NoError(t, err)
			require.NotNil(t, result)
			if scenario.success {
				require.Equal(t, transaction.TxStatusSuccess, result.Status)
			} else {
				require.NotEqual(t, transaction.TxStatusSuccess, result.Status)
			}
			require.Equal(t, roots, f.roots())
			require.Equal(t, beforeOwner, f.account(owner.Address))
			require.Equal(t, beforeReceiver, f.account(receiver.Address))
			require.Equal(t, beforeContract, f.account(contract))
			afterStorage, _, err := facade.GetKeyValuePairs(f.address(contract).Bech32, apiData.AccountQueryOptions{})
			require.NoError(t, err)
			require.Equal(t, storage, afterStorage)
			// Positive control: the exact same signed transaction can still execute once.
			executed := f.execute(tx)
			if scenario.success {
				require.Equal(t, transaction.TxStatusSuccess, executed.Status)
			} else {
				requireExecutionFailure(t, executed)
				require.Equal(t, beforeContract.RootHash, f.account(contract).RootHash)
				require.Equal(t, beforeContract.Balance, f.account(contract).Balance)
			}
			require.Equal(t, beforeOwner.Nonce+1, f.account(owner.Address).Nonce)
			if scenario.name == "transfer" {
				require.Equal(t, new(big.Int).Add(amount(t, beforeReceiver.Balance), big.NewInt(100)).String(), f.account(receiver.Address).Balance)
			}
			if scenario.name == "storage-write" {
				require.Equal(t, big.NewInt(1), f.query(contract, "getSum"))
			}
		})
	}
}

func TestSupernovaSmokeContractOwnershipAndRewards(t *testing.T) {
	f := newFixture(t)
	owner, newOwner := f.wallet(0), f.wallet(0)
	contract, code := f.deploy(owner)
	f.success(f.tx(owner, contract, 0, "add@01", 10_000_000))
	before := f.account(contract)
	require.Positive(t, amount(t, before.DeveloperReward).Sign(), "contract call must actually earn developer rewards")
	// Failed authorization is a processed transaction (fee/nonce consumed), not admission rejection.
	unauthorized := f.execute(f.tx(newOwner, contract, 0, core.BuiltInFunctionClaimDeveloperRewards, 10_000_000))
	requireExecutionFailure(t, unauthorized)
	require.Equal(t, before.DeveloperReward, f.account(contract).DeveloperReward)
	f.success(f.tx(owner, contract, 0, core.BuiltInFunctionChangeOwnerAddress+"@"+hexArg(newOwner.Address), 10_000_000))
	require.Equal(t, f.address(newOwner.Address).Bech32, f.account(contract).OwnerAddress)
	claimableAfterTransfer := f.account(contract).DeveloperReward
	requireExecutionFailure(t, f.execute(f.tx(owner, contract, 0, core.BuiltInFunctionClaimDeveloperRewards, 10_000_000)))
	require.Equal(t, claimableAfterTransfer, f.account(contract).DeveloperReward)
	oldOwner := f.execute(f.tx(owner, contract, 0, "upgradeContract@"+code+"@0102@02", 100_000_000))
	requireExecutionFailure(t, oldOwner)
	require.Equal(t, before.CodeHash, f.account(contract).CodeHash)
	require.Equal(t, big.NewInt(1), f.query(contract, "getSum"))
	f.success(f.tx(newOwner, contract, 0, "upgradeContract@"+code+"@0102@02", 100_000_000))
	require.Equal(t, f.address(newOwner.Address).Bech32, f.account(contract).OwnerAddress)
	require.Equal(t, big.NewInt(2), f.query(contract, "getSum"))
	claimable := amount(t, f.account(contract).DeveloperReward)
	ownerBefore := f.account(newOwner.Address)
	claim := f.success(f.tx(newOwner, contract, 0, core.BuiltInFunctionClaimDeveloperRewards, 10_000_000))
	require.Equal(t, "0", f.account(contract).DeveloperReward)
	expected := new(big.Int).Sub(new(big.Int).Add(amount(t, ownerBefore.Balance), claimable), amount(t, claim.Fee))
	require.Equal(t, expected.String(), f.account(newOwner.Address).Balance)
}
