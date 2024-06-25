package operationmodes

import (
	"fmt"
	"strings"
)

// constants that define the operation mode of the node
const (
	OperationModeFullArchive          = "full-archive"
	OperationModeDbLookupExtension    = "db-lookup-extension"
	OperationModeHistoricalBalances   = "historical-balances"
	OperationModeSnapshotlessObserver = "snapshotless-observer"
)

// ParseOperationModes will check and parse the operation modes
func ParseOperationModes(operationModeList string) ([]string, error) {
	if len(operationModeList) == 0 {
		return []string{}, nil
	}

	modes := strings.Split(operationModeList, ",")
	for _, mode := range modes {
		err := checkOperationModeValidity(mode)
		if err != nil {
			return []string{}, err
		}
	}

	// db lookup extension and historical balances
	isInvalid := sliceContainsBothElements(modes, OperationModeHistoricalBalances, OperationModeDbLookupExtension)
	if isInvalid {
		return []string{}, fmt.Errorf("operation-mode flag cannot contain both db-lookup-extension and historical-balances")
	}

	// snapshotless observer and historical balances
	isInvalid = sliceContainsBothElements(modes, OperationModeSnapshotlessObserver, OperationModeHistoricalBalances)
	if isInvalid {
		return []string{}, fmt.Errorf("operation-mode flag cannot contain both snapshotless-observer and historical-balances")
	}

	// snapshotless observer and full archive
	isInvalid = sliceContainsBothElements(modes, OperationModeSnapshotlessObserver, OperationModeFullArchive)
	if isInvalid {
		return []string{}, fmt.Errorf("operation-mode flag cannot contain both snapshotless-observer and full-archive")
	}

	return modes, nil
}

func checkOperationModeValidity(mode string) error {
	switch mode {
	case OperationModeFullArchive, OperationModeDbLookupExtension, OperationModeHistoricalBalances, OperationModeSnapshotlessObserver:
		return nil
	default:
		return fmt.Errorf("invalid operation mode <%s>", mode)
	}
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
