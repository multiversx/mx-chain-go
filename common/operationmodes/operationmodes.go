package operationmodes

import "fmt"

const (
	OperationModeFullArchive        = "full-archive"
	OperationModeDbLookupExtension  = "db-lookup-extension"
	OperationModeHistoricalBalances = "historical-balances"
	OperationModeLiteObserver       = "lite-observer"
)

// CheckOperationModes will check the compatibility of the provided operation modes and return an error if any
func CheckOperationModes(modes []string) error {
	if len(modes) == 0 {
		return nil
	}

	// db lookup extension and historical balances
	isInvalid := sliceContainsBothElements(modes, OperationModeHistoricalBalances, OperationModeDbLookupExtension)
	if isInvalid {
		return fmt.Errorf("operation-mode flag cannot contain both db-lookup-extension and historical-balances")
	}

	// lite observer and historical balances
	isInvalid = sliceContainsBothElements(modes, OperationModeLiteObserver, OperationModeHistoricalBalances)
	if isInvalid {
		return fmt.Errorf("operation-mode flag cannot contain both lite-observer and historical-balances")
	}

	// lite observer and full archive
	isInvalid = sliceContainsBothElements(modes, OperationModeLiteObserver, OperationModeFullArchive)
	if isInvalid {
		return fmt.Errorf("operation-mode flag cannot contain both lite-observer and full-archive")
	}

	return nil
}

func sliceContainsBothElements(elements []string, first string, second string) bool {
	containsFirstElement := SliceContainsElement(elements, first)
	containsSecondElement := SliceContainsElement(elements, second)

	return containsFirstElement && containsSecondElement
}

// SliceContainsElement will return true if the provided slice contains the provided element
func SliceContainsElement(elements []string, element string) bool {
	for _, el := range elements {
		if el == element {
			return true
		}
	}

	return false
}
