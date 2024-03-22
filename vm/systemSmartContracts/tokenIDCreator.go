package systemSmartContracts

import (
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/vm"
)

type tokenIDCreator struct {
	eei    vm.SystemEI
	hasher hashing.Hasher
}

func NewTokenIDCreator(eei vm.SystemEI, hasher hashing.Hasher) (*tokenIDCreator, error) {
	return &tokenIDCreator{
		eei:    eei,
		hasher: hasher,
	}, nil
}

func (tic *tokenIDCreator) CreateNewTokenIdentifier(caller []byte, ticker []byte) ([]byte, error) {
	newRandomBase := append(caller, tic.eei.BlockChainHook().CurrentRandomSeed()...)
	newRandom := tic.hasher.Compute(string(newRandomBase))
	newRandomForTicker := newRandom[:tickerRandomSequenceLength]

	tickerPrefix := append(ticker, []byte(tickerSeparator)...)
	newRandomAsBigInt := big.NewInt(0).SetBytes(newRandomForTicker)

	one := big.NewInt(1)
	for i := 0; i < numOfRetriesForIdentifier; i++ {
		encoded := fmt.Sprintf("%06x", newRandomAsBigInt)
		newIdentifier := append(tickerPrefix, encoded...)
		buff := tic.eei.GetStorage(newIdentifier)
		if len(buff) == 0 {
			return newIdentifier, nil
		}
		newRandomAsBigInt.Add(newRandomAsBigInt, one)
	}

	return nil, vm.ErrCouldNotCreateNewTokenIdentifier
}

func (tic *tokenIDCreator) IsInterfaceNil() bool {
	return tic == nil
}
