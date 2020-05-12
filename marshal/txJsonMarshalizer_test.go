package marshal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTxStruct struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Nonce    uint64 `json:"nonce"`
	Data     string `json:"data"`
}

func TestTxJsonMarshalizer_MarshalUnmarshalWithCharactersThatCouldBeEncoded(t *testing.T) {
	t.Parallel()

	tx := &TestTxStruct{
		Sender:   "sndr",
		Receiver: "rcvr",
		Nonce:    10,
		Data:     "data@~`!@#$^&*()_=[]{};'<>?,./|<>><!!!!!",
	}

	tjm := TxJsonMarshalizer{}

	// custom json marshalizer
	marshaledTx1, err := tjm.Marshal(tx)
	assert.NotNil(t, marshaledTx1)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(marshaledTx1), tx.Data))

	var resTx1 *TestTxStruct
	err = tjm.Unmarshal(&resTx1, marshaledTx1)
	assert.Equal(t, tx, resTx1)
	assert.NoError(t, err)
}
