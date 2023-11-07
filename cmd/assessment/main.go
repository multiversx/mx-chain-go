package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/cmd/assessment/benchmarks"
	"github.com/multiversx/mx-chain-go/cmd/assessment/benchmarks/factory"
	"github.com/multiversx/mx-chain-go/cmd/assessment/hostParameters"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

const hostPlaceholder = "%host"
const timestampPlaceholder = "%time"

var (
	nodeHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`

	// outputFile defines a flag for the benchmarks output file. Data will be written in csv format.
	outputFile = cli.StringFlag{
		Name:  "output-file",
		Usage: "The output file format where benchmarks will be written in csv format.",
		Value: "./output-" + hostPlaceholder + "-" + timestampPlaceholder + ".csv",
	}

	log = logger.GetOrCreate("main")
)

func main() {
	_ = logger.SetDisplayByteSlice(logger.ToHexShort)

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "MultiversX Node Assessment Tool"
	machineID := core.GetAnonymizedMachineID(app.Name)

	app.Version = fmt.Sprintf("assessment-%s/%s-%s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Usage = "This tool is used to measure the host's performance on some certain tasks used by a MultiversX node. It " +
		"produces anonymized host parameters along with a list of benchmarks results. More details can be found in the README.md file."
	app.Flags = []cli.Flag{
		outputFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startAssessment(c, app.Version, machineID)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startAssessment(c *cli.Context, version string, machineID string) error {
	outputFileName := c.GlobalString(outputFile.Name)
	outputFileName = strings.Replace(outputFileName, hostPlaceholder, machineID, 1)
	outputFileName = strings.Replace(outputFileName, timestampPlaceholder, fmt.Sprintf("%d", time.Now().Unix()), 1)

	log.Info("Saving benchmarks result", "file", outputFileName)
	log.Info("Starting host assessment process...")
	sw := core.NewStopWatch()
	sw.Start("whole process")
	defer func() {
		sw.Stop("whole process")
		log.Debug("assessment process time measurement", sw.GetMeasurements()...)
	}()
	log.Info("Benchmark in progress. Please wait!")

	run, err := factory.NewRunner("./testdata")
	if err != nil {
		return err
	}

	hpg := hostParameters.NewHostParameterGetter(version)
	hostInfo := hpg.GetHostInfo()
	benchmarkResult := run.RunAllTests()

	log.Info("Host's anonymized info:\n" + hostInfo.ToDisplayTable())
	log.Info("Host's performance info:\n" + benchmarkResult.ToDisplayTable())

	printFinalResult(benchmarkResult)

	err = saveToFile(hostInfo, benchmarkResult, outputFileName)

	return err
}

func printFinalResult(results *benchmarks.TestResults) {
	if results.Error != nil {
		log.Error("The Node Under Test (NUT) performance can not be determined due to encountered errors")
		return
	}

	if results.EnoughComputingPower {
		log.Info("The Node Under Test (NUT) has enough computing power")
		return
	}

	log.Error("The Node Under Test (NUT) does not have enough computing power",
		"maximum accepted", benchmarks.ThresholdEnoughComputingPower,
		"obtained", results.TotalDuration)
}

func saveToFile(hi *hostParameters.HostInfo, results *benchmarks.TestResults, outputFileName string) error {
	buff := bytes.NewBuffer(make([]byte, 0))
	csvWriter := csv.NewWriter(buff)
	err := csvWriter.WriteAll(hi.ToStrings())
	if err != nil {
		return err
	}
	err = csvWriter.WriteAll(results.ToStrings())
	if err != nil {
		return err
	}

	return os.WriteFile(outputFileName, buff.Bytes(), core.FileModeReadWrite)
}
