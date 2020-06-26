package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type databaseClient struct {
	dbClient *elasticsearch.Client
}

func newDatabaseWriter(cfg elasticsearch.Config) (*databaseClient, error) {
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &databaseClient{dbClient: es}, nil
}

// CheckAndCreateIndex will check if a index exits and if dont will create a new one
func (dc *databaseClient) CheckAndCreateIndex(index string, body io.Reader) error {
	var res *esapi.Response
	var err error
	defer func() {
		closeESResponseBody(res)
	}()

	res, err = dc.dbClient.Indices.Exists([]string{index})
	if err != nil {
		return err
	}

	// Indices.Exists actually does a HEAD request to the elastic index.
	// A status code of 200 actually means the index exists so we
	//  don't need to do anything.
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// A status code of 404 means the index does not exist so we create it
	if res.StatusCode == http.StatusNotFound {
		err = dc.createDatabaseIndex(index, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *databaseClient) createDatabaseIndex(index string, body io.Reader) error {
	var err error
	var res *esapi.Response
	defer func() {
		closeESResponseBody(res)
	}()

	if body != nil {
		res, err = dc.dbClient.Indices.Create(index, dc.dbClient.Indices.Create.WithBody(body))
	} else {
		res, err = dc.dbClient.Indices.Create(index)
	}

	if err != nil {
		return err
	}

	if res.IsError() {
		// Resource already exists
		if res.StatusCode == http.StatusBadRequest {
			return nil
		}

		log.Warn("indexer: resource already exists", "error", res.String())
		return ErrCannotCreateIndex
	}

	return nil
}

// DoRequest will do a request to elastic server
func (dc *databaseClient) DoRequest(req *esapi.IndexRequest) error {
	var err error
	var res *esapi.Response
	defer func() {
		closeESResponseBody(res)
	}()

	res, err = req.Do(context.Background(), dc.dbClient)
	if err != nil {
		return err
	}

	if res.IsError() {
		log.Warn("indexer do request", "error", res.String())
	}

	return nil
}

// DoBulkRequest will do a bulk of request to elastic server
func (dc *databaseClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	var err error
	var res *esapi.Response
	defer func() {
		closeESResponseBody(res)
	}()

	res, err = dc.dbClient.Bulk(reader, dc.dbClient.Bulk.WithIndex(index))
	if err != nil {
		log.Warn("indexer do bulk request no response ",
			"error", err.Error())
		return err
	}

	if res.IsError() {
		log.Warn("indexer do bulk request",
			"error", res.String())
		return fmt.Errorf("do bulk request %s", res.String())
	}

	return nil
}

// DoMultiGet wil do a multi get request to elaticsearch server
func (dc *databaseClient) DoMultiGet(obj object, index string) (object, error) {
	body, err := encode(obj)
	if err != nil {
		return nil, err
	}

	var res *esapi.Response
	defer func() {
		closeESResponseBody(res)
	}()

	res, err = dc.dbClient.Mget(
		&body,
		dc.dbClient.Mget.WithIndex(index),
	)
	if err != nil {
		log.Warn("indexer: cannot do multi get no response", "error", err)
		return nil, err
	}

	if res.IsError() {
		log.Warn("indexer: cannot do multi get", "error", res.String())
		return nil, fmt.Errorf("do multi get %s", res.String())
	}

	var responseBody []byte
	responseBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn("indexer:cannot read from response body", "error", err,
			"body", string(responseBody))
		return nil, err
	}

	var decodedBody object
	if err := json.Unmarshal(responseBody, &decodedBody); err != nil {
		log.Warn("indexer cannot decode body", "error", err,
			"body", string(responseBody))
		return nil, err
	}

	return decodedBody, nil
}

func closeESResponseBody(res *esapi.Response) {
	if res != nil && res.Body != nil {
		err := res.Body.Close()
		if err != nil {
			log.Warn("cannot close elastic search response body", "error", err)
		}
	}
}
