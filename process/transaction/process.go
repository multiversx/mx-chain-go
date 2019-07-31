package transaction

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.DefaultLogger()

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	scProcessor      process.SmartContractProcessor
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(
	accounts state.AccountsAdapter,
	hasher hashing.Hasher,
	addressConv state.AddressConverter,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	scProcessor process.SmartContractProcessor,
) (*txProcessor, error) {

	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if addressConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if scProcessor == nil {
		return nil, process.ErrNilSmartContractProcessor
	}

	return &txProcessor{
		accounts:         accounts,
		hasher:           hasher,
		adrConv:          addressConv,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
		scProcessor:      scProcessor,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction, roundIndex uint32) error {
	if tx == nil {
		return process.ErrNilTransaction
	}

	adrSrc, adrDst, err := txProc.getAddresses(tx)
	if err != nil {
		return err
	}

	txType, err := txProc.scProcessor.ComputeTransactionType(tx)
	if err != nil {
		return err
	}

	switch txType {
	case process.MoveBalance:
		return txProc.processMoveBalance(tx, adrSrc, adrDst)
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, adrSrc, roundIndex)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, adrSrc, adrDst, roundIndex)
	}

	return process.ErrWrongTransaction
}

func (txProc *txProcessor) processMoveBalance(
	tx *transaction.Transaction,
	adrSrc, adrDst state.AddressContainer,
) error {

	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	value := tx.Value
	// is sender address in node shard
	if acntSrc != nil {
		err = txProc.checkTxValues(acntSrc, value, tx.Nonce)
		if err != nil {
			return err
		}
	}

	err = txProc.moveBalances(acntSrc, acntDst, value)
	if err != nil {
		return err
	}

	// is sender address in node shard
	if acntSrc != nil {
		err = txProc.increaseNonce(acntSrc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) processSCDeployment(
	tx *transaction.Transaction,
	adrSrc state.AddressContainer,
	roundIndex uint32,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, err := txProc.getAccountFromAddress(adrSrc)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.DeploySmartContract(tx, acntSrc, roundIndex)
	return err
}

func (txProc *txProcessor) processSCInvoking(
	tx *transaction.Transaction,
	adrSrc, adrDst state.AddressContainer,
	roundIndex uint32,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst, roundIndex)
	return err
}

func (txProc *txProcessor) getAddresses(tx *transaction.Transaction) (adrSrc, adrDst state.AddressContainer, err error) {
	//for now we assume that the address = public key
	adrSrc, err = txProc.adrConv.CreateAddressFromPublicKeyBytes(tx.SndAddr)
	if err != nil {
		return
	}
	adrDst, err = txProc.adrConv.CreateAddressFromPublicKeyBytes(tx.RcvAddr)
	return
}

func (txProc *txProcessor) getAccounts(adrSrc, adrDst state.AddressContainer,
) (acntSrc, acntDst *state.Account, err error) {

	shardForCurrentNode := txProc.shardCoordinator.SelfId()
	shardForSrc := txProc.shardCoordinator.ComputeId(adrSrc)
	shardForDst := txProc.shardCoordinator.ComputeId(adrDst)

	srcInShard := shardForSrc == shardForCurrentNode
	dstInShard := shardForDst == shardForCurrentNode

	if srcInShard && adrSrc == nil ||
		dstInShard && adrDst == nil {
		return nil, nil, process.ErrNilAddressContainer
	}

	if bytes.Equal(adrSrc.Bytes(), adrDst.Bytes()) {
		acntWrp, err := txProc.accounts.GetAccountWithJournal(adrSrc)
		if err != nil {
			return nil, nil, err
		}

		account, ok := acntWrp.(*state.Account)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		return account, account, nil
	}

	if srcInShard {
		acntSrcWrp, err := txProc.accounts.GetAccountWithJournal(adrSrc)
		if err != nil {
			return nil, nil, err
		}

		account, ok := acntSrcWrp.(*state.Account)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		acntSrc = account
	}

	if dstInShard {
		acntDstWrp, err := txProc.accounts.GetAccountWithJournal(adrDst)
		if err != nil {
			return nil, nil, err
		}

		account, ok := acntDstWrp.(*state.Account)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		acntDst = account
	}

	return
}

func (txProc *txProcessor) getAccountFromAddress(adrSrc state.AddressContainer) (state.AccountHandler, error) {
	shardForCurrentNode := txProc.shardCoordinator.SelfId()
	shardForSrc := txProc.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := txProc.accounts.GetAccountWithJournal(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

func (txProc *txProcessor) checkTxValues(acntSrc *state.Account, value *big.Int, nonce uint64) error {
	if acntSrc.Nonce < nonce {
		return process.ErrHigherNonceInTransaction
	}

	if acntSrc.Nonce > nonce {
		return process.ErrLowerNonceInTransaction
	}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate
	if acntSrc.Balance.Cmp(value) < 0 {
		return process.ErrInsufficientFunds
	}

	return nil
}

func (txProc *txProcessor) moveBalances(acntSrc, acntDst *state.Account,
	value *big.Int,
) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	// is sender address in node shard
	if acntSrc != nil {
		err := acntSrc.SetBalanceWithJournal(operation1.Sub(acntSrc.Balance, value))
		if err != nil {
			return err
		}
	}

	// is receiver address in node shard
	if acntDst != nil {
		err := acntDst.SetBalanceWithJournal(operation2.Add(acntDst.Balance, value))
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) increaseNonce(acntSrc *state.Account) error {
	return acntSrc.SetNonceWithJournal(acntSrc.Nonce + 1)
}
