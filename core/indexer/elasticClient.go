package indexer

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type elasticClient struct {
	es *elasticsearch.Client
}

func newElasticClient(cfg elasticsearch.Config) (*elasticClient, error) {
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ec := &elasticClient{
		es: es,
	}

	return ec, nil
}

// NodesInfo returns info about the elastic cluster nodes
func (ec *elasticClient) NodesInfo() (*esapi.Response, error) {
	return ec.es.Nodes.Info()
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) IndexExists(index string) bool {
	res, err := ec.es.Indices.Exists([]string{index})
	defer func() {
		if res.Body != nil {
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

	if res.StatusCode == http.StatusOK {
		return true
	}

	if res.StatusCode == http.StatusNotFound {
		return false
	}

	log.Warn("elasticClient.IndexExists", "invalid status code returned by the elastic nodes:", res.StatusCode)

	return false
}

func (ec *elasticClient) parseResponse(res *esapi.Response, dest interface{}) error {
	defer func() {
		if res.Body != nil {
			err := res.Body.Close()
			if err != nil {
				log.Warn("elasticClient.parseResponse", "could not close body: ", err.Error())
			}
		}
	}()

	if res.StatusCode != http.StatusOK {
		log.Warn("elasticClient.parseResponse", "error returned by elastic API:", res.StatusCode)
		return ErrBackOff
	}

	err := ec.loadResponseBody(res.Body, dest)
	if err != nil {
		log.Warn("elasticClient.parseResponse", "could not load response body:", err.Error())
		return ErrBackOff
	}

	return nil
}

func (ec *elasticClient) loadResponseBody(body io.ReadCloser, dest interface{}) error {
	if dest == nil {
		_, err := io.Copy(ioutil.Discard, body)
		return err
	}

	err := json.NewDecoder(body).Decode(dest)
	return err
}
