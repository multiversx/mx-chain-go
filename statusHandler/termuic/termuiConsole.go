package termuic

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/statusHandler/termuic/termuiRenders"
	ui "github.com/gizak/termui/v3"
)

//refreshInterval is used for a ticker that refresh termui console at a specific interval
const refreshInterval = time.Second

//maxLogLines is used to specify how many lines of logs need to store in slice
var maxLogLines = 100

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	termuiConsoleMetrics *sync.Map
	logLines             []string
	mutLogLineWrite      sync.RWMutex
	consoleRender        TermuiRender
	grid                 *termuiRenders.DrawableContainer
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(metricData *sync.Map) *TermuiConsole {
	tc := TermuiConsole{
		termuiConsoleMetrics: metricData,
	}

	return &tc
}

func (tc *TermuiConsole) Write(p []byte) (n int, err error) {
	go func(p []byte) {
		logLine := string(p)
		stringSlice := strings.Split(logLine, "\n")

		tc.mutLogLineWrite.Lock()
		for _, line := range stringSlice {
			line = strings.Replace(line, "\r", "", len(line))
			if line != "" {
				tc.logLines = append(tc.logLines, line)
			}
		}

		startPos := len(tc.logLines) - maxLogLines
		if startPos < 0 {
			startPos = 0
		}
		tc.logLines = tc.logLines[startPos:len(tc.logLines)]

		tc.mutLogLineWrite.Unlock()
	}(p)

	return len(p), nil
}

// Start method - will start termui console
func (tc *TermuiConsole) Start() error {
	if err := ui.Init(); err != nil {
		return err
	}

	go func() {
		defer ui.Close()
		tc.eventLoop()
	}()

	return nil
}

func (tc *TermuiConsole) eventLoop() {
	tc.grid = termuiRenders.NewDrawableContainer()
	tc.consoleRender = termuiRenders.NewWidgetsRender(tc.termuiConsoleMetrics, tc.grid)

	termWidth, termHeight := ui.TerminalDimensions()
	tc.grid.SetRectangle(0, 0, termWidth, termHeight)

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	tc.consoleRender.RefreshData(tc.logLines)

	for {
		select {
		case <-time.After(refreshInterval):
			tc.consoleRender.RefreshData(tc.logLines)
			ui.Clear()
			ui.Render(tc.grid.TopLeft(), tc.grid.TopRight(), tc.grid.Bottom())

		case <-sigTerm:
			ui.Clear()
			return
		case e := <-uiEvents:
			tc.processUiEvents(e)
		}
	}
}

func (tc *TermuiConsole) processUiEvents(e ui.Event) {
	switch e.ID {
	case "<Resize>":
		payload := e.Payload.(ui.Resize)
		tc.grid.SetRectangle(0, 0, payload.Width, payload.Height)
		ui.Clear()
		ui.Render(tc.grid.TopLeft(), tc.grid.TopRight(), tc.grid.Bottom())

	case "<C-c>":
		ui.Close()
		StopApplication()
		return
	}
}
