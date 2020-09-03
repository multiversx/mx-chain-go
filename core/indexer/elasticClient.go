package indexer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
	"net/http"
)

type responseErrorHandler func(res *esapi.Response) error

type elasticClient struct {
	elasticBaseUrl string
	es             *elasticsearch.Client
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
		es:             es,
		elasticBaseUrl: cfg.Addresses[0],
	}

	return ec, nil
}

// TemplateExists checks weather a template is already created
func (ec *elasticClient) TemplateExists(index string) bool {
	res, err := ec.es.Indices.ExistsTemplate([]string{index})
	return exists(res, err)
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) IndexExists(index string) bool {
	res, err := ec.es.Indices.Exists([]string{index})
	return exists(res, err)
}

// PolicyExists checks if a policy was already created
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
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		log.Warn("elasticClient.PolicyExists", "error returned by kibana api", err.Error())
		return false
	}

	return existsRes.Ok
}

// AliasExists checks if an index alias already exists
func (ec *elasticClient) AliasExists(alias string) bool {
	aliasRoute := fmt.Sprintf(
		"%s/_alias/%s",
		ec.elasticBaseUrl,
		alias,
	)

	req, err := newRequest(http.MethodHead, aliasRoute, nil)
	if err != nil {
		log.Warn("elasticClient.AliasExists", "could not create request object", err.Error())
		return false
	}

	res, err := ec.es.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.AliasExists", "error performing request", err.Error())
		return false
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return exists(response, nil)
}

// CreateIndex creates an elasticsearch index
func (ec *elasticClient) CreateIndex(index string) error {
	res, err := ec.es.Indices.Create(index)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreatePolicy creates a new policy for elastic indexes. Policies define rollover parameters
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
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		return err
	}

	if !existsRes.Ok {
		return ErrCouldNotCreatePolicy
	}

	return nil
}

// CreateIndexTemplate creates an elasticsearch index template
func (ec *elasticClient) CreateIndexTemplate(templateName string, template io.Reader) error {
	res, err := ec.es.Indices.PutTemplate(template, templateName)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreateAlias creates an index alias
func (ec *elasticClient) CreateAlias(alias string, index string) error {
	res, err := ec.es.Indices.PutAlias([]string{index}, alias)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CheckAndCreateTemplate creates an index template if it does not already exist
func (ec *elasticClient) CheckAndCreateTemplate(templateName string, template io.Reader) error {
	if ec.TemplateExists(templateName) {
		return nil
	}

	return ec.CreateIndexTemplate(templateName, template)
}

// CheckAndCreatePolicy creates a new index policy if it does not already exist
func (ec *elasticClient) CheckAndCreatePolicy(policyName string, policy io.Reader) error {
	if ec.PolicyExists(policyName) {
		return nil
	}

	return ec.CreatePolicy(policyName, policy)
}

// CheckAndCreateIndex creates a new index if it does not already exist
func (ec *elasticClient) CheckAndCreateIndex(indexName string) error {
	if ec.IndexExists(indexName) {
		return nil
	}

	return ec.CreateIndex(indexName)
}

// CheckAndCreateAlias creates a new alias if it does not already exist
func (ec *elasticClient) CheckAndCreateAlias(alias string, indexName string) error {
	if ec.AliasExists(alias) {
		return nil
	}

	return ec.CreateAlias(alias, indexName)
}

// DoRequest will do a request to elastic server
func (ec *elasticClient) DoRequest(req *esapi.IndexRequest) error {
	res, err := req.Do(context.Background(), ec.es)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// DoBulkRequest will do a bulk of request to elastic server
func (ec *elasticClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	res, err := ec.es.Bulk(reader, ec.es.Bulk.WithIndex(index))
	if err != nil {
		log.Warn("elasticClient.DoMultiGet",
			"indexer do bulk request no response", err.Error())
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// DoMultiGet wil do a multi get request to elaticsearch server
func (ec *elasticClient) DoMultiGet(obj object, index string) (object, error) {
	body, err := encode(obj)
	if err != nil {
		return nil, err
	}

	res, err := ec.es.Mget(
		&body,
		ec.es.Mget.WithIndex(index),
	)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet", "cannot do multi get no response", err.Error())
		return nil, err
	}

	var decodedBody object
	err = parseResponse(res, &decodedBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet", "error parsing response", err.Error())
		return nil, err
	}

	return decodedBody, nil
}
