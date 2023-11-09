package groups

import (
	"encoding/hex"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
)

func parseBoolUrlParam(c *gin.Context, name string) (bool, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return false, nil
	}

	return strconv.ParseBool(param)
}

func parseUint32UrlParam(c *gin.Context, name string) (core.OptionalUint32, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return core.OptionalUint32{}, nil
	}

	value, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return core.OptionalUint32{}, err
	}

	return core.OptionalUint32{Value: uint32(value), HasValue: true}, nil
}

func parseUint64UrlParam(c *gin.Context, name string) (core.OptionalUint64, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return core.OptionalUint64{}, nil
	}

	value, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return core.OptionalUint64{}, err
	}

	return core.OptionalUint64{Value: value, HasValue: true}, nil
}

func parseHexBytesUrlParam(c *gin.Context, name string) ([]byte, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return nil, nil
	}

	decoded, err := hex.DecodeString(param)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
