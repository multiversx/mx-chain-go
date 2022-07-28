package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts"
	"github.com/ElrondNetwork/elrond-go/outport/process/transactionsfee"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type ArgOutportDataProviderFactory struct {
	AddressConverter       core.PubkeyConverter
	AccountsDB             state.AccountsAdapter
	Marshalizer            marshal.Marshalizer
	EsdtDataStorageHandler vmcommon.ESDTNFTStorageHandler
	TransactionsStorer     storage.Storer
	TxFeeCalculator        transactionsfee.FeesProcessorHandler
	process.ArgOutportDataProvider
}

// CreateOutportDataProvider will create a new instance of outport.DataProviderOutport
func CreateOutportDataProvider(arg ArgOutportDataProviderFactory) (outport.DataProviderOutport, error) {
	alteredAccountsProvider, err := alteredaccounts.NewAlteredAccountsProvider(alteredaccounts.ArgsAlteredAccountsProvider{
		ShardCoordinator:       arg.ShardCoordinator,
		AddressConverter:       arg.AddressConverter,
		AccountsDB:             arg.AccountsDB,
		Marshalizer:            arg.Marshalizer,
		EsdtDataStorageHandler: arg.EsdtDataStorageHandler,
	})
	if err != nil {
		return nil, err
	}

	transactionsFeeProc, err := transactionsfee.NewTransactionFeeProcessor(transactionsfee.ArgTransactionsFeeProcessor{
		Marshalizer:        arg.Marshalizer,
		TransactionsStorer: arg.TransactionsStorer,
		ShardCoordinator:   arg.ShardCoordinator,
		TxFeeCalculator:    arg.TxFeeCalculator,
	})
	if err != nil {
		return nil, err
	}

	return process.NewOutportDataProvider(process.ArgOutportDataProvider{
		ShardCoordinator:         arg.ShardCoordinator,
		AlteredAccountsProvider:  alteredAccountsProvider,
		TransactionsFeeProcessor: transactionsFeeProc,
		TxCoordinator:            arg.TxCoordinator,
		NodesCoordinator:         arg.NodesCoordinator,
		GasConsumedProvider:      arg.GasConsumedProvider,
		EconomicsData:            arg.EconomicsData,
	})
}
