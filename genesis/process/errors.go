package process

import "errors"

var errNilGenesisBlockCreator = errors.New("nil genesis block creator provided")

var errCouldNotGenerateInitialESDTTransfers = errors.New("could not generate initial esdt transfers")
