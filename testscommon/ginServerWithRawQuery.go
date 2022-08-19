package testscommon

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// CreateGinContextWithRawQuery creates a test gin context
func CreateGinContextWithRawQuery(rawQuery string) *gin.Context {
	return &gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: rawQuery,
			},
		},
	}
}
