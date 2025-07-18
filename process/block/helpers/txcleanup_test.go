package helpers

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
)

func getMockBody(hash []byte, numtxsPerPool int) *block.Body{
	body := &block.Body{}
	selfShard := 0
	numShards := 3
	
	for i := 0; i < numShards; i++ {
		txHashes := [][]byte{}
		
		for j := 0; j < numtxsPerPool; j++ {
			txHashes = append(txHashes, fmt.Appendf(nil, "hash-%d-%d", i, j))
		}
		
		if(i == 0) {
			txHashes = append(txHashes, hash)
		}
		
		body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{
			SenderShardID:   uint32(selfShard),
			ReceiverShardID: uint32(i),
			TxHashes:        txHashes,
		})
	}
	
	return body		
}

func TestComputeRandomnessForTxCleaning(t *testing.T) {
	t.Parallel()

	t.Run("randomness", func(t *testing.T) {
		body := getMockBody([]byte("alice1-hash"), 5)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(3))
		body = getMockBody([]byte("bob1-hash"), 5)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(5))
		body = getMockBody([]byte("bob4-hash"), 3)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(0))
		body = getMockBody([]byte("alice1-hash"), 17)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(3))
		body = getMockBody([]byte("bob1-hash"), 17)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(5))
		body = getMockBody([]byte("48602391228527b14364f15eb9145e62ad0a338816b4d8bee01344e158508911"), 180)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(9))
		body = getMockBody([]byte("8b5fa093cb63c61977a55856086e34ef7f8a00c8187f85b1ab2286714016768b"), 180)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(14))
		body = getMockBody([]byte("9ac874ca88e071db6fa64e654d95480761a1ea10aac0e8e99db7df5cfdfd3157"), 25000)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(19779))
		body = getMockBody([]byte("4f8a89dd647ee6f936949eb407935901a2e2fd84f6b76c532fea0c839fe01f19"), 25000)
		assert.Equal(t, ComputeRandomnessForCleanup(body), uint64(2701))
	})
}
