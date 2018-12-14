package config

type cacheConfig struct {
	Size uint32 `json:"size"`
	Type string `json:"type"`
}

type dbConfig struct {
	File string `json:"file"`
	Type string `json:"type"`
}

type bloomFilterConfig struct {
	Size     uint     `json:"size"`
	HashFunc []string `json:"hashFunc"`
}

type storageConfig struct {
	Cache cacheConfig       `json:"cache"`
	DB    dbConfig          `json:"db"`
	Bloom bloomFilterConfig `json:"bloom"`
}

type Config struct {
	TxBlockBodyStorage    storageConfig `json:"blockStorage"`
	StateBlockBodyStorage storageConfig `json:"stateBlockBodyStorage"`
	PeerBlockBodyStorage  storageConfig `json:"peerBlockBodyStorage"`
	BlockHeaderStorage    storageConfig `json:"blockHeaderStorage"`
	TxStorage             storageConfig `json:"txStorage"`
	BadBlocksCache        cacheConfig   `json:"badBlocksCache"`
	TxPoolStorage         cacheConfig   `json:"txPoolStorage"`
	BlockPoolStorage      cacheConfig   `json:"blockPoolStorage"`
	Logger                struct {
		Path            string `json:"path"`
		StackTraceDepth int    `json:"stackTraceDepth"`
	} `json:"logger"`
}
