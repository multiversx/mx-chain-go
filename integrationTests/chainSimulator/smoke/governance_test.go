package smoke

import (
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
)

func TestSupernovaSmokeGovernanceConfigAuthorization(t *testing.T) {
	f := newFixture(t)
	caller := f.wallet(0)
	facade := f.cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler()
	address := f.address(vm.GovernanceSCAddress).Bech32
	before, _, err := facade.GetKeyValuePairs(address, apiData.AccountQueryOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, before, "governance contract must be initialized")
	beforeAccount := f.account(vm.GovernanceSCAddress)
	beforeCaller := f.account(caller.Address)
	// Well-formed configuration arguments, signed by an ordinary funded account.
	args := []string{"1000", "100", "5000", "3300", "5000"}
	for i := range args {
		args[i] = hexArg([]byte(args[i]))
	}
	result := f.execute(f.tx(caller, vm.GovernanceSCAddress, 0, "changeConfig@"+strings.Join(args, "@"), 100_000_000))
	requireExecutionFailure(t, result)
	messages := ""
	for _, scr := range result.SmartContractResults {
		messages += scr.ReturnMessage
	}
	if result.Logs != nil {
		for _, event := range result.Logs.Events {
			for _, topic := range event.Topics {
				messages += string(topic)
			}
		}
	}
	require.Contains(t, messages, "changeConfig can be called only by owner")
	after, _, err := facade.GetKeyValuePairs(address, apiData.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, before, after, "rejected configuration update must preserve all governance storage")
	require.Equal(t, beforeAccount.RootHash, f.account(vm.GovernanceSCAddress).RootHash)
	require.Equal(t, beforeAccount.Balance, f.account(vm.GovernanceSCAddress).Balance)
	require.Equal(t, beforeCaller.Nonce+1, f.account(caller.Address).Nonce)
	require.Positive(t, amount(t, result.Fee).Sign())
	require.Equal(t, new(big.Int).Sub(amount(t, beforeCaller.Balance), amount(t, result.Fee)).String(), f.account(caller.Address).Balance)
}

func TestSupernovaSmokeGovernanceLifecycle(t *testing.T) {
	f := newFixture(t, func(cfg *config.Configs) {
		stakingConfig(cfg)
		cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 20
		cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.ProposalCost = "1000"
		cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.LostProposalFee = "100"
		cfg.SystemSCConfig.GovernanceSystemSCConfig.Active.MinQuorum = 0.02
		cfg.EpochConfig.EnableEpochs.GovernanceFixesEnableEpoch = 0
	})
	voter, outsider := f.wallet(0), f.wallet(1)
	contract := f.delegation(voter)
	power := f.queryAmount(vm.GovernanceSCAddress, "viewVotingPower", voter.Address)
	require.Equal(t, egld(2500), power)
	start, endA, endB := f.epoch()+3, f.epoch()+12, f.epoch()+15
	view := func(nonce byte) [][]byte { return f.queryValues(vm.GovernanceSCAddress, "viewProposal", []byte{nonce}) }
	for index, end := range []uint32{endA, endB} {
		commit := []byte(strings.Repeat(string(rune('a'+index)), 40))
		data := "proposal@" + hexArg(commit) + "@" + hexArg(new(big.Int).SetUint64(uint64(start)).Bytes()) + "@" + hexArg(new(big.Int).SetUint64(uint64(end)).Bytes())
		f.success(f.tx(voter, vm.GovernanceSCAddress, 1000, data, 100_000_000))
		values := view(byte(index + 1))
		require.Len(t, values, 13)
		require.Equal(t, commit, values[1])
		require.Equal(t, voter.Address, values[3])
	}
	initialA, initialB := view(1), view(2)
	requireExecutionFailure(t, f.execute(f.tx(voter, vm.GovernanceSCAddress, 0, "vote@01@796573", 100_000_000)))
	require.Equal(t, initialA, view(1))
	require.Equal(t, initialB, view(2))
	f.atEpoch(start)
	requireExecutionFailure(t, f.execute(f.tx(outsider, vm.GovernanceSCAddress, 0, "vote@01@796573", 100_000_000)))
	require.Equal(t, initialA, view(1))
	require.Equal(t, initialB, view(2))
	f.success(f.tx(voter, vm.GovernanceSCAddress, 0, "vote@01@796573", 100_000_000))
	votedA := view(1)
	require.Equal(t, power, new(big.Int).SetBytes(votedA[7]))
	require.Equal(t, egld(2500), new(big.Int).SetBytes(votedA[6]))
	// Pending proposals do not rewrite an already recorded vote when stake changes.
	f.success(f.call(voter, contract, egld(100), "delegate"))
	require.Equal(t, egld(2600), f.queryAmount(vm.GovernanceSCAddress, "viewVotingPower", voter.Address))
	require.Equal(t, votedA, view(1))
	key, signature := f.validatorKey(voter.Address)
	f.success(f.call(voter, vm.ValidatorSCAddress, egld(2550), "stake@01@"+key+"@"+signature))
	require.Equal(t, egld(5150), f.queryAmount(vm.GovernanceSCAddress, "viewVotingPower", voter.Address))
	require.Equal(t, votedA, view(1))
	f.success(f.tx(voter, contract, 0, "unDelegate@"+hexArg(egld(40).Bytes()), 100_000_000))
	require.Equal(t, egld(5110), f.queryAmount(vm.GovernanceSCAddress, "viewVotingPower", voter.Address))
	f.success(f.tx(voter, vm.ValidatorSCAddress, 0, "unStakeTokens@"+hexArg(egld(20).Bytes()), 100_000_000))
	require.Equal(t, egld(5090), f.queryAmount(vm.GovernanceSCAddress, "viewVotingPower", voter.Address))
	require.Equal(t, votedA, view(1))
	require.Equal(t, initialB, view(2))
	f.success(f.tx(voter, vm.GovernanceSCAddress, 0, "vote@02@796573", 100_000_000))
	votedB := view(2)
	require.Equal(t, egld(5090), new(big.Int).SetBytes(votedB[7]))
	require.Equal(t, egld(5090), new(big.Int).SetBytes(votedB[6]))
	f.success(f.call(voter, contract, egld(10), "delegate"))
	for _, nonce := range []byte{1, 2} {
		requireExecutionFailure(t, f.execute(f.tx(voter, vm.GovernanceSCAddress, 0, "vote@"+hexArg([]byte{nonce})+"@796573", 100_000_000)))
	}
	require.Equal(t, votedA, view(1))
	require.Equal(t, votedB, view(2))
	requireExecutionFailure(t, f.execute(f.tx(voter, vm.GovernanceSCAddress, 0, "closeProposal@01", 100_000_000)))
	f.atEpoch(endA + 1)
	f.success(f.tx(voter, vm.GovernanceSCAddress, 0, "closeProposal@01", 100_000_000))
	closedA := view(1)
	require.Equal(t, []byte("true"), closedA[11])
	require.Equal(t, []byte("true"), closedA[12])
	require.Equal(t, votedA[6:11], closedA[6:11])
	require.Equal(t, votedB, view(2), "closing proposal A must not close or change overlapping proposal B")
	// A new delegator can still vote on B while A is closed.
	f.success(f.call(outsider, contract, egld(10), "delegate"))
	f.success(f.tx(outsider, vm.GovernanceSCAddress, 0, "vote@02@6e6f", 100_000_000))
	require.Equal(t, closedA, view(1))
	finalB := view(2)
	require.Equal(t, egld(5090), new(big.Int).SetBytes(finalB[7]))
	require.Equal(t, egld(10), new(big.Int).SetBytes(finalB[8]))
	require.Equal(t, egld(5100), new(big.Int).SetBytes(finalB[6]))
	f.atEpoch(endB + 1)
	f.success(f.tx(voter, vm.GovernanceSCAddress, 0, "closeProposal@02", 100_000_000))
	closedB := view(2)
	require.Equal(t, []byte("true"), closedB[11])
	require.Equal(t, []byte("true"), closedB[12])
	require.Equal(t, finalB[6:11], closedB[6:11])
	require.Equal(t, closedA, view(1))
}
