package process

import (
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process/transactionsfee"
)

// AlteredAccountsProviderHandler defines the functionality needed for provisioning of altered accounts when indexing data
type AlteredAccountsProviderHandler interface {
	ExtractAlteredAccountsFromPool(txPool *outport.Pool) (map[string]*outport.AlteredAccount, error)
	IsInterfaceNil() bool
}

type TransactionsFeeHandler interface {
	IsInterfaceNil() bool
	PutFeeAndGasUsed(pool *outport.Pool) error
}

type GasConsumedProvider interface {
	TotalGasProvided() uint64
	TotalGasProvidedWithScheduled() uint64
	TotalGasRefunded() uint64
	TotalGasPenalized() uint64
	IsInterfaceNil() bool
}

type EconomicsDataHandler interface {
	transactionsfee.FeesProcessorHandler
	MaxGasLimitPerBlock(shardID uint32) uint64
}
