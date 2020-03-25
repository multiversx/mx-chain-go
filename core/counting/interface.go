package counting

// Countable is an interface to get counts
type Countable interface {
	GetCounts() Counts
	IsInterfaceNil() bool
}

// Counts is an interface to interact with counts
type Counts interface {
	GetTotal() int64
	String() string
	IsInterfaceNil() bool
}
