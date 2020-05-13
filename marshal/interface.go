package marshal

// Sizer contains method Size that is needed to get the size of a proto obj
type Sizer interface {
	Size() (n int)
}
