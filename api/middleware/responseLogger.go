package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

const prefixDurationTooLong = "[too long]"
const prefixBadRequest = "[bad request]"
const prefixInternalError = "[internal error]"

type responseLoggerMiddleware struct {
	thresholdDurationForLoggingRequest time.Duration
}

// NewResponseLoggerMiddleware returns a new instance of responseLoggerMiddleware
func NewResponseLoggerMiddleware(thresholdDurationForLoggingRequest time.Duration) *responseLoggerMiddleware {
	return &responseLoggerMiddleware{
		thresholdDurationForLoggingRequest: thresholdDurationForLoggingRequest,
	}
}

// MonitoringMiddleware logs detail about a request if it is not successful or it's duration is higher than a threshold
func (rl *responseLoggerMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		latency := time.Since(t)
		status := c.Writer.Status()
		if latency > rl.thresholdDurationForLoggingRequest || c.Writer.Status() != http.StatusOK {
			responseBodyString := removeWhitespacesFromString(bw.body.String())
			logRequestAndResponse(c, latency, status, responseBodyString)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rl *responseLoggerMiddleware) IsInterfaceNil() bool {
	return rl == nil
}

func logRequestAndResponse(c *gin.Context, duration time.Duration, status int, response string) {
	request := "n/a"

	if c.Request.Body != nil {
		reqBody := c.Request.Body
		reqBodyBytes, err := ioutil.ReadAll(reqBody)
		if err != nil {
			log.Error(err.Error())
			return
		}

		if len(reqBodyBytes) > 0 {
			request = removeWhitespacesFromString(string(reqBodyBytes))
		}
	}

	logPrefix := prefixDurationTooLong
	if status == http.StatusBadRequest {
		logPrefix = prefixBadRequest
	} else if status == http.StatusInternalServerError {
		logPrefix = prefixInternalError
	}

	log.Warn(logPrefix+" api request",
		"path", c.Request.RequestURI,
		"duration", duration,
		"status", status,
		"request", request,
		"response", response,
	)
}

func removeWhitespacesFromString(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
