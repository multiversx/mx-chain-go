package transaction

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.NewDefaultLogger()

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

	if bytes.Equal(adrDst.Bytes(), state.RegistrationAddress.Bytes()) {
		regAccount, err := txProc.accounts.GetJournalizedAccount(state.RegistrationAddress)
		if err != nil {
			return err
		}

		regData := &state.RegistrationData{}
		err = txProc.marshalizer.Unmarshal(regData, tx.Data)
		if err != nil {
			return err
		}
		regData.OriginatorPubKey = adrSrc.Bytes()
		regData.RoundIndex = roundIndex

		err = regAccount.AppendDataRegistrationWithJournal(regData)
		if err != nil {
			return err
		}
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
	if acntDst != nil && acntDst.Code() != nil {
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
		err2 := txProc.accounts.RevertToSnapshot(0)

		if err2 != nil {
			log.Error(err2.Error())
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

	account, err := txProc.accounts.GetJournalizedAccount(addrContainer)

	if err != nil {
		return err
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
) (acntSrc, acntDst state.JournalizedAccountWrapper, err error) {
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
		acnt, err := txProc.accounts.GetJournalizedAccount(adrSrc)
		if err != nil {
			return nil, nil, err
		}

		return acnt, acnt, nil
	}

	if srcInShard {
		acntSrc, err = txProc.accounts.GetJournalizedAccount(adrSrc)
		if err != nil {
			return nil, nil, err
		}
	}

	if dstInShard {
		acntDst, err = txProc.accounts.GetJournalizedAccount(adrDst)
		if err != nil {
			return nil, nil, err
		}
	}

	return acntSrc, acntDst, nil
}

func (txProc *txProcessor) callSCHandler(tx *transaction.Transaction) error {
	if txProc.scHandler == nil {
		return process.ErrNoVM
	}

	return txProc.scHandler(txProc.accounts, tx)
}

func (txProc *txProcessor) checkTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	//TODO: undo this for nonce checking and un-skip tests
	//if acntSrc.BaseAccount().Nonce < nonce {
	//	return process.ErrHigherNonceInTransaction
	//}
	//
	//if acntSrc.BaseAccount().Nonce > nonce {
	//	return process.ErrLowerNonceInTransaction
	//}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate

	if acntSrc.BaseAccount().Balance.Cmp(value) < 0 {
		return process.ErrInsufficientFunds
	}

	return nil
}

func (txProc *txProcessor) moveBalances(acntSrc, acntDst state.JournalizedAccountWrapper,
	value *big.Int,
) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	// is sender address in node shard
	if acntSrc != nil {
		err := acntSrc.SetBalanceWithJournal(operation1.Sub(acntSrc.BaseAccount().Balance, value))
		if err != nil {
			return err
		}
	}

	// is receiver address in node shard
	if acntDst != nil {
		err := acntDst.SetBalanceWithJournal(operation2.Add(acntDst.BaseAccount().Balance, value))
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) increaseNonce(acntSrc state.JournalizedAccountWrapper) error {
	return acntSrc.SetNonceWithJournal(acntSrc.BaseAccount().Nonce + 1)
}
