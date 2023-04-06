package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestGetEncodedAddress(t *testing.T) {
	t.Parallel()

	address := []byte("12345678901234567890123456789012")
	expectedEncodedAddr := "erd1xyerxdp4xcmnswfsxyerxdp4xcmnswfsxyerxdp4xcmnswfsxyeqlrqt99"
	txUnmarshalledHandler := &txUnmarshaller{
		addressPubKeyConverter: &testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return len(address)
			},
			SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
				require.Equal(t, pkBytes, address)
				return expectedEncodedAddr
			},
		},
	}

	encodedAddr := txUnmarshalledHandler.getEncodedAddress(address)
	require.Equal(t, expectedEncodedAddr, encodedAddr)

	encodedAddr = txUnmarshalledHandler.getEncodedAddress([]byte("abc"))
	require.Empty(t, encodedAddr)
}

func TestBigIntToStr(t *testing.T) {
	t.Parallel()

	val := bigIntToStr(big.NewInt(123))
	require.Equal(t, "123", val)

	val = bigIntToStr(nil)
	require.Empty(t, val)
}
