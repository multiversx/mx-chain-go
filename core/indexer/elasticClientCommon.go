package indexer

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"net/url"

	"github.com/ElrondNetwork/elrond-go/core"
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

	respBodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Warn("elasticClient.parseResponse",
		"error returned by elastic API", res.StatusCode,
		"body", string(respBodyBytes))

	return ErrBackOff
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

func computeBalanceAsFloat(balance *big.Int, denomination int, numberDecimals int) float64 {
	balanceCopy := big.NewInt(0).Set(balance)
	balanceBigFloat := big.NewFloat(0).SetInt(balanceCopy)
	denominationPow := math.Pow(10, float64(denomination))
	numberToUseForReachingDecimals := math.Pow(10, float64(numberDecimals))

	balanceWithoutDenomination := balanceBigFloat.Quo(balanceBigFloat, big.NewFloat(0).SetInt(big.NewInt(int64(denominationPow))))
	balanceFloat, _ := balanceWithoutDenomination.Float64()

	balanceFloatWith4Decimals := math.Round(balanceFloat*numberToUseForReachingDecimals) / numberToUseForReachingDecimals

	return core.MaxFloat64(balanceFloatWith4Decimals, 0)
}
