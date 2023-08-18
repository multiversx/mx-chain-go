package compatibility

import "sort"

// Sort sorts data.
// It makes one call to data.Len to determine n and O(n*log(n)) calls to
// data.Less and data.Swap. The sort is not guaranteed to be stable.
func Sort(data sort.Interface) {
	n := data.Len()
	quickSort_func(data, 0, n, maxDepth(n))
}
