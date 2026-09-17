package smoke

import (
	"testing"

	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func TestSupernovaSmokeValidatorRewardAddress(t *testing.T) {
	f := newFixture(t, stakingConfig)
	owner, receiver, outsider := f.genesisValidatorOwner(), f.wallet(1), f.wallet(0)
	f.fund(owner, egld(10000))
	// Positive control: the existing validator produces rewards before changing it.
	before := amount(t, f.account(owner.Address).Balance)
	f.until(100, func() bool { return amount(t, f.account(owner.Address).Balance).Cmp(before) > 0 })
	f.success(f.tx(owner, vm.ValidatorSCAddress, 0, "changeRewardAddress@"+hexArg(receiver.Address), 100_000_000))
	requireExecutionFailure(t, f.execute(f.tx(outsider, vm.ValidatorSCAddress, 0, "changeRewardAddress@"+hexArg(outsider.Address), 100_000_000)))
	// Drain rewards already assigned before the change, then measure a full epoch.
	f.atEpoch(f.epoch() + 1)
	require.NoError(t, f.cs.GenerateBlocks(8))
	oldBalance, newBalance, outsiderBalance := f.account(owner.Address).Balance, amount(t, f.account(receiver.Address).Balance), f.account(outsider.Address).Balance
	f.atEpoch(f.epoch() + 1)
	require.NoError(t, f.cs.GenerateBlocks(8))
	require.Equal(t, oldBalance, f.account(owner.Address).Balance)
	require.Greater(t, amount(t, f.account(receiver.Address).Balance).Cmp(newBalance), 0)
	require.Equal(t, outsiderBalance, f.account(outsider.Address).Balance)
}
