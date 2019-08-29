package termuic

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/termuic/termuiRenders"
	ui "github.com/gizak/termui/v3"
)

//refreshInterval is used for a ticker that refresh termui console at a specific interval
const refreshInterval = time.Second

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	presenter     statusHandler.PresenterInterface
	consoleRender TermuiRender
	grid          *termuiRenders.DrawableContainer
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(presenter statusHandler.PresenterInterface) *TermuiConsole {
	tc := TermuiConsole{
		presenter: presenter,
	}

	return &tc
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
	tc.consoleRender = termuiRenders.NewWidgetsRender(tc.presenter, tc.grid)

	termWidth, termHeight := ui.TerminalDimensions()
	tc.grid.SetRectangle(0, 0, termWidth, termHeight)

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	tc.consoleRender.RefreshData()

	for {
		select {
		case <-time.After(refreshInterval):
			tc.consoleRender.RefreshData()
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
