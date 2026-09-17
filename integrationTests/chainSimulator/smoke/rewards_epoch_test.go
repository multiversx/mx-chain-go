package smoke

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/vm"
)

// Regression for PR #8011: the epoch-change proposal must expose the epoch being
// prepared to the system contracts. Otherwise delegation rewards are stored one
// epoch behind. Service-fee changes make successive reward snapshots distinct.
// This intentionally fails on master fbb1284, before that fix.
func TestSupernovaSmokeDelegationRewardEpoch(t *testing.T) {
	f := newFixture(t, stakingConfig)
	// Convert a genesis validator instead of adding a fourth node to a network
	// with three eligible slots: a rotating waiting node need not earn each epoch.
	owner := f.genesisValidatorOwner()
	f.fund(owner, egld(10000))
	f.success(f.tx(owner, vm.DelegationManagerSCAddress, 0, "makeNewContractFromValidatorData@00@03e8", 510_000_000))
	contracts := f.queryValues(vm.DelegationManagerSCAddress, "getAllContractAddresses")
	require.Len(t, contracts, 1)
	contract := contracts[0]
	require.Equal(t, owner.Address, f.queryValues(contract, "getContractConfig")[0])
	user := f.wallet(1)
	f.success(f.call(user, contract, egld(10), "delegate"))
	totalActive := f.queryAmount(contract, "getTotalActiveStake")
	require.GreaterOrEqual(t, totalActive.Cmp(egld(2510)), 0)
	// Positive control: the validator must actually earn rewards before checking
	// their epoch index. This also runs beyond the PR's default fix epoch (3).
	f.until(240, func() bool {
		return f.epoch() >= 4 && f.queryAmount(contract, "getClaimableRewards", user.Address).Sign() > 0
	})
	require.Greater(t, f.epoch(), uint32(3))
	record := func(epoch uint32) [][]byte {
		t.Helper()
		values := f.queryValues(contract, "getRewardData", new(big.Int).SetUint64(uint64(epoch)).Bytes())
		require.Len(t, values, 3)
		require.Positive(t, new(big.Int).SetBytes(values[0]).Sign())
		require.Equal(t, totalActive, new(big.Int).SetBytes(values[1]))
		return values
	}
	var previous [][]byte
	var previousEpoch uint32
	for _, fee := range []int64{10000, 0} {
		f.success(f.tx(owner, contract, 0, "changeServiceFee@"+hexArg(big.NewInt(fee).Bytes()), 100_000_000))
		require.Equal(t, big.NewInt(fee), new(big.Int).SetBytes(f.queryValues(contract, "getContractConfig")[1]))
		ownerBefore := f.queryAmount(contract, "getTotalCumulatedRewardsForUser", owner.Address)
		userBefore := f.queryAmount(contract, "getTotalCumulatedRewardsForUser", user.Address)
		target := f.epoch() + 1
		f.atEpoch(target)
		// Let the epoch-start execution settle, without crossing another epoch
		// and accidentally accepting a record produced one epoch late.
		require.NoError(t, f.cs.GenerateBlocks(8))
		require.Equal(t, target, f.epoch())
		t.Logf("expecting reward snapshot for epoch %d with service fee %d", target, fee)
		current := record(target)
		require.Equal(t, big.NewInt(fee), new(big.Int).SetBytes(current[2]))
		ownerEarned := new(big.Int).Sub(f.queryAmount(contract, "getTotalCumulatedRewardsForUser", owner.Address), ownerBefore)
		userEarned := new(big.Int).Sub(f.queryAmount(contract, "getTotalCumulatedRewardsForUser", user.Address), userBefore)
		reward := new(big.Int).SetBytes(current[0])
		if fee == 10000 {
			require.Zero(t, userEarned.Sign(), "100%% service fee leaves no new delegator reward")
			require.Equal(t, reward, ownerEarned)
		} else {
			expectedUser := new(big.Int).Div(new(big.Int).Mul(reward, egld(10)), totalActive)
			expectedOwner := new(big.Int).Div(new(big.Int).Mul(reward, new(big.Int).Sub(totalActive, egld(10))), totalActive)
			require.Positive(t, expectedUser.Sign())
			require.Equal(t, expectedUser, userEarned)
			require.Equal(t, expectedOwner, ownerEarned)
		}
		if previous != nil {
			require.Equal(t, previous, record(previousEpoch), "later rewards must preserve the prior epoch's snapshot")
		}
		previous, previousEpoch = current, target
	}
	for _, wallet := range []*integrationTests.TestWalletAccount{user, owner} {
		before := amount(t, f.account(wallet.Address).Balance)
		claimed := f.success(f.tx(wallet, contract, 0, "claimRewards", 100_000_000))
		paid := eventAmount(t, claimed, "claimRewards")
		require.Positive(t, paid.Sign())
		require.Equal(t, new(big.Int).Sub(new(big.Int).Add(before, paid), amount(t, claimed.Fee)).String(), f.account(wallet.Address).Balance)
	}
}
