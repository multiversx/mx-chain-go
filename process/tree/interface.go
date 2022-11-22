package tree

// IntervalTree defines a tree able to handle intervals
type IntervalTree interface {
	Contains(value uint64) bool
	IsInterfaceNil() bool
}
