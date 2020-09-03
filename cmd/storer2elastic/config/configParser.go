package config

// Config holds the toml configuration for the levelDB to elastic tool
type Config struct {
	General       GeneralConfig       `toml:"general"`
	ElasticSearch ElasticSearchConfig `toml:"elasticSearch"`
}

// GeneralConfig holds basic configuration
type GeneralConfig struct {
	DBPathWithChainID  string `toml:"dbPathWithChainID"`
	NodeConfigFilePath string `toml:"nodeConfigPath"`
	ChainID            string `toml:"chainID"`
}

// ElasticSearchConfig holds the elastic search configuration
type ElasticSearchConfig struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// DBConfig will map the db configuration
type DBConfig struct {
	FilePath          string `toml:"filePath"`
	Type              string `toml:"type"`
	BatchDelaySeconds int    `toml:"batchDelaySeconds"`
	MaxBatchSize      int    `toml:"maxBatchSize"`
	MaxOpenFiles      int    `toml:"maxOpenFiles"`
}
