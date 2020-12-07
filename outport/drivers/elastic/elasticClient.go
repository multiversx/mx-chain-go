package elastic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const errPolicyAlreadyExists = "document already exists"

type responseErrorHandler func(res *esapi.Response) error

type elasticClient struct {
	elasticBaseUrl string
	es             *elasticsearch.Client
}

// NewElasticClient will create a new instance of elasticClient
func NewElasticClient(cfg elasticsearch.Config) (*elasticClient, error) {
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

// CheckAndCreateTemplate creates an index template if it does not already exist
func (ec *elasticClient) CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error {
	if ec.templateExists(templateName) {
		return nil
	}

	return ec.createIndexTemplate(templateName, template)
}

// CheckAndCreatePolicy creates a new index policy if it does not already exist
func (ec *elasticClient) CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error {
	if ec.PolicyExists(policyName) {
		return nil
	}

	return ec.createPolicy(policyName, policy)
}

// CheckAndCreateIndex creates a new index if it does not already exist
func (ec *elasticClient) CheckAndCreateIndex(indexName string) error {
	if ec.indexExists(indexName) {
		return nil
	}

	return ec.createIndex(indexName)
}

// CheckAndCreateAlias creates a new alias if it does not already exist
func (ec *elasticClient) CheckAndCreateAlias(alias string, indexName string) error {
	if ec.aliasExists(alias) {
		return nil
	}

	return ec.createAlias(alias, indexName)
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
		log.Warn("elasticClient.DoBulkRequest",
			"indexer do bulk request no response", err.Error())
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// DoMultiGet wil do a multi get request to elaticsearch server
func (ec *elasticClient) DoMultiGet(obj objectsMap, index string) (objectsMap, error) {
	body, err := encode(obj)
	if err != nil {
		return nil, err
	}

	res, err := ec.es.Mget(
		&body,
		ec.es.Mget.WithIndex(index),
	)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet",
			"cannot do multi get no response", err.Error())
		return nil, err
	}

	var decodedBody objectsMap
	err = parseResponse(res, &decodedBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet",
			"error parsing response", err.Error())
		return nil, err
	}

	return decodedBody, nil
}

// DoBulkRemove will do a bulk remove to elasticsearch server
func (ec *elasticClient) DoBulkRemove(index string, hashes []string) error {
	obj := prepareHashesForBulkRemove(hashes)
	body, err := encode(obj)
	if err != nil {
		return err
	}

	res, err := ec.es.DeleteByQuery(
		[]string{index},
		&body,
		ec.es.DeleteByQuery.WithIgnoreUnavailable(true),
	)

	if err != nil {
		log.Warn("elasticClient.DoBulkRemove",
			"cannot do bulk remove", err.Error())
		return err
	}

	var decodedBody objectsMap
	err = parseResponse(res, &decodedBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoBulkRemove",
			"error parsing response", err.Error())
		return err
	}

	return nil
}

// TemplateExists checks weather a template is already created
func (ec *elasticClient) templateExists(index string) bool {
	res, err := ec.es.Indices.ExistsTemplate([]string{index})
	return exists(res, err)
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) indexExists(index string) bool {
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
		log.Warn("elasticClient.PolicyExists",
			"could not create request objectsMap", err.Error())
		return false
	}

	res, err := ec.es.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.PolicyExists",
			"error performing request", err.Error())
		return false
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	existsRes := &kibanaResponse{}
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		log.Warn("elasticClient.PolicyExists",
			"error returned by kibana api", err.Error())
		return false
	}

	return existsRes.Ok
}

// AliasExists checks if an index alias already exists
func (ec *elasticClient) aliasExists(alias string) bool {
	aliasRoute := fmt.Sprintf(
		"/_alias/%s",
		alias,
	)

	req, err := newRequest(http.MethodHead, aliasRoute, nil)
	if err != nil {
		log.Warn("elasticClient.AliasExists",
			"could not create request objectsMap", err.Error())
		return false
	}

	res, err := ec.es.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.AliasExists",
			"error performing request", err.Error())
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
func (ec *elasticClient) createIndex(index string) error {
	res, err := ec.es.Indices.Create(index)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreatePolicy creates a new policy for elastic indexes. Policies define rollover parameters
func (ec *elasticClient) createPolicy(policyName string, policy *bytes.Buffer) error {
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

	existsRes := &kibanaResponse{}
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		return err
	}

	if !existsRes.Ok && !strings.Contains(existsRes.Error, errPolicyAlreadyExists) {
		return ErrCouldNotCreatePolicy
	}

	return nil
}

// CreateIndexTemplate creates an elasticsearch index template
func (ec *elasticClient) createIndexTemplate(templateName string, template io.Reader) error {
	res, err := ec.es.Indices.PutTemplate(template, templateName)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreateAlias creates an index alias
func (ec *elasticClient) createAlias(alias string, index string) error {
	res, err := ec.es.Indices.PutAlias([]string{index}, alias)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *elasticClient) IsInterfaceNil() bool {
	return ec == nil
}
