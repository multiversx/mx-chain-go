package common

import "github.com/ElrondNetwork/elrond-go-core/data/outport"

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

type AlteredAccountTokenAPIResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AlteredAccountAPIResponse struct {
	Address string                      `json:"address"`
	Balance string                      `json:"balance"`
	Tokens  []*outport.AccountTokenData `json:"tokens,omitempty"`
}

type AlteredAccountsForBlockAPIResponse struct {
	Accounts []*AlteredAccountAPIResponse `json:"accounts"`
}
