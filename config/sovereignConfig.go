package config

// SovereignConfig holds sovereign config
type SovereignConfig struct {
	ExtendedShardHdrNonceHashStorage StorageConfig
	ExtendedShardHeaderStorage       StorageConfig
	MainChainNotarization            MainChainNotarization    `toml:"MainChainNotarization"`
	OutgoingSubscribedEvents         OutgoingSubscribedEvents `toml:"OutgoingSubscribedEvents"`
	OutGoingBridge                   OutGoingBridge           `toml:"OutGoingBridge"`
	NotifierConfig                   NotifierConfig           `toml:"NotifierConfig"`
	GenesisConfig                    GenesisConfig            `toml:"GenesisConfig"`
	OutGoingBridgeCertificate        OutGoingBridgeCertificate
}

// OutgoingSubscribedEvents holds config for outgoing subscribed events
type OutgoingSubscribedEvents struct {
	TimeToWaitForUnconfirmedOutGoingOperationInSeconds uint32            `toml:"TimeToWaitForUnconfirmedOutGoingOperationInSeconds"`
	SubscribedEvents                                   []SubscribedEvent `toml:"SubscribedEvents"`
}

// MainChainNotarization defines necessary data to start main chain notarization on a sovereign shard
type MainChainNotarization struct {
	MainChainNotarizationStartRound uint64 `toml:"MainChainNotarizationStartRound"`
}

// OutGoingBridge holds config for grpc client to send outgoing bridge txs
type OutGoingBridge struct {
	GRPCHost string `toml:"GRPCHost"`
	GRPCPort string `toml:"GRPCPort"`
}

// OutGoingBridgeCertificate holds config for outgoing bridge certificate paths
type OutGoingBridgeCertificate struct {
	CertificatePath   string
	CertificatePkPath string
}

// NotifierConfig holds sovereign notifier configuration
type NotifierConfig struct {
	SubscribedEvents []SubscribedEvent `toml:"SubscribedEvents"`
	WebSocketConfig  WebSocketConfig   `toml:"WebSocket"`
}

// SubscribedEvent holds subscribed events config
type SubscribedEvent struct {
	Identifier string   `toml:"Identifier"`
	Addresses  []string `toml:"Addresses"`
}

// WebSocketConfig holds web socket config
type WebSocketConfig struct {
	Url                string `toml:"Url"`
	MarshallerType     string `toml:"MarshallerType"`
	RetryDuration      uint32 `toml:"RetryDuration"`
	BlockingAckOnError bool   `toml:"BlockingAckOnError"`
	HasherType         string `toml:"HasherType"`
	Mode               string `toml:"Mode"`
	WithAcknowledge    bool   `toml:"WithAcknowledge"`
	AcknowledgeTimeout int    `toml:"AcknowledgeTimeout"`
	Version            uint32 `toml:"Version"`
}

// GenesisConfig should hold all sovereign genesis related configs
type GenesisConfig struct {
	NativeESDT string `toml:"NativeESDT"`
}
