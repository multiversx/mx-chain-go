package indexer

import (
	"bytes"
	"encoding/json"
	"fmt"
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
				log.Warn("elasticClient.exists", "could not close body: ", err.Error())
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
	if body == nil {
		return nil
	}
	if dest == nil {
		_, err := io.Copy(ioutil.Discard, body)
		return err
	}

	err := json.NewDecoder(body).Decode(dest)
	return err
}

func elasticDefaultErrorResponseHandler(res *esapi.Response) error {
	responseBody := map[string]interface{}{}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w cannot read elastic response body bytes", err)
	}

	err = json.Unmarshal(bodyBytes, &responseBody)
	if err != nil {
		errToReturn := err
		isBackOffError := strings.Contains(string(bodyBytes), fmt.Sprintf("%d", http.StatusForbidden)) ||
			strings.Contains(string(bodyBytes), fmt.Sprintf("%d", http.StatusTooManyRequests))
		if isBackOffError {
			errToReturn = ErrBackOff
		}

		return fmt.Errorf("%w, cannot unmarshal elastic response body to map[string]interface{}, "+
			"decode error: %s, body response: %s", errToReturn, err.Error(), string(bodyBytes))
	}

	if res.IsError() {
		if errIsAlreadyExists(responseBody) {
			return nil
		}
		if isErrAliasAlreadyExists(responseBody) {
			log.Debug("alias already exists", "response", fmt.Sprintf("%v", responseBody))
			return nil
		}
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("error while parsing the response: code returned: %v, body: %v, bodyBytes: %v",
		res.StatusCode, responseBody, bodyBytes)
}

func errIsAlreadyExists(response map[string]interface{}) bool {
	alreadyExistsMessage := "resource_already_exists_exception"
	errKey := "error"
	typeKey := "type"

	errMapI, ok := response[errKey]
	if !ok {
		return false
	}

	errMap, ok := errMapI.(map[string]interface{})
	if !ok {
		return false
	}

	existsString, ok := errMap[typeKey].(string)
	if !ok {
		return false
	}

	return existsString == alreadyExistsMessage
}

func isErrAliasAlreadyExists(response map[string]interface{}) bool {
	aliasExistsMessage := "invalid_alias_name_exception"
	errKey := "error"
	typeKey := "type"

	errMapI, ok := response[errKey]
	if !ok {
		return false
	}

	errMap, ok := errMapI.(map[string]interface{})
	if !ok {
		return false
	}

	existsString, ok := errMap[typeKey].(string)
	if !ok {
		return false
	}

	return existsString == aliasExistsMessage
}

func kibanaResponseErrorHandler(res *esapi.Response) error {
	errorRes := &kibanaResponse{}
	decodeErr := loadResponseBody(res.Body, errorRes)
	if decodeErr != nil {
		return decodeErr
	}

	log.Warn("elasticClient.parseResponse",
		"error returned by elastic API", errorRes.Error,
		"code", res.StatusCode)
	return ErrBackOff
}

func newRequest(method, path string, body *bytes.Buffer) (*http.Request, error) {
	r := http.Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if body != nil {
		r.Body = ioutil.NopCloser(body)
		r.ContentLength = int64(body.Len())
	}

	return &r, nil
}

/**
 * parseResponse will check and load the elastic/kibana api response into the destination objectsMap. Custom errorHandler
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
					"could not close body", err.Error())
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
