package common

// GetProofResponse is a struct that stores the response of a GetProof API request
type GetProofResponse struct {
	Proof    [][]byte
	Value    []byte
	RootHash string
}

// TransactionsPoolAPIResponse is a struct that holds the data to be returned when getting the transaction pool from an API call
type TransactionsPoolAPIResponse struct {
	RegularTransactions  []string `json:"regularTransactions"`
	SmartContractResults []string `json:"smartContractResults"`
	Rewards              []string `json:"rewards"`
}

// TransactionsForSenderApiResponse is a struct that holds the data to be returned when getting the transactions for a sender from an API call
type TransactionsForSenderApiResponse struct {
	Sender       string   `json:"sender"`
	Transactions []string `json:"transactions"`
}
