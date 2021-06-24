package api

// AccountResponse is the data transfer object to be returned on API when requesting an address data
type AccountResponse struct {
	Address         string `json:"address"`
	Nonce           uint64 `json:"nonce"`
	Balance         string `json:"balance"`
	Username        string `json:"username"`
	Code            string `json:"code"`
	CodeHash        []byte `json:"codeHash"`
	RootHash        []byte `json:"rootHash"`
	CodeMetadata    []byte `json:"codeMetadata"`
	DeveloperReward string `json:"developerReward"`
	OwnerAddress    string `json:"ownerAddress"`
}
