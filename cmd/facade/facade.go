package facade

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

type Facade interface {
	StartNode() error
	StopNode() error
	StartNTP(clockSyncPeriod int)
	WaitForStartTime(t time.Time)
	StartBackgroundServices(wg *sync.WaitGroup)
	SetLogger(logger *logger.Logger)
	IsNodeRunning() bool
	GetBalance(address string) (*big.Int, error)
	GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error)
	GetTransaction(hash string) (*transaction.Transaction, error)
}
