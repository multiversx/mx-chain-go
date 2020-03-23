package logger

import (
	"encoding/json"
	"fmt"
)

// Profile holds global logger options
type Profile struct {
	LogLevelPatterns string
	WithCorrelation  bool
	WithLoggerName   bool
}

// GetCurrentProfile gets the current logger profile
func GetCurrentProfile() Profile {
	return Profile{
		LogLevelPatterns: GetLogLevelPattern(),
		WithCorrelation:  IsEnabledCorrelation(),
		WithLoggerName:   IsEnabledLoggerName(),
	}
}

// UnmarshalProfile deserializes into a Profile object
func UnmarshalProfile(data []byte) (Profile, error) {
	profile := Profile{}
	err := json.Unmarshal(data, &profile)
	if err != nil {
		return Profile{}, err
	}

	return profile, nil
}

// Marshal serializes the Profile object
func (profile *Profile) Marshal() ([]byte, error) {
	data, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Apply sets the global logger options
func (profile *Profile) Apply() error {
	err := SetLogLevel(profile.LogLevelPatterns)
	if err != nil {
		return err
	}

	ToggleCorrelation(profile.WithCorrelation)
	ToggleLoggerName(profile.WithLoggerName)
	return nil
}

func (profile *Profile) String() string {
	return fmt.Sprintf("[pattern=%s, with correlation=%t, with logger name=%t]",
		profile.LogLevelPatterns,
		profile.WithCorrelation,
		profile.WithLoggerName,
	)
}
