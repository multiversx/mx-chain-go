package check

// IfNil tests if the provided interface pointer or underlying object is nil
func IfNil(checker NilInterfaceChecker) bool {
	if checker == nil {
		return true
	}
	return checker.IsInterfaceNil()
}
