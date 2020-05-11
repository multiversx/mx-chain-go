package marshal

// Sizer contains method size that is needed get size of a proto obj
type Sizer interface {
	Size() (n int)
}
