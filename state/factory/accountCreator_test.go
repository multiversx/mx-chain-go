package factory_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/stretchr/testify/assert"
)

func TestAccountCreator_CreateAccountNilAddress(t *testing.T) {
	t.Parallel()

	accF := factory.NewAccountCreator()

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)
	assert.False(t, check.IfNil(accF))

	acc, err := accF.CreateAccount(nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

var decimalPlaces = "000000000000000000"
func TestBigInt_test(t *testing.T){
	sum, _ := big.NewInt(0).SetString("20000000" + decimalPlaces, 10)
	fmt.Println("len", len(sum.Bytes()))
	fmt.Printf("%s\n", hex.EncodeToString(sum.Bytes()))
	fmt.Printf("%v\n", sum.Bytes())
}

func TestAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF := factory.NewAccountCreator()

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(make([]byte, 32))

	assert.Nil(t, err)
	assert.False(t, check.IfNil(acc))
}
