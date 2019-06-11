package process

// BlockHeaderState specifies which is the state of the block header received
type BlockHeaderState int

const (
	// BHReceived defines ID of a received block header
	BHReceived BlockHeaderState = iota
	// BHProcessed defines ID of a processed block header
	BHProcessed
	// BHProposed defines ID of a proposed block header
	BHProposed
)

// TransactionType specifies the type of the transaction
type TransactionType int

const (
	// MoveBalance defines ID of a payment transaction - moving balances
	MoveBalance TransactionType = iota
	// SCDeployment defines ID of a transaction to store a smart contract
	SCDeployment
	// SCInvoking defines ID of a transaction of type smart contract call
	SCInvoking
)

const ShardBlockFinality = 1
const MetablockFinality = 1
