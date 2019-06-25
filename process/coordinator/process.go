package coordinator

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

type transactionCoordinator struct {
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
}

func NewTransactionCoordinator(
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
) (*transactionCoordinator, error) {
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	tc := &transactionCoordinator{
		adrConv:          adrConv,
		shardCoordinator: shardCoordinator,
		accounts:         accounts,
	}

	return tc, nil
}

func (tc *transactionCoordinator) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	err := tc.checkTxValidity(tx)
	if err != nil {
		return 0, err
	}

	isEmptyAddress := tc.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment, nil
		}
		return 0, process.ErrWrongTransaction
	}

	acntDst, err := tc.getAccountFromAddress(tx.GetRecvAddress())
	if err != nil {
		return 0, err
	}

	if acntDst == nil {
		return process.MoveBalance, nil
	}

	if !acntDst.IsInterfaceNil() && len(acntDst.GetCode()) > 0 {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (tc *transactionCoordinator) IsDataPreparedForProcessing(haveTime func() time.Duration) {
	panic("implement me")
}

func (tc *transactionCoordinator) SaveBlockDataToStorage(body block.Body) error {
	panic("implement me")
}

func (tc *transactionCoordinator) RestoreBlockDataFromStorage(body block.Body) (int, map[int][]byte, error) {
	panic("implement me")
}

func (tc *transactionCoordinator) RemoveBlockDataFromStorage(body block.Body) error {
	panic("implement me")
}

func (tc *transactionCoordinator) ProcessBlockTransaction(body block.Body, round uint32, haveTime func() time.Duration) error {
	panic("implement me")
}

func (tc *transactionCoordinator) CreateAndProcessCrossShardTransactions(header data.HeaderHandler, round uint32, haveTime func()) (bool, error) {
	panic("implement me")
}

func (tc *transactionCoordinator) GetAllCurrentUsedTxs(transactionType process.TransactionType) map[string]data.TransactionHandler {
	panic("implement me")
}

func (tc *transactionCoordinator) checkTxValidity(tx data.TransactionHandler) error {
	if tx == nil || tx.IsInterfaceNil() {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := tc.adrConv.AddressLen() != len(tx.GetRecvAddress())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

func (tc *transactionCoordinator) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRecvAddress(), make([]byte, tc.adrConv.AddressLen()))
	return isEmptyAddress
}

func (tc *transactionCoordinator) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
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
