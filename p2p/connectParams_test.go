package p2p_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

func TestFailNewBadConnectParams(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	//invalid port
	p2p.NewConnectParamsFromPort(65536)
}

func TestNotFailNewConnectParams(t *testing.T) {
	cp := p2p.NewConnectParamsFromPort(65535)

	buff, err := cp.PrivKey.Bytes()
	assert.Nil(t, err)

	fmt.Printf("Private key: %v\n", buff)

	buff, err = cp.PubKey.Bytes()
	assert.Nil(t, err)

	fmt.Printf("Public key: %v\n", buff)
	fmt.Printf("ID: %v\n", cp.ID.Pretty())

}

func TestNewConnectParams(t *testing.T) {
	buffPrivKey := []byte{8, 2, 18, 32, 240, 44, 132, 237, 70,
		30, 188, 118, 0, 25, 28, 224, 190, 134, 240, 66, 58, 63,
		181, 131, 208, 151, 28, 19, 89, 49, 67, 184, 225, 63, 248, 166}

	buffPubKey := []byte{8, 2, 18, 33, 2, 177, 16, 21, 115, 117, 145,
		182, 92, 142, 155, 26, 135, 89, 80, 140, 70, 129, 67, 40, 43, 71,
		196, 19, 170, 252, 70, 103, 157, 161, 72, 124, 36}

	pid := "16Uiu2HAm7LrNF9uTDVBPxQovFGcYJGqu8ZEndNADitpeQh52yCN7"

	prv, err := crypto.UnmarshalPrivateKey(buffPrivKey)
	assert.Nil(t, err)

	params := p2p.NewConnectParams("0.0.0.0", 4000, prv)

	buffPrivKeyComputed, err := prv.Bytes()
	assert.Nil(t, err)

	assert.Equal(t, 0, bytes.Compare(buffPrivKeyComputed, buffPrivKey))

	buffPubKeyComputed, err := params.PubKey.Bytes()
	assert.Nil(t, err)

	assert.Equal(t, 0, bytes.Compare(buffPrivKeyComputed, buffPrivKey))
	assert.Equal(t, 0, bytes.Compare(buffPubKeyComputed, buffPubKey))

	assert.Equal(t, pid, params.ID.Pretty())
}

func TestSignVerify(t *testing.T) {
	params := p2p.NewConnectParamsFromPort(4000)

	bPrivKey, _ := params.PrivKey.Bytes()
	fmt.Printf("Priv key: %v\n", hex.EncodeToString(bPrivKey))

	bPubKey, _ := params.PubKey.Bytes()
	fmt.Printf("Pub key: %v\n", hex.EncodeToString(bPubKey))

	fmt.Printf("ID: %v\n", params.ID.Pretty())

	buffSig, err := params.PrivKey.Sign([]byte{65, 66, 67})
	assert.Nil(t, err)
	fmt.Printf("Sig: %v\n", base64.StdEncoding.EncodeToString(buffSig))

	buffPubKey, err := crypto.MarshalPublicKey(params.PubKey)
	fmt.Printf("Marshaled pub key: %v\n", base64.StdEncoding.EncodeToString(buffPubKey))

	//recovery and verify
	pubKeyVerif, err := crypto.UnmarshalPublicKey(buffPubKey)
	assert.Nil(t, err)

	signed, err := pubKeyVerif.Verify([]byte{65, 66, 67}, buffSig)
	assert.Nil(t, err)

	fmt.Printf("Signed \"ABC\"? %v\n", signed)
	idVerif, err := peer.IDFromPublicKey(pubKeyVerif)
	assert.Nil(t, err)
	fmt.Printf("Signed by %v\n", idVerif.Pretty())

	assert.Equal(t, idVerif, params.ID)

}
