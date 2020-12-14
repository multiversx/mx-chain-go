package examples

import (
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/stretchr/testify/require"
)

const blsPemContents = `
-----BEGIN PRIVATE KEY for cb66c844dc64e0cab854997df9b3fd1b1c071ee25e6634e6b95bdb15f0acb38e7c5ac20d02f231a03fde3abb8220951328a7f550915ff9da4a1960c8005d6dfabc50f776bd17433f29e8d0566871380b439024a70dfade593173c6462ad49318-----
NGYwYmRkNWEyMjM3Y2I2MWU0OTU3NjNkOTNjM2E4NTU3NzQyMGZhZjUxZTRiYmI0
MGQ1YTAzZTkxNWYxZTIxZg==
-----END PRIVATE KEY for cb66c844dc64e0cab854997df9b3fd1b1c071ee25e6634e6b95bdb15f0acb38e7c5ac20d02f231a03fde3abb8220951328a7f550915ff9da4a1960c8005d6dfabc50f776bd17433f29e8d0566871380b439024a70dfade593173c6462ad49318-----
`

var blsSigner = &mclsig.BlsSingleSigner{}
var keyGen = signing.NewKeyGenerator(&mcl.SuiteBLS12{})

func TestBLSSigning(t *testing.T) {
	privateKey, publicKey, publicKeyAsHex := loadSkPk(t)

	message := "message to be signed"
	signature := sign(t, []byte(message), privateKey)

	fmt.Printf("message: %s\npublicKey: %s\nsignature: %s\n\n",
		message,
		publicKeyAsHex,
		hex.EncodeToString(signature),
	)

	verify(t, []byte(message), publicKey, signature)
}

func loadSkPk(t *testing.T) (crypto.PrivateKey, crypto.PublicKey, string) {
	buff := []byte(blsPemContents)

	blkRecovered, _ := pem.Decode(buff)
	require.NotNil(t, blkRecovered)

	blockType := blkRecovered.Type
	header := "PRIVATE KEY for "
	require.True(t, strings.Index(blockType, header) == 0)
	blockTypeString := blockType[len(header):]

	skHex := blkRecovered.Bytes
	skBuff, err := hex.DecodeString(string(skHex))
	require.Nil(t, err)

	sk, err := keyGen.PrivateKeyFromByteArray(skBuff)
	require.Nil(t, err)

	pk := sk.GeneratePublic()
	pkBuff, err := pk.ToByteArray()
	require.Nil(t, err)
	require.Equal(t, blockTypeString, hex.EncodeToString(pkBuff))

	return sk, pk, blockTypeString
}

func sign(t *testing.T, message []byte, privKey crypto.PrivateKey) []byte {
	signature, err := blsSigner.Sign(privKey, message)
	require.Nil(t, err)

	return signature
}

func verify(t *testing.T, message []byte, publicKey crypto.PublicKey, signature []byte) {
	err := blsSigner.Verify(publicKey, message, signature)
	require.Nil(t, err)
}
