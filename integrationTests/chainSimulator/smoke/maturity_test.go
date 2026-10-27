package smoke

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func (f *fixture) atEpoch(epoch uint32) {
	f.t.Helper()
	require.LessOrEqual(f.t, f.epoch(), epoch, "test must not miss the requested epoch")
	f.until(200, func() bool { return f.epoch() >= epoch })
	require.Equal(f.t, epoch, f.epoch())
}

func TestSupernovaSmokeStaggeredMaturity(t *testing.T) {
	for _, kind := range []string{"validator", "delegation"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t, func(cfg *config.Configs) {
				stakingConfig(cfg)
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 4
			})
			owner, user := f.wallet(0), f.wallet(1)
			contract, unstake, withdraw, list := vm.ValidatorSCAddress, "unStakeTokens", "unBondTokens", "getUnStakedTokensList"
			if kind == "validator" {
				f.fund(user, egld(10000))
				// Send validator rewards to a separate wallet so principal accounting is exact.
				key, signature := f.validatorKey(user.Address)
				f.success(f.call(user, contract, egld(2506), "stake@01@"+key+"@"+signature+"@"+hexArg(owner.Address)))
			} else {
				contract = f.delegation(owner)
				unstake, withdraw, list = "unDelegate", "withdraw", "getUserUnDelegatedList"
				f.success(f.call(user, contract, egld(6), "delegate"))
			}
			initial := amount(t, f.account(user.Address).Balance)
			fees := new(big.Int)
			first := f.epoch() + 1
			amounts := []int64{1, 2, 3}
			checkPending := func(firstIndex int, epoch uint32) {
				values := f.queryValues(contract, list, user.Address)
				require.Len(t, values, 2*(len(amounts)-firstIndex))
				for i := firstIndex; i < len(amounts); i++ {
					offset := 2 * (i - firstIndex)
					require.Equal(t, egld(amounts[i]), new(big.Int).SetBytes(values[offset]))
					maturity := first + uint32(i) + 4
					remaining := uint32(0)
					if maturity > epoch {
						remaining = maturity - epoch
					}
					require.Equal(t, new(big.Int).SetUint64(uint64(remaining)), new(big.Int).SetBytes(values[offset+1]))
				}
			}
			for i, value := range amounts {
				epoch := first + uint32(i)
				f.atEpoch(epoch)
				tx := f.success(f.tx(user, contract, 0, unstake+"@"+hexArg(egld(value).Bytes()), 100_000_000))
				fees.Add(fees, amount(t, tx.Fee))
				require.Equal(t, epoch, f.epoch(), "unstakes must occur in distinct known epochs")
			}
			// M-1: none of the three entries is mature, and a processed withdrawal pays nothing.
			f.atEpoch(first + 3)
			checkPending(0, first+3)
			early := f.execute(f.tx(user, contract, 0, withdraw, 100_000_000))
			fees.Add(fees, amount(t, early.Fee))
			require.Equal(t, first+3, f.epoch())
			require.Equal(t, new(big.Int).Sub(initial, fees).String(), f.account(user.Address).Balance)
			checkPending(0, first+3)
			recovered := new(big.Int)
			for i, value := range amounts {
				// M, M+1 and M+2: each entry matures independently, exactly one epoch apart.
				epoch := first + uint32(i) + 4
				f.atEpoch(epoch)
				checkPending(i, epoch)
				tx := f.success(f.tx(user, contract, 0, withdraw, 100_000_000))
				fees.Add(fees, amount(t, tx.Fee))
				recovered.Add(recovered, egld(value))
				require.Equal(t, epoch, f.epoch(), "withdrawal must execute in the maturity epoch")
				require.Equal(t, new(big.Int).Sub(new(big.Int).Add(initial, recovered), fees).String(), f.account(user.Address).Balance)
				if i+1 < len(amounts) {
					checkPending(i+1, epoch)
				}
			}
			if kind == "validator" {
				require.Empty(t, f.queryValues(contract, list, user.Address))
			} else {
				require.Zero(t, f.queryAmount(contract, "getTotalUnStaked").Sign())
				require.Equal(t, egld(2500), f.queryAmount(contract, "getTotalActiveStake"))
			}
			// One epoch after the final maturity, a fresh signed withdrawal cannot pay again.
			f.atEpoch(first + 7)
			repeated := f.execute(f.tx(user, contract, 0, withdraw, 100_000_000))
			fees.Add(fees, amount(t, repeated.Fee))
			require.Equal(t, new(big.Int).Sub(new(big.Int).Add(initial, egld(6)), fees).String(), f.account(user.Address).Balance)
		})
	}
}
