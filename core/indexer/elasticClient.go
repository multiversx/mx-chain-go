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

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type elasticClient struct {
	elasticBaseUrl string
	es *elasticsearch.Client
}

func newElasticClient(cfg elasticsearch.Config) (*elasticClient, error) {
	if len(cfg.Addresses) == 0 {
		return nil, ErrNoElasticUrlProvided
	}

 	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ec := &elasticClient{
		es: es,
		elasticBaseUrl: cfg.Addresses[0],
	}

	return ec, nil
}

// NodesInfo returns info about the elastic cluster nodes
func (ec *elasticClient) NodesInfo() (*esapi.Response, error) {
	return ec.es.Nodes.Info()
}

// TemplateExists -
func (ec *elasticClient) TemplateExists(index string) bool {
	res, err := ec.es.Indices.ExistsTemplate([]string{index})
	return ec.exists(res, err)
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) IndexExists(index string) bool {
	res, err := ec.es.Indices.Exists([]string{index})
	return ec.exists(res, err)
}

// PolicyExists -
func (ec *elasticClient) PolicyExists(policy string) bool {
	policyRoute := fmt.Sprintf(
	"%s/%s/ism/policies/%s",
			ec.elasticBaseUrl,
			kibanaPluginPath,
			policy,
	)

	req, err := newRequest(http.MethodGet, policyRoute, nil)
	if err != nil {
		log.Warn("elasticClient.PolicyExists", "could not create request object", err.Error())
		return false
	}

	res, err := ec.es.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.PolicyExists", "error performing request", err.Error())
		return false
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	existsRes := &kibanaExistsResponse{}
	err = ec.parseResponse(response, existsRes)
	if err != nil {
		log.Warn("elasticClient.PolicyExists", "error returned by kibana api", err.Error())
		return false
	}

	return existsRes.Ok
}

// CreateIndex -
func (ec *elasticClient) CreateIndex(index string) error {
	res, err := ec.es.Indices.Create(index)
	if err != nil {
		return err
	}

	return ec.parseResponse(res, nil)
}

// CreatePolicy -
func (ec *elasticClient) CreatePolicy(policyName string, policy io.Reader) error {
	policyRoute := fmt.Sprintf(
		"%s/%s/ism/policies/%s",
		ec.elasticBaseUrl,
		kibanaPluginPath,
		policyName,
	)

	req, err := newRequest(http.MethodPut, policyRoute, policy)
	if err != nil {
		return err
	}

	req.Header[headerContentType] = headerContentTypeJSON
	req.Header[headerXSRF] = []string{"false"}
	res, err := ec.es.Transport.Perform(req)
	if err != nil {
		return err
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	existsRes := &kibanaExistsResponse{}
	err = ec.parseResponse(response, existsRes)
	if err != nil {
		return err
	}

	if !existsRes.Ok {
		return ErrCouldNotCreatePolicy
	}

	return nil
}

// CreateIndexTemplate -
func (ec *elasticClient) CreateIndexTemplate(templateName string, template io.Reader) error {
	res, err := ec.es.Indices.PutTemplate(template, templateName)
	if err != nil {
		return err
	}

	return ec.parseResponse(res, nil)
}

// CheckAndCreateTemplate -
func (ec *elasticClient) CheckAndCreateTemplate(templateName string, template io.Reader) error {
	if ec.TemplateExists(templateName) {
		return nil
	}

	return ec.CreateIndexTemplate(templateName, template)
}

// CheckAndCreatePolicy -
func (ec *elasticClient) CheckAndCreatePolicy(policyName string, policy io.Reader) error {
	if ec.PolicyExists(policyName) {
		return nil
	}

	return ec.CreatePolicy(policyName, policy)
}

// CheckAndCreateIndex -
func (ec *elasticClient) CheckAndCreateIndex(indexName string) error {
	if ec.IndexExists(indexName) {
		return nil
	}

	return ec.CreateIndex(indexName)
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
		errorRes := &kibanaErrorResponse{}
		decodeErr := ec.loadResponseBody(res.Body, errorRes)
		if decodeErr != nil {
			return decodeErr
		}
		log.Warn("elasticClient.parseResponse", "error returned by elastic API:", errorRes.Error, "code", res.StatusCode)
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

func (ec *elasticClient) exists(res *esapi.Response, err error) bool {
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