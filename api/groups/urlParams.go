package groups

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseBoolUrlParam(c *gin.Context, name string) (bool, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return false, nil
	}

	return strconv.ParseBool(param)
}

func parseUintUrlParam(c *gin.Context, name string) (uint64, error) {
	param := c.Request.URL.Query().Get(name)
	if param == "" {
		return 0, nil
	}

	return strconv.ParseUint(param, 10, 64)
}
