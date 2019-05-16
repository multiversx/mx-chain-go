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
	MiniBlocksStorage    StorageConfig
	PeerBlockBodyStorage StorageConfig
	BlockHeaderStorage   StorageConfig
	TxStorage            StorageConfig

	ShardDataStorage StorageConfig
	MetaBlockStorage StorageConfig
	PeerDataStorage  StorageConfig

	AccountsTrieStorage StorageConfig
	BadBlocksCache      CacheConfig

	TxBlockBodyDataPool       CacheConfig
	StateBlockBodyDataPool    CacheConfig
	PeerBlockBodyDataPool     CacheConfig
	BlockHeaderDataPool       CacheConfig
	BlockHeaderNoncesDataPool CacheConfig
	TxDataPool                CacheConfig
	MetaBlockBodyDataPool     CacheConfig

	MiniBlockHeaderHashesDataPool CacheConfig
	ShardHeadersDataPool          CacheConfig
	MetaHeaderNoncesDataPool      CacheConfig

	Logger         LoggerConfig
	Address        AddressConfig
	Hasher         TypeConfig
	MultisigHasher TypeConfig
	Marshalizer    TypeConfig

	ResourceStats   ResourceStatsConfig
	Heartbeat       HeartbeatConfig
	GeneralSettings GeneralSettingsConfig
	Consensus       TypeConfig
}

// NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port int
	Seed string
}

// MdnsPeerDiscoveryConfig will hold the mdns discovery config settings
type MdnsPeerDiscoveryConfig struct {
	Enabled              bool
	RefreshIntervalInSec int
	ServiceTag           string
}

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig struct {
	Enabled              bool
	RefreshIntervalInSec int
	RandezVous           string
	InitialPeerList      []string
}

// P2PConfig will hold all the P2P settings
type P2PConfig struct {
	Node                NodeConfig
	MdnsPeerDiscovery   MdnsPeerDiscoveryConfig
	KadDhtPeerDiscovery KadDhtPeerDiscoveryConfig
}

// ResourceStatsConfig will hold all resource stats settings
type ResourceStatsConfig struct {
	Enabled              bool
	RefreshIntervalInSec int
}

// HeartbeatConfig will hold all heartbeat settings
type HeartbeatConfig struct {
	Enabled                             bool
	MinTimeToWaitBetweenBroadcastsInSec int
	MaxTimeToWaitBetweenBroadcastsInSec int
	DurationInSecToConsiderUnresponsive int
}

// GeneralSettingsConfig will hold the general settings for a node
type GeneralSettingsConfig struct {
	DestinationShardAsObserver string
}
