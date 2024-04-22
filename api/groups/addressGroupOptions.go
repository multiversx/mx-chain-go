package groups

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/data/api"
	customErrors "github.com/multiversx/mx-chain-go/api/errors"
)

func extractAccountQueryOptions(c *gin.Context) (api.AccountQueryOptions, error) {
	options, err := parseAccountQueryOptions(c)
	if err != nil {
		return api.AccountQueryOptions{}, fmt.Errorf("%w: %v", customErrors.ErrBadUrlParams, err)
	}

	err = checkAccountQueryOptions(options)
	if err != nil {
		return api.AccountQueryOptions{}, fmt.Errorf("%w: %v", customErrors.ErrBadUrlParams, err)
	}

	return options, nil
}

func parseAccountQueryOptions(c *gin.Context) (api.AccountQueryOptions, error) {
	onFinalBlock, err := parseBoolUrlParam(c, urlParamOnFinalBlock)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	onStartOfEpoch, err := parseUint32UrlParam(c, urlParamOnStartOfEpoch)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	blockNonce, err := parseUint64UrlParam(c, urlParamBlockNonce)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	blockHash, err := parseHexBytesUrlParam(c, urlParamBlockHash)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	blockRootHash, err := parseHexBytesUrlParam(c, urlParamBlockRootHash)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	hintEpoch, err := parseUint32UrlParam(c, urlParamHintEpoch)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	options := api.AccountQueryOptions{
		OnFinalBlock:   onFinalBlock,
		OnStartOfEpoch: onStartOfEpoch,
		BlockNonce:     blockNonce,
		BlockHash:      blockHash,
		BlockRootHash:  blockRootHash,
		HintEpoch:      hintEpoch,
		WithKeys:       false,
	}
	return options, nil
}

func checkAccountQueryOptions(options api.AccountQueryOptions) error {
	numSpecifiedBlockCoordinates := 0

	if options.BlockNonce.HasValue {
		numSpecifiedBlockCoordinates++
	}
	if len(options.BlockHash) > 0 {
		numSpecifiedBlockCoordinates++
	}
	if len(options.BlockRootHash) > 0 {
		numSpecifiedBlockCoordinates++
	}

	if numSpecifiedBlockCoordinates > 1 {
		return errors.New("only one block coordinate (blockNonce vs. blockHash vs. blockRootHash) can be specified at a time")
	}
	if options.OnFinalBlock && numSpecifiedBlockCoordinates > 0 {
		return errors.New("onFinalBlock is not compatible with any other block coordinates")
	}
	if options.OnStartOfEpoch.HasValue && numSpecifiedBlockCoordinates > 0 {
		return errors.New("onStartOfEpoch is not compatible with any other block coordinates")
	}
	if options.HintEpoch.HasValue && len(options.BlockRootHash) == 0 {
		return errors.New("hintEpoch is optional, but only compatible with blockRootHash")
	}

	return nil
}
