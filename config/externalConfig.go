package config

// ExternalConfig will hold the configurations for external tools, such as Explorer or Elastic Search
type ExternalConfig struct {
	Explorer               ExplorerConfig
	ElasticSearchConnector ElasticSearchConfig
}

// ExplorerConfig will hold the configuration for the explorer indexer
type ExplorerConfig struct {
	Enabled    bool
	IndexerURL string
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Username string
	Password string
}
