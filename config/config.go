package config

// CacheConfig will map the json cache configuration
type CacheConfig struct {
	Size   uint32 `json:"size"`
	Type   string `json:"type"`
	Shards uint32 `json:"shards"`
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

// MarshalizerConfig
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

	ShardDataStorage StorageConfig
	BootstrapStorage StorageConfig
	MetaBlockStorage StorageConfig
	PeerDataStorage  StorageConfig

	AccountsTrieStorage     StorageConfig
	PeerAccountsTrieStorage StorageConfig
	BadBlocksCache          CacheConfig

	TxBlockBodyDataPool         CacheConfig
	StateBlockBodyDataPool      CacheConfig
	PeerBlockBodyDataPool       CacheConfig
	BlockHeaderDataPool         CacheConfig
	BlockHeaderNoncesDataPool   CacheConfig
	TxDataPool                  CacheConfig
	UnsignedTransactionDataPool CacheConfig
	RewardTransactionDataPool   CacheConfig
	MetaBlockBodyDataPool       CacheConfig

	MiniBlockHeaderHashesDataPool CacheConfig
	ShardHeadersDataPool          CacheConfig
	MetaHeaderNoncesDataPool      CacheConfig

	EpochStartConfig EpochStartConfig
	Logger           LoggerConfig
	Address          AddressConfig
	BLSPublicKey     AddressConfig
	Hasher           TypeConfig
	MultisigHasher   TypeConfig
	Marshalizer      MarshalizerConfig

	ResourceStats   ResourceStatsConfig
	Heartbeat       HeartbeatConfig
	GeneralSettings GeneralSettingsConfig
	Consensus       TypeConfig
	Explorer        ExplorerConfig

	NTPConfig NTPConfig
}

// NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port            int
	Seed            string
	TargetPeerCount int
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
	NetworkID                  string
	StatusPollingIntervalSec   int
}

// ExplorerConfig will hold the configuration for the explorer indexer
type ExplorerConfig struct {
	Enabled    bool
	IndexerURL string
}

// ServersConfig will hold all the confidential settings for servers
type ServersConfig struct {
	ElasticSearch ElasticSearchConfig
	Prometheus    PrometheusConfig
}

// PrometheusConfig will hold configuration for prometheus, such as the join URL
type PrometheusConfig struct {
	PrometheusBaseURL string
	JoinRoute         string
	StatusRoute       string
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Username string
	Password string
}

// FacadeConfig will hold different configuration option that will be passed to the main ElrondFacade
type FacadeConfig struct {
	RestApiInterface  string
	PprofEnabled      bool
	Prometheus        bool
	PrometheusJoinURL string
	PrometheusJobName string
}
