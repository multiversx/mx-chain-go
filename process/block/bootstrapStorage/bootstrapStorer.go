//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. bootstrapData.proto
package bootstrapStorage

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/frozen"
	"github.com/multiversx/mx-chain-go/storage"
)

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilBootStorer signals that an operation has been attempted to or with a nil storer implementation
var ErrNilBootStorer = errors.New("nil boot storer")

type bootstrapStorer struct {
	store       storage.Storer
	marshalizer marshal.Marshalizer
	lastRound   int64
}

// NewBootstrapStorer will return an instance of bootstrap storer
func NewBootstrapStorer(
	marshalizer marshal.Marshalizer,
	store storage.Storer,
) (*bootstrapStorer, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(store) {
		return nil, ErrNilBootStorer
	}

	bootStorer := &bootstrapStorer{
		store:       store,
		marshalizer: marshalizer,
	}
	bootStorer.lastRound = bootStorer.GetHighestRound()

	return bootStorer, nil
}

// Put will save bootData in storage
func (bs *bootstrapStorer) Put(round int64, bootData BootstrapData) error {
	bootData.LastRound = atomic.LoadInt64(&bs.lastRound)

	// save bootstrap round information
	bootDataBytes, err := bs.marshalizer.Marshal(&bootData)
	if err != nil {
		return err
	}

	key := []byte(strconv.FormatInt(round, 10))
	err = bs.store.Put(key, bootDataBytes)
	if err != nil {
		return err
	}

	// save round with a static key
	roundBytes, err := bs.marshalizer.Marshal(&RoundNum{Num: round})
	if err != nil {
		return err
	}

	err = bs.store.Put([]byte(common.HighestRoundFromBootStorage), roundBytes)
	if err != nil {
		return err
	}

	atomic.StoreInt64(&bs.lastRound, round)

	return nil
}

// Get will read data from storage
func (bs *bootstrapStorer) Get(round int64) (BootstrapData, error) {
	key := []byte(strconv.FormatInt(round, 10))
	bootstrapDataBytes, err := bs.store.Get(key)
	if err != nil {
		return BootstrapData{}, err
	}

	var bootData BootstrapData
	err = bs.marshalizer.Unmarshal(&bootData, bootstrapDataBytes)
	if err != nil {
		return BootstrapData{}, err
	}

	return bootData, nil
}

// GetHighestRound will return the highest round saved in storage
func (bs *bootstrapStorer) GetHighestRound() int64 {
	if frozen.ShouldOverrideHighestRound {
		return frozen.HighestRound
	}

	roundBytes, err := bs.store.Get([]byte(common.HighestRoundFromBootStorage))
	if err != nil {
		return 0
	}

	var round RoundNum
	err = bs.marshalizer.Unmarshal(&round, roundBytes)
	if err != nil {
		return 0
	}

	return round.Num
}

// SaveLastRound will save the last round
func (bs *bootstrapStorer) SaveLastRound(round int64) error {
	atomic.StoreInt64(&bs.lastRound, round)

	// save round with a static key
	roundBytes, err := bs.marshalizer.Marshal(&RoundNum{Num: round})
	if err != nil {
		return err
	}

	err = bs.store.Put([]byte(common.HighestRoundFromBootStorage), roundBytes)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bs *bootstrapStorer) IsInterfaceNil() bool {
	return bs == nil
}
