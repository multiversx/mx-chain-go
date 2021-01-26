package main

import (
	"fmt"
	"os"
	"runtime"

	logger "github.com/ElrondNetwork/elrond-go-logger"
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
	app.Flags = []cli.Flag{}
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

func startAssessment(_ *cli.Context, version string) error {
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

	benchmarkResult := run.GetStringTableAfterRun()

	hpg := hostParameters.NewHostParameterGetter(version)
	log.Info("Host's anonymized info:\n" + hpg.GetParameterStringTable())
	log.Info("Host's performance info:\n" + benchmarkResult)

	return nil
}
