package transaction

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.DefaultLogger()

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	hasher           hashing.Hasher
	scHandler        func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
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

	return &txProcessor{
		accounts:         accounts,
		hasher:           hasher,
		adrConv:          addressConv,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
	}, nil
}

// SCHandler returns the smart contract execution function
func (txProc *txProcessor) SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	return txProc.scHandler
}

// SetSCHandler sets the smart contract execution function
func (txProc *txProcessor) SetSCHandler(f func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	txProc.scHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction, roundIndex int32) error {
	if tx == nil {
		return process.ErrNilTransaction
	}

	adrSrc, adrDst, err := txProc.getAddresses(tx)
	if err != nil {
		return err
	}

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

	// is receiver address in node shard and this address contains a SC code
	if acntDst != nil && acntDst.GetCode() != nil {
		err = txProc.callSCHandler(tx)
		if err != nil {
			// TODO: Revert state if SC execution failed and substract only some fee needed for SC job done
			return err
		}
	}

	return nil
}

// SetBalancesToTrie adds balances to trie
func (txProc *txProcessor) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	if txProc.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if accBalance == nil {
		return nil, process.ErrNilValue
	}

	for i, v := range accBalance {
		err := txProc.setBalanceToTrie([]byte(i), v)

		if err != nil {
			return nil, err
		}
	}

	rootHash, err = txProc.accounts.Commit()
	if err != nil {
		errToLog := txProc.accounts.RevertToSnapshot(0)
		if errToLog != nil {
			log.Error(errToLog.Error())
		}

		return nil, err
	}

	return rootHash, err
}

func (txProc *txProcessor) setBalanceToTrie(addr []byte, balance *big.Int) error {
	if addr == nil {
		return process.ErrNilValue
	}

	addrContainer, err := txProc.adrConv.CreateAddressFromPublicKeyBytes(addr)
	if err != nil {
		return err
	}
	if addrContainer == nil {
		return process.ErrNilAddressContainer
	}

	accWrp, err := txProc.accounts.GetAccountWithJournal(addrContainer)
	if err != nil {
		return err
	}

	account, ok := accWrp.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return account.SetBalanceWithJournal(balance)
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

func (txProc *txProcessor) callSCHandler(tx *transaction.Transaction) error {
	if txProc.scHandler == nil {
		return process.ErrNoVM
	}

	return txProc.scHandler(txProc.accounts, tx)
}

func (txProc *txProcessor) checkTxValues(acntSrc *state.Account, value *big.Int, nonce uint64) error {
	// TODO order transactions - than uncomment this
	//if acntSrc.Nonce < nonce {
	//	return process.ErrHigherNonceInTransaction
	//}

	//if acntSrc.Nonce > nonce {
	//	return process.ErrLowerNonceInTransaction
	//}

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
