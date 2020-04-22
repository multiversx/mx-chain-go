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

// PubkeyConfig will map the json public key configuration
type PubkeyConfig struct {
	Length          int    `json:"length"`
	Type            string `json:"type"`
	SignatureLength int
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
	SyncPeriodSeconds   int
	Version             int
}

// EvictionWaitingListConfig will hold the configuration for the EvictionWaitingList
type EvictionWaitingListConfig struct {
	Size uint     `json:"size"`
	DB   DBConfig `json:"db"`
}

// EpochStartConfig will hold the configuration of EpochStart settings
type EpochStartConfig struct {
	MinRoundsBetweenEpochs      int64
	RoundsPerEpoch              int64
	ShuffledOutRestartThreshold float64
}

// BlockSizeThrottleConfig will hold the configuration for adaptive block size throttle
type BlockSizeThrottleConfig struct {
	MinSizeInBytes uint32
	MaxSizeInBytes uint32
}

// Config will hold the entire application configuration parameters
type Config struct {
	MiniBlocksStorage          StorageConfig
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

	AccountsTrieStorage      StorageConfig
	PeerAccountsTrieStorage  StorageConfig
	TrieSnapshotDB           DBConfig
	EvictionWaitingList      EvictionWaitingListConfig
	StateTriesConfig         StateTriesConfig
	TrieStorageManagerConfig TrieStorageManagerConfig
	BadBlocksCache           CacheConfig

	TxBlockBodyDataPool         CacheConfig
	PeerBlockBodyDataPool       CacheConfig
	TxDataPool                  CacheConfig
	UnsignedTransactionDataPool CacheConfig
	RewardTransactionDataPool   CacheConfig
	TrieNodesDataPool           CacheConfig
	WhiteListPool               CacheConfig
	EpochStartConfig            EpochStartConfig
	AddressPubkeyConverter      PubkeyConfig
	ValidatorPubkeyConverter    PubkeyConfig
	Hasher                      TypeConfig
	MultisigHasher              TypeConfig
	Marshalizer                 MarshalizerConfig
	VmMarshalizer               TypeConfig
	TxSignMarshalizer           TypeConfig

	PublicKeyShardId CacheConfig
	PublicKeyPeerId  CacheConfig
	PeerIdShardId    CacheConfig

	Antiflood       AntifloodConfig
	ResourceStats   ResourceStatsConfig
	Heartbeat       HeartbeatConfig
	GeneralSettings GeneralSettingsConfig
	Consensus       TypeConfig
	StoragePruning  StoragePruningConfig
	TxLogsStorage   StorageConfig

	NTPConfig               NTPConfig
	HeadersPoolConfig       HeadersPoolConfig
	BlockSizeThrottleConfig BlockSizeThrottleConfig
	VirtualMachineConfig    VirtualMachineConfig

	Hardfork HardforkConfig
}

// StoragePruningConfig will hold settings relates to storage pruning
type StoragePruningConfig struct {
	Enabled             bool
	FullArchive         bool
	NumEpochsToKeep     uint64
	NumActivePersisters uint64
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
	HbmiRefreshInterval                 uint32
	HeartbeatStorage                    StorageConfig
}

// GeneralSettingsConfig will hold the general settings for a node
type GeneralSettingsConfig struct {
	StatusPollingIntervalSec int
	MaxComputableRounds      uint64
	StartInEpochEnabled      bool
}

// FacadeConfig will hold different configuration option that will be passed to the main ElrondFacade
type FacadeConfig struct {
	RestApiInterface string
	PprofEnabled     bool
}

// StateTriesConfig will hold information about state tries
type StateTriesConfig struct {
	CheckpointRoundsModulus     uint
	AccountsStatePruningEnabled bool
	PeerStatePruningEnabled     bool
}

// TrieStorageManagerConfig will hold config information about trie storage manager
type TrieStorageManagerConfig struct {
	PruningBufferLen   uint32
	SnapshotsBufferLen uint32
	MaxSnapshots       uint8
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

// TopicMaxMessagesConfig will hold the maximum number of messages/sec per topic value
type TopicMaxMessagesConfig struct {
	Topic             string
	NumMessagesPerSec uint32
}

// TopicAntifloodConfig will hold the maximum values per second to be used in certain topics
type TopicAntifloodConfig struct {
	DefaultMaxMessagesPerSec uint32
	MaxMessages              []TopicMaxMessagesConfig
}

// TxAccumulatorConfig will hold the tx accumulator config values
type TxAccumulatorConfig struct {
	MaxAllowedTimeInMilliseconds   uint32
	MaxDeviationTimeInMilliseconds uint32
}

// AntifloodConfig will hold all p2p antiflood parameters
type AntifloodConfig struct {
	Enabled                   bool
	NumConcurrentResolverJobs int32
	NetworkMaxInput           AntifloodLimitsConfig
	PeerMaxInput              AntifloodLimitsConfig
	PeerMaxOutput             AntifloodLimitsConfig
	Cache                     CacheConfig
	BlackList                 BlackListConfig
	WebServer                 WebServerAntifloodConfig
	Topic                     TopicAntifloodConfig
	TxAccumulator             TxAccumulatorConfig
}

// AntifloodLimitsConfig will hold the maximum antiflood limits in both number of messages and total
// size of the messages
type AntifloodLimitsConfig struct {
	MessagesPerSecond  uint32
	TotalSizePerSecond uint64
}

// VirtualMachineConfig holds configuration for the Virtual Machine(s)
type VirtualMachineConfig struct {
	OutOfProcessEnabled bool
	OutOfProcessConfig  VirtualMachineOutOfProcessConfig
}

// VirtualMachineOutOfProcessConfig holds configuration for out-of-process virtual machine(s)
type VirtualMachineOutOfProcessConfig struct {
	LogsMarshalizer     string
	MessagesMarshalizer string
	MaxLoopTime         int
}

// HardforkConfig holds the configuration for the hardfork trigger
type HardforkConfig struct {
	EnableTrigger         bool
	EnableTriggerFromP2P  bool
	PublicKeyToListenFrom string
}
