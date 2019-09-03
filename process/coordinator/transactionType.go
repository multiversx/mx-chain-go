package coordinator

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type txTypeHandler struct {
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
}

// NewTxTypeHandler creates a transaction type handler
func NewTxTypeHandler(
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
) (*txTypeHandler, error) {
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	tc := &txTypeHandler{
		adrConv:          adrConv,
		shardCoordinator: shardCoordinator,
		accounts:         accounts,
	}

	return tc, nil
}

// ComputeTransactionType calculates the transaction type
func (tc *txTypeHandler) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	err := tc.checkTxValidity(tx)
	if err != nil {
		return process.InvalidTransaction, err
	}

	_, isTxfee := tx.(*feeTx.FeeTx)
	if isTxfee {
		return process.TxFee, nil
	}

	isEmptyAddress := tc.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment, nil
		}
		return process.InvalidTransaction, process.ErrWrongTransaction
	}

	acntDst, err := tc.getAccountFromAddress(tx.GetRecvAddress())
	if err != nil {
		return process.InvalidTransaction, err
	}

	if acntDst == nil {
		return process.MoveBalance, nil
	}

	if !acntDst.IsInterfaceNil() && len(acntDst.GetCode()) > 0 {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (tc *txTypeHandler) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRecvAddress(), make([]byte, tc.adrConv.AddressLen()))
	return isEmptyAddress
}

func (tc *txTypeHandler) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	adrSrc, err := tc.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := tc.shardCoordinator.SelfId()
	shardForSrc := tc.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := tc.accounts.GetAccountWithJournal(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

func (tc *txTypeHandler) checkTxValidity(tx data.TransactionHandler) error {
	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := tc.adrConv.AddressLen() != len(tx.GetRecvAddress())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tx *txTypeHandler) IsInterfaceNil() bool {
	if tx == nil {
		return true
	}
	return false
}
