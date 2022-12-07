package factory

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts"
	"github.com/ElrondNetwork/elrond-go/outport/process/transactionsfee"
)

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

func checkArgOutportDataProviderFactory(arg ArgOutportDataProviderFactory) error {
	if check.IfNil(arg.AddressConverter) {
		return alteredaccounts.ErrNilPubKeyConverter
	}
	if check.IfNil(arg.AccountsDB) {
		return alteredaccounts.ErrNilAccountsDB
	}
	if check.IfNil(arg.Marshaller) {
		return transactionsfee.ErrNilMarshaller
	}
	if check.IfNil(arg.EsdtDataStorageHandler) {
		return alteredaccounts.ErrNilESDTDataStorageHandler
	}
	if check.IfNil(arg.TransactionsStorer) {
		return transactionsfee.ErrNilStorage
	}
	if check.IfNil(arg.EconomicsData) {
		return transactionsfee.ErrNilTransactionFeeCalculator
	}
	if check.IfNil(arg.ShardCoordinator) {
		return transactionsfee.ErrNilShardCoordinator
	}
	if check.IfNil(arg.TxCoordinator) {
		return process.ErrNilTransactionCoordinator
	}
	if check.IfNil(arg.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.GasConsumedProvider) {
		return process.ErrNilGasConsumedProvider
	}
	if check.IfNil(arg.Hasher) {
		return ErrNilHasher
	}

	return nil
}
