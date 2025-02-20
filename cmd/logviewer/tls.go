package main

import (
	"crypto/tls"
	"crypto/x509"
)

func loadTLSClientConfig(certFile string, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certPool, err := createCertPool(cert)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}, nil
}

func createCertPool(cert tls.Certificate) (*x509.CertPool, error) {
	certLeaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certLeaf)

	return certPool, nil
}
