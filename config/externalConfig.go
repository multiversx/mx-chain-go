package config

// ExternalConfig will hold the configurations for external tools, such as Explorer or Elasticsearch
type ExternalConfig struct {
	ElasticSearchConnector ElasticSearchConfig
	EventNotifierConnector EventNotifierConfig
	HostDriversConfig      []HostDriversConfig
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Enabled                   bool
	IndexerCacheSize          int
	BulkRequestMaxSizeInBytes int
	URL                       string
	UseKibana                 bool
	Username                  string
	Password                  string
	EnabledIndexes            []string
}

// EventNotifierConfig will hold the configuration for the events notifier driver
type EventNotifierConfig struct {
	Enabled           bool
	UseAuthorization  bool
	ProxyUrl          string
	Username          string
	Password          string
	RequestTimeoutSec int
	MarshallerType    string
}

// CovalentConfig will hold the configurations for covalent indexer
type CovalentConfig struct {
	Enabled              bool
	URL                  string
	RouteSendData        string
	RouteAcknowledgeData string
}

// HostDriversConfig will hold the configuration for WebSocket driver
type HostDriversConfig struct {
	Enabled                    bool
	WithAcknowledge            bool
	BlockingAckOnError         bool
	DropMessagesIfNoConnection bool
	URL                        string
	MarshallerType             string
	Mode                       string
	RetryDurationInSec         int
	AcknowledgeTimeoutInSec    int
	Version                    uint32
}
