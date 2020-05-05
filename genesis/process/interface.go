package process

import "github.com/ElrondNetwork/elrond-go/genesis"

type deployProcessor interface {
	Deploy(sc genesis.InitialSmartContractHandler) error
}
