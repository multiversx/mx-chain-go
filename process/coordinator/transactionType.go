package coordinator

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
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
	if check.IfNil(adrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
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
func (tth *txTypeHandler) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	err := tth.checkTxValidity(tx)
	if err != nil {
		return process.InvalidTransaction, err
	}

	_, isRewardTx := tx.(*rewardTx.RewardTx)
	if isRewardTx {
		return process.RewardTx, nil
	}

	isEmptyAddress := tth.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment, nil
		}
		return process.InvalidTransaction, process.ErrWrongTransaction
	}

	acntDst, err := tth.getAccountFromAddress(tx.GetRecvAddress())
	if err != nil {
		return process.InvalidTransaction, err
	}

	if check.IfNil(acntDst) {
		return process.MoveBalance, nil
	}

	if len(tx.GetData()) > 0 && core.IsSmartContractAddress(tx.GetRecvAddress()) {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (tth *txTypeHandler) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRecvAddress(), make([]byte, tth.adrConv.AddressLen()))
	return isEmptyAddress
}

func (tth *txTypeHandler) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	adrSrc, err := tth.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := tth.shardCoordinator.SelfId()
	shardForSrc := tth.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := tth.accounts.LoadAccount(adrSrc)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

func (tth *txTypeHandler) checkTxValidity(tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := tth.adrConv.AddressLen() != len(tx.GetRecvAddress())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tth *txTypeHandler) IsInterfaceNil() bool {
	return tth == nil
}
