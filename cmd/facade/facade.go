package facade

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

type Facade interface {
	//Starts the underlying node
	StartNode() error

	//Stops the underlying node
	StopNode() error

	//Starts the NTP clock with a set sync period
	StartNTP(clockSyncPeriod int)

	//Waits for the startTime to arrive and only after proceeds
	WaitForStartTime(t time.Time)

	//Starts all background services needed for the correct functionality of the node
	StartBackgroundServices(wg *sync.WaitGroup)

	//Sets the current logger
	SetLogger(logger *logger.Logger)

	//Gets if the underlying node is running
	IsNodeRunning() bool

	//Gets the current balance for a specified address
	GetBalance(address string) (*big.Int, error)

	//Generates a transaction from a sender, receiver, value and data
	GenerateTransaction(sender string, receiver string, value big.Int, data string) (*transaction.Transaction, error)

	//Gets the transaction with a specified hash
	GetTransaction(hash string) (*transaction.Transaction, error)
}
