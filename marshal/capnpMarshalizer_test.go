package marshal_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/stretchr/testify/assert"
	"math/big"
)

type dataGenerator interface {
	// GenerateDummyArray generates an array of data of the implementer type
	// The implementer needs to implement CapnpHelper as well
	GenerateDummyArray() []data.CapnpHelper
}

type TxBlockBody struct {
	block.TxBlockBody
}

type StateBlockBody struct {
	block.StateBlockBody
}

type PeerBlockBody struct {
	block.PeerBlockBody
}

type Header struct {
	block.Header
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
		m.Marshal(dArray[i%l])
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

		copy(t, mar)
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

func BenchmarkCapnprotoStateBlocksMarshal(b *testing.B) {
	bl := &StateBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, bl)
}

func BenchmarkJsonStateBlocksMarshal(b *testing.B) {
	bl := &StateBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, bl)
}

func BenchmarkCapnprotoStateBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &StateBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, false)
}

func BenchmarkJsonStateBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &StateBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, false)
}

func BenchmarkCapnprotoStateBlocksUnmarshalValidate(b *testing.B) {
	bl := &StateBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkJsonStateBlocksUnmarshalValidate(b *testing.B) {
	bl := &StateBlockBody{}
	cmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkCapnprotoPeerBlocksMarshal(b *testing.B) {
	bl := &PeerBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, bl)
}

func BenchmarkJsonPeerBlocksMarshal(b *testing.B) {
	bl := &PeerBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, bl)
}

func BenchmarkCapnprotoPeerBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &PeerBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, false)
}

func BenchmarkJsonPeerBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &PeerBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, false)
}

func BenchmarkCapnprotoPeerBlocksUnmarshalValidate(b *testing.B) {
	bl := &PeerBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkJsonPeerBlocksUnmarshalValidate(b *testing.B) {
	bl := &PeerBlockBody{}
	cmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkCapnprotoStateTxBlocksUnmarshalValidate(b *testing.B) {
	bl := &StateBlockBody{}
	jmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, jmr, bl, true)
}

func BenchmarkJsonStateTxBlocksUnmarshalValidate(b *testing.B) {
	bl := &StateBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, true)
}

func BenchmarkCapnprotoTxBlocksMarshal(b *testing.B) {
	bl := &TxBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, bl)
}

func BenchmarkJsonTxBlocksMarshal(b *testing.B) {
	bl := &TxBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, bl)
}

func BenchmarkCapnprotoTxBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &TxBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, false)
}

func BenchmarkJsonTxBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &TxBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, false)
}

func BenchmarkCapnprotoTxBlocksUnmarshalValidate(b *testing.B) {
	bl := &TxBlockBody{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkJsonTxBlocksUnmarshalValidate(b *testing.B) {
	bl := &TxBlockBody{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, true)
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

// GenerateDummyArray is used to generate an array of PeerBlockBody with dummy data
func (peerBlock *PeerBlockBody) GenerateDummyArray() []data.CapnpHelper {
	peerBlocks := make([]data.CapnpHelper, 0, 1000)

	changes := make([]block.PeerChange, 0, 1000)

	for i := 0; i < 1000; i++ {
		changes = append(changes, block.PeerChange{
			PubKey:      []byte(RandomStr(32)),
			ShardIdDest: uint32(rand.Intn(100)),
		})
	}

	for i := 0; i < 1000; i++ {
		peerBlocks = append(peerBlocks, &PeerBlockBody{
			PeerBlockBody: block.PeerBlockBody{
				StateBlockBody: block.StateBlockBody{
					RootHash: []byte(RandomStr(32)),
					ShardID: uint32(rand.Intn(1000)),
				},
				Changes: changes,
			},
		})
	}

	return peerBlocks
}

// GenerateDummyArray is used to generate an array of StateBlockBody with dummy data
func (sBlock *StateBlockBody) GenerateDummyArray() []data.CapnpHelper {
	sBlocks := make([]data.CapnpHelper, 0, 1000)

	for i := 0; i < 1000; i++ {
		sBlocks = append(sBlocks, &StateBlockBody{
			StateBlockBody: block.StateBlockBody{
				RootHash: []byte(RandomStr(32)),
				ShardID: uint32(rand.Intn(1000)),
			},
		})
	}

	return sBlocks
}

// GenerateDummyArray is used to generate an array of TxBlockBody with dummy data
func (txBlk *TxBlockBody) GenerateDummyArray() []data.CapnpHelper {
	blocks := make([]data.CapnpHelper, 0, 100)
	for i := 0; i < 100; i++ {
		lenMini := rand.Intn(20) + 1
		miniblocks := make([]block.MiniBlock, 0, lenMini)
		for j := 0; j < lenMini; j++ {
			lenTxHashes := rand.Intn(20) + 1
			txHashes := make([][]byte, 0, lenTxHashes)
			for k := 0; k < lenTxHashes; k++ {
				txHashes = append(txHashes, []byte(RandomStr(32)))
			}
			miniblock := block.MiniBlock{
				ShardID:  uint32(rand.Intn(20)),
				TxHashes: txHashes,
			}

			miniblocks = append(miniblocks, miniblock)
		}
		bl := TxBlockBody{
			TxBlockBody: block.TxBlockBody{
				StateBlockBody: block.StateBlockBody{
					RootHash: []byte(RandomStr(32)),
				},
				MiniBlocks: miniblocks,
			},
		}
		blocks = append(blocks, &bl)
	}

	return blocks
}

// GenerateDummyArray is used to generate an array of block headers with dummy data
func (h *Header) GenerateDummyArray() []data.CapnpHelper {
	headers := make([]data.CapnpHelper, 0, 1000)

	for i := 0; i < 1000; i++ {
		headers = append(headers, &Header{
			Header: block.Header{
				Nonce:         uint64(rand.Int63n(10000)),
				PrevHash:      []byte(RandomStr(32)),
				ShardId:       uint32(rand.Intn(20)),
				TimeStamp:     uint64(rand.Intn(20)),
				Round:         uint32(rand.Intn(20000)),
				Epoch:		   uint32(rand.Intn(20000)),
				BlockBodyHash: []byte(RandomStr(32)),
				BlockBodyType: block.TxBlock,
				Signature:     []byte(RandomStr(32)),
				Commitment:    []byte(RandomStr(32)),
				PubKeysBitmap: []byte(RandomStr(10)),
			},
		})
	}

	return headers
}

// GenerateDummyArray is used to generate an array of transactions with dummy data
func (tx *Transaction) GenerateDummyArray() []data.CapnpHelper {
	transactions := make([]data.CapnpHelper, 0, 1000)

	val := big.NewInt(0)
	val.GobDecode([]byte(RandomStr(32)))

	for i := 0; i < 1000; i++ {
		transactions = append(transactions, &Transaction{
			Transaction: transaction.Transaction{
				Nonce:     uint64(rand.Int63n(10000)),
				Value:     *val,
				RcvAddr:   []byte(RandomStr(32)),
				SndAddr:   []byte(RandomStr(32)),
				GasPrice:  uint64(rand.Int63n(10000)),
				GasLimit:  uint64(rand.Int63n(10000)),
				Data:      []byte(RandomStr(32)),
				Signature: []byte(RandomStr(32)),
				Challenge: []byte(RandomStr(32)),
			},
		})
	}

	return transactions
}
