package preprocess

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountsTracker_NewAccountsTrackerShouldWork(t *testing.T) {
	at := newAccountsTracker()
	assert.NotNil(t, at)
}

func TestAccountsTracker_InitGetSetShouldWork(t *testing.T) {
	at := newAccountsTracker()

	address := []byte("X")
	accntInfo := accountInfo{
		nonce:   1,
		balance: big.NewInt(1),
	}

	at.setAccountInfo(address, accntInfo)

	ai, found := at.getAccountInfo(address)
	assert.True(t, found)
	assert.Equal(t, accntInfo, ai)

	at.init()

	ai, found = at.getAccountInfo(address)
	assert.False(t, found)
	assert.Equal(t, accountInfo{}, ai)
}
