package dtos

// WalletKey holds the public and the private key of a wallet bey
type WalletKey struct {
	Address       string `json:"address"`
	PrivateKeyHex string `json:"privateKeyHex"`
}

// InitialWalletKeys holds the initial wallet keys
type InitialWalletKeys struct {
	InitialWalletWithStake *WalletKey            `json:"initialWalletWithStake"`
	ShardWallets           map[uint32]*WalletKey `json:"shardWallets"`
}
