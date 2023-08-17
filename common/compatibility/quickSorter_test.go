package compatibility

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

var testInts = [...]int{74, 59, 238, -784, 9845, 959, 905, 0, 0, 42, 7586, -5467984, 7586}
var testFloat64s = [...]float64{74.3, 59.0, math.Inf(1), 238.2, -784.0, 2.3, math.NaN(), math.NaN(), math.Inf(-1), 9845.768, -959.7485, 905, 7.8, 7.8}
var testStrings = [...]string{"", "Hello", "foo", "bar", "foo", "f00", "%*&^*&^&", "***"}

func TestSortIntSlice(t *testing.T) {
	data := testInts
	a := sort.IntSlice(data[0:])
	Sort(a)
	if !sort.IsSorted(a) {
		t.Errorf("sorted %v", testInts)
		t.Errorf("   got %v", data)
	}
}

func TestSortFloat64Slice(t *testing.T) {
	data := testFloat64s
	a := sort.Float64Slice(data[0:])
	Sort(a)
	if !sort.IsSorted(a) {
		t.Errorf("sorted %v", testFloat64s)
		t.Errorf("   got %v", data)
	}
}

func TestSortStringSlice(t *testing.T) {
	data := testStrings
	a := sort.StringSlice(data[0:])
	Sort(a)
	if !sort.IsSorted(a) {
		t.Errorf("sorted %v", testStrings)
		t.Errorf("   got %v", data)
	}
}

func TestInts(t *testing.T) {
	data := testInts
	sort.Ints(data[0:])
	if !sort.IntsAreSorted(data[0:]) {
		t.Errorf("sorted %v", testInts)
		t.Errorf("   got %v", data)
	}
}

func TestFloat64s(t *testing.T) {
	data := testFloat64s
	sort.Float64s(data[0:])
	if !sort.Float64sAreSorted(data[0:]) {
		t.Errorf("sorted %v", testFloat64s)
		t.Errorf("   got %v", data)
	}
}

func TestStrings(t *testing.T) {
	data := testStrings
	sort.Strings(data[0:])
	if !sort.StringsAreSorted(data[0:]) {
		t.Errorf("sorted %v", testStrings)
		t.Errorf("   got %v", data)
	}
}

func TestSlice(t *testing.T) {
	data := testStrings
	sort.Slice(data[:], func(i, j int) bool {
		return data[i] < data[j]
	})
	if !sort.SliceIsSorted(data[:], func(i, j int) bool { return data[i] < data[j] }) {
		t.Errorf("sorted %v", testStrings)
		t.Errorf("   got %v", data)
	}
}

func TestSortLarge_Random(t *testing.T) {
	n := 1000000
	if testing.Short() {
		n /= 100
	}
	data := make([]int, n)
	for i := 0; i < len(data); i++ {
		data[i] = rand.Intn(100)
	}
	if sort.IntsAreSorted(data) {
		t.Fatalf("terrible rand.rand")
	}
	sort.Ints(data)
	if !sort.IntsAreSorted(data) {
		t.Errorf("sort didn't sort - 1M ints")
	}
}

func TestReverseSortIntSlice(t *testing.T) {
	data := testInts
	data1 := testInts
	a := sort.IntSlice(data[0:])
	Sort(a)
	r := sort.IntSlice(data1[0:])
	Sort(sort.Reverse(r))
	for i := 0; i < len(data); i++ {
		if a[i] != r[len(data)-1-i] {
			t.Errorf("reverse sort didn't sort")
		}
		if i > len(data)/2 {
			break
		}
	}
}

type nonDeterministicTestingData struct {
	r *rand.Rand
}

func (t *nonDeterministicTestingData) Len() int {
	return 500
}
func (t *nonDeterministicTestingData) Less(i, j int) bool {
	if i < 0 || j < 0 || i >= t.Len() || j >= t.Len() {
		panic("nondeterministic comparison out of bounds")
	}
	return t.r.Float32() < 0.5
}
func (t *nonDeterministicTestingData) Swap(i, j int) {
	if i < 0 || j < 0 || i >= t.Len() || j >= t.Len() {
		panic("nondeterministic comparison out of bounds")
	}
}

func TestNonDeterministicComparison(t *testing.T) {
	// Ensure that sort.Sort does not panic when Less returns inconsistent results.
	// See https://golang.org/issue/14377.
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	td := &nonDeterministicTestingData{
		r: rand.New(rand.NewSource(0)),
	}

	for i := 0; i < 10; i++ {
		Sort(td)
	}
}

const (
	_Sawtooth = iota
	_Rand
	_Stagger
	_Plateau
	_Shuffle
	_NDist
)

const (
	_Copy = iota
	_Reverse
	_ReverseFirstHalf
	_ReverseSecondHalf
	_Sorted
	_Dither
	_NMode
)

type testingData struct {
	desc        string
	t           *testing.T
	data        []int
	maxswap     int // number of swaps allowed
	ncmp, nswap int
}

func (d *testingData) Len() int { return len(d.data) }
func (d *testingData) Less(i, j int) bool {
	d.ncmp++
	return d.data[i] < d.data[j]
}
func (d *testingData) Swap(i, j int) {
	if d.nswap >= d.maxswap {
		d.t.Fatalf("%s: used %d swaps sorting slice of %d", d.desc, d.nswap, len(d.data))
	}
	d.nswap++
	d.data[i], d.data[j] = d.data[j], d.data[i]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func lg(n int) int {
	i := 0
	for 1<<uint(i) < n {
		i++
	}
	return i
}

func testBentleyMcIlroy(t *testing.T, sortFunc func(sort.Interface), maxswap func(int) int) {
	sizes := []int{100, 1023, 1024, 1025}
	if testing.Short() {
		sizes = []int{100, 127, 128, 129}
	}
	dists := []string{"sawtooth", "rand", "stagger", "plateau", "shuffle"}
	modes := []string{"copy", "reverse", "reverse1", "reverse2", "sort", "dither"}
	var tmp1, tmp2 [1025]int
	for _, n := range sizes {
		for m := 1; m < 2*n; m *= 2 {
			for dist := 0; dist < _NDist; dist++ {
				j := 0
				k := 1
				data := tmp1[0:n]
				for i := 0; i < n; i++ {
					switch dist {
					case _Sawtooth:
						data[i] = i % m
					case _Rand:
						data[i] = rand.Intn(m)
					case _Stagger:
						data[i] = (i*m + i) % n
					case _Plateau:
						data[i] = min(i, m)
					case _Shuffle:
						if rand.Intn(m) != 0 {
							j += 2
							data[i] = j
						} else {
							k += 2
							data[i] = k
						}
					}
				}

				mdata := tmp2[0:n]
				for mode := 0; mode < _NMode; mode++ {
					switch mode {
					case _Copy:
						for i := 0; i < n; i++ {
							mdata[i] = data[i]
						}
					case _Reverse:
						for i := 0; i < n; i++ {
							mdata[i] = data[n-i-1]
						}
					case _ReverseFirstHalf:
						for i := 0; i < n/2; i++ {
							mdata[i] = data[n/2-i-1]
						}
						for i := n / 2; i < n; i++ {
							mdata[i] = data[i]
						}
					case _ReverseSecondHalf:
						for i := 0; i < n/2; i++ {
							mdata[i] = data[i]
						}
						for i := n / 2; i < n; i++ {
							mdata[i] = data[n-(i-n/2)-1]
						}
					case _Sorted:
						for i := 0; i < n; i++ {
							mdata[i] = data[i]
						}
						// Ints is known to be correct
						// because mode Sort runs after mode _Copy.
						sort.Ints(mdata)
					case _Dither:
						for i := 0; i < n; i++ {
							mdata[i] = data[i] + i%5
						}
					}

					desc := fmt.Sprintf("n=%d m=%d dist=%s mode=%s", n, m, dists[dist], modes[mode])
					d := &testingData{desc: desc, t: t, data: mdata[0:n], maxswap: maxswap(n)}
					sortFunc(d)
					// Uncomment if you are trying to improve the number of compares/swaps.
					//t.Logf("%s: ncmp=%d, nswp=%d", desc, d.ncmp, d.nswap)

					// If we were testing C qsort, we'd have to make a copy
					// of the slice and sort it ourselves and then compare
					// x against it, to ensure that qsort was only permuting
					// the data, not (for example) overwriting it with zeros.
					//
					// In go, we don't have to be so paranoid: since the only
					// mutating method Sort can call is TestingData.swap,
					// it suffices here just to check that the final slice is sorted.
					if !sort.IntsAreSorted(mdata) {
						t.Fatalf("%s: ints not sorted\n\t%v", desc, mdata)
					}
				}
			}
		}
	}
}

func TestSortBM(t *testing.T) {
	testBentleyMcIlroy(t, Sort, func(n int) int { return n * lg(n) * 12 / 10 })
}

// This is based on the "antiquicksort" implementation by M. Douglas McIlroy.
// See https://www.cs.dartmouth.edu/~doug/mdmspe.pdf for more info.
type adversaryTestingData struct {
	t         *testing.T
	data      []int // item values, initialized to special gas value and changed by Less
	maxcmp    int   // number of comparisons allowed
	ncmp      int   // number of comparisons (calls to Less)
	nsolid    int   // number of elements that have been set to non-gas values
	candidate int   // guess at current pivot
	gas       int   // special value for unset elements, higher than everything else
}

func (d *adversaryTestingData) Len() int { return len(d.data) }

func (d *adversaryTestingData) Less(i, j int) bool {
	if d.ncmp >= d.maxcmp {
		d.t.Fatalf("used %d comparisons sorting adversary data with size %d", d.ncmp, len(d.data))
	}
	d.ncmp++

	if d.data[i] == d.gas && d.data[j] == d.gas {
		if i == d.candidate {
			// freeze i
			d.data[i] = d.nsolid
			d.nsolid++
		} else {
			// freeze j
			d.data[j] = d.nsolid
			d.nsolid++
		}
	}

	if d.data[i] == d.gas {
		d.candidate = i
	} else if d.data[j] == d.gas {
		d.candidate = j
	}

	return d.data[i] < d.data[j]
}

func (d *adversaryTestingData) Swap(i, j int) {
	d.data[i], d.data[j] = d.data[j], d.data[i]
}

func newAdversaryTestingData(t *testing.T, size int, maxcmp int) *adversaryTestingData {
	gas := size - 1
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = gas
	}
	return &adversaryTestingData{t: t, data: data, maxcmp: maxcmp, gas: gas}
}

func TestAdversary(t *testing.T) {
	const size = 10000            // large enough to distinguish between O(n^2) and O(n*log(n))
	maxcmp := size * lg(size) * 4 // the factor 4 was found by trial and error
	d := newAdversaryTestingData(t, size, maxcmp)
	Sort(d) // This should degenerate to heapsort.
	// Check data is fully populated and sorted.
	for i, v := range d.data {
		if v != i {
			t.Fatalf("adversary data not fully sorted")
		}
	}
}

var countOpsSizes = []int{1e2, 3e2, 1e3, 3e3, 1e4, 3e4, 1e5, 3e5, 1e6}

func countOps(t *testing.T, algo func(sort.Interface), name string) {
	sizes := countOpsSizes
	if testing.Short() {
		sizes = sizes[:5]
	}
	if !testing.Verbose() {
		t.Skip("Counting skipped as non-verbose mode.")
	}
	for _, n := range sizes {
		td := testingData{
			desc:    name,
			t:       t,
			data:    make([]int, n),
			maxswap: 1<<31 - 1,
		}
		for i := 0; i < n; i++ {
			td.data[i] = rand.Intn(n / 5)
		}
		algo(&td)
		t.Logf("%s %8d elements: %11d Swap, %10d Less", name, n, td.nswap, td.ncmp)
	}
}

func TestCountSortOps(t *testing.T) { countOps(t, Sort, "Sort  ") }
