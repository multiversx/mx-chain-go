package marshal_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dataGenerator interface {
	// GenerateDummyArray generates an array of data of the implementer type
	// The implementer needs to implement GogoProtoHelper as well
	GenerateDummyArray() []interface{}

	// resett the content of an object
	Reset()
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
	totalOut := uint64(0)
	b.ReportAllocs()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buf, _ := m.Marshal(dArray[i%l])
		totalOut += uint64(len(buf))
	}

	b.ReportMetric(float64(totalOut)/float64(b.N), "stremB/op")
}

func benchUnmarshal(b *testing.B, m marshal.Marshalizer, obj interface{}, validate bool) {
	b.StopTimer()
	dArray := obj.(dataGenerator).GenerateDummyArray()
	l := len(dArray)
	serialized := make([][]byte, l)
	var err error

	for i, obj := range dArray {
		serialized[i], err = m.Marshal(obj)
		require.Nil(b, err)
	}

	b.ReportAllocs()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		n := i % l
		obj.(dataGenerator).Reset()
		err := m.Unmarshal(obj, serialized[n])
		assert.Nil(b, err)
		if validate {
			orig := dArray[n]
			require.Equal(b, orig, obj)
		}
	}
}

// benchmarks
func BenchmarkMarshal(b *testing.B) {

	hdr := &Header{}
	mb := &MiniBlock{}
	tx := &Transaction{}

	gmsr := &marshal.GogoProtoMarshalizer{}
	jmsr := &marshal.JsonMarshalizer{}

	benchData := []struct {
		name string
		obj  dataGenerator
		msr  marshal.Marshalizer
	}{
		{name: "HdrGogo", obj: hdr, msr: gmsr},
		{name: "HdrJSON", obj: hdr, msr: jmsr},
		{name: "MbGogo", obj: mb, msr: gmsr},
		{name: "MbJSON", obj: mb, msr: jmsr},
		{name: "TxGogo", obj: tx, msr: gmsr},
		{name: "TxJSON", obj: tx, msr: jmsr},
	}
	for _, bd := range benchData {
		b.Run(bd.name, func(sb *testing.B) {
			benchMarshal(sb, bd.msr, bd.obj)
		})
	}
}

func BenchmarkUnarshal(b *testing.B) {

	benchmarkUnarshal(b, false)
}

func BenchmarkUnarshalValidate(b *testing.B) {
	benchmarkUnarshal(b, true)
}

func benchmarkUnarshal(b *testing.B, validate bool) {
	hdr := &Header{}
	mb := &MiniBlock{}
	tx := &Transaction{}

	gmsr := &marshal.GogoProtoMarshalizer{}
	jmsr := &marshal.JsonMarshalizer{}

	benchData := []struct {
		name string
		obj  dataGenerator
		msr  marshal.Marshalizer
	}{
		{name: "HdrGogo", obj: hdr, msr: gmsr},
		{name: "HdrJSON", obj: hdr, msr: jmsr},
		{name: "MbGogo", obj: mb, msr: gmsr},
		{name: "MbJSON", obj: mb, msr: jmsr},
		{name: "TxGogo", obj: tx, msr: gmsr},
		{name: "TxJSON", obj: tx, msr: jmsr},
	}
	for _, bd := range benchData {
		b.Run(bd.name, func(sb *testing.B) {
			benchUnmarshal(sb, bd.msr, bd.obj, validate)
		})
	}
}

// GenerateDummyArray is used to generate an array of MiniBlockHeaders with dummy data
func (sBlock *MiniBlock) GenerateDummyArray() []interface{} {
	sBlocks := make([]interface{}, 0, 1000)

	for i := 0; i < 1000; i++ {
		lenTxHashes := rand.Intn(200) + 1
		txHashes := make([][]byte, 0, lenTxHashes)
		for k := 0; k < lenTxHashes; k++ {
			txHashes = append(txHashes, []byte(RandomStr(32)))
		}
		sBlocks = append(sBlocks, &MiniBlock{
			MiniBlock: block.MiniBlock{
				TxHashes:        txHashes,
				ReceiverShardID: uint32(rand.Intn(1000)),
				SenderShardID:   uint32(rand.Intn(1000)),
				Type:            0,
			},
		})
	}

	return sBlocks
}

// GenerateDummyArray is used to generate an array of block headers with dummy data
func (h *Header) GenerateDummyArray() []interface{} {
	headers := make([]interface{}, 0, 1000)

	mbhs := make([]block.MiniBlockHeader, 25)
	pcs := make([]block.PeerChange, 25)

	for i := 0; i < 25; i++ {
		mbhs[i] = block.MiniBlockHeader{
			Hash:            []byte("mini block header"),
			SenderShardID:   uint32(1),
			ReceiverShardID: uint32(i),
		}

		pcs[i] = block.PeerChange{
			PubKey:      []byte("public key"),
			ShardIdDest: uint32(i),
		}
	}

	for i := 0; i < 1000; i++ {
		headers = append(headers, &Header{
			Header: block.Header{
				Nonce:            uint64(rand.Int63n(10000)),
				PrevHash:         []byte(RandomStr(32)),
				PrevRandSeed:     []byte(RandomStr(32)),
				RandSeed:         []byte(RandomStr(32)),
				PubKeysBitmap:    []byte(RandomStr(10)),
				ShardID:          uint32(rand.Intn(20)),
				TimeStamp:        uint64(rand.Intn(20)),
				Round:            uint64(rand.Intn(20000)),
				Epoch:            uint32(rand.Intn(20000)),
				BlockBodyType:    block.TxBlock,
				Signature:        []byte(RandomStr(32)),
				LeaderSignature:  nil,
				MiniBlockHeaders: mbhs,
				PeerChanges:      pcs,
				RootHash:         []byte("root hash"),
				MetaBlockHashes:  nil,
				TxCount:          0,
			},
		})
	}

	return headers
}

// GenerateDummyArray is used to generate an array of transactions with dummy data
func (tx *Transaction) GenerateDummyArray() []interface{} {
	transactions := make([]interface{}, 0, 1000)

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
				Data:      []byte(RandomStr(32)),
				Signature: []byte(RandomStr(32)),
			},
		})
	}

	return transactions
}

func (h *Header) Reset() {
	if h != nil {
		*h = Header{}
	}
}

func (mb *MiniBlock) Reset() {
	if mb != nil {
		*mb = MiniBlock{}
	}
}

func (tx *Transaction) Reset() {
	if tx != nil {
		*tx = Transaction{}
	}
}
