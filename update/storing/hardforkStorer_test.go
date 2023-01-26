package storing

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/assert"
)

func createDefaultArg() ArgHardforkStorer {
	return ArgHardforkStorer{
		KeysStore:   genericMocks.NewStorerMock(),
		KeyValue:    genericMocks.NewStorerMock(),
		Marshalizer: &mock.MarshalizerMock{},
	}
}

func TestNewHardforkStorer_NilKeysStorerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	arg.KeysStore = nil
	hs, err := NewHardforkStorer(arg)

	assert.True(t, check.IfNil(hs))
	assert.True(t, errors.Is(err, update.ErrNilStorage))
}

func TestNewHardforkStorer_NilKeyValStorerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	arg.KeyValue = nil
	hs, err := NewHardforkStorer(arg)

	assert.True(t, check.IfNil(hs))
	assert.True(t, errors.Is(err, update.ErrNilStorage))
}

func TestNewHardforkStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	arg.Marshalizer = nil
	hs, err := NewHardforkStorer(arg)

	assert.True(t, check.IfNil(hs))
	assert.True(t, errors.Is(err, update.ErrNilMarshalizer))
}

func TestNewHardforkStorer_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	hs, err := NewHardforkStorer(arg)

	assert.False(t, check.IfNil(hs))
	assert.Nil(t, err)
}

func TestHardforkStorer_WriteReadTests(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	hs, _ := NewHardforkStorer(arg)

	expectedValues := map[string][]string{
		"trie@tr@0@8": {"trie_key_0", "trie_key_1, trie_key_2"},
		"trie@tr@0@10@fe35692992485b5e16ded6cd5d22134e0f243dac798212d894428dcbc44b44c2": {
			"727440373437323430333034303331333034303636363533333335333633393332333933393332333433383335363233353635333133363634363536343336363336343335363433323332333133333334363533303636333233343333363436313633333733393338333233313332363433383339333433343332333836343633363236333334333436323334333436333332",
			"747240304031304037323635363736393733373437323631373436393666366535663633366637333734",
		},
		"trie@tr@0@10@f68370f1ba68471f8b5d46405c33e08d24611d70bee92f7ce6e62d186d0495ee": {
			"727440373437323430333034303331333034303636333633383333333733303636333136323631333633383334333733313636333836323335363433343336333433303335363333333333363533303338363433323334333633313331363433373330363236353635333933323636333736333635333636353336333236343331333833363634333033343339333536353635",
			"747240304031304037323635363736393733373437323631373436393666366535663633366637333734",
		},
		"trie@tr@0@10@ed45a0c396936ae610731632f534fad98adf6b89319795556e6a69f677e4880a": {
			"727440373437323430333034303331333034303635363433343335363133303633333333393336333933333336363136353336333133303337333333313336333333323636333533333334363636313634333933383631363436363336363233383339333333313339333733393335333533353336363533363631333633393636333633373337363533343338333833303631",
			"747240304031304037323635363736393733373437323631373436393666366535663633366637333734",
		},
		"transactions": {
			"7478407277644030363935316635303832316135653033353566373130303164313236383534303731356462343735343864643062656136356464373231303734326162393134",
		},
		"miniBlocks": {
			"6d624061383230383939366335666236663161633235383534636630353133646138613831653262646336353434366262363736313232653238373266646238366339",
		},
		"metaBlock": {
			"6d65746140696e746567726174696f6e20746573747320636861696e2049444061376133653334396164373565356132663862326362363338616466313161633939623531386332306134363035613937653038346638623330313236326632",
		},
	}

	for identifier, keys := range expectedValues {
		for _, key := range keys {
			err := hs.Write(identifier, []byte(key), make([]byte, 0))
			assert.Nil(t, err)
		}

		err := hs.FinishedIdentifier(identifier)
		assert.Nil(t, err)
	}

	//load and check
	mutRecovered := sync.Mutex{}
	recovered := make(map[string][]string)
	hs.RangeKeys(func(identifier string, keys [][]byte) bool {
		mutRecovered.Lock()
		defer mutRecovered.Unlock()

		for _, key := range keys {
			recovered[identifier] = append(recovered[identifier], string(key))
		}

		return true
	})

	assert.Equal(t, expectedValues, recovered)
}

func TestHardforkStorer_Get(t *testing.T) {
	t.Parallel()

	getCalled := 0
	expectedResult := []byte("expected result")
	arg := createDefaultArg()
	providedKey := []byte("key")
	identifier := "identifier"
	arg.KeyValue = &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if bytes.Equal(key, append([]byte(identifier), providedKey...)) {
				getCalled++
				return expectedResult, nil
			}

			return nil, fmt.Errorf("unexpected error")
		},
	}
	hs, _ := NewHardforkStorer(arg)

	val, err := hs.Get(identifier, providedKey)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, val)
	assert.Equal(t, 1, getCalled)
}

func TestHardforkStorer_CloseKeysCloseErrors(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("error keys store")
	numCloseCalled := 0
	arg := createDefaultArg()
	arg.KeyValue = &storageStubs.StorerStub{
		CloseCalled: func() error {
			numCloseCalled++
			return errExpected
		},
	}
	arg.KeysStore = &storageStubs.StorerStub{
		CloseCalled: func() error {
			numCloseCalled++
			return nil
		},
	}
	hs, _ := NewHardforkStorer(arg)
	err := hs.Close()

	assert.Equal(t, errExpected, err)
	assert.Equal(t, 2, numCloseCalled)
}

func TestHardforkStorer_CloseKeyValueCloseErrors(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("error keys store")
	numCloseCalled := 0
	arg := createDefaultArg()
	arg.KeyValue = &storageStubs.StorerStub{
		CloseCalled: func() error {
			numCloseCalled++
			return nil
		},
	}
	arg.KeysStore = &storageStubs.StorerStub{
		CloseCalled: func() error {
			numCloseCalled++
			return errExpected
		},
	}
	hs, _ := NewHardforkStorer(arg)
	err := hs.Close()

	assert.Equal(t, errExpected, err)
	assert.Equal(t, 2, numCloseCalled)
}

func TestHardforkStorer_RangeKeysNilHandlerShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArg()
	rangeKeysCalled := false
	arg.KeysStore = &storageStubs.StorerStub{
		RangeKeysCalled: func(_ func(key []byte, val []byte) bool) {
			rangeKeysCalled = true
		},
	}
	hs, _ := NewHardforkStorer(arg)

	hs.RangeKeys(nil)

	assert.False(t, rangeKeysCalled)
}
