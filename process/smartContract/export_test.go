package smartContract

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (sc *scProcessor) CreateSCRTransactions(
	crossOutAccs []*vmcommon.OutputAccount,
	tx *transaction.Transaction,
	txHash []byte,
) ([]data.TransactionHandler, error) {
	return sc.processSCOutputAccounts(crossOutAccs, tx, txHash)
}
