package copy_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/benchmark-broadcast/copy"
	"github.com/satori/go.uuid"
)

func TestPeersToPeerMap(t *testing.T) {
	var dest []uuid.UUID
	var src = []uuid.UUID{
		uuid.Must(uuid.NewV4(), nil),
		uuid.Must(uuid.NewV4(), nil),
		uuid.Must(uuid.NewV4(), nil)}

	copy.PeersToPeerMap(&dest, src)

	if len(src) != len(dest) {
		t.Errorf("Src and dest have different lengths")
	}

	for i := range src {
		if len(src) != len(dest) {
			t.Fatalf("Src and dest have different lengths")

		}
		if src[i] != dest[i] {
			t.Errorf("Value from source map was not found in destination map")
		}

	}
}

func TestPeerPath(t *testing.T) {
	var dest []int
	var src = []int{2, 3, 4, 1, 5}

	copy.PeerPath(&dest, src, 6)

	if len(src)+1 != len(dest) {
		t.Fatalf("Src and dest have different lengths")

	}

	for i := range src {
		if src[i] != dest[i] {
			t.Errorf("Src and dest have different values")
		}
	}

}
