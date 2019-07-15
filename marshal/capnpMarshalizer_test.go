package marshal_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

type dataGenerator interface {
	// GenerateDummyArray generates an array of data of the implementer type
	// The implementer needs to implement CapnpHelper as well
	GenerateDummyArray() []data.CapnpHelper
}

type Header struct {
	block.Header
}

type MiniBlock struct {
	block.MiniBlock
}

type Transaction struct {
	transaction.Transaction
}

// RandomStr generates random strings of set length
func RandomStr(l int) string {
	buf := make([]byte, l)

	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}

func benchMarshal(b *testing.B, m marshal.Marshalizer, obj dataGenerator) {
	b.StopTimer()

	dArray := obj.GenerateDummyArray()
	l := len(dArray)
	b.ReportAllocs()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m.Marshal(dArray[i%l])
	}
}

func benchUnmarshal(b *testing.B, m marshal.Marshalizer, obj interface{}, validate bool) {
	b.StopTimer()
	dArray := obj.(dataGenerator).GenerateDummyArray()
	l := len(dArray)
	serialized := make([][]byte, l)

	for i, obj := range dArray {
		mar, _ := m.Marshal(obj)
		t := make([]byte, len(mar))

		_ = copy(t, mar)
		serialized[i] = t
	}

	b.ReportAllocs()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n := i % l
		err := m.Unmarshal(obj.(data.CapnpHelper), serialized[n])

		assert.Nil(b, err)

		// Check unmarshalled data as expected
		if validate {
			orig := dArray[n]
			assert.Equal(b, orig, obj)
		}
	}
}

// benchmarks
func BenchmarkCapnprotoTransactionMarshal(b *testing.B) {
	tx := &Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, tx)
}

func BenchmarkJsonTransactionMarshal(b *testing.B) {
	tx := &Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, tx)
}

func BenchmarkCapnprotoTransactionUnmarshalNoValidate(b *testing.B) {
	tx := &Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, tx, false)
}

func BenchmarkJsonTransactionUnmarshalNoValidate(b *testing.B) {
	tx := &Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, tx, false)
}

func BenchmarkCapnprotoTransactionUnmarshalValidate(b *testing.B) {
	tx := &Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, tx, true)
}

func BenchmarkJsonTransactionUnmarshalValidate(b *testing.B) {
	tx := &Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, tx, true)
}

func BenchmarkCapnprotoMiniBlocksMarshal(b *testing.B) {
	bl := &MiniBlock{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, bl)
}

func BenchmarkJsonMiniBlocksMarshal(b *testing.B) {
	bl := &MiniBlock{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, bl)
}

func BenchmarkCapnprotoMiniBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &MiniBlock{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, false)
}

func BenchmarkJsonMiniBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &MiniBlock{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, false)
}

func BenchmarkCapnprotoMiniBlocksUnmarshalValidate(b *testing.B) {
	bl := &MiniBlock{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkJsonMiniBlocksUnmarshalValidate(b *testing.B) {
	bl := &MiniBlock{}
	cmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkCapnprotoHeaderMarshal(b *testing.B) {
	h := &Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, h)
}

func BenchmarkJsonHeaderMarshal(b *testing.B) {
	h := &Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, h)
}

func BenchmarkCapnprotoHeaderUnmarshalNoValidate(b *testing.B) {
	h := &Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, h, false)
}

func BenchmarkJsonHeaderUnmarshalNoValidate(b *testing.B) {
	h := &Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, h, false)
}

func BenchmarkCapnprotoHeaderUnmarshalValidate(b *testing.B) {
	h := &Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, h, true)
}

func BenchmarkJsonHeaderUnmarshalValidate(b *testing.B) {
	h := &Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, h, true)
}

// GenerateDummyArray is used to generate an array of MiniBlockHeaders with dummy data
func (sBlock *MiniBlock) GenerateDummyArray() []data.CapnpHelper {
	sBlocks := make([]data.CapnpHelper, 0, 1000)

	for i := 0; i < 1000; i++ {
		lenTxHashes := rand.Intn(20) + 1
		txHashes := make([][]byte, 0, lenTxHashes)
		for k := 0; k < lenTxHashes; k++ {
			txHashes = append(txHashes, []byte(RandomStr(32)))
		}
		sBlocks = append(sBlocks, &MiniBlock{
			MiniBlock: block.MiniBlock{
				TxHashes:        txHashes,
				ReceiverShardID: uint32(rand.Intn(1000)),
				SenderShardID:   uint32(rand.Intn(1000)),
			},
		})
	}

	return sBlocks
}

// GenerateDummyArray is used to generate an array of block headers with dummy data
func (h *Header) GenerateDummyArray() []data.CapnpHelper {
	headers := make([]data.CapnpHelper, 0, 1000)

	mbh := block.MiniBlockHeader{
		Hash:            []byte("mini block header"),
		SenderShardID:   uint32(1),
		ReceiverShardID: uint32(0),
	}

	pc := block.PeerChange{
		PubKey:      []byte("public key"),
		ShardIdDest: uint32(0),
	}

	for i := 0; i < 1000; i++ {
		headers = append(headers, &Header{
			Header: block.Header{
				Nonce:            uint64(rand.Int63n(10000)),
				PrevHash:         []byte(RandomStr(32)),
				PrevRandSeed:     []byte(RandomStr(32)),
				RandSeed:         []byte(RandomStr(32)),
				ShardId:          uint32(rand.Intn(20)),
				TimeStamp:        uint64(rand.Intn(20)),
				Round:            uint32(rand.Intn(20000)),
				Epoch:            uint32(rand.Intn(20000)),
				BlockBodyType:    block.TxBlock,
				Signature:        []byte(RandomStr(32)),
				PubKeysBitmap:    []byte(RandomStr(10)),
				MiniBlockHeaders: []block.MiniBlockHeader{mbh},
				PeerChanges:      []block.PeerChange{pc},
				RootHash:         []byte("root hash"),
			},
		})
	}

	return headers
}

// GenerateDummyArray is used to generate an array of transactions with dummy data
func (tx *Transaction) GenerateDummyArray() []data.CapnpHelper {
	transactions := make([]data.CapnpHelper, 0, 1000)

	val := big.NewInt(0)
	_ = val.GobDecode([]byte(RandomStr(32)))

	for i := 0; i < 1000; i++ {
		transactions = append(transactions, &Transaction{
			Transaction: transaction.Transaction{
				Nonce:     uint64(rand.Int63n(10000)),
				Value:     val,
				RcvAddr:   []byte(RandomStr(32)),
				SndAddr:   []byte(RandomStr(32)),
				GasPrice:  uint64(rand.Int63n(10000)),
				GasLimit:  uint64(rand.Int63n(10000)),
				Data:      RandomStr(32),
				Signature: []byte(RandomStr(32)),
				Challenge: []byte(RandomStr(32)),
			},
		})
	}

	return transactions
}
