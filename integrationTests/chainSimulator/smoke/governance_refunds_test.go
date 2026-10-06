package smoke

import (
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func TestSupernovaSmokeGovernanceRefunds(t *testing.T) {
	for _, vote := range []string{"yes", "no-quorum", "veto", "abstain"} {
		t.Run(vote, func(t *testing.T) {
			f := newFixture(t, func(cfg *config.Configs) {
				stakingConfig(cfg)
				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 10
				cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.ProposalCost = "1000"
				cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.LostProposalFee = "100"
				cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.MinQuorum = 0.02
				cfg.EpochConfig.EnableEpochs.GovernanceFixesEnableEpoch = 0
			})
			issuer, closer := f.wallet(0), f.wallet(1)
			f.delegation(issuer)
			start, end := f.epoch()+2, f.epoch()+4
			data := "proposal@" + hexArg([]byte(strings.Repeat("c", 40))) + "@" + hexArg(big.NewInt(int64(start)).Bytes()) + "@" + hexArg(big.NewInt(int64(end)).Bytes())
			f.success(f.tx(issuer, vm.GovernanceSCAddress, 1000, data, 100_000_000))
			f.atEpoch(start)
			if vote != "no-quorum" {
				f.success(f.tx(issuer, vm.GovernanceSCAddress, 0, "vote@01@"+hexArg([]byte(vote)), 100_000_000))
			}
			f.atEpoch(end + 1)
			beforeIssuer, beforeCloser := f.account(issuer.Address), f.account(closer.Address)
			beforeContract := amount(t, f.account(vm.GovernanceSCAddress).Balance)
			// A third party can close, but only the proposal issuer receives the deposit.
			closed := f.success(f.tx(closer, vm.GovernanceSCAddress, 0, "closeProposal@01", 100_000_000))
			view := f.queryValues(vm.GovernanceSCAddress, "viewProposal", []byte{1})
			require.Equal(t, []byte("true"), view[11])
			refund, passed := int64(900), "false"
			if vote == "yes" {
				refund, passed = 1000, "true"
			}
			require.Equal(t, []byte(passed), view[12])
			require.Equal(t, new(big.Int).Add(amount(t, beforeIssuer.Balance), big.NewInt(refund)).String(), f.account(issuer.Address).Balance)
			require.Equal(t, new(big.Int).Sub(amount(t, beforeCloser.Balance), amount(t, closed.Fee)).String(), f.account(closer.Address).Balance)
			require.Equal(t, new(big.Int).Sub(beforeContract, big.NewInt(refund)).String(), f.account(vm.GovernanceSCAddress).Balance)
			beforeIssuer = f.account(issuer.Address)
			beforeCloser = f.account(closer.Address)
			contractBalance := f.account(vm.GovernanceSCAddress).Balance
			repeated := f.execute(f.tx(closer, vm.GovernanceSCAddress, 0, "closeProposal@01", 100_000_000))
			requireExecutionFailure(t, repeated)
			require.Equal(t, view, f.queryValues(vm.GovernanceSCAddress, "viewProposal", []byte{1}))
			require.Equal(t, beforeIssuer.Balance, f.account(issuer.Address).Balance)
			require.Equal(t, contractBalance, f.account(vm.GovernanceSCAddress).Balance)
			require.Equal(t, new(big.Int).Sub(amount(t, beforeCloser.Balance), amount(t, repeated.Fee)).String(), f.account(closer.Address).Balance)
		})
	}
}
