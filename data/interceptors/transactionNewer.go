package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type TransactionNewer struct {
	*transaction.Transaction
}
