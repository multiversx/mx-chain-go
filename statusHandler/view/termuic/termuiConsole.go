package termuic

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view/termuic/termuiRenders"
	ui "github.com/gizak/termui/v3"
)

//refreshInterval is used for a ticker that refresh termui console at a specific interval
const refreshInterval = time.Second

// numOfTicksBeforeRedrawing represents the number of ticks which have to pass until a fake resize will be made
// in order to clean the unwanted appeared characters
const numOfTicksBeforeRedrawing = 10

var log = logger.GetOrCreate("statushandler/termui")

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	presenter     view.Presenter
	consoleRender TermuiRender
	grid          *termuiRenders.DrawableContainer
	mutRefresh    *sync.RWMutex
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(presenter view.Presenter) (*TermuiConsole, error) {
	if presenter == nil || presenter.IsInterfaceNil() {
		return nil, statusHandler.ErrNilPresenterInterface
	}

	tc := TermuiConsole{
		presenter:  presenter,
		mutRefresh: &sync.RWMutex{},
	}

	return &tc, nil
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
	if tc.grid == nil {
		log.Debug("cannot render termui console", "error", statusHandler.ErrNilGrid.Error())
		return
	}

	var err error
	tc.consoleRender, err = termuiRenders.NewWidgetsRender(tc.presenter, tc.grid)
	if err != nil {
		log.Debug("nil console render", "error", err.Error())
		return
	}

	termWidth, termHeight := ui.TerminalDimensions()
	tc.grid.SetRectangle(0, 0, termWidth, termHeight)

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	tc.consoleRender.RefreshData()
	ticksCounter := uint32(0)

	for {
		select {
		case <-time.After(refreshInterval):
			tc.doChanges(&ticksCounter)
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
		tc.doResizeEvent(e)
	case "<C-c>":
		ui.Close()
		stopApplication()
		return
	}
}

func (tc *TermuiConsole) doChanges(counter *uint32) {
	atomic.AddUint32(counter, 1)
	if atomic.LoadUint32(counter) > numOfTicksBeforeRedrawing {
		tc.doResize(ui.TerminalDimensions())
		atomic.StoreUint32(counter, 0)
	} else {
		tc.refreshWindow()
	}
}

func (tc *TermuiConsole) doResizeEvent(e ui.Event) {
	payload := e.Payload.(ui.Resize)
	tc.doResize(payload.Width, payload.Height)
}

func (tc *TermuiConsole) doResize(width int, height int) {
	tc.grid.SetRectangle(0, 0, width, height)
	tc.refreshWindow()
}

func (tc *TermuiConsole) refreshWindow() {
	tc.mutRefresh.Lock()
	defer tc.mutRefresh.Unlock()

	tc.consoleRender.RefreshData()
	ui.Clear()
	ui.Render(tc.grid.TopLeft(), tc.grid.TopRight(), tc.grid.Bottom())
}
