package smartContract

import (
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/process"
)

type argumentParser struct {
	callParser    process.CallArgumentsParser
	deployParser  process.DeployArgumentsParser
	storageParser process.StorageArgumentsParser
}

// NewArgumentParser creates a full argument parser component
func NewArgumentParser() *argumentParser {
	callArgsParser := parsers.NewCallArgsParser()
	deployArgsParser := parsers.NewDeployArgsParser()
	storageArgsParser := parsers.NewStorageUpdatesParser()

	a := &argumentParser{
		callParser:    callArgsParser,
		deployParser:  deployArgsParser,
		storageParser: storageArgsParser,
	}

	return a
}

// ParseCallData returns parsed data for contract calls
func (a *argumentParser) ParseCallData(data string) (string, [][]byte, error) {
	return a.callParser.ParseData(data)
}

// ParseDeployData returns parsed data for deploy data
func (a *argumentParser) ParseDeployData(data string) (*parsers.DeployArgs, error) {
	return a.deployParser.ParseData(data)
}

// CreateDataFromStorageUpdate creates contract call data from storage update
func (a *argumentParser) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string {
	return a.storageParser.CreateDataFromStorageUpdate(storageUpdates)
}

// GetStorageUpdates returns storage updates from contract call data
func (a *argumentParser) GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error) {
	return a.storageParser.GetStorageUpdates(data)
}

// IsInterfaceNil return true if underlying object is nil
func (a *argumentParser) IsInterfaceNil() bool {
	return a == nil
}
