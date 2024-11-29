package main

import (
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

var (
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogTrace.String(),
	}
	grpcEnabled = cli.BoolFlag{
		Name:  "grpc-enabled",
		Usage: "Boolean option for enabling GRPC server.",
	}
	sovereignBridgeCertificateFile = cli.StringFlag{
		Name:  "certificate",
		Usage: "The path for sovereign grpc bridge service certificate file.",
		Value: "~/MultiversX/testnet/node/config/certificate.crt",
	}
	sovereignBridgeCertificatePkFile = cli.StringFlag{
		Name:  "certificate-pk",
		Usage: "The path for sovereign grpc bridge private key certificate file.",
		Value: "~/MultiversX/testnet/node/config/private_key.pem",
	}
)
