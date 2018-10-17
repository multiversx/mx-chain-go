package marshal_test

import (
	"testing"

	"reflect"

	"fmt"
	"math/rand"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func benchMarshalTransactions(b *testing.B, m marshal.Marshalizer) {
	b.StopTimer()

	dArray := GenerateDummyTxArray()
	l := len(dArray)
	b.ReportAllocs()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		m.Marshal(dArray[i%l])
	}

}

func benchMarshalHeaders(b *testing.B, m marshal.Marshalizer) {
	b.StopTimer()

	dArray := GenerateDummyHeaderArray()
	l := len(dArray)
	b.ReportAllocs()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		m.Marshal(dArray[i%l])
	}
}

func benchMarshalBlocks(b *testing.B, m marshal.Marshalizer) {
	b.StopTimer()

	dArray := GenerateDummyBlockArray()
	l := len(dArray)
	b.ReportAllocs()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		m.Marshal(dArray[i%l])
	}
}

func benchUnmarshalTransactions(b *testing.B, m marshal.Marshalizer, validate bool) {
	b.StopTimer()
	dArray := GenerateDummyTxArray()
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
		var tx transaction.Transaction
		err := m.Unmarshal(&tx, serialized[n])

		if err != nil {
			b.Fatalf("error unmarshalling %s : %s", serialized[n], err)
		}

		// Check unmarshalled data as expected
		if validate {
			orig := dArray[n]
			valid := reflect.DeepEqual(orig, &tx)
			if !valid {
				b.Fatalf("unmarshaled data different than expected: \n%v\n%v\n", orig, &tx)
			}
		}
	}
}

func benchUnmarshalHeaders(b *testing.B, m marshal.Marshalizer, validate bool) {
	b.StopTimer()
	dArray := GenerateDummyHeaderArray()
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
		var h block.Header
		err := m.Unmarshal(&h, serialized[n])

		if err != nil {
			b.Fatalf("error unmarshalling %s : %s", serialized[n], err)
		}

		// Check unmarshalled data as expected
		if validate {
			orig := dArray[n]
			valid := reflect.DeepEqual(orig, &h)
			if !valid {
				b.Fatalf("unmarshaled data different than expected: \n%v\n%v\n", orig, &h)
			}
		}
	}
}

func benchUnmarshalBlocks(b *testing.B, m marshal.Marshalizer, validate bool) {
	b.StopTimer()
	dArray := GenerateDummyBlockArray()
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
		var bl block.Block
		err := m.Unmarshal(&bl, serialized[n])

		if err != nil {
			b.Fatalf("error unmarshalling %s : %s", serialized[n], err)
		}

		// Check unmarshalled data as expected
		if validate {
			orig := dArray[n]
			valid := reflect.DeepEqual(orig, &bl)
			if !valid {
				b.Fatalf("unmarshaled data different than expected: \n%v\n%v\n", orig, &bl)
			}
		}
	}
}

// benchmarks
func BenchmarkCapnprotoTransactionMarshal(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshalTransactions(b, cmr)
}

func BenchmarkJsonTransactionMarshal(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchMarshalTransactions(b, jmr)
}

func BenchmarkCapnprotoTransactionUnmarshalNoValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalTransactions(b, cmr, false)
}

func BenchmarkJsonTransactionUnmarshalNoValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalTransactions(b, jmr, false)
}

func BenchmarkCapnprotoTransactionUnmarshalValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalTransactions(b, cmr, true)
}

func BenchmarkJsonTransactionUnmarshalValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalTransactions(b, jmr, true)
}

func BenchmarkCapnprotoBlocksMarshal(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshalBlocks(b, cmr)
}

func BenchmarkJsonBlocksMarshal(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchMarshalBlocks(b, jmr)
}

func BenchmarkCapnprotoBlocksUnmarshalNoValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalBlocks(b, cmr, false)
}

func BenchmarkJsonBlocksUnmarshalNoValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalBlocks(b, jmr, false)
}

func BenchmarkCaonprotoBlocksUnmarshalValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalBlocks(b, cmr, true)
}

func BenchmarkJsonBlocksUnmarshalValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalBlocks(b, jmr, true)
}

func BenchmarkCapnprotoHeaderMarshal(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshalHeaders(b, cmr)
}

func BenchmarkJsonHeaderMarshal(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchMarshalHeaders(b, jmr)
}

func BenchmarkCapnprotoHeaderUnmarshalNoValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalHeaders(b, cmr, false)
}

func BenchmarkJsonHeaderUnmarshalNoValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalHeaders(b, jmr, false)
}

func BenchmarkCapnprotoHeaderUnmarshalValidate(b *testing.B) {
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshalHeaders(b, cmr, true)
}

func BenchmarkJsonHeaderUnmarshalValidate(b *testing.B) {
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshalHeaders(b, jmr, true)
}

func GenerateDummyTxArray() []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0, 1000)

	for i := 0; i < 1000; i++ {
		transactions = append(transactions, &transaction.Transaction{
			Nonce:     []byte(randomStr(6)),
			Value:     []byte(randomStr(6)),
			RcvAddr:   []byte(randomStr(32)),
			SndAddr:   []byte(randomStr(32)),
			GasPrice:  []byte(randomStr(6)),
			GasLimit:  []byte(randomStr(6)),
			Data:      []byte(randomStr(32)),
			Signature: []byte(randomStr(32)),
			Challenge: []byte(randomStr(32)),
			PubKey:    []byte(randomStr(32)),
		})
	}

	return transactions
}

func GenerateDummyHeaderArray() []*block.Header {
	headers := make([]*block.Header, 0, 1000)

	pkList := make([][]byte, 0, 21)

	for i := 0; i < 21; i++ {
		pkList = append(pkList, []byte(randomStr(32)))
	}

	for i := 0; i < 1000; i++ {
		headers = append(headers, &block.Header{
			Nonce:      []byte(randomStr(4)),
			PrevHash:   []byte(randomStr(32)),
			ShardId:    uint32(rand.Intn(20)),
			TimeStamp:  []byte(randomStr(20)),
			Round:      uint32(rand.Intn(20000)),
			BlockHash:  []byte(randomStr(32)),
			Signature:  []byte(randomStr(32)),
			Commitment: []byte(randomStr(32)),
			PubKeys:    pkList,
		})
	}

	return headers
}

func GenerateDummyBlockArray() []*block.Block {
	blocks := make([]*block.Block, 0, 100)
	for i := 0; i < 100; i++ {
		lenMini := rand.Intn(20) + 1
		miniblocks := make([]block.MiniBlock, 0, lenMini)
		for j := 0; j < lenMini; j++ {
			lenTxHashes := rand.Intn(20) + 1
			txHashes := make([][]byte, 0, lenTxHashes)
			for k := 0; k < lenTxHashes; k++ {
				txHashes = append(txHashes, []byte(randomStr(32)))
			}
			miniblock := block.MiniBlock{
				DestShardID: uint32(rand.Intn(20)),
				TxHashes:    txHashes,
			}

			miniblocks = append(miniblocks, miniblock)
		}
		bl := block.Block{MiniBlocks: miniblocks}
		blocks = append(blocks, &bl)
	}

	return blocks
}

func randomStr(l int) string {
	buf := make([]byte, l)

	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}
