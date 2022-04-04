package helpers

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
)

const (
	patternItemsSeparator = ","
	divisibleByChar       = "%"
)

type customDatabaseRemover struct {
	shouldRemoveHandler func(dbIdentifier string, epoch uint32) bool
}

// NewCustomDatabaseRemover returns a new instance of customDatabaseRemover
func NewCustomDatabaseRemover(cfg config.StoragePruningConfig) (*customDatabaseRemover, error) {
	if len(cfg.AccountsTrieSkipRemovalCustomPattern) == 0 {
		return &customDatabaseRemover{
			shouldRemoveHandler: func(_ string, _ uint32) bool {
				return true
			},
		}, nil
	}

	shouldRemoveFunc, err := parseCustomPattern(cfg.AccountsTrieSkipRemovalCustomPattern)
	if err != nil {
		return nil, fmt.Errorf("%w while creating a custom database remover", err)
	}

	return &customDatabaseRemover{
		shouldRemoveHandler: shouldRemoveFunc,
	}, nil
}

// ShouldRemove returns the result of the inner handler function
func (c *customDatabaseRemover) ShouldRemove(dbIdentifier string, epoch uint32) bool {
	return c.shouldRemoveHandler(dbIdentifier, epoch)
}

func parseCustomPattern(pattern string) (func(dbIdentifier string, epoch uint32) bool, error) {
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

	return func(_ string, epoch uint32) bool {
		for _, divisibleEpoch := range epochs {
			if epoch%divisibleEpoch == 0 {
				return false
			}
		}

		return true
	}, nil
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
