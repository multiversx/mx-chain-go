package core

// GetTrimmedPk returns a trimmed string to the pkPrefixSize value
func GetTrimmedPk(pk string) string {
	if len(pk) > pkPrefixSize {
		pk = pk[:pkPrefixSize]
	}

	return pk
}
