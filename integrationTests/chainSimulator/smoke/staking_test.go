package smoke

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	chainTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/vm"
)

func stakingConfig(cfg *config.Configs) {
	cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
	cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod = 40
	cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodSupernova = 40
	cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
	configs.SetMaxNumberOfNodesInConfigs(cfg, 12, 0, numShards)
}

func egld(value int64) *big.Int { return new(big.Int).Mul(chainTests.OneEGLD, big.NewInt(value)) }

func (f *fixture) validatorKey(address []byte) (string, string) {
	f.t.Helper()
	privateKeys, publicKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.NoError(f.t, err)
	require.NoError(f.t, f.cs.AddValidatorKeys(privateKeys))
	privateKey, err := signing.NewKeyGenerator(mcl.NewSuiteBLS12()).PrivateKeyFromByteArray(privateKeys[0])
	require.NoError(f.t, err)
	signature, err := mclsig.NewBlsSigner().Sign(privateKey, address)
	require.NoError(f.t, err)
	return publicKeys[0], hexArg(signature)
}

func TestSupernovaSmokeValidatorLifecycle(t *testing.T) {
	f := newFixture(t, stakingConfig)
	owner := f.wallet(0)
	f.fund(owner, egld(10000))
	key, signature := f.validatorKey(owner.Address)
	decoded, err := hex.DecodeString(key)
	require.NoError(t, err)
	initial := amount(t, f.account(owner.Address).Balance)
	fees := new(big.Int)
	// Re-register the same key with the recovered funds, without re-minting.
	for cycle := 0; cycle < 2; cycle++ {
		t.Logf("staking cycle %d", cycle+1)
		stake := f.success(f.call(owner, vm.ValidatorSCAddress, egld(2500), "stake@01@"+key+"@"+signature))
		fees.Add(fees, amount(t, stake.Fee))
		require.Equal(t, "staked", staking.GetBLSKeyStatus(t, f.cs.GetNodeHandler(core.MetachainShardId), decoded))
		unstake := f.success(f.tx(owner, vm.ValidatorSCAddress, 0, "unStake@"+key, 100_000_000))
		fees.Add(fees, amount(t, unstake.Fee))
		require.Equal(t, "unStaked", staking.GetBLSKeyStatus(t, f.cs.GetNodeHandler(core.MetachainShardId), decoded))
		before := f.account(owner.Address)
		early := f.execute(f.tx(owner, vm.ValidatorSCAddress, 0, "unBondTokens", 100_000_000))
		fees.Add(fees, amount(t, early.Fee))
		require.Equal(t, new(big.Int).Sub(amount(t, before.Balance), amount(t, early.Fee)).String(), f.account(owner.Address).Balance)
		maturity := f.epoch() + 4
		f.until(160, func() bool { return f.epoch() >= maturity })
		unbond := f.success(f.tx(owner, vm.ValidatorSCAddress, 0, "unBondNodes@"+key, 100_000_000))
		fees.Add(fees, amount(t, unbond.Fee))
		before = f.account(owner.Address)
		withdrawn := f.success(f.tx(owner, vm.ValidatorSCAddress, 0, "unBondTokens", 100_000_000))
		fees.Add(fees, amount(t, withdrawn.Fee))
		require.Equal(t, new(big.Int).Sub(new(big.Int).Add(amount(t, before.Balance), egld(2500)), amount(t, withdrawn.Fee)).String(), f.account(owner.Address).Balance)
		repeated := f.execute(f.tx(owner, vm.ValidatorSCAddress, 0, "unBondTokens", 100_000_000))
		fees.Add(fees, amount(t, repeated.Fee))
		require.Equal(t, new(big.Int).Sub(initial, fees).String(), f.account(owner.Address).Balance, "each complete cycle returns the principal exactly once")
	}
}

func (f *fixture) delegation(owner *integrationTests.TestWalletAccount) []byte {
	f.t.Helper()
	f.fund(owner, egld(10000))
	f.success(f.call(owner, vm.DelegationManagerSCAddress, egld(2500), "createNewDelegationContract@00@03e8"))
	contracts := f.queryValues(vm.DelegationManagerSCAddress, "getAllContractAddresses")
	require.NotEmpty(f.t, contracts)
	contract := contracts[len(contracts)-1]
	require.Equal(f.t, owner.Address, f.queryValues(contract, "getContractConfig")[0])
	return contract
}

func TestSupernovaSmokeDelegationLifecycleAndOwnership(t *testing.T) {
	f := newFixture(t, stakingConfig)
	owner, user, newOwner := f.wallet(0), f.wallet(1), f.wallet(0)
	contract := f.delegation(owner)
	key, signature := f.validatorKey(contract)
	f.success(f.tx(owner, contract, 0, "addNodes@"+key+"@"+signature, 500_000_000))
	f.success(f.tx(owner, contract, 0, "stakeNodes@"+key, 100_000_000))
	f.success(f.call(user, contract, egld(10), "delegate"))
	require.Equal(t, egld(10), f.queryAmount(contract, "getUserActiveStake", user.Address))
	require.Equal(t, egld(2510), f.queryAmount(contract, "getTotalActiveStake"))
	f.until(240, func() bool { return f.queryAmount(contract, "getClaimableRewards", user.Address).Sign() > 0 })
	before := f.account(user.Address)
	claimed := f.success(f.tx(user, contract, 0, "claimRewards", 100_000_000))
	paid := new(big.Int).Add(new(big.Int).Sub(amount(t, f.account(user.Address).Balance), amount(t, before.Balance)), amount(t, claimed.Fee))
	require.Positive(t, paid.Sign())
	require.Equal(t, paid, eventAmount(t, claimed, "claimRewards"))
	require.Equal(t, new(big.Int).Add(paid, f.queryAmount(contract, "getClaimableRewards", user.Address)), f.queryAmount(contract, "getTotalCumulatedRewardsForUser", user.Address))
	f.until(160, func() bool { return f.queryAmount(contract, "getClaimableRewards", user.Address).Sign() > 0 })
	activeBefore := f.queryAmount(contract, "getUserActiveStake", user.Address)
	before = f.account(user.Address)
	redelegated := f.success(f.tx(user, contract, 0, "reDelegateRewards", 100_000_000))
	activeAfter := f.queryAmount(contract, "getUserActiveStake", user.Address)
	redelegatedAmount := new(big.Int).Sub(activeAfter, activeBefore)
	require.Positive(t, redelegatedAmount.Sign())
	require.Equal(t, redelegatedAmount, eventAmount(t, redelegated, "delegate"))
	require.Equal(t, new(big.Int).Sub(amount(t, before.Balance), amount(t, redelegated.Fee)).String(), f.account(user.Address).Balance)
	require.Equal(t, new(big.Int).Add(egld(2500), activeAfter), f.queryAmount(contract, "getTotalActiveStake"))
	// Owner transfer moves the original owner's fund record and administrative authority.
	configBefore := f.queryValues(contract, "getContractConfig")
	f.success(f.tx(owner, contract, 0, "changeOwner@"+hexArg(newOwner.Address), 100_000_000))
	require.Equal(t, newOwner.Address, f.queryValues(contract, "getContractConfig")[0])
	require.Equal(t, egld(2500), f.queryAmount(contract, "getUserActiveStake", newOwner.Address))
	require.Equal(t, activeAfter, f.queryAmount(contract, "getUserActiveStake", user.Address))
	requireExecutionFailure(t, f.execute(f.tx(owner, contract, 0, "changeServiceFee@07d0", 100_000_000)))
	require.Equal(t, configBefore[1], f.queryValues(contract, "getContractConfig")[1])
	f.success(f.tx(newOwner, contract, 0, "changeServiceFee@07d0", 100_000_000))
	require.Equal(t, big.NewInt(2000), new(big.Int).SetBytes(f.queryValues(contract, "getContractConfig")[1]))
	f.success(f.tx(user, contract, 0, "unDelegate@"+hexArg(activeAfter.Bytes()), 100_000_000))
	require.Zero(t, f.queryAmount(contract, "getUserActiveStake", user.Address).Sign())
	require.Equal(t, activeAfter, f.queryAmount(contract, "getUserUnStakedValue", user.Address))
	before = f.account(user.Address)
	early := f.execute(f.tx(user, contract, 0, "withdraw", 100_000_000))
	requireExecutionFailure(t, early)
	require.Equal(t, new(big.Int).Sub(amount(t, before.Balance), amount(t, early.Fee)).String(), f.account(user.Address).Balance)
	f.until(160, func() bool { return f.queryAmount(contract, "getUserUnBondable", user.Address).Cmp(activeAfter) == 0 })
	before = f.account(user.Address)
	withdrawn := f.success(f.tx(user, contract, 0, "withdraw", 100_000_000))
	require.Equal(t, new(big.Int).Sub(new(big.Int).Add(amount(t, before.Balance), activeAfter), amount(t, withdrawn.Fee)).String(), f.account(user.Address).Balance)
	before = f.account(user.Address)
	repeated := f.execute(f.tx(user, contract, 0, "withdraw", 100_000_000))
	requireExecutionFailure(t, repeated)
	require.Equal(t, new(big.Int).Sub(amount(t, before.Balance), amount(t, repeated.Fee)).String(), f.account(user.Address).Balance)
}

// Compare the amount declared by the contract with independently read account state.
func eventAmount(t *testing.T, result *transaction.ApiTransactionResult, name string) *big.Int {
	t.Helper()
	require.NotNil(t, result.Logs)
	for _, event := range result.Logs.Events {
		if event.Identifier == name {
			require.NotEmpty(t, event.Topics)
			return new(big.Int).SetBytes(event.Topics[0])
		}
	}
	t.Fatalf("missing %s event", name)
	return nil
}

func (f *fixture) genesisValidatorOwner() *integrationTests.TestWalletAccount {
	f.t.Helper()
	genesisOwner := f.cs.GetInitialWalletKeys().StakeWallets[0]
	owner := integrationTests.CreateTestWalletAccount(f.cs.GetNodeHandler(0).GetShardCoordinator(), 0)
	privateKey, err := hex.DecodeString(genesisOwner.PrivateKeyHex)
	require.NoError(f.t, err)
	owner.SkTxSign, err = owner.KeygenTxSign.PrivateKeyFromByteArray(privateKey)
	require.NoError(f.t, err)
	owner.Address = genesisOwner.Address.Bytes
	return owner
}
