// Package debitauthorization supplies an observational EGLD debit invariant for
// integration tests. It does not change consensus or reject account mutations.
package debitauthorization

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// VerifyTransaction must verify the exact transaction's signatures, chain,
// version, activation rules and account/guardian policy. A no-op verifier is
// suitable only for observer self-tests, never for a signature-coverage claim.
type VerifyTransaction func(*transaction.Transaction) error

// AccountResult separates spending from credits so a credit cannot hide a debit.
// Limits are signed ceilings, not an oracle for the exact fee charged.
type AccountResult struct {
	Before, After, Debits, Credits *big.Int
	ValueLimit, FeeLimit           *big.Int
}

// Report describes one observed execution. Events includes reverted writes;
// RevertedEvents counts those excluded from the committed-spending invariant.
type Report struct {
	ExecutionID    string
	Accounts       map[string]AccountResult
	Events         int
	RevertedEvents int
}

type balanceEvent struct {
	address  string
	delta    *big.Int
	index    int // -1 until saved to the accounts journal
	reverted bool
}

type allowance struct {
	value *big.Int
	fee   *big.Int
}

type execution struct {
	encoded    []byte
	tx         *transaction.Transaction
	allowances map[string]allowance
	before     map[string]*big.Int
	events     []*balanceEvent
	problems   []error
	generation uint64
}

// Observer decorates the real accounts adapter before node and VM construction.
// Executions must be serial, as with the integration-test transaction processor.
// Finish must run before Commit; snapshots may be reverted during an execution.
type Observer struct {
	state.AccountsAdapter
	active     *execution
	generation uint64
}

// New returns an account observer. Fixture initialization outside Begin/Finish
// is intentionally unobserved; the decorated adapter must be used by all tested
// consumers, and its underlying adapter must not be mutated directly.
func New(accounts state.AccountsAdapter) *Observer {
	return &Observer{AccountsAdapter: accounts}
}

// Begin verifies and copies a transaction and returns the copy to execute.
// The initial profile supports ordinary and relayed-v3 transactions. Legacy
// relayed envelopes fail explicitly until their funding rules are modeled.
func (o *Observer) Begin(tx *transaction.Transaction, verify VerifyTransaction) (*transaction.Transaction, error) {
	if o.active != nil {
		return nil, errors.New("debit authorization: execution already active")
	}
	if tx == nil || tx.Value == nil || tx.Value.Sign() < 0 || verify == nil || len(tx.Signature) == 0 {
		return nil, errors.New("debit authorization: invalid transaction, value, signature or verifier")
	}
	encoded, err := tx.Marshal()
	if err != nil {
		return nil, err
	}
	verified := &transaction.Transaction{}
	if err = verified.Unmarshal(encoded); err != nil {
		return nil, err
	}
	if err = verify(verified); err != nil {
		return nil, fmt.Errorf("debit authorization: verification: %w", err)
	}
	afterVerify, err := verified.Marshal()
	if err != nil || !bytes.Equal(encoded, afterVerify) {
		return nil, errors.New("debit authorization: verifier changed transaction")
	}
	function := strings.SplitN(string(verified.Data), "@", 2)[0]
	if function == core.RelayedTransaction || function == core.RelayedTransactionV2 {
		return nil, errors.New("debit authorization: legacy relayed profile unsupported")
	}
	if core.IsSmartContractAddress(verified.SndAddr) {
		return nil, errors.New("debit authorization: sender is not an ordinary account")
	}
	acc, err := o.AccountsAdapter.LoadAccount(verified.SndAddr)
	if err != nil {
		return nil, err
	}
	if acc.GetNonce() != verified.Nonce {
		return nil, errors.New("debit authorization: nonce does not match execution state")
	}
	feePayer := verified.SndAddr
	if len(verified.RelayerAddr) != 0 {
		if len(verified.RelayerSignature) == 0 || core.IsSmartContractAddress(verified.RelayerAddr) {
			return nil, errors.New("debit authorization: missing relayer authorization")
		}
		feePayer = verified.RelayerAddr
	}
	limits := map[string]allowance{
		string(verified.SndAddr): {value: new(big.Int).Set(verified.Value), fee: new(big.Int)},
	}
	fee := new(big.Int).Mul(new(big.Int).SetUint64(verified.GasLimit), new(big.Int).SetUint64(verified.GasPrice))
	payerLimit, ok := limits[string(feePayer)]
	if !ok {
		payerLimit = allowance{value: new(big.Int), fee: new(big.Int)}
	}
	payerLimit.fee.Set(fee)
	limits[string(feePayer)] = payerLimit
	o.generation++
	o.active = &execution{
		encoded: encoded, tx: verified, allowances: limits,
		before: make(map[string]*big.Int), generation: o.generation,
	}
	return verified, nil
}

// Finish checks all saved, non-reverted debit events and reconciles them with
// the actual account balances. Refunds are credits; they never refill limits.
// The caller must also assert exact fees, recipients, outcomes and state roots
// against independent expectations appropriate to its transaction profile.
func (o *Observer) Finish() (*Report, error) {
	run := o.active
	if run == nil {
		return nil, errors.New("debit authorization: no active execution")
	}
	defer func() { o.active = nil }()
	report := &Report{
		ExecutionID: fmt.Sprintf("%x", sha256.Sum256(run.encoded)),
		Accounts:    make(map[string]AccountResult), Events: len(run.events),
	}
	encoded, err := run.tx.Marshal()
	if err != nil || !bytes.Equal(encoded, run.encoded) {
		run.problems = append(run.problems, errors.New("executed transaction changed after verification"))
	}
	for address, before := range run.before {
		after, balanceErr := o.balance([]byte(address))
		if balanceErr != nil {
			run.problems = append(run.problems, balanceErr)
			continue
		}
		result := AccountResult{
			Before: new(big.Int).Set(before), After: after,
			Debits: new(big.Int), Credits: new(big.Int),
			ValueLimit: new(big.Int), FeeLimit: new(big.Int),
		}
		if limit, ok := run.allowances[address]; ok {
			result.ValueLimit.Set(limit.value)
			result.FeeLimit.Set(limit.fee)
		}
		report.Accounts[address] = result
	}
	for _, event := range run.events {
		if event.reverted {
			report.RevertedEvents++
		}
		if event.index < 0 || event.reverted {
			continue
		}
		result, ok := report.Accounts[event.address]
		if !ok {
			continue
		}
		if event.delta.Sign() < 0 {
			result.Debits.Sub(result.Debits, event.delta)
		} else {
			result.Credits.Add(result.Credits, event.delta)
		}
	}
	for address, result := range report.Accounts {
		limit := new(big.Int).Add(result.ValueLimit, result.FeeLimit)
		if result.Debits.Cmp(limit) > 0 {
			run.problems = append(run.problems, fmt.Errorf("account %x debited %s, signed ceiling %s", address, result.Debits, limit))
		}
		expected := new(big.Int).Add(result.Before, result.Credits)
		expected.Sub(expected, result.Debits)
		if expected.Cmp(result.After) != 0 {
			run.problems = append(run.problems, fmt.Errorf("account %x balance does not reconcile: observed %s, actual %s", address, expected, result.After))
		}
	}
	return report, errors.Join(run.problems...)
}

func (o *Observer) balance(address []byte) (*big.Int, error) {
	acc, err := o.AccountsAdapter.GetExistingAccount(address)
	if errors.Is(err, state.ErrAccNotFound) {
		return new(big.Int), nil
	}
	if err != nil {
		return nil, err
	}
	user, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, errors.New("debit authorization: unexpected account type")
	}
	return new(big.Int).Set(user.GetBalance()), nil
}

func (o *Observer) wrap(acc vmcommon.AccountHandler, err error) (vmcommon.AccountHandler, error) {
	if err != nil || check.IfNil(acc) {
		return acc, err
	}
	user, ok := acc.(state.UserAccountHandler)
	if !ok || core.IsSmartContractAddress(acc.AddressBytes()) {
		return acc, nil
	}
	return &observedAccount{UserAccountHandler: user, observer: o, address: bytes.Clone(acc.AddressBytes()), last: new(big.Int).Set(user.GetBalance())}, nil
}

// LoadAccount returns an account with its balance methods observed.
func (o *Observer) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	return o.wrap(o.AccountsAdapter.LoadAccount(address))
}

// GetExistingAccount returns an account with its balance methods observed.
func (o *Observer) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	return o.wrap(o.AccountsAdapter.GetExistingAccount(address))
}

// GetAccountFromBytes also observes accounts materialized by a consumer.
func (o *Observer) GetAccountFromBytes(address, encoded []byte) (vmcommon.AccountHandler, error) {
	return o.wrap(o.AccountsAdapter.GetAccountFromBytes(address, encoded))
}

func (o *Observer) remember(address []byte) {
	if o.active == nil {
		return
	}
	key := string(address)
	if _, ok := o.active.before[key]; ok {
		return
	}
	balance, err := o.balance(address)
	if err != nil {
		o.active.problems = append(o.active.problems, err)
		return
	}
	o.active.before[key] = balance
}

// SaveAccount preserves the real account type for serialization and associates
// observed changes with the real journal position at which they became durable.
func (o *Observer) SaveAccount(acc vmcommon.AccountHandler) error {
	observed, wrapped := acc.(*observedAccount)
	if wrapped {
		if observed.observer != o && o.active != nil {
			o.active.problems = append(o.active.problems, errors.New("account saved through a different observer"))
		}
		observed.capture()
		acc = observed.UserAccountHandler
	} else if o.active != nil && !check.IfNil(acc) && !core.IsSmartContractAddress(acc.AddressBytes()) {
		o.active.problems = append(o.active.problems, errors.New("ordinary account saved outside the observation wrapper"))
	}
	err := o.AccountsAdapter.SaveAccount(acc)
	if err == nil && wrapped && o.active != nil {
		for _, event := range observed.pending {
			event.index = o.AccountsAdapter.JournalLen()
		}
		observed.pending = nil
	}
	return err
}

// RemoveAccount observes destruction of an ordinary account's remaining funds.
func (o *Observer) RemoveAccount(address []byte) error {
	var event *balanceEvent
	if o.active != nil && !core.IsSmartContractAddress(address) {
		o.remember(address)
		balance, err := o.balance(address)
		if err != nil {
			return err
		}
		event = &balanceEvent{address: string(address), delta: new(big.Int).Neg(balance), index: -1}
	}
	err := o.AccountsAdapter.RemoveAccount(address)
	if err == nil && event != nil {
		event.index = o.AccountsAdapter.JournalLen()
		o.active.events = append(o.active.events, event)
	}
	return err
}

// RevertToSnapshot marks reverted spending as journal restoration. It does not
// create a new credit or allow a refund to replenish an authorization budget.
func (o *Observer) RevertToSnapshot(snapshot int) error {
	err := o.AccountsAdapter.RevertToSnapshot(snapshot)
	if err != nil {
		if o.active != nil {
			o.active.problems = append(o.active.problems, fmt.Errorf("state revert failed: %w", err))
		}
		return err
	}
	if o.active != nil {
		for _, event := range o.active.events {
			if event.index < 0 || event.index > snapshot {
				event.reverted = true
			}
		}
	}
	return nil
}

// Commit requires the caller to evaluate Finish before journal indexes reset.
func (o *Observer) Commit() ([]byte, error) {
	if o.active != nil {
		return nil, errors.New("debit authorization: Finish must precede Commit")
	}
	return o.AccountsAdapter.Commit()
}

// CommitInEpoch has the same observation boundary as Commit.
func (o *Observer) CommitInEpoch(current, target uint32) ([]byte, error) {
	if o.active != nil {
		return nil, errors.New("debit authorization: Finish must precede CommitInEpoch")
	}
	return o.AccountsAdapter.CommitInEpoch(current, target)
}

type observedAccount struct {
	state.UserAccountHandler
	observer   *Observer
	address    []byte
	last       *big.Int
	pending    []*balanceEvent
	generation uint64
}

// AccountDataHandler preserves the VM-common account interface used by built-ins
// and guardian validation; storage behavior remains on the underlying account.
func (a *observedAccount) AccountDataHandler() vmcommon.AccountDataHandler {
	return a.UserAccountHandler.(vmcommon.UserAccountHandler).AccountDataHandler()
}

var _ vmcommon.UserAccountHandler = (*observedAccount)(nil)
var _ state.AccountsAdapter = (*Observer)(nil)

func (a *observedAccount) capture() {
	run := a.observer.active
	current := a.UserAccountHandler.GetBalance()
	if run == nil {
		a.last.Set(current)
		a.pending = nil
		return
	}
	if a.generation != run.generation {
		a.pending = nil
		a.generation = run.generation
	}
	a.observer.remember(a.address)
	if !bytes.Equal(a.address, a.UserAccountHandler.AddressBytes()) {
		run.problems = append(run.problems, errors.New("observed account identity changed"))
	}
	delta := new(big.Int).Sub(current, a.last)
	if delta.Sign() != 0 {
		event := &balanceEvent{address: string(a.address), delta: delta, index: -1}
		run.events = append(run.events, event)
		a.pending = append(a.pending, event)
	}
	a.last.Set(current)
}

func (a *observedAccount) AddToBalance(value *big.Int) error {
	a.capture()
	err := a.UserAccountHandler.AddToBalance(value)
	a.capture()
	return err
}

func (a *observedAccount) SubFromBalance(value *big.Int) error {
	a.capture()
	err := a.UserAccountHandler.SubFromBalance(value)
	a.capture()
	return err
}
