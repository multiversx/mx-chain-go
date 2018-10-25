package p2p_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshal(t *testing.T) {
	mrsh := &mock.MarshalizerMock{}

	m1 := p2p.NewMessage("p1", []byte("ABCDEF"), mrsh)
	fmt.Println("Original:")
	fmt.Println(m1)

	buff, err := m1.ToByteArray()
	assert.Nil(t, err)
	fmt.Println("Marshaled:")
	fmt.Println(string(buff))

	m2, err := p2p.CreateFromByteArray(mrsh, buff)
	assert.Nil(t, err)
	fmt.Println("Unmarshaled:")
	fmt.Println(*m2)

	assert.Equal(t, m1, m2)
}

func TestAddHop(t *testing.T) {
	m1 := p2p.NewMessage("p1", []byte("ABCDEF"), &mock.MarshalizerMock{})

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

	p2p.NewMessage("", []byte{}, nil)
}

func TestMessageWithNilsMarshalizers(t *testing.T) {
	m := p2p.NewMessage("", []byte{}, &mock.MarshalizerMock{})

	m.SetMarshalizer(nil)

	_, err := m.ToByteArray()
	assert.NotNil(t, err)

	_, err = p2p.CreateFromByteArray(nil, []byte{})
	assert.NotNil(t, err)
}

func TestMessage_Sign_NilParams_ShouldErr(t *testing.T) {
	param := &p2p.ConnectParams{}

	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	err := mes.Sign(param.PrivKey)
	assert.NotNil(t, err)

}

func TestMessage_Sign_Values_ShouldWork(t *testing.T) {
	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())

	fmt.Printf("Payload: %v\n", string(mes.Payload))
	fmt.Printf("Pub key: %v\n", hex.EncodeToString(mes.PubKey))
	fmt.Printf("Sig: %v\n", base64.StdEncoding.EncodeToString(mes.Sig))

}

func TestMessage_Verify_NilParams_ShouldFalse(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}
	mes.SetSigned(true)

	err := mes.VerifyAndSetSigned()
	assert.Nil(t, err)

	assert.False(t, mes.Signed())

}

func TestMessage_Verify_EmptyParams_ShouldFalse(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}
	mes.Sig = make([]byte, 0)
	mes.PubKey = make([]byte, 0)
	mes.SetSigned(true)

	err := mes.VerifyAndSetSigned()
	assert.Nil(t, err)

	assert.False(t, mes.Signed())

}

func TestMessage_Verify_EmptyPeers_ShouldFalse(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())

	err = mes.VerifyAndSetSigned()
	assert.Nil(t, err)

	assert.False(t, mes.Signed())
}

func TestMessage_Verify_WrongPubKey_ShouldErr(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())
	mes.PubKey = []byte{65, 66, 67}
	mes.AddHop(param.ID.Pretty())

	err = mes.VerifyAndSetSigned()
	assert.NotNil(t, err)

	assert.False(t, mes.Signed())
}

func TestMessage_Verify_MismatchID_ShouldErr(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	param2, err := p2p.NewConnectParamsFromPort(4001)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())
	mes.PubKey, err = crypto.MarshalPublicKey(param2.PubKey)
	assert.Nil(t, err)
	mes.AddHop(param.ID.Pretty())

	err = mes.VerifyAndSetSigned()
	assert.NotNil(t, err)

	assert.False(t, mes.Signed())

}

func TestMessage_Verify_WrongSig_ShouldErr(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())
	mes.Sig = []byte{65, 66, 67}
	assert.Nil(t, err)
	mes.AddHop(param.ID.Pretty())

	err = mes.VerifyAndSetSigned()
	assert.NotNil(t, err)

	assert.False(t, mes.Signed())

}

func TestMessage_Verify_TamperedPayload_ShouldErr(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())
	mes.Payload = []byte{65, 66, 68}
	assert.Nil(t, err)
	mes.AddHop(param.ID.Pretty())

	err = mes.VerifyAndSetSigned()
	assert.NotNil(t, err)

	assert.False(t, mes.Signed())

}

func TestMessage_Verify_Values_ShouldWork(t *testing.T) {
	mes := p2p.Message{Payload: []byte{65, 66, 67}, Type: "string"}

	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	err = mes.Sign(param.PrivKey)
	assert.Nil(t, err)
	assert.True(t, mes.Signed())
	assert.Nil(t, err)
	mes.AddHop(param.ID.Pretty())

	err = mes.VerifyAndSetSigned()
	assert.Nil(t, err)

	assert.True(t, mes.Signed())

}

func BenchmarkMessage_Sign(b *testing.B) {
	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(b, err)

	for i := 0; i < b.N; i++ {
		mes := p2p.Message{Payload: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" + strconv.Itoa(i))}
		err := mes.Sign(param.PrivKey)
		assert.Nil(b, err)
	}
}

func BenchmarkMessage_SignVerif(b *testing.B) {
	param, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(b, err)

	for i := 0; i < b.N; i++ {
		mes := p2p.Message{Payload: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" + strconv.Itoa(i))}
		err := mes.Sign(param.PrivKey)
		assert.Nil(b, err)

		err = mes.VerifyAndSetSigned()
		assert.Nil(b, err)
	}
}
