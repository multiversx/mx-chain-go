package dtos

// AddressState will hold the address state
type AddressState struct {
	Address string `json:"address"`
	// ShardID: This field is needed for the system account address (it is the same on all shards).
	ShardID          uint32            `json:"shardID,omitempty"`
	Nonce            uint64            `json:"nonce,omitempty"`
	Balance          string            `json:"balance,omitempty"`
	Code             string            `json:"code,omitempty"`
	RootHash         string            `json:"rootHash,omitempty"`
	CodeMetadata     string            `json:"codeMetadata,omitempty"`
	CodeHash         string            `json:"codeHash,omitempty"`
	DeveloperRewards string            `json:"developerReward,omitempty"`
	Owner            string            `json:"ownerAddress,omitempty"`
	Keys             map[string]string `json:"keys,omitempty"`
}
