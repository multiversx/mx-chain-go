package txsFee

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data/block"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	processorTransaction "github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/debitauthorization"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

type signedDebitFixture struct {
	ctx      *vm.VMTestContext
	observer *debitauthorization.Observer
}

func newSignedDebitFixture(t *testing.T) *signedDebitFixture {
	t.Helper()
	f := &signedDebitFixture{}
	ctx, err := vm.CreatePreparedTxProcessorWithAccountsDecorator(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}, 1, func(accounts state.AccountsAdapter) state.AccountsAdapter {
		f.observer = debitauthorization.New(accounts)
		return f.observer
	})
	require.NoError(t, err)
	f.ctx = ctx
	t.Cleanup(ctx.Close)
	t.Cleanup(func() { require.NoError(t, ctx.Accounts.Close()) })
	return f
}

func (f *signedDebitFixture) wallet(t *testing.T, balance int64) *integrationTests.TestWalletAccount {
	t.Helper()
	wallet := integrationTests.CreateTestWalletAccount(f.ctx.ShardCoordinator, f.ctx.ShardCoordinator.SelfId())
	_, err := vm.CreateAccount(f.ctx.Accounts, wallet.Address, 0, big.NewInt(balance))
	require.NoError(t, err)
	return wallet
}

func signedDebitTx(sender, receiver []byte, value int64) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		SndAddr: bytes.Clone(sender), RcvAddr: bytes.Clone(receiver), Value: big.NewInt(value),
		GasPrice: 10, GasLimit: 100, Data: []byte("memo"),
		ChainID: bytes.Clone(integrationTests.ChainID), Version: 2,
	}
}

func signDebitTransaction(t *testing.T, tx *dataTransaction.Transaction, sender, guardian, relayer *integrationTests.TestWalletAccount) {
	t.Helper()
	message, err := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	require.NoError(t, err)
	tx.Signature, err = sender.SingleSigner.Sign(sender.SkTxSign, message)
	require.NoError(t, err)
	if guardian != nil {
		tx.GuardianSignature, err = guardian.SingleSigner.Sign(guardian.SkTxSign, message)
		require.NoError(t, err)
	}
	if relayer != nil {
		tx.RelayerSignature, err = relayer.SingleSigner.Sign(relayer.SkTxSign, message)
		require.NoError(t, err)
	}
}

func (f *signedDebitFixture) verify(tx *dataTransaction.Transaction) error {
	encoded, err := integrationTests.TestMarshalizer.Marshal(tx)
	if err != nil {
		return err
	}
	intercepted, err := processorTransaction.NewInterceptedTransaction(
		encoded, integrationTests.TestMarshalizer, integrationTests.TestTxSignMarshalizer,
		integrationTests.TestHasher, integrationTests.TestKeyGenForAccounts, integrationTests.TestSingleSigner,
		integrationTests.TestAddressPubkeyConverter, f.ctx.ShardCoordinator, f.ctx.EconomicsData,
		&testscommon.WhiteListHandlerStub{}, // always misses: real signatures must run
		smartContract.NewArgumentParser(), integrationTests.ChainID, true, integrationTests.TestTxSignHasher,
		versioning.NewTxVersionChecker(1), f.ctx.EnableEpochsHandler,
	)
	if err != nil {
		return err
	}
	if err = intercepted.CheckValidity(); err != nil {
		return err
	}
	acc, err := f.ctx.Accounts.LoadAccount(tx.SndAddr)
	if err != nil {
		return err
	}
	// Reuse the stateful guardian policy as well as cryptographic verification.
	guardianVerifier := f.ctx.TxProcessor.(interface {
		VerifyGuardian(*dataTransaction.Transaction, state.UserAccountHandler) error
	})
	return guardianVerifier.VerifyGuardian(tx, acc.(state.UserAccountHandler))
}

func (f *signedDebitFixture) execute(t *testing.T, tx *dataTransaction.Transaction) (*debitauthorization.Report, vmcommon.ReturnCode, error) {
	t.Helper()
	verified, err := f.observer.Begin(tx, f.verify)
	require.NoError(t, err)
	code, processErr := f.ctx.TxProcessor.ProcessTransaction(verified)
	report, invariantErr := f.observer.Finish()
	require.NoError(t, invariantErr)
	_, err = f.ctx.Accounts.Commit()
	require.NoError(t, err)
	return report, code, processErr
}

func TestSignedDebitOrdinaryTransfer(t *testing.T) {
	f := newSignedDebitFixture(t)
	sender := f.wallet(t, 10000)
	receiver := f.wallet(t, 0)
	bystander := f.wallet(t, 10000)
	tx := signedDebitTx(sender.Address, receiver.Address, 100)
	signDebitTransaction(t, tx, sender, nil, nil)
	report, code, err := f.execute(t, tx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.Equal(t, "150", report.Accounts[string(sender.Address)].Debits.String()) // value 100 + (base 1 + 4 bytes) * price 10
	require.Equal(t, "50", f.ctx.TxFeeHandler.GetAccumulatedFees().String())
	vm.TestAccount(t, f.ctx.Accounts, sender.Address, 1, big.NewInt(9850))
	vm.TestAccount(t, f.ctx.Accounts, receiver.Address, 0, big.NewInt(100))
	vm.TestAccount(t, f.ctx.Accounts, bystander.Address, 0, big.NewInt(10000))
	_, err = f.observer.Begin(tx, f.verify)
	require.ErrorContains(t, err, "nonce")
}

func TestSignedDebitRelayedV3Transfer(t *testing.T) {
	f := newSignedDebitFixture(t)
	sender := f.wallet(t, 10000)
	receiver := f.wallet(t, 0)
	relayer := f.wallet(t, 10000)
	tx := signedDebitTx(sender.Address, receiver.Address, 100)
	tx.RelayerAddr = bytes.Clone(relayer.Address)
	signDebitTransaction(t, tx, sender, nil, relayer)
	report, code, err := f.execute(t, tx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.Equal(t, "100", report.Accounts[string(sender.Address)].Debits.String())
	require.Equal(t, "60", report.Accounts[string(relayer.Address)].Debits.String()) // ordinary fee plus relayed base cost
	require.Zero(t, report.Accounts[string(sender.Address)].FeeLimit.Sign())
	require.Zero(t, report.Accounts[string(relayer.Address)].ValueLimit.Sign())
	vm.TestAccount(t, f.ctx.Accounts, sender.Address, 1, big.NewInt(9900))
	vm.TestAccount(t, f.ctx.Accounts, relayer.Address, 0, big.NewInt(9940))
	vm.TestAccount(t, f.ctx.Accounts, receiver.Address, 0, big.NewInt(100))
}

func TestSignedDebitFailedExecutionRetainsOnlyAuthorizedFee(t *testing.T) {
	f := newSignedDebitFixture(t)
	sender := f.wallet(t, 100)
	receiver := f.wallet(t, 0)
	tx := signedDebitTx(sender.Address, receiver.Address, 100)
	tx.GasPrice, tx.GasLimit = 1, 20
	signDebitTransaction(t, tx, sender, nil, nil)
	report, _, err := f.execute(t, tx)
	require.ErrorIs(t, err, process.ErrFailedTransaction)
	require.Equal(t, "20", report.Accounts[string(sender.Address)].Debits.String())
	require.Equal(t, "20", f.ctx.TxFeeHandler.GetAccumulatedFees().String())
	vm.TestAccount(t, f.ctx.Accounts, sender.Address, 1, big.NewInt(80))
	vm.TestAccount(t, f.ctx.Accounts, receiver.Address, 0, big.NewInt(0))
}

func TestSignedDebitSignatureAndStateValidationBeforeExecution(t *testing.T) {
	for _, kind := range []string{"missing sender signature", "changed signed value", "invalid relayer signature", "wrong chain", "future nonce"} {
		t.Run(kind, func(t *testing.T) {
			f := newSignedDebitFixture(t)
			sender := f.wallet(t, 10000)
			receiver := f.wallet(t, 0)
			tx := signedDebitTx(sender.Address, receiver.Address, 100)
			var relayer *integrationTests.TestWalletAccount
			if kind == "invalid relayer signature" {
				relayer = f.wallet(t, 10000)
				tx.RelayerAddr = bytes.Clone(relayer.Address)
			}
			if kind == "wrong chain" {
				tx.ChainID = []byte("other-chain")
			}
			if kind == "future nonce" {
				tx.Nonce = 1
			}
			signDebitTransaction(t, tx, sender, nil, relayer)
			switch kind {
			case "missing sender signature":
				tx.Signature = nil
			case "changed signed value":
				tx.Value.Add(tx.Value, big.NewInt(1))
			case "invalid relayer signature":
				tx.RelayerSignature[0] ^= 1
			}
			root, err := f.ctx.Accounts.RootHash()
			require.NoError(t, err)
			_, err = f.observer.Begin(tx, f.verify)
			require.Error(t, err)
			after, err := f.ctx.Accounts.RootHash()
			require.NoError(t, err)
			require.Equal(t, root, after)
			vm.TestAccount(t, f.ctx.Accounts, sender.Address, 0, big.NewInt(10000))
		})
	}
}

func TestSignedDebitGuardedAccount(t *testing.T) {
	f := newSignedDebitFixture(t)
	sender := f.wallet(t, 10000000)
	receiver := f.wallet(t, 0)
	guardian := f.wallet(t, 0)
	acc, err := f.ctx.Accounts.LoadAccount(sender.Address)
	require.NoError(t, err)
	user := acc.(state.UserAccountHandler)
	require.NoError(t, f.ctx.GuardedAccountsHandler.SetGuardian(acc.(vmcommon.UserAccountHandler), guardian.Address, nil, []byte("test-service")))
	user.SetCodeMetadata((&vmcommon.CodeMetadata{Guarded: true}).ToBytes())
	require.NoError(t, f.ctx.Accounts.SaveAccount(user))
	f.ctx.EpochNotifier.CheckEpoch(&block.Header{Epoch: vm.EpochGuardianDelay})
	_, err = f.ctx.Accounts.Commit()
	require.NoError(t, err)

	tx := signedDebitTx(sender.Address, receiver.Address, 100)
	tx.GasLimit = 100000
	tx.Options = dataTransaction.MaskGuardedTransaction
	tx.GuardianAddr = bytes.Clone(guardian.Address)
	signDebitTransaction(t, tx, sender, guardian, nil)
	validGuardianSignature := bytes.Clone(tx.GuardianSignature)
	tx.GuardianSignature[0] ^= 1
	_, err = f.observer.Begin(tx, f.verify)
	require.Error(t, err)
	tx.GuardianSignature = validGuardianSignature
	report, code, err := f.execute(t, tx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.Positive(t, report.Accounts[string(sender.Address)].Debits.Sign())
	vm.TestAccount(t, f.ctx.Accounts, receiver.Address, 0, big.NewInt(100))
	vm.TestAccount(t, f.ctx.Accounts, guardian.Address, 0, big.NewInt(0))
}

func TestSignedDebitWasmCall(t *testing.T) {
	f := newSignedDebitFixture(t)
	// Deployment is fixture setup, outside the signed-call observation interval.
	contract, _ := utils.DoDeployNoChecks(t, f.ctx, "../wasm/testdata/counter/output/counter.wasm")
	f.ctx.CreateBlockStarted()
	sender := f.wallet(t, 1000000000)
	tx := signedDebitTx(sender.Address, contract, 0)
	tx.GasLimit = 100000
	tx.Data = []byte("increment")
	signDebitTransaction(t, tx, sender, nil, nil)
	report, code, err := f.execute(t, tx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.NotContains(t, report.Accounts, string(contract))
	result := report.Accounts[string(sender.Address)]
	require.Positive(t, result.Debits.Sign())
	require.Equal(t, big.NewInt(2), vm.GetIntValueFromSC(nil, f.ctx.Accounts, contract, "get"))
	charged := new(big.Int).Sub(result.Before, result.After)
	require.Equal(t, f.ctx.TxFeeHandler.GetAccumulatedFees(), charged)
}

func TestSignedDebitRollbackAndReexecution(t *testing.T) {
	f := newSignedDebitFixture(t)
	sender := f.wallet(t, 10000)
	receiver := f.wallet(t, 0)
	root, err := f.ctx.Accounts.RootHash()
	require.NoError(t, err)
	tx := signedDebitTx(sender.Address, receiver.Address, 100)
	signDebitTransaction(t, tx, sender, nil, nil)
	verified, err := f.observer.Begin(tx, f.verify)
	require.NoError(t, err)
	snapshot := f.ctx.Accounts.JournalLen()
	code, err := f.ctx.TxProcessor.ProcessTransaction(verified)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.NoError(t, f.ctx.Accounts.RevertToSnapshot(snapshot))
	report, err := f.observer.Finish()
	require.NoError(t, err)
	require.Positive(t, report.RevertedEvents)
	require.Zero(t, report.Accounts[string(sender.Address)].Debits.Sign())
	after, err := f.ctx.Accounts.RootHash()
	require.NoError(t, err)
	require.Equal(t, root, after)
	f.ctx.CreateBlockStarted() // reset the fixture's non-state fee/result collectors
	report, code, err = f.execute(t, tx)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, code)
	require.Equal(t, "150", report.Accounts[string(sender.Address)].Debits.String())
}
