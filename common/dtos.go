package common

import (
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
)

// GetProofResponse is a struct that stores the response of a GetProof API request
type GetProofResponse struct {
	Proof    [][]byte
	Value    []byte
	RootHash string
}

// TransactionsPoolAPIResponse is a struct that holds the data to be returned when getting the transaction pool from an API call
type TransactionsPoolAPIResponse struct {
	RegularTransactions  []Transaction `json:"regularTransactions"`
	SmartContractResults []Transaction `json:"smartContractResults"`
	Rewards              []Transaction `json:"rewards"`
}

// Transaction is a struct that holds transaction fields to be returned when getting the transactions from pool
type Transaction struct {
	TxFields map[string]interface{} `json:"txFields"`
}

// TransactionsPoolForSenderApiResponse is a struct that holds the data to be returned when getting the transactions for a sender from an API call
type TransactionsPoolForSenderApiResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// NonceGapApiResponse is a struct that holds a nonce gap from transactions pool
// From - first unknown nonce
// To   - last unknown nonce
type NonceGapApiResponse struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

// TransactionsPoolNonceGapsForSenderApiResponse is a struct that holds the data to be returned when getting the nonce gaps from transactions pool for a sender from an API call
type TransactionsPoolNonceGapsForSenderApiResponse struct {
	Sender string                `json:"sender"`
	Gaps   []NonceGapApiResponse `json:"gaps"`
}

// DelegationDataAPI will be used when requesting the genesis balances from API
type DelegationDataAPI struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

// InitialAccountAPI represents the structure to be returned when requesting the genesis balances from API
type InitialAccountAPI struct {
	Address      string            `json:"address"`
	Supply       string            `json:"supply"`
	Balance      string            `json:"balance"`
	StakingValue string            `json:"stakingvalue"`
	Delegation   DelegationDataAPI `json:"delegation"`
}

// EpochStartDataAPI holds fields from the first block in a given epoch
type EpochStartDataAPI struct {
	Nonce             uint64 `json:"nonce"`
	Round             uint64 `json:"round"`
	Timestamp         int64  `json:"timestamp"`
	Epoch             uint32 `json:"epoch"`
	Shard             uint32 `json:"shard"`
	PrevBlockHash     string `json:"prevBlockHash"`
	StateRootHash     string `json:"stateRootHash"`
	ScheduledRootHash string `json:"scheduledRootHash"`
	AccumulatedFees   string `json:"accumulatedFees,omitempty"`
	DeveloperFees     string `json:"developerFees,omitempty"`
}

// AlteredAccountsForBlockAPIResponse holds the altered accounts for a certain block
type AlteredAccountsForBlockAPIResponse struct {
	Accounts []*alteredAccount.AlteredAccount `json:"accounts"`
}

// AuctionNode holds data needed for a node in auction to respond to API calls
type AuctionNode struct {
	BlsKey    string `json:"blsKey"`
	Qualified bool   `json:"qualified"`
}

// AuctionListValidatorAPIResponse holds the data needed for an auction node validator for responding to API calls
type AuctionListValidatorAPIResponse struct {
	Owner          string         `json:"owner"`
	NumStakedNodes int64          `json:"numStakedNodes"`
	TotalTopUp     string         `json:"totalTopUp"`
	TopUpPerNode   string         `json:"topUpPerNode"`
	QualifiedTopUp string         `json:"qualifiedTopUp"`
	AuctionList    []*AuctionNode `json:"auctionList"`
}
