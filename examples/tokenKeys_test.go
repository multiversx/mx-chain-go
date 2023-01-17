package examples

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

func TestComputeTokenStorageKey(t *testing.T) {
	t.Parallel()

	require.Equal(t, hex.EncodeToString([]byte("ELRONDesdtWEGLD-bd4d79")), computeStorageKey("WEGLD-bd4d79", 0))
	require.Equal(t, hex.EncodeToString([]byte("ELRONDesdtMYNFT-aaabbbF")), computeStorageKey("MYNFT-aaabbb", 70))
}

func TestDecodeTokenFromProtoBytes(t *testing.T) {
	t.Parallel()

	keyHex := "12020001221108013a0d6e657720617474726962757465"
	valueBytes, err := hex.DecodeString(keyHex)
	require.NoError(t, err)

	marshaller := marshal.GogoProtoMarshalizer{}
	recoveredToken := esdt.ESDigitalToken{}

	err = marshaller.Unmarshal(&recoveredToken, valueBytes)
	require.NoError(t, err)

	fmt.Println(spew.Sdump(recoveredToken))
}

func computeStorageKey(tokenIdentifier string, tokenNonce uint64) string {
	key := []byte(core.ProtectedKeyPrefix)
	key = append(key, core.ESDTKeyIdentifier...)
	key = append(key, []byte(tokenIdentifier)...)

	if tokenNonce > 0 {
		nonceBI := big.NewInt(int64(tokenNonce))

		key = append(key, nonceBI.Bytes()...)
	}

	return hex.EncodeToString(key)
}
