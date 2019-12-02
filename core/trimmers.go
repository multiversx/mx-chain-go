package core

import "github.com/ElrondNetwork/elrond-go/core/constants"

// GetTrimmedPk returns a trimmed string to the pkPrefixSize value
func GetTrimmedPk(pk string) string {
	if len(pk) > constants.PkPrefixSize {
		pk = pk[:constants.PkPrefixSize]
	}

	return pk
}
