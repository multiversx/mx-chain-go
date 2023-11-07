package config

import "github.com/multiversx/mx-chain-go/config"

// SovereignConfig holds sovereign node config
type SovereignConfig struct {
	*config.Configs
	SovereignExtraConfig *config.SovereignConfig
	NotifierConfig       *NotifierConfig
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
