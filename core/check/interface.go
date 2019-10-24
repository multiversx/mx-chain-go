package check

// NilInterfaceChecker checks if an interface's underlying object is nil
type NilInterfaceChecker interface {
	IsInterfaceNil() bool
}
