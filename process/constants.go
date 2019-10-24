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
	// BHNotarized defines ID of a notarized block header
	BHNotarized
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
	// RewardTx defines ID of a reward transaction
	RewardTx
	// InvalidTransaction defines unknown transaction type
	InvalidTransaction
)

const ShardBlockFinality = 1
const MetaBlockFinality = 1
const MaxHeaderRequestsAllowed = 10
const MaxItemsInBlock = 15000
const MinItemsInBlock = 25
const MaxNoncesDifference = 5

// TODO - calculate exactly in case of the VM, for every VM to have a similar constant, operations / seconds
const MaxGasLimitPerMiniBlock = uint64(1500000)
const MaxRequestsWithTimeoutAllowed = 3

// MaxHeadersToRequestInAdvance defines the maximum number of headers which will be requested in advance if they are missing
const MaxHeadersToRequestInAdvance = 10

// RoundModulusTrigger defines a round modulus on which a trigger for an action will be released
const RoundModulusTrigger = 5

// MaxOccupancyPercentageAllowed defines the maximum occupancy percentage allowed to be used,
// from the full pool capacity, for the received data which are not needed in the near future
const MaxOccupancyPercentageAllowed = float64(0.9)
