package storing

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var log = logger.GetOrCreate("update/storing")

type hardforkStorer struct {
	keysStore   storage.Storer
	keyValue    storage.Storer
	marshalizer marshal.Marshalizer

	mut  sync.Mutex
	keys map[string][][]byte
}

func NewHardforkStorer(keys storage.Storer, keyValue storage.Storer, marshalizer marshal.Marshalizer) (*hardforkStorer, error) {
	if check.IfNil(keys) {
		return nil, fmt.Errorf("%w for keys", update.ErrNilStorage)
	}
	if check.IfNil(keyValue) {
		return nil, fmt.Errorf("%w for key-values", update.ErrNilStorage)
	}
	if check.IfNil(marshalizer) {
		return nil, update.ErrNilMarshalizer
	}

	return &hardforkStorer{
		keysStore:   keys,
		keyValue:    keyValue,
		marshalizer: marshalizer,
		keys:        make(map[string][][]byte),
	}, nil
}

func (hs *hardforkStorer) Write(identifier string, key []byte, value []byte) error {
	hs.mut.Lock()
	defer hs.mut.Unlock()

	hs.keys[identifier] = append(hs.keys[identifier], key)

	log.Trace("hardforkStorer.Write",
		"key", key,
		"value", value,
	)

	return hs.keyValue.Put(key, value)
}

func (hs *hardforkStorer) FinishedIdentifier(identifier string) error {
	hs.mut.Lock()
	defer hs.mut.Unlock()

	log.Trace("hardforkStorer.FinishedIdentifier", "identifier", identifier)

	vals := hs.keys[identifier]
	if len(vals) == 0 {
		return nil
	}

	b := &batch.Batch{
		Data: vals,
	}

	buff, err := hs.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	delete(hs.keys, identifier)

	return hs.keysStore.Put([]byte(identifier), buff)
}

func (hs *hardforkStorer) RangeKeys(handler func(identifier string, keys [][]byte)) {
	if handler == nil {
		return
	}

	chIterate := hs.keysStore.Iterate()
	for kv := range chIterate {
		b := &batch.Batch{}
		err := hs.marshalizer.Unmarshal(b, kv.Val())
		if err != nil {
			log.Warn("error reading identifiers",
				"key", string(kv.Key()),
				"error", err,
			)
			continue
		}

		handler(string(kv.Key()), b.Data)
	}
}

func (hs *hardforkStorer) Get(key []byte) ([]byte, error) {
	return hs.keyValue.Get(key)
}

func (hs *hardforkStorer) Close() error {
	err1 := hs.keysStore.Close()
	err2 := hs.keyValue.Close()

	if err1 != nil {
		return err1
	}

	return err2
}

// IsInterfaceNil returns true if there is no value under the interface
func (hs *hardforkStorer) IsInterfaceNil() bool {
	return hs == nil
}
