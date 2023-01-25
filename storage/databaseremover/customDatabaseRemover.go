package databaseremover

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	patternItemsSeparator = ","
	divisibleByChar       = "%"
)

type customDatabaseRemover struct {
	divisibleEpochs []uint32
}

// NewCustomDatabaseRemover returns a new instance of customDatabaseRemover
func NewCustomDatabaseRemover(cfg config.StoragePruningConfig) (*customDatabaseRemover, error) {
	if len(cfg.AccountsTrieSkipRemovalCustomPattern) == 0 {
		return &customDatabaseRemover{
			divisibleEpochs: make([]uint32, 0),
		}, nil
	}

	divisibleEpochs, err := parseCustomPattern(cfg.AccountsTrieSkipRemovalCustomPattern)
	if err != nil {
		return nil, fmt.Errorf("%w while creating a custom database remover", err)
	}

	return &customDatabaseRemover{
		divisibleEpochs: divisibleEpochs,
	}, nil
}

// ShouldRemove returns the result of the inner handler function
func (c *customDatabaseRemover) ShouldRemove(_ string, epoch uint32) bool {
	return c.shouldRemoveConsideringDivisibleEpochs(epoch)
}

func (c *customDatabaseRemover) shouldRemoveConsideringDivisibleEpochs(epoch uint32) bool {
	for _, divisibleEpoch := range c.divisibleEpochs {
		if epoch%divisibleEpoch == 0 {
			return false
		}
	}

	return true
}

func parseCustomPattern(pattern string) ([]uint32, error) {
	epochs := make([]uint32, 0)
	splitArgs := strings.Split(pattern, patternItemsSeparator)
	for _, arg := range splitArgs {
		if len(arg) == 0 {
			return nil, errEmptyPatternArgument
		}

		epoch, err := extractEpochFromArg(arg)
		if err != nil {
			return nil, err
		}

		epochs = append(epochs, epoch)
	}

	return epochs, nil
}

func extractEpochFromArg(arg string) (uint32, error) {
	splitStr := strings.Split(arg, divisibleByChar)
	if len(splitStr) != 2 {
		return 0, errInvalidPatternArgument
	}

	epochBI, ok := big.NewInt(0).SetString(splitStr[1], 10)
	if !ok {
		return 0, errCannotDecodeEpochNumber
	}

	epochNum := uint32(epochBI.Uint64())

	if epochNum == 0 {
		return 0, errEpochCannotBeZero
	}

	return epochNum, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *customDatabaseRemover) IsInterfaceNil() bool {
	return c == nil
}
