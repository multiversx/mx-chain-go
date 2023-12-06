package dtos

// AddressState will hold the address state
type AddressState struct {
	Address      string            `json:"address"`
	Nonce        uint64            `json:"nonce,omitempty"`
	Balance      string            `json:"balance,omitempty"`
	Code         string            `json:"code,omitempty"`
	RootHash     string            `json:"rootHash,omitempty"`
	CodeMetadata string            `json:"codeMetadata,omitempty"`
	Owner        string            `json:"owner,omitempty"`
	Keys         map[string]string `json:"keys,omitempty"`
}
