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

// LoggerConfig will map the json logger configuration
type LoggerConfig struct {
	Path            string `json:"path"`
	StackTraceDepth int    `json:"stackTraceDepth"`
}

// AddressConfig will map the json address configuration
type AddressConfig struct {
	Length int    `json:"length"`
	Prefix string `json:"prefix"`
}

// TypeConfig will map the json string type configuration
type TypeConfig struct {
	Type string `json:"type"`
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

	Logger         LoggerConfig  `json:"logger"`
	Address        AddressConfig `json:"address"`
	Hasher         TypeConfig    `json:"hasher"`
	MultisigHasher TypeConfig    `json:"multisigHasher"`
	Marshalizer    TypeConfig    `json:"marshalizer"`
}

//NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port int
	Seed string
}

//MdnsPeerDiscoveryConfig will hold the mdns discovery config settings
type MdnsPeerDiscoveryConfig struct {
	Enabled              bool
	RefreshIntervalInSec int
	ServiceTag           string
}

//KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig struct {
	Enabled              bool
	RefreshIntervalInSec int
	RandezVous           string
	InitialPeerList      []string
}

//P2PConfig will hold all the P2P settings
type P2PConfig struct {
	Node                NodeConfig
	MdnsPeerDiscovery   MdnsPeerDiscoveryConfig
	KadDhtPeerDiscovery KadDhtPeerDiscoveryConfig
}
