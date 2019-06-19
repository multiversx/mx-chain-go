package bits_test

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/benchmark-broadcast/bits"
)

func TestXORBytes(t *testing.T) {

	var tests = []struct {
		a []byte
		b []byte
		c []byte
	}{
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, []byte{0x00, 0xff, 0x00, 0xff, 0x00}, []byte{0xff, 0x00, 0xff, 0x00, 0xff}},
		{[]byte{0x01, 0x01, 0x01, 0x01, 0x01}, []byte{0x00, 0x00, 0x00, 0x00, 0x00}, []byte{0x01, 0x01, 0x01, 0x01, 0x01}},
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, []byte{0xff, 0xff, 0xff, 0xff, 0xff}, []byte{0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, test := range tests {
		result, err := bits.XOR(test.a, test.b)

		if err != nil {
			t.Errorf("Error in calculating XOR")
		}

		if !reflect.DeepEqual(result, test.c) {
			t.Errorf("%v XOR %v expected %v actual %v \n", test.a, test.b, test.c, result)
		}

	}

}

func TestCountSetBits(t *testing.T) {
	var tests = []struct {
		nr        []byte
		nrOfBytes int
	}{
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, 40},
		{[]byte{0x00, 0xff, 0x00, 0xff, 0x00}, 16},
		{[]byte{0x0f, 0x0f, 0x0f, 0x0f, 0x0f}, 20},
	}

	for _, test := range tests {
		result := bits.CountSet(test.nr)

		if result != test.nrOfBytes {
			t.Errorf("Slice %v expected %v actual %v \n", test.nr, test.nrOfBytes, result)
		}
	}
}
