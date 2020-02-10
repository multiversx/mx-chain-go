package fuzzer

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	fuzz "github.com/google/gofuzz"
)

type Fuzzer struct {
	// holds the first build error encounter (all builder methods should giveup if this is not nil)
	err error

	// the number of distinct accounts to be used
	numAccounts int

	// then actual address book
	addressBook [][]byte

	// transaction data bounds
	minTxVal *big.Int
	maxTxVal *big.Int

	// data size bounds
	minDataSz     int
	maxDataSz     int
	dataNilChance float64
}

func New() *Fuzzer {
	return &Fuzzer{
		err:           nil,
		numAccounts:   10,
		addressBook:   nil,
		minTxVal:      big.NewInt(0),
		maxTxVal:      big.NewInt(math.MaxInt64),
		minDataSz:     0,
		maxDataSz:     (1 << 20),
		dataNilChance: .05,
	}
}

func (f *Fuzzer) GetError() error {
	return f.err
}

func (f *Fuzzer) SetNumAcc(numAcc int) *Fuzzer {
	if f.err != nil {
		return f
	}
	if numAcc < 2 {
		f.err = fmt.Errorf("The number of accounts (%d) should be at least 2", numAcc)
	}
	f.addressBook = nil
	f.numAccounts = numAcc
	return f
}

func (f *Fuzzer) SetAddrBook(addrBook [][]byte) *Fuzzer {
	if f.err != nil {
		return f
	}
	if len(addrBook) < 2 {
		f.err = fmt.Errorf("The number of accounts (%d) should be at least 2", len(addrBook))
	}
	f.addressBook = addrBook
	f.numAccounts = len(addrBook)
	return f
}

func (f *Fuzzer) SetValueBounds(min, max *big.Int) *Fuzzer {
	if f.err != nil {
		return f
	}

	if min.Cmp(max) > 1 {
		f.err = fmt.Errorf("Value min (%s) should be less that max (%s)", min, max)
	}

	f.minTxVal = min
	f.maxTxVal = max

	return f
}

func (f *Fuzzer) SetDataSzBounds(min, max int) *Fuzzer {
	if f.err != nil {
		return f
	}

	if min > max {
		f.err = fmt.Errorf("Data size min (%d) should be less that max (%d)", min, max)
	}

	if min < 0 {
		f.err = fmt.Errorf("Data size min (%d) should be positive", min)
	}

	f.minDataSz = min
	f.maxDataSz = max

	return f
}

func (f *Fuzzer) GetTransactions(count uint64) ([]*transaction.Transaction, error) {

	gf := fuzz.New()
	if f.err != nil {
		return nil, f.err
	}

	if f.numAccounts != len(f.addressBook) {
		f.addressBook = make([][]byte, f.numAccounts)
		gf.NumElements(32, 32)
		gf.NilChance(0)
		for i := range f.addressBook {
			f.addressBook[i] = make([]byte, 32)
			gf.Fuzz(&f.addressBook[i])
		}
	}

	ret := make([]*transaction.Transaction, count)

	gf.NumElements(f.minDataSz, f.maxDataSz)
	gf.NilChance(f.dataNilChance)
	for i := range ret {
		tmp := 0
		gf.Fuzz(&tmp)

		if tmp < 0 {
			tmp *= -1
		}
		recvIdx := tmp % f.numAccounts
		tmp /= f.numAccounts
		sndIdx := tmp % f.numAccounts
		tmp /= f.numAccounts

		t := &transaction.Transaction{
			Nonce:     uint64(tmp),
			Value:     &big.Int{}, // set in a second pass
			RcvAddr:   f.addressBook[recvIdx],
			SndAddr:   f.addressBook[sndIdx],
			Signature: nil, // keep nil for now
		}
		gf.Fuzz(&t.GasPrice)
		gf.Fuzz(&t.GasLimit)
		gf.Fuzz(&t.Data)
		ret[i] = t
	}

	delta := new(big.Int).Sub(f.maxTxVal, f.minTxVal)
	// if we should work with more than 63 bits
	if delta.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		// to keep it simple we generate a number of bits at least min value need and as most as max
		minBytes := len(f.minTxVal.Bytes())
		maxBytes := len(f.maxTxVal.Bytes())

		gf.NumElements(minBytes, maxBytes)
		gf.NilChance(0)

		for i := range ret {
			var tmp []byte
			gf.Fuzz(&tmp)
			ti := new(big.Int).SetBytes(tmp)
			if ti.Cmp(f.maxTxVal) > 0 {
				ti.Set(f.maxTxVal)
			} else if ti.Cmp(f.minTxVal) < 0 {
				ti.Set(f.minTxVal)
			}
			ret[i].Value = ti
		}

	} else {
		idelta := delta.Int64()
		for i := range ret {
			ret[i].Value = new(big.Int).Set(f.minTxVal)
			if idelta > 0 {
				var t int64
				gf.Fuzz(&t)
				if t < 0 {
					t *= -1
				}
				ret[i].Value.Add(ret[i].Value, big.NewInt(t%idelta))
			}
		}
	}
	return ret, nil
}
