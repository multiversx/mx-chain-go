package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks"
	"github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks/factory"
	"github.com/ElrondNetwork/elrond-go/cmd/assessment/hostParameters"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/denisbrodbeck/machineid"
	"github.com/urfave/cli"
)

const maxMachineIDLen = 10

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
		Usage: "The output file where benchmarks will be written in csv format.",
		Value: "./output.csv",
	}

	log = logger.GetOrCreate("main")
)

func main() {
	_ = logger.SetDisplayByteSlice(logger.ToHexShort)

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node Assessment Tool"
	machineID, err := machineid.ProtectedID(app.Name)
	if err != nil {
		log.Warn("error fetching machine id", "error", err)
		machineID = "unknown"
	}
	if len(machineID) > maxMachineIDLen {
		machineID = machineID[:maxMachineIDLen]
	}

	app.Version = fmt.Sprintf("assessment-%s/%s-%s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Usage = "This tool is used to measure the host's performance on some certain tasks used by an elrond node. It " +
		"produces anonymized host parameters along with a list of benchmarks results. More details can be found in the README.md file."
	app.Flags = []cli.Flag{
		outputFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startAssessment(c, app.Version)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startAssessment(c *cli.Context, version string) error {
	outputFileName := c.GlobalString(outputFile.Name)
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

	err = saveToFile(hostInfo, benchmarkResult, outputFileName)

	return err
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

	return ioutil.WriteFile(outputFileName, buff.Bytes(), os.ModePerm)
}
