package check

// ForNil tests if the provided interface pointer or underlying object is nil
func ForNil(checker NilInterfaceChecker) bool {
	if checker == nil {
		return true
	}

	return checker.IsInterfaceNil()
}
