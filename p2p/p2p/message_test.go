package p2p

import (
	"fmt"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	m1 := NewMessage("p1", "ABCDEF")
	fmt.Println("Original:")
	fmt.Println(m1)

	str := m1.ToJson()
	fmt.Println("Marshaled:")
	fmt.Println(str)

	m2 := FromJson(str)
	fmt.Println("Unmarshaled:")
	fmt.Println(*m2)

	if (m1.Payload != m2.Payload) || (m1.Hops != m2.Hops) {
		t.Fatal("Error un-marshaling!")
	}
}

func TestAddHop(t *testing.T) {
	m1 := NewMessage("p1", "ABCDEF")

	if (len(m1.Peers) != 1) || (m1.Hops != 0) {
		t.Fatal("Should have been 1 peer and 0 hops")
	}

	m1.AddHop("p2")

	if (len(m1.Peers) != 2) || (m1.Hops != 1) {
		t.Fatal("Should have been 2 peers and 1 hop")
	}

}
