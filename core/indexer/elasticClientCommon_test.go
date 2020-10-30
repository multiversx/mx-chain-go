package indexer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadResponseBody_NilBodyNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, nil)
	require.NoError(t, err)
}

func TestLoadResponseBody_NilBodyNotNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, struct{}{})
	require.NoError(t, err)
}

func TestElasticDefaultErrorResponseHandler_ReadAllFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	resp := &esapi.Response{
		StatusCode: 0,
		Header:     nil,
		Body: &mock.ReadCloserStub{
			ReadCalled: func(p []byte) (n int, err error) {
				return 0, expectedErr
			},
		},
	}

	err := elasticDefaultErrorResponseHandler(resp)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestElasticDefaultErrorResponseHandler_UnmarshalFailsWithHttpForbiddenErrorShouldSignalBackOffErr(t *testing.T) {
	t.Parallel()

	httpErrString := fmt.Sprintf("%v Request throttled due to too many requests", http.StatusForbidden)
	resp := createMockEsapiResponseWithText(httpErrString)
	err := elasticDefaultErrorResponseHandler(resp)

	assert.True(t, errors.Is(err, ErrBackOff))
}

func TestElasticDefaultErrorResponseHandler_UnmarshalFailsWithHttpTooManyRequestsErrorShouldSignalBackOffErr(t *testing.T) {
	t.Parallel()

	httpErrString := fmt.Sprintf("%v Request throttled due to too many requests", http.StatusTooManyRequests)
	resp := createMockEsapiResponseWithText(httpErrString)
	err := elasticDefaultErrorResponseHandler(resp)

	assert.True(t, errors.Is(err, ErrBackOff))
}

func TestElasticDefaultErrorResponseHandler_UnmarshalFailsWithGenericError(t *testing.T) {
	t.Parallel()

	httpErrString := fmt.Sprintf("Generic error")
	resp := createMockEsapiResponseWithText(httpErrString)
	err := elasticDefaultErrorResponseHandler(resp)

	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid character 'G'"))
}

func TestElasticDefaultErrorResponseHandler_AlreadyExistsShouldRetNil(t *testing.T) {
	t.Parallel()

	data := `{
	"error": {
		"type": "resource_already_exists_exception"
	}
}`
	resp := createMockEsapiResponseWithText(data)
	resp.StatusCode = 300 // this is an error
	err := elasticDefaultErrorResponseHandler(resp)

	require.Nil(t, err)
}

func TestElasticDefaultErrorResponseHandler_StatusCodeOkShouldRetNil(t *testing.T) {
	t.Parallel()

	data := `{}`
	resp := createMockEsapiResponseWithText(data)
	resp.StatusCode = http.StatusOK
	err := elasticDefaultErrorResponseHandler(resp)

	require.Nil(t, err)
}

func TestElasticDefaultErrorResponseHandler_StatusCodeNotOkShouldErr(t *testing.T) {
	t.Parallel()

	errorString := "unique error string"
	data := `{
	"data": "` + errorString + `"
}`
	resp := createMockEsapiResponseWithText(data)
	resp.StatusCode = http.StatusBadRequest
	err := elasticDefaultErrorResponseHandler(resp)

	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), errorString))
}

func createMockEsapiResponseWithText(str string) *esapi.Response {
	return &esapi.Response{
		StatusCode: 0,
		Header:     nil,
		Body: &mock.ReadCloserStub{
			ReadCalled: func(p []byte) (n int, err error) {
				//dump contents into provided byte slice
				for i := 0; i < len(p); i++ {
					if i < len(str) {
						p[i] = str[i]
					} else {
						break
					}
				}

				return len(str), io.EOF
			},
		},
	}
}
