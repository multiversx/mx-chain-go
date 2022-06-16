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

// AuctionListValidatorAPIResponse holds the data needed for an auction node validator for responding to API calls
type AuctionListValidatorAPIResponse struct {
	Owner   string `json:"owner"`
	NodeKey string `json:"nodeKey"`
	TopUp   string `json:"topUp"`
}
