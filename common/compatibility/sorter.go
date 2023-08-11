package compatibility

type lessSwap struct {
	less func(a, b int) bool
	swap func(a, b int)
}

// Less calls the inner less handler
func (ls *lessSwap) Less(a, b int) bool {
	return ls.less(a, b)
}

// Swap calls the inner swap handler
func (ls *lessSwap) Swap(a, b int) {
	ls.swap(a, b)
}

// SortSlice prepares and calls the quick sorter algorithm
func SortSlice(swap func(a, b int), less func(a, b int) bool, length int) {
	wrapper := &lessSwap{
		less: less,
		swap: swap,
	}

	quickSort_func(wrapper, 0, length, maxDepth(length))
}
