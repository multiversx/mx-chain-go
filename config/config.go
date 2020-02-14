package config

// CacheConfig will map the json cache configuration
type CacheConfig struct {
	Type        string `json:"type"`
	Size        uint32 `json:"size"`
	SizeInBytes uint32 `json:"sizeInBytes"`
	Shards      uint32 `json:"shards"`
}

//HeadersPoolConfig will map the headers cache configuration
type HeadersPoolConfig struct {
	MaxHeadersPerShard            int
	NumElementsToRemoveOnEviction int
}

// DBConfig will map the json db configuration
type DBConfig struct {
	FilePath          string `json:"file"`
	Type              string `json:"type"`
	BatchDelaySeconds int    `json:"batchDelaySeconds"`
	MaxBatchSize      int    `json:"maxBatchSize"`
	MaxOpenFiles      int    `json:"maxOpenFiles"`
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

// AddressConfig will map the json address configuration
type AddressConfig struct {
	Length int    `json:"length"`
	Prefix string `json:"prefix"`
}

// TypeConfig will map the json string type configuration
type TypeConfig struct {
	Type string `json:"type"`
}

// MarshalizerConfig holds the marshalizer related configuration
type MarshalizerConfig struct {
	Type           string `json:"type"`
	SizeCheckDelta uint32 `json:"sizeCheckDelta"`
}

// NTPConfig will hold the configuration for NTP queries
type NTPConfig struct {
	Hosts               []string
	Port                int
	TimeoutMilliseconds int
	Version             int
}

// EvictionWaitingListConfig will hold the configuration for the EvictionWaitingList
type EvictionWaitingListConfig struct {
	Size uint     `json:"size"`
	DB   DBConfig `json:"db"`
}

// EpochStartConfig will hold the configuration of EpochStart settings
type EpochStartConfig struct {
	MinRoundsBetweenEpochs int64
	RoundsPerEpoch         int64
}

// Config will hold the entire application configuration parameters
type Config struct {
	MiniBlocksStorage          StorageConfig
	MiniBlockHeadersStorage    StorageConfig
	PeerBlockBodyStorage       StorageConfig
	BlockHeaderStorage         StorageConfig
	TxStorage                  StorageConfig
	UnsignedTransactionStorage StorageConfig
	RewardTxStorage            StorageConfig
	ShardHdrNonceHashStorage   StorageConfig
	MetaHdrNonceHashStorage    StorageConfig
	StatusMetricsStorage       StorageConfig

	BootstrapStorage StorageConfig
	MetaBlockStorage StorageConfig

	AccountsTrieStorage     StorageConfig
	PeerAccountsTrieStorage StorageConfig
	TrieSnapshotDB          DBConfig
	EvictionWaitingList     EvictionWaitingListConfig
	StateTrieConfig         StateTrieConfig
	BadBlocksCache          CacheConfig

	TxBlockBodyDataPool         CacheConfig
	PeerBlockBodyDataPool       CacheConfig
	TxDataPool                  CacheConfig
	UnsignedTransactionDataPool CacheConfig
	RewardTransactionDataPool   CacheConfig
	TrieNodesDataPool           CacheConfig
	EpochStartConfig            EpochStartConfig
	Address                     AddressConfig
	BLSPublicKey                AddressConfig
	Hasher                      TypeConfig
	MultisigHasher              TypeConfig
	Marshalizer                 MarshalizerConfig

	Antiflood       AntifloodConfig
	ResourceStats   ResourceStatsConfig
	Heartbeat       HeartbeatConfig
	GeneralSettings GeneralSettingsConfig
	Consensus       TypeConfig
	Explorer        ExplorerConfig
	StoragePruning  StoragePruningConfig

	NTPConfig         NTPConfig
	HeadersPoolConfig HeadersPoolConfig
}

// NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port            int
	Seed            string
	TargetPeerCount int
}

// StoragePruningConfig will hold settings relates to storage pruning
type StoragePruningConfig struct {
	Enabled             bool
	FullArchive         bool
	NumEpochsToKeep     uint64
	NumActivePersisters uint64
}

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig struct {
	Enabled                          bool
	RefreshIntervalInSec             uint32
	RandezVous                       string
	InitialPeerList                  []string
	BucketSize                       uint32
	RoutingTableRefreshIntervalInSec uint32
}

// P2PConfig will hold all the P2P settings
type P2PConfig struct {
	Node                NodeConfig
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
	HeartbeatStorage                    StorageConfig
}

// GeneralSettingsConfig will hold the general settings for a node
type GeneralSettingsConfig struct {
	DestinationShardAsObserver string
	StatusPollingIntervalSec   int
	MaxComputableRounds        uint64
}

// ExplorerConfig will hold the configuration for the explorer indexer
type ExplorerConfig struct {
	Enabled    bool
	IndexerURL string
}

// ServersConfig will hold all the confidential settings for servers
type ServersConfig struct {
	ElasticSearch ElasticSearchConfig
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Username string
	Password string
}

// FacadeConfig will hold different configuration option that will be passed to the main ElrondFacade
type FacadeConfig struct {
	RestApiInterface string
	PprofEnabled     bool
}

// StateTrieConfig will hold information about state trie
type StateTrieConfig struct {
	RoundsModulus  uint
	PruningEnabled bool
}

// WebServerAntifloodConfig will hold the anti-lflooding parameters for the web server
type WebServerAntifloodConfig struct {
	SimultaneousRequests         uint32
	SameSourceRequests           uint32
	SameSourceResetIntervalInSec uint32
}

// BlackListConfig will hold the p2p peer black list threshold values
type BlackListConfig struct {
	ThresholdNumMessagesPerSecond uint32
	ThresholdSizePerSecond        uint64
	NumFloodingRounds             uint32
	PeerBanDurationInSeconds      uint32
}

// TopicAntifloodConfig will hold the maximum values per second to be used in certain topics
type TopicAntifloodConfig struct {
	DefaultMaxMessagesPerSec   uint32
	HeartbeatMaxMessagesPerSec uint32
	HeadersRequestsPerSec      uint32
}

// AntifloodConfig will hold all p2p antiflood parameters
type AntifloodConfig struct {
	Enabled                   bool
	Cache                     CacheConfig
	BlackList                 BlackListConfig
	PeerMaxMessagesPerSecond  uint32
	PeerMaxTotalSizePerSecond uint64
	MaxMessagesPerSecond      uint32
	MaxTotalSizePerSecond     uint64
	WebServer                 WebServerAntifloodConfig
	Topic                     TopicAntifloodConfig
}
