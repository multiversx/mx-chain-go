package config

import "github.com/multiversx/mx-chain-go/config"

// SovereignConfig holds sovereign node config
type SovereignConfig struct {
	*config.Configs
	NotifierConfig *NotifierConfig
}

// NotifierConfig holds sovereign notifier configuration
type NotifierConfig struct {
	NumOfMainShards     uint32          `toml:"NumMainShards"`
	SubscribedAddresses []string        `toml:"SubscribedAddresses"`
	WebSocketConfig     WebSocketConfig `toml:"WebSocket"`
}

// WebSocketConfig holds web socket config
type WebSocketConfig struct {
	Url                string `toml:"Url"`
	MarshallerType     string `toml:"MarshallerType"`
	RetryDuration      uint32 `toml:"RetryDuration"`
	BlockingAckOnError bool   `toml:"BlockingAckOnError"`
	HasherType         string `toml:"HasherType"`
}
