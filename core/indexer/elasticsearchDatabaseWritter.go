package indexer

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type databaseWriter struct {
	client *elasticsearch.Client
}

func newDatabaseWriter(cfg elasticsearch.Config) (*databaseWriter, error) {
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &databaseWriter{client: es}, nil
}

// CheckAndCreateIndex will check if a index exits and if dont will create a new one
func (dw *databaseWriter) checkAndCreateIndex(index string, body io.Reader) error {
	res, err := dw.client.Indices.Exists([]string{index})
	if err != nil {
		return err
	}

	defer closeESResponseBody(res)
	// Indices.Exists actually does a HEAD request to the elastic index.
	// A status code of 200 actually means the index exists so we
	//  don't need to do anything.
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// A status code of 404 means the index does not exist so we create it
	if res.StatusCode == http.StatusNotFound {
		err = dw.createDatabaseIndex(index, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dw *databaseWriter) createDatabaseIndex(index string, body io.Reader) error {
	var err error
	var res *esapi.Response

	if body != nil {
		res, err = dw.client.Indices.Create(
			index,
			dw.client.Indices.Create.WithBody(body))
	} else {
		res, err = dw.client.Indices.Create(index)
	}

	defer closeESResponseBody(res)

	if err != nil {
		return err
	}

	if res.IsError() {
		// Resource already exists
		if res.StatusCode == badRequest {
			return nil
		}

		log.Warn("indexer: resource already exists", "error", res.String())
		return ErrCannotCreateIndex
	}

	return nil
}

// DoRequest will do a request to elastic server
func (dw *databaseWriter) doRequest(req esapi.IndexRequest) error {
	res, err := req.Do(context.Background(), dw.client)
	if err != nil {
		return err
	}

	if res.IsError() {
		log.Warn("indexer", "error", res.String())
	}

	return nil
}

// DoBulkRequest will do a bulk of request to elastic server
func (dw *databaseWriter) doBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	res, err := dw.client.Bulk(reader, dw.client.Bulk.WithIndex(index))
	if err != nil {
		return err
	}

	if res.IsError() {
		log.Warn("indexer", "error", res.String())
	}

	return nil
}
