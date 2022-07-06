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
