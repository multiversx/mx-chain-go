package config

// Preferences will hold the configuration related to node's preferences
type Preferences struct {
	Preferences   PreferencesConfig
	NamedIdentity []NamedIdentity
}

// PreferencesConfig will hold the fields which are node specific such as the display name
type PreferencesConfig struct {
	DestinationShardAsObserver string
	NodeDisplayName            string
	Identity                   string
	RedundancyLevel            int64
	PreferredConnections       []string
	ConnectionWatcherType      string
	FullArchive                bool
}

// NamedIdentity will hold the fields which are node named identities
type NamedIdentity struct {
	Identity string
	NodeName string
	BLSKeys  []string
}
