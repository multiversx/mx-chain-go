package core

// GetTrimmedPk returns a trimmed string to the pkPrefixSize value
func GetTrimmedPk(pk string) string {
	if len(pk) > pkPrefixSize {
		pk = pk[:pkPrefixSize]
	}

	return pk
}

// TrimSoftwareVersion returns a trimmed byte array of the software version
func TrimSoftwareVersion(version string) string {
	softwareVersionLength := MinInt(MaxSoftwareVersionLengthInBytes, len(version))
	return version[:softwareVersionLength]
}
