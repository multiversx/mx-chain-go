package firehose

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/alteredAccount"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/stretchr/testify/require"
)

func TestFirehoseIndexer_SaveBlock(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}

	blk := &firehose.FirehoseBlock{
		HeaderHash:      []byte("hasdsada2132131h"),
		HeaderType:      "879bdsajhk3",
		AlteredAccounts: []*alteredAccount.AlteredAccount{{Nonce: 4444}},
	}
	blkBytes, err := marshaller.Marshal(blk)
	require.Nil(t, err)

	_ = blkBytes
	//_, err = fmt.Fprintf(os.Stdout, "%x", blkBytes)
	hexDecodedBytes, err := hex.DecodeString("120b383739626473616a686b331a10686173647361646132313332313331682a0610dc221a0100")
	require.Nil(t, err)

	newBlk := &firehose.FirehoseBlock{}
	err = marshaller.Unmarshal(newBlk, hexDecodedBytes)
	require.Nil(t, err)

}
