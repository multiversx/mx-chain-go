package facade

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

type Facade interface {
	//StartNode starts the underlying node
	StartNode() error

	//StopNode stops the underlying node
	StopNode() error

	//StartNTP starts the NTP clock with a set sync period
	StartNTP(clockSyncPeriod int)

	//WaitForStartTime waits for the startTime to arrive and only after proceeds
	WaitForStartTime(t time.Time)

	//StartBackgroundServices starts all background services needed for the correct functionality of the node
	StartBackgroundServices(wg *sync.WaitGroup)

	//SetLogger sets the current logger
	SetLogger(logger *logger.Logger)

	//IsNodeRunning gets if the underlying node is running
	IsNodeRunning() bool

	//GetBalance gets the current balance for a specified address
	GetBalance(address string) (*big.Int, error)

	//GenerateTransaction generates a transaction from a sender, receiver, value and data
	GenerateTransaction(sender string, receiver string, value big.Int, data string) (*transaction.Transaction, error)

	//GetTransaction gets the transaction with a specified hash
	GetTransaction(hash string) (*transaction.Transaction, error)
}
