package middleware

import (
	"bytes"
	"fmt"
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
const responseMaxLength = 100

type responseLoggerMiddleware struct {
	thresholdDurationForLoggingRequest time.Duration
	printRequestFunc                   func(title string, path string, duration time.Duration, status int, request string, response string)
}

// NewResponseLoggerMiddleware returns a new instance of responseLoggerMiddleware
func NewResponseLoggerMiddleware(thresholdDurationForLoggingRequest time.Duration) *responseLoggerMiddleware {
	rlm := &responseLoggerMiddleware{
		thresholdDurationForLoggingRequest: thresholdDurationForLoggingRequest,
	}

	rlm.printRequestFunc = rlm.printRequest

	return rlm
}

// MonitoringMiddleware logs detail about a request if it is not successful or it's duration is higher than a threshold
func (rlm *responseLoggerMiddleware) MiddlewareHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		latency := time.Since(t)
		status := c.Writer.Status()

		shouldLogRequest := latency > rlm.thresholdDurationForLoggingRequest || c.Writer.Status() != http.StatusOK
		if shouldLogRequest {
			responseBodyString := removeWhitespacesFromString(bw.body.String())
			rlm.logRequestAndResponse(c, latency, status, responseBodyString)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rlm *responseLoggerMiddleware) IsInterfaceNil() bool {
	return rlm == nil
}

func (rlm *responseLoggerMiddleware) logRequestAndResponse(c *gin.Context, duration time.Duration, status int, response string) {
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

	title := rlm.computeLogTitle(status)

	rlm.printRequestFunc(title, c.Request.RequestURI, duration, status, request, response)
}

func (rlm *responseLoggerMiddleware) computeLogTitle(status int) string {
	logPrefix := prefixDurationTooLong
	if status == http.StatusBadRequest {
		logPrefix = prefixBadRequest
	} else if status == http.StatusInternalServerError {
		logPrefix = prefixInternalError
	} else if status != http.StatusOK {
		logPrefix = fmt.Sprintf("http code %d", status)
	}

	return fmt.Sprintf("%s api request", logPrefix)
}

func (rlm *responseLoggerMiddleware) printRequest(title string, path string, duration time.Duration, status int, request string, response string) {
	if len(response) > responseMaxLength {
		response = response[:responseMaxLength] + "..."
	}

	log.Debug(title,
		"path", path,
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
