package bootstrapStorage

import (
	"errors"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// HighestRoundFromBootStorage is the key for the highest round that is saved in storage
const highestRoundFromBootStorage = "highestRoundFromBootStorage"

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilBootStorer signals that an operation has been attempted to or with a nil storer implementation
var ErrNilBootStorer = errors.New("nil boot storer")

//BootstrapHeaderInfo is struct used to store information about a header
type BootstrapHeaderInfo struct {
	ShardId uint32
	Nonce   uint64
	Hash    []byte
}

// BootstrapData is struct used to store information that are needed for bootstrap
type BootstrapData struct {
	HeaderInfo           BootstrapHeaderInfo
	LastNotarizedHeaders []BootstrapHeaderInfo
	LastFinals           []BootstrapHeaderInfo
	ProcessedMiniBlocks  map[string]map[string]struct{}
	HighestFinalNonce    uint64
	LastRound            int64
}

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
	bootData.LastRound = bs.lastRound

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
	roundBytes, err := bs.marshalizer.Marshal(&round)
	if err != nil {
		return err
	}

	err = bs.store.Put([]byte(highestRoundFromBootStorage), roundBytes)
	if err != nil {
		return err
	}

	bs.lastRound = round

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

// GetHighestRound will return highest round saved in storage
func (bs *bootstrapStorer) GetHighestRound() int64 {
	roundBytes, err := bs.store.Get([]byte(highestRoundFromBootStorage))
	if err != nil {
		return 0
	}

	var round int64
	err = bs.marshalizer.Unmarshal(&round, roundBytes)
	if err != nil {
		return 0
	}

	return round
}

// SaveLastRound will save the last round
func (bs *bootstrapStorer) SaveLastRound(round int64) error {
	bs.lastRound = round

	// save round with a static key
	roundBytes, err := bs.marshalizer.Marshal(&round)
	if err != nil {
		return err
	}

	err = bs.store.Put([]byte(highestRoundFromBootStorage), roundBytes)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bs *bootstrapStorer) IsInterfaceNil() bool {
	if bs == nil {
		return true
	}
	return false
}
