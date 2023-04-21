package config

import "github.com/multiversx/mx-chain-go/config"

type SovereignConfig struct {
	*config.Configs
	NotifierConfig *NotifierConfig
}

// NotifierConfig holds sovereign notifier configuration
type NotifierConfig struct {
	NumOfMainShards     uint32          `toml:"NumOfMainShards"`
	SubscribedAddresses []string        `toml:"SubscribedAddresses"`
	WebSocketConfig     WebSocketConfig `toml:"WebSocketConfig"`
}

// WebSocketConfig holds web sockets config
type WebSocketConfig struct {
	Url                string `toml:"Url"`
	MarshallerType     string `toml:"MarshallerType"`
	RetryDuration      uint32 `toml:"RetryDuration"`
	BlockingAckOnError bool   `toml:"BlockingAckOnError"`
}
