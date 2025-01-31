package disabled

type topicsChecker struct {
}

// NewDisabledTopicsChecker -
func NewDisabledTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

// CheckValidity -
func (tc *topicsChecker) CheckValidity(_ [][]byte) error {
	return nil
}

// IsInterfaceNil - returns true if there is no value under the interface
func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
