package dtos

// WalletKey holds the public and the private key of a wallet
type WalletKey struct {
	Address       WalletAddress `json:"address"`
	PrivateKeyHex string        `json:"privateKeyHex"`
}

// InitialWalletKeys holds the initial wallet keys
type InitialWalletKeys struct {
	StakeWallets   []*WalletKey          `json:"stakeWallets"`
	BalanceWallets map[uint32]*WalletKey `json:"balanceWallets"`
}

// WalletAddress holds the address in multiple formats
type WalletAddress struct {
	Bech32 string `json:"bech32"`
	Bytes  []byte `json:"bytes"`
}

// BLSKey holds the BLS key in multiple formats
type BLSKey struct {
	Hex   string
	Bytes []byte
}
