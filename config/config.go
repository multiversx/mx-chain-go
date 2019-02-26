package config

// CacheConfig will map the json cache configuration
type CacheConfig struct {
	Size uint32 `json:"size"`
	Type string `json:"type"`
}

// DBConfig will map the json db configuration
type DBConfig struct {
	FilePath string `json:"file"`
	Type     string `json:"type"`
}

// BloomFilterConfig will map the json bloom filter configuration
type BloomFilterConfig struct {
	Size     uint     `json:"size"`
	HashFunc []string `json:"hashFunc"`
}

// StorageConfig will map the json storage unit configuration
type StorageConfig struct {
	Cache CacheConfig       `json:"cache"`
	DB    DBConfig          `json:"db"`
	Bloom BloomFilterConfig `json:"bloom"`
}

// Config will hold the entire application configuration parameters
type Config struct {
	MiniBlocksStorage    StorageConfig `json:"miniBlocksStorage"`
	PeerBlockBodyStorage StorageConfig `json:"peerBlockBodyStorage"`
	BlockHeaderStorage   StorageConfig `json:"blockHeaderStorage"`
	TxStorage            StorageConfig `json:"txStorage"`

	AccountsTrieStorage StorageConfig `json:"accountsTrieStorage"`
	BadBlocksCache      CacheConfig   `json:"badBlocksCache"`

	TxBlockBodyDataPool       CacheConfig `json:"txBlockBodyDataPool"`
	StateBlockBodyDataPool    CacheConfig `json:"stateBlockBodyDataPool"`
	PeerBlockBodyDataPool     CacheConfig `json:"peerBlockBodyDataPool"`
	BlockHeaderDataPool       CacheConfig `json:"blockHeaderDataPool"`
	BlockHeaderNoncesDataPool CacheConfig `json:"blockHeaderNoncesDataPool"`
	TxDataPool                CacheConfig `json:"txDataPool"`

	Logger struct {
		Path            string `json:"path"`
		StackTraceDepth int    `json:"stackTraceDepth"`
	} `json:"logger"`
	Address struct {
		Length int    `json:"length"`
		Prefix string `json:"prefix"`
	} `json:"address"`
	Hasher struct {
		Type string `json:"type"`
	} `json:"hasher"`
	MultisigHasher struct {
		Type string `json:"type"`
	} `json:"multisigHasher"`
	Marshalizer struct {
		Type string `json:"type"`
	} `json:"marshalizer"`
}
