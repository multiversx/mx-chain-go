package ntp

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
)

// NewNTPGoogleConfig creates an NTPConfig object that configures NTP to use a predefined list of hosts. This is useful
// for tests, for example, to avoid loading a configuration file just to have a NTPConfig
func NewNTPGoogleConfig() config.NTPConfig {
	return config.NTPConfig{
		Hosts:               []string{"time.google.com", "time.cloudflare.com", "time.apple.com", "time.windows.com"},
		Port:                123,
		Version:             0,
		TimeoutMilliseconds: 100,
		SyncPeriodSeconds:   3600,
	}
}

// NTPOptions defines configuration options for a NTP query
type NTPOptions struct {
	Hosts        []string
	Version      int
	LocalAddress string
	Timeout      time.Duration
	Port         int
}

// NewNTPOptions creates a new NTPOptions object
func NewNTPOptions(ntpConfig config.NTPConfig) NTPOptions {
	ntpConfig.TimeoutMilliseconds = core.MaxInt(minTimeout, ntpConfig.TimeoutMilliseconds)
	timeout := time.Duration(ntpConfig.TimeoutMilliseconds) * time.Millisecond

	return NTPOptions{
		Hosts:        ntpConfig.Hosts,
		Port:         ntpConfig.Port,
		Version:      ntpConfig.Version,
		LocalAddress: "",
		Timeout:      timeout,
	}
}
