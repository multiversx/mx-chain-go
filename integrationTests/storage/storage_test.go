package storage

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &marshal.JsonMarshalizer{}
var testHasher = sha256.Sha256{}

var rcvAddr = make([]byte, 32)
var sndAddr = make([]byte, 32)
var sig = make([]byte, 64)

func createStoredData(nonce uint64) ([]byte, []byte) {
	tx := &transaction.Transaction{
		Nonce:     nonce,
		GasLimit:  0,
		GasPrice:  0,
		RcvAddr:   rcvAddr,
		SndAddr:   sndAddr,
		Signature: sig,
		Value:     big.NewInt(1),
	}

	txBuff, _ := testMarshalizer.Marshal(tx)
	txHash := testHasher.Compute(string(txBuff))

	return txHash, txBuff
}

func createStorage() storage.Storer {
	db, _ := leveldb.NewDB("Transactions", 10, 30000)
	cacher, _ := lrucache.NewCache(100000)
	store, _ := storageUnit.NewStorageUnit(
		cacher,
		db,
	)

	return store
}

func TestWriteContinously(t *testing.T) {
	t.Skip("infinite test")
	rand.Reader.Read(rcvAddr)
	rand.Reader.Read(sndAddr)
	rand.Reader.Read(sig)

	store := createStorage()
	defer func() {
		store.DestroyUnit()
	}()

	startTime := time.Now()
	for i := 0; ; i++ {
		written := 10000
		if i%written == 0 {
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			fmt.Printf("Written %d, total %d in %f s\n", written, i, diff.Seconds())
			startTime = time.Now()
		}

		key, val := createStoredData(uint64(i))

		err := store.Put(key, val)
		assert.Nil(t, err)
	}
}

func TestWriteReadAnDeleteContinously(t *testing.T) {
	t.Skip("infinite test")
	chWriteStop := make(chan struct{}, 1)
	chRemoveStop := make(chan struct{}, 1)
	chReadStop := make(chan struct{}, 1)
	chErrors := make(chan error, 2)

	rand.Reader.Read(rcvAddr)
	rand.Reader.Read(sndAddr)
	rand.Reader.Read(sig)

	store := createStorage()
	defer func() {
		store.DestroyUnit()
	}()

	var maxWritten = uint64(0)

	//write anonymous func
	go func() {
		written := 10000
		initTime := time.Now()
		startTime := time.Now()
		for counter := 0; ; counter++ {
			select {
			case <-chWriteStop:
				return
			default:
			}

			if counter%written == 0 {
				endTime := time.Now()
				diff := endTime.Sub(startTime)
				cumul := endTime.Sub(initTime)
				fmt.Printf("Written %d, total %d in %f s\nCumulativeTime %f\n", written, counter, diff.Seconds(), cumul.Seconds())
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

				fmt.Printf("timestamp: %d, num go: %d, go mem: %s, sys mem: %s, total mem: %s, num GC: %d\n",
					time.Now().Unix(),
					runtime.NumGoroutine(),
					core.ConvertBytes(memStats.Alloc),
					core.ConvertBytes(memStats.Sys),
					core.ConvertBytes(memStats.TotalAlloc),
					memStats.NumGC,
				)
				startTime = time.Now()
			}

			key, val := createStoredData(uint64(counter))

			errPut := store.Put(key, val)
			if errPut != nil {
				chErrors <- errPut
				return
			}

			atomic.StoreUint64(&maxWritten, uint64(counter))
		}
	}()

	//read anonymous func
	go func() {
		maxRoutines := make(chan struct{}, 5000)

		for {
			select {
			case <-chReadStop:
				return
			case <-time.After(time.Microsecond * 10):
				//reads happen less often than writes
			}
			maxRoutines <- struct{}{}
			if atomic.LoadUint64(&maxWritten) == 0 {
				//not written yet
				continue
			}

			go func() {
				defer func() {
					<-maxRoutines
				}()
				randOp, _ := rand.Int(rand.Reader, big.NewInt(10))

				var errGet error
				if randOp.Cmp(big.NewInt(7)) < 0 {
					//read an existing value
					maxWrittenBigInt := big.NewInt(0).SetUint64(atomic.LoadUint64(&maxWritten))
					existingNonce, _ := rand.Int(rand.Reader, maxWrittenBigInt)

					key, _ := createStoredData(existingNonce.Uint64())
					_, errGet = store.Get(key)
					if errGet == storage.ErrKeyNotFound {
						//fmt.Printf("Not getting tx with nonce %d, maxwritten: %d\n", existingNonce.Uint64(),
						//	maxWrittenBigInt.Uint64())
						errGet = nil
					}

				} else {
					//read a not existing hash
					_, errGet = store.Get(make([]byte, 32))
					if errGet == storage.ErrKeyNotFound {
						errGet = nil
					}
				}

				if errGet != nil {
					chErrors <- errGet
					return
				}
			}()
		}
	}()

	go func() {
		for {
			select {
			case <-chRemoveStop:
				return
			case <-time.After(time.Millisecond * 100):
				//remove happen less often than writes
			}

			if atomic.LoadUint64(&maxWritten) == 0 {
				//not written yet
				continue
			}

			maxWrittenBigInt := big.NewInt(0).SetUint64(atomic.LoadUint64(&maxWritten))
			existingNonce, _ := rand.Int(rand.Reader, maxWrittenBigInt)
			key, _ := createStoredData(existingNonce.Uint64())
			errRemove := store.Remove(key)

			if errRemove != nil {
				chErrors <- errRemove
				return
			}

		}
	}()

	select {
	case err := <-chErrors:
		assert.Nil(t, err)
	}

	chWriteStop <- struct{}{}
	chReadStop <- struct{}{}
	chRemoveStop <- struct{}{}
}
