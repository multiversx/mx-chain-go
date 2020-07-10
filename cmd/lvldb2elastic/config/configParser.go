package config

// Config holds the toml configuration for the levelDB to elastic tool
type Config struct {
	General       GeneralConfig       `toml:"general"`
	ElasticSearch ElasticSearchConfig `toml:"elasticSearch"`
}

// GeneralConfig holds basic configuration
type GeneralConfig struct {
	DBPath  string `toml:"dbPath"`
	Timeout int    `toml:"timeout"`
}

// ElasticSearchConfig holds the elastic search configuration
type ElasticSearchConfig struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}
