package main

import "fmt"

const (
	operationModeImportDb           = "import-db"
	operationModeFullArchive        = "full-archive"
	operationModeDbLookupExtension  = "db-lookup-extension"
	operationModeHistoricalBalances = "historical-balances"
)

func checkOperationModes(modes []string) error {
	// db lookup extension and historical balances
	isInvalid := sliceContainsBothElements(modes, operationModeHistoricalBalances, operationModeDbLookupExtension)
	if isInvalid {
		return fmt.Errorf("operation-mode flag cannot contain both db-lookup-extension and historical-balances")
	}

	return nil
}

func sliceContainsBothElements(elements []string, first string, second string) bool {
	firstFound := false
	secondFound := false
	for _, element := range elements {
		if element == first {
			firstFound = true
		}
		if element == second {
			secondFound = true
		}

		if firstFound && secondFound {
			return true
		}
	}

	return false
}

func sliceContainsElement(elements []string, element string) bool {
	for _, el := range elements {
		if el == element {
			return true
		}
	}

	return false
}
