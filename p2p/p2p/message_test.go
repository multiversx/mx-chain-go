package p2p

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	mrsh := service.GetMarshalizerService()

	m1 := NewMessage("p1", []byte("ABCDEF"), mrsh)
	fmt.Println("Original:")
	fmt.Println(m1)

	buff, err := m1.ToByteArray()
	assert.Nil(t, err)
	fmt.Println("Marshaled:")
	fmt.Println(string(buff))

	m2, err := CreateFromByteArray(mrsh, buff)
	assert.Nil(t, err)
	fmt.Println("Unmarshaled:")
	fmt.Println(*m2)

	assert.Equal(t, m1, m2)
}

func TestAddHop(t *testing.T) {
	m1 := NewMessage("p1", []byte("ABCDEF"), service.GetMarshalizerService())

	if (len(m1.Peers) != 1) || (m1.Hops != 0) {
		assert.Fail(t, "Should have been 1 peer and 0 hops")
	}

	m1.AddHop("p2")

	if (len(m1.Peers) != 2) || (m1.Hops != 1) {
		assert.Fail(t, "Should have been 2 peers and 1 hop")
	}

}

func TestNewMessageWithNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "Code did not panic on creating new message with nil marshalizer!")
		}
	}()

	NewMessage("", []byte{}, nil)
}

func TestMessageWithNilsMarshalizers(t *testing.T) {
	m := NewMessage("", []byte{}, service.GetMarshalizerService())

	m.marsh = nil

	_, err := m.ToByteArray()
	assert.NotNil(t, err)

	_, err = CreateFromByteArray(nil, []byte{})
	assert.NotNil(t, err)

}
