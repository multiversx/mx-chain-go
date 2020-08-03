package config

// ExternalConfig will hold the configurations for external tools, such as Explorer or Elastic Search
type ExternalConfig struct {
	ElasticSearchConnector ElasticSearchConfig
	Outport                OutportConfig
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Enabled  bool
	URL      string
	Username string
	Password string
}

// OutportConfig holds configuration for the outport
type OutportConfig struct {
	Enabled             bool
	MessagesMarshalizer string
}
