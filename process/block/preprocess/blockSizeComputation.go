package preprocess

import (
	"crypto/rand"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

// maxAllowedSizeInBytes defines how many bytes are allowed as payload in a message. We can not have 1MB of data
// as there are cases when extra data is needed (consensus data fields and p2p message fields)
const maxAllowedSizeInBytes = uint32(core.MegabyteSize * 95 / 100)

// blockSizeComputation is able to estimate the size in bytes of a block body given the number of contained
// transactions hashes and the number of miniblocks. It uses the marshalizer to compute the size as precise as possible.
type blockSizeComputation struct {
	miniblockSize uint32
	txSize        uint32

	numMiniBlocks uint32
	numTxs        uint32
}

// NewBlockSizeComputation creates a blockSizeComputation instance
func NewBlockSizeComputation(marshalizer marshal.Marshalizer) (*blockSizeComputation, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	bsc := &blockSizeComputation{}
	err := bsc.precomputeValues(marshalizer)
	if err != nil {
		return nil, err
	}

	return bsc, nil
}

func (bsc *blockSizeComputation) precomputeValues(marshalizer marshal.Marshalizer) error {
	oneEmptyMiniblockSize, err := bsc.computeSize(marshalizer, bsc.generateDummyBlockbody(1, 0))
	if err != nil {
		return err
	}

	oneMiniblockSizeWithTenTxs, err := bsc.computeSize(marshalizer, bsc.generateDummyBlockbody(1, 10))
	if err != nil {
		return err
	}

	oneMiniblockSizeWithTwentyTxs, err := bsc.computeSize(marshalizer, bsc.generateDummyBlockbody(1, 20))
	if err != nil {
		return err
	}

	tenMiniblocksWithTenTxs, err := bsc.computeSize(marshalizer, bsc.generateDummyBlockbody(10, 10))
	if err != nil {
		return err
	}

	bsc.txSize = (oneMiniblockSizeWithTwentyTxs - oneMiniblockSizeWithTenTxs) / 10
	bsc.miniblockSize = core.MaxUint32(oneEmptyMiniblockSize, oneMiniblockSizeWithTenTxs-10*bsc.txSize)
	bsc.miniblockSize = core.MaxUint32(bsc.miniblockSize, (tenMiniblocksWithTenTxs-100*bsc.txSize)/10)

	return nil
}

func (bsc *blockSizeComputation) computeSize(marshalizer marshal.Marshalizer, object interface{}) (uint32, error) {
	buff, err := marshalizer.Marshal(object)
	if err != nil {
		return 0, err
	}

	return uint32(len(buff)), nil
}

func (bsc *blockSizeComputation) generateDummyBlockbody(numMiniblocks int, numTxHashesPerMiniblock int) block.Body {
	body := make(block.Body, numMiniblocks)
	for i := 0; i < numMiniblocks; i++ {
		body[i] = bsc.generateDummyMiniblock(numTxHashesPerMiniblock)
	}

	return body
}

func (bsc *blockSizeComputation) generateDummyMiniblock(numTxHashes int) *block.MiniBlock {
	mb := &block.MiniBlock{
		TxHashes:        nil,
		ReceiverShardID: 999,
		SenderShardID:   999,
		Type:            255,
	}

	mb.Type = block.TxBlock
	mb.TxHashes = make([][]byte, numTxHashes)
	for i := 0; i < numTxHashes; i++ {
		mb.TxHashes[i] = make([]byte, 32)
		_, _ = rand.Reader.Read(mb.TxHashes[i])
	}

	return mb
}

// Init reset the stored values of accumulated numTxs and numMiniBlocks
func (bsc *blockSizeComputation) Init() {
	atomic.StoreUint32(&bsc.numTxs, 0)
	atomic.StoreUint32(&bsc.numMiniBlocks, 0)
}

// AddNumMiniBlocks adds the provided value to numMiniBlocks in a concurrent safe manner
func (bsc *blockSizeComputation) AddNumMiniBlocks(numMiniBlocks int) {
	atomic.AddUint32(&bsc.numMiniBlocks, uint32(numMiniBlocks))
}

// AddNumTxs adds the provided value to numTxs in a concurrent safe manner
func (bsc *blockSizeComputation) AddNumTxs(numTxs int) {
	atomic.AddUint32(&bsc.numTxs, uint32(numTxs))
}

// IsMaxBlockSizeReached returns true if the provided number of
func (bsc *blockSizeComputation) IsMaxBlockSizeReached(numNewMiniBlocks int, numNewTxs int) bool {
	totalMiniBlocks := atomic.LoadUint32(&bsc.numMiniBlocks) + uint32(numNewMiniBlocks)
	totalTxs := atomic.LoadUint32(&bsc.numTxs) + uint32(numNewTxs)

	return bsc.isMaxBlockSizeReached(totalMiniBlocks, totalTxs)
}

func (bsc *blockSizeComputation) isMaxBlockSizeReached(totalMiniBlocks uint32, totalTxs uint32) bool {
	miniblocksSize := bsc.miniblockSize * totalMiniBlocks
	txsSize := bsc.txSize * totalTxs

	return miniblocksSize+txsSize > maxAllowedSizeInBytes
}

// MaxTransactionsInOneMiniblock returns the maximum transactions in a single miniblock
func (bsc *blockSizeComputation) MaxTransactionsInOneMiniblock() int {
	return int((maxAllowedSizeInBytes - bsc.miniblockSize) / bsc.txSize)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsc *blockSizeComputation) IsInterfaceNil() bool {
	return bsc == nil
}
