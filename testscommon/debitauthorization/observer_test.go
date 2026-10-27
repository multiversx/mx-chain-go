package debitauthorization_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/debitauthorization"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/stretchr/testify/require"
)

var sender = bytes.Repeat([]byte{1}, 32)
var recipient = bytes.Repeat([]byte{2}, 32)
var relayer = bytes.Repeat([]byte{3}, 32)

// This verifier intentionally models an already-verified fixture. Real signature
// verification is covered by the signedDebit integration tests, not this stub.
func fixtureVerified(_ *transaction.Transaction) error { return nil }

func fixtureTx() *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: bytes.Clone(sender), RcvAddr: bytes.Clone(recipient),
		Value: big.NewInt(10), GasLimit: 2, GasPrice: 1,
		ChainID: []byte("test"), Version: 1, Signature: []byte("fixture only"),
	}
}

func setup(t *testing.T) *debitauthorization.Observer {
	t.Helper()
	accounts := integrationtests.CreateAccountsDB(integrationtests.CreateMemUnit(), enableEpochsHandlerMock.NewEnableEpochsHandlerStub())
	t.Cleanup(func() { require.NoError(t, accounts.Close()) })
	o := debitauthorization.New(accounts)
	for _, address := range [][]byte{sender, recipient, relayer} {
		acc := load(t, o, address)
		require.NoError(t, acc.AddToBalance(big.NewInt(100)))
		require.NoError(t, o.SaveAccount(acc))
	}
	_, err := o.Commit()
	require.NoError(t, err)
	return o
}

func load(t *testing.T, o *debitauthorization.Observer, address []byte) state.UserAccountHandler {
	t.Helper()
	acc, err := o.LoadAccount(address)
	require.NoError(t, err)
	return acc.(state.UserAccountHandler)
}

func TestObserverRecordsGrossDebitsEvenWhenCreditsExceedThem(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, recipient)
	require.NoError(t, acc.AddToBalance(big.NewInt(20)))
	require.NoError(t, acc.SubFromBalance(big.NewInt(5)))
	require.NoError(t, o.SaveAccount(acc))
	report, err := o.Finish()
	require.ErrorContains(t, err, "signed ceiling 0")
	result := report.Accounts[string(recipient)]
	require.Equal(t, "115", result.After.String())
	require.Equal(t, "5", result.Debits.String())
	require.Equal(t, "20", result.Credits.String())
}

func TestObserverAuthorizedDebitAndNegativeAddition(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(10)))
	require.NoError(t, acc.AddToBalance(big.NewInt(-2)))
	require.NoError(t, o.SaveAccount(acc))
	report, err := o.Finish()
	require.NoError(t, err)
	require.Equal(t, "12", report.Accounts[string(sender)].Debits.String())
	require.Equal(t, "88", report.Accounts[string(sender)].After.String())
}

func TestObserverCreditsDoNotRenewSignedBudget(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(12)))
	require.NoError(t, acc.AddToBalance(big.NewInt(12)))
	require.NoError(t, acc.SubFromBalance(big.NewInt(1)))
	require.NoError(t, o.SaveAccount(acc))
	_, err = o.Finish()
	require.ErrorContains(t, err, "debited 13, signed ceiling 12")
}

func TestObserverRelayedV3SeparatesAccountsAndCopiesLimits(t *testing.T) {
	o := setup(t)
	tx := fixtureTx()
	tx.RelayerAddr = bytes.Clone(relayer)
	tx.RelayerSignature = []byte("fixture only")
	executed, err := o.Begin(tx, fixtureVerified)
	require.NoError(t, err)
	tx.Value.SetInt64(999)
	tx.SndAddr[0] = 9
	require.Equal(t, "10", executed.Value.String())
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(10)))
	require.NoError(t, o.SaveAccount(acc))
	payer := load(t, o, relayer)
	require.NoError(t, payer.SubFromBalance(big.NewInt(2)))
	require.NoError(t, o.SaveAccount(payer))
	report, err := o.Finish()
	require.NoError(t, err)
	require.Zero(t, report.Accounts[string(sender)].FeeLimit.Sign())
	require.Zero(t, report.Accounts[string(relayer)].ValueLimit.Sign())
}

func TestObserverRevertRestoresBudgetWithoutTreatingRestorationAsSpending(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	snapshot := o.JournalLen()
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(12)))
	require.NoError(t, o.SaveAccount(acc))
	require.NoError(t, o.RevertToSnapshot(snapshot))
	acc = load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(12)))
	require.NoError(t, o.SaveAccount(acc))
	report, err := o.Finish()
	require.NoError(t, err)
	require.Equal(t, 1, report.RevertedEvents)
	require.Equal(t, "12", report.Accounts[string(sender)].Debits.String())
}

func TestObserverFailedSubtractionAndUnsavedChangesAreNotCommittedDebits(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, recipient)
	require.Error(t, acc.SubFromBalance(big.NewInt(101)))
	require.NoError(t, acc.SubFromBalance(big.NewInt(5))) // deliberately never saved
	report, err := o.Finish()
	require.NoError(t, err)
	require.Zero(t, report.Accounts[string(recipient)].Debits.Sign())
	require.Equal(t, "100", report.Accounts[string(recipient)].After.String())
}

func TestObserverPartialRollbackPreservesPreviouslyChargedFee(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(2)))
	require.NoError(t, o.SaveAccount(acc))
	snapshot := o.JournalLen()
	require.NoError(t, acc.SubFromBalance(big.NewInt(10)))
	require.NoError(t, o.SaveAccount(acc))
	require.NoError(t, o.RevertToSnapshot(snapshot))
	report, err := o.Finish()
	require.NoError(t, err)
	require.Equal(t, "2", report.Accounts[string(sender)].Debits.String())
	require.Equal(t, "98", report.Accounts[string(sender)].After.String())
	require.Equal(t, 1, report.RevertedEvents)
}

func TestObserverRelayedV3DoesNotGiveSenderTheRelayerFeeBudget(t *testing.T) {
	o := setup(t)
	tx := fixtureTx()
	tx.RelayerAddr = bytes.Clone(relayer)
	tx.RelayerSignature = []byte("fixture only")
	_, err := o.Begin(tx, fixtureVerified)
	require.NoError(t, err)
	acc := load(t, o, sender)
	require.NoError(t, acc.SubFromBalance(big.NewInt(11)))
	require.NoError(t, o.SaveAccount(acc))
	_, err = o.Finish()
	require.ErrorContains(t, err, "debited 11, signed ceiling 10")
}

func TestObserverReportsUnwrappedAccountSaves(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc, err := o.AccountsAdapter.LoadAccount(sender)
	require.NoError(t, err)
	require.NoError(t, o.SaveAccount(acc))
	_, err = o.Finish()
	require.ErrorContains(t, err, "outside the observation wrapper")
}

func TestObserverAccountRemovalRequiresAuthorization(t *testing.T) {
	o := setup(t)
	_, err := o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	require.NoError(t, o.RemoveAccount(recipient))
	_, err = o.Finish()
	require.ErrorContains(t, err, "debited 100, signed ceiling 0")
}

func TestObserverContractAccountIsOutsideTheInvariant(t *testing.T) {
	o := setup(t)
	sc := make([]byte, 32)
	sc[8] = 5
	sc[31] = 1
	require.True(t, core.IsSmartContractAddress(sc))
	acc := load(t, o, sc)
	require.NoError(t, acc.AddToBalance(big.NewInt(100)))
	require.NoError(t, o.SaveAccount(acc))
	_, err := o.Commit()
	require.NoError(t, err)
	_, err = o.Begin(fixtureTx(), fixtureVerified)
	require.NoError(t, err)
	acc = load(t, o, sc)
	require.NoError(t, acc.SubFromBalance(big.NewInt(20)))
	require.NoError(t, o.SaveAccount(acc))
	report, err := o.Finish()
	require.NoError(t, err)
	require.NotContains(t, report.Accounts, string(sc))
}

func TestObserverRejectsInvalidLifecycleAndTransactionIdentity(t *testing.T) {
	t.Run("verification failure creates no authority", func(t *testing.T) {
		o := setup(t)
		_, err := o.Begin(fixtureTx(), func(*transaction.Transaction) error { return errors.New("invalid signature") })
		require.ErrorContains(t, err, "invalid signature")
		_, err = o.Finish()
		require.ErrorContains(t, err, "no active execution")
	})
	t.Run("nonce must match", func(t *testing.T) {
		o := setup(t)
		tx := fixtureTx()
		tx.Nonce++
		_, err := o.Begin(tx, fixtureVerified)
		require.ErrorContains(t, err, "nonce")
	})
	t.Run("changed verifier input", func(t *testing.T) {
		o := setup(t)
		_, err := o.Begin(fixtureTx(), func(tx *transaction.Transaction) error { tx.Value.SetInt64(50); return nil })
		require.ErrorContains(t, err, "verifier changed transaction")
	})
	t.Run("changed execution input", func(t *testing.T) {
		o := setup(t)
		tx, err := o.Begin(fixtureTx(), fixtureVerified)
		require.NoError(t, err)
		tx.Value.SetInt64(50)
		_, err = o.Finish()
		require.ErrorContains(t, err, "changed after verification")
	})
	t.Run("finish before commit", func(t *testing.T) {
		o := setup(t)
		_, err := o.Begin(fixtureTx(), fixtureVerified)
		require.NoError(t, err)
		_, err = o.Begin(fixtureTx(), fixtureVerified)
		require.ErrorContains(t, err, "already active")
		_, err = o.Commit()
		require.ErrorContains(t, err, "Finish must precede")
		_, err = o.Finish()
		require.NoError(t, err)
	})
	t.Run("unmodeled legacy funding fails explicitly", func(t *testing.T) {
		o := setup(t)
		tx := fixtureTx()
		tx.Data = []byte("relayedTx@00")
		_, err := o.Begin(tx, fixtureVerified)
		require.ErrorContains(t, err, "legacy relayed profile unsupported")
	})
}
