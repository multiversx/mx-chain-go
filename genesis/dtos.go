package genesis

import "github.com/ElrondNetwork/elrond-go-core/data"

// IndexingData specifies transactions sets that will be used for indexing
type IndexingData struct {
	DelegationTxs      []data.TransactionHandler
	ScrsTxs            map[string]data.TransactionHandler
	StakingTxs         []data.TransactionHandler
	DeploySystemScTxs  []data.TransactionHandler
	DeployInitialScTxs []data.TransactionHandler
}
