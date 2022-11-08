package config

// Preferences will hold the configuration related to node's preferences
type Preferences struct {
	Preferences PreferencesConfig
}

// PreferencesConfig will hold the fields which are node specific such as the display name
type PreferencesConfig struct {
	DestinationShardAsObserver  string
	NodeDisplayName             string
	Identity                    string
	RedundancyLevel             int64
	PreferredConnections        []string
	ConnectionWatcherType       string
	OverridableConfigTomlValues []OverridableConfig
	FullArchive                 bool
}

// OverridableConfig holds the path and the new value to be updated in the configuration
type OverridableConfig struct {
	Path  string
	Value string
}
