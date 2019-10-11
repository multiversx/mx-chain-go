package config

// ConfigPreferences will hold the configuration related to node's preferences
type ConfigPreferences struct {
	Preferences PreferencesConfig
}

// PreferencesConfig will hold the fields which are node specific such as the display name
type PreferencesConfig struct {
	NodeDisplayName string
}
