package process

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/stretchr/testify/require"
)

func TestPrepareBufferMiniblocks(t *testing.T) {
	var buff bytes.Buffer

	meta := []byte("test1")
	serializedData := []byte("test2")

	buff = prepareBufferMiniblocks(buff, meta, serializedData)

	var expectedBuff bytes.Buffer
	serializedData = append(serializedData, "\n"...)
	expectedBuff.Grow(len(meta) + len(serializedData))
	_, _ = expectedBuff.Write(meta)
	_, _ = expectedBuff.Write(serializedData)

	require.Equal(t, expectedBuff, buff)
}

func TestSerializeScResults(t *testing.T) {
	t.Parallel()

	scResult1 := &types.ScResult{
		Hash:     "hash1",
		Nonce:    1,
		GasPrice: 10,
		GasLimit: 50,
	}
	scResult2 := &types.ScResult{
		Hash:     "hash2",
		Nonce:    2,
		GasPrice: 10,
		GasLimit: 50,
	}
	scrs := []*types.ScResult{scResult1, scResult2}

	res, err := serializeScResults(scrs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "hash1" } }
{"nonce":1,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
{ "index" : { "_id" : "hash2" } }
{"nonce":2,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeReceipts(t *testing.T) {
	t.Parallel()

	rec1 := &types.Receipt{
		Hash:   "recHash1",
		Sender: "sender1",
		TxHash: "txHash1",
	}
	rec2 := &types.Receipt{
		Hash:   "recHash2",
		Sender: "sender2",
		TxHash: "txHash2",
	}

	recs := []*types.Receipt{rec1, rec2}

	res, err := serializeReceipts(recs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "recHash1" } }
{"value":"","sender":"sender1","txHash":"txHash1","timestamp":0}
{ "index" : { "_id" : "recHash2" } }
{"value":"","sender":"sender2","txHash":"txHash2","timestamp":0}
`
	require.Equal(t, expectedRes, res[0].String())
}
