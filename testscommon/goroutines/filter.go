package goroutines

import "strings"

var ignorable = []string{
	"github.com/libp2p/go-libp2p-asn-util",
	"go.opencensus.io",
	"github.com/libp2p/go-nat",
}

// AllPassFilter returns true for all provided strings
func AllPassFilter(_ string) bool {
	return true
}

// TestsRelevantGoRoutines returns false for the goroutines that contains ignorable strings
func TestsRelevantGoRoutines(goRoutineData string) bool {
	for _, str := range ignorable {
		if strings.Contains(goRoutineData, str) {
			return false
		}
	}

	return true
}
