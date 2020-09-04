package indexer

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func exists(res *esapi.Response, err error) bool {
	defer func() {
		if res != nil && res.Body != nil {
			_, _ = io.Copy(ioutil.Discard, res.Body)
			err = res.Body.Close()
			if err != nil {
				log.Warn("elasticClient.IndexExists", "could not close body: ", err.Error())
			}
		}
	}()

	if err != nil {
		log.Warn("elasticClient.IndexExists", "could not check index on the elastic nodes:", err.Error())
		return false
	}

	switch res.StatusCode {
	case http.StatusOK:
		return true
	case http.StatusNotFound:
		return false
	default:
		log.Warn("elasticClient.exists", "invalid status code returned by the elastic nodes:", res.StatusCode)
		return false
	}
}

func loadResponseBody(body io.ReadCloser, dest interface{}) error {
	if dest == nil {
		_, err := io.Copy(ioutil.Discard, body)
		return err
	}

	err := json.NewDecoder(body).Decode(dest)
	return err
}

func elasticDefaultErrorResponseHandler(res *esapi.Response) error {
	responseBody := map[string]interface{}{}
	decodeErr := loadResponseBody(res.Body, &responseBody)
	if decodeErr != nil {
		return decodeErr
	}

	if res.IsError() {
		if errIsAlreadyExists(responseBody) {
			return nil
		}
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		return nil
	}

	log.Warn("elasticClient.parseResponse",
		"error returned by elastic API:", res.StatusCode)
	log.Warn("elasticClient.parseResponse",
		"error returned by elastic API:", res.Body)
	return ErrBackOff
}

func errIsAlreadyExists(response map[string]interface{}) bool {
	alreadyExistsMessage := "resource_already_exists_exception"

	if errMapI, ok := response["error"]; ok {
		if errMap, ok := errMapI.(map[string]interface{}); ok {
			if existsString, ok := errMap["type"].(string); ok {
				return existsString == alreadyExistsMessage
			}
		}
	}

	return false
}

func kibanaResponseErrorHandler(res *esapi.Response) error {
	errorRes := &kibanaErrorResponse{}
	decodeErr := loadResponseBody(res.Body, errorRes)
	if decodeErr != nil {
		return decodeErr
	}

	log.Warn("elasticClient.parseResponse",
		"error returned by elastic API:", errorRes.Error,
		"code", res.StatusCode)
	return ErrBackOff
}

func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	r := http.Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if body != nil {
		switch b := body.(type) {
		case *bytes.Buffer:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		case *bytes.Reader:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		case *strings.Reader:
			r.Body = ioutil.NopCloser(body)
			r.ContentLength = int64(b.Len())
		default:
			r.Body = ioutil.NopCloser(body)
		}
	}

	return &r, nil
}

/**
 * parseResponse will check and load the elastic/kibana api response into the destination object. Custom errorHandler
 *  can be passed for special requests that want to handle StatusCode != 200. Every responseErrorHandler
 *  implementation should call loadResponseBody or consume the response body in order to be able to
 *  reuse persistent TCP connections: https://github.com/elastic/go-elasticsearch#usage
 */
func parseResponse(res *esapi.Response, dest interface{}, errorHandler responseErrorHandler) error {
	defer func() {
		if res != nil && res.Body != nil {
			err := res.Body.Close()
			if err != nil {
				log.Warn("elasticClient.parseResponse",
					"could not close body: ", err.Error())
			}
		}
	}()

	if errorHandler == nil {
		errorHandler = elasticDefaultErrorResponseHandler
	}

	if res.StatusCode != http.StatusOK {
		return errorHandler(res)
	}

	err := loadResponseBody(res.Body, dest)
	if err != nil {
		log.Warn("elasticClient.parseResponse",
			"could not load response body:", err.Error())
		return ErrBackOff
	}

	return nil
}
