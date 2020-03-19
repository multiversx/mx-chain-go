package rand_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand"
	"github.com/stretchr/testify/assert"
)

func TestNewSeedRandReader_NilSeedShouldErr(t *testing.T) {
	t.Parallel()

	srr, err := rand.NewSeedRandReader(nil)

	assert.Nil(t, srr)
	assert.Equal(t, p2p.ErrEmptySeed, err)
}

func TestNewSeedRandReader_ShouldWork(t *testing.T) {
	t.Parallel()

	seed := []byte("seed")
	srr, err := rand.NewSeedRandReader(seed)

	assert.NotNil(t, srr)
	assert.Nil(t, err)
}

func TestSeedRandReader_ReadNilBufferShouldErr(t *testing.T) {
	t.Parallel()

	seed := []byte("seed")
	srr, _ := rand.NewSeedRandReader(seed)

	n, err := srr.Read(nil)

	assert.Equal(t, 0, n)
	assert.Equal(t, err, p2p.ErrEmptyBuffer)
}

func TestSeedRandReader_ReadShouldWork(t *testing.T) {
	t.Parallel()

	seed := []byte("seed")
	srr, _ := rand.NewSeedRandReader(seed)

	testTbl := []struct {
		pSize int
		p     []byte
		n     int
		err   error
		name  string
	}{
		{pSize: 1, p: []byte{15}, n: 1, err: nil, name: "1 character"},
		{pSize: 2, p: []byte{15, 210}, n: 2, err: nil, name: "2 characters"},
		{pSize: 4, p: []byte{15, 210, 236, 97}, n: 4, err: nil, name: "4 characters"},
		{pSize: 5, p: []byte{15, 210, 236, 97, 112}, n: 5, err: nil, name: "5 characters"},
		{pSize: 8, p: []byte{15, 210, 236, 97, 112, 165, 91, 186}, n: 8, err: nil, name: "8 characters"},
	}

	for _, tc := range testTbl {
		t.Run(tc.name, func(t *testing.T) {
			p := make([]byte, tc.pSize)

			n, err := srr.Read(p)

			assert.Equal(t, tc.p, p)
			assert.Equal(t, tc.n, n)
			assert.Equal(t, tc.err, err)
		})
	}
}
