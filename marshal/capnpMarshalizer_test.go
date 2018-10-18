package marshal_test

import (
	"testing"

	"reflect"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func benchMarshal(b *testing.B, m marshal.Marshalizer, obj data.DataGenerator) {
	b.StopTimer()

	dArray := obj.GenerateDummyArray()
	l := len(dArray)
	b.ReportAllocs()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		m.Marshal(dArray[i%l])
	}
}

func benchUnmarshal(b *testing.B, m marshal.Marshalizer, obj data.DataGenerator, validate bool) {
	b.StopTimer()
	dArray := obj.GenerateDummyArray()
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

// benchmarks
func BenchmarkCapnprotoTransactionMarshal(b *testing.B) {
	tx := &transaction.Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, tx)
}

func BenchmarkJsonTransactionMarshal(b *testing.B) {
	tx := &transaction.Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, tx)
}

func BenchmarkCapnprotoTransactionUnmarshalNoValidate(b *testing.B) {
	tx := &transaction.Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, tx, false)
}

func BenchmarkJsonTransactionUnmarshalNoValidate(b *testing.B) {
	tx := &transaction.Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, tx, false)
}

func BenchmarkCapnprotoTransactionUnmarshalValidate(b *testing.B) {
	tx := &transaction.Transaction{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, tx, true)
}

func BenchmarkJsonTransactionUnmarshalValidate(b *testing.B) {
	tx := &transaction.Transaction{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, tx, true)
}

func BenchmarkCapnprotoBlocksMarshal(b *testing.B) {
	bl := &block.Block{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, bl)
}

func BenchmarkJsonBlocksMarshal(b *testing.B) {
	bl := &block.Block{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, bl)
}

func BenchmarkCapnprotoBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &block.Block{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, false)
}

func BenchmarkJsonBlocksUnmarshalNoValidate(b *testing.B) {
	bl := &block.Block{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, false)
}

func BenchmarkCaonprotoBlocksUnmarshalValidate(b *testing.B) {
	bl := &block.Block{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, bl, true)
}

func BenchmarkJsonBlocksUnmarshalValidate(b *testing.B) {
	bl := &block.Block{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, bl, true)
}

func BenchmarkCapnprotoHeaderMarshal(b *testing.B) {
	h := &block.Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchMarshal(b, cmr, h)
}

func BenchmarkJsonHeaderMarshal(b *testing.B) {
	h := &block.Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchMarshal(b, jmr, h)
}

func BenchmarkCapnprotoHeaderUnmarshalNoValidate(b *testing.B) {
	h := &block.Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, h, false)
}

func BenchmarkJsonHeaderUnmarshalNoValidate(b *testing.B) {
	h := &block.Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, h, false)
}

func BenchmarkCapnprotoHeaderUnmarshalValidate(b *testing.B) {
	h := &block.Header{}
	cmr := &marshal.CapnpMarshalizer{}
	benchUnmarshal(b, cmr, h, true)
}

func BenchmarkJsonHeaderUnmarshalValidate(b *testing.B) {
	h := &block.Header{}
	jmr := &marshal.JsonMarshalizer{}
	benchUnmarshal(b, jmr, h, true)
}
