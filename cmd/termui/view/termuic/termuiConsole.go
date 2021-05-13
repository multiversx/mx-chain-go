package termuic

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/termui/view"
	"github.com/ElrondNetwork/elrond-go/cmd/termui/view/termuic/termuiRenders"
	ui "github.com/gizak/termui/v3"
)

// numOfTicksBeforeRedrawing represents the number of ticks which have to pass until a fake resize will be made
// in order to clean the unwanted appeared characters
const numOfTicksBeforeRedrawing = 10

var log = logger.GetOrCreate("statushandler/view/termuic")

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	presenter                 view.Presenter
	consoleRender             TermuiRender
	grid                      *termuiRenders.DrawableContainer
	mutRefresh                *sync.RWMutex
	chanNodeIsStarting        chan struct{}
	refreshTimeInMilliseconds int
}

// NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(presenter view.Presenter, refreshTimeInMilliseconds int, chanNodeIsStarting chan struct{}) (*TermuiConsole, error) {
	if presenter == nil {
		return nil, view.ErrNilPresenterInterface
	}
	if refreshTimeInMilliseconds < 1 {
		return nil, view.ErrInvalidRefreshTimeInMilliseconds
	}
	if chanNodeIsStarting == nil {
		return nil, view.ErrNilChanNodeIsStarting
	}

	tc := TermuiConsole{
		presenter:                 presenter,
		mutRefresh:                &sync.RWMutex{},
		refreshTimeInMilliseconds: refreshTimeInMilliseconds,
		chanNodeIsStarting:        chanNodeIsStarting,
	}

	return &tc, nil
}

// Start method - will start termui console
func (tc *TermuiConsole) Start() error {
	go func() {
		defer func() {
			log.Debug("closing termui ui")
			ui.Close()
		}()
		_ = ui.Init()
		tc.eventLoop()
	}()

	return nil
}

func (tc *TermuiConsole) eventLoop() {
	tc.grid = termuiRenders.NewDrawableContainer()
	if tc.grid == nil {
		log.Debug("cannot presenter termui console", "error", view.ErrNilGrid.Error())
		return
	}

	var err error
	tc.consoleRender, err = termuiRenders.NewWidgetsRender(tc.presenter, tc.grid)
	if err != nil {
		log.Debug("nil console presenter", "error", err.Error())
		return
	}

	termWidth, termHeight := ui.TerminalDimensions()
	tc.grid.SetRectangle(0, 0, termWidth, termHeight)

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	tc.consoleRender.RefreshData(tc.refreshTimeInMilliseconds)
	ticksCounter := uint32(0)

	for {
		select {
		case <-time.After(time.Millisecond * time.Duration(tc.refreshTimeInMilliseconds)):
			tc.doChanges(&ticksCounter, tc.refreshTimeInMilliseconds)
		case <-sigTerm:
			ui.Clear()
			return
		case e := <-uiEvents:
			tc.processUiEvents(e, tc.refreshTimeInMilliseconds)
		case <-tc.chanNodeIsStarting:
			tc.presenter.InvalidateCache()
		}
	}
}

func (tc *TermuiConsole) processUiEvents(e ui.Event, numMillisecondsRefreshTime int) {
	switch e.ID {
	case "<Resize>":
		tc.doResizeEvent(e, numMillisecondsRefreshTime)
	case "<C-c>":
		ui.Close()
		stopApplication()
		return
	}
}

func (tc *TermuiConsole) doChanges(counter *uint32, numMillisecondsRefreshTime int) {
	atomic.AddUint32(counter, 1)
	if atomic.LoadUint32(counter) > numOfTicksBeforeRedrawing {
		width, height := ui.TerminalDimensions()
		tc.doResize(width, height, numMillisecondsRefreshTime)
		atomic.StoreUint32(counter, 0)
	} else {
		tc.refreshWindow(numMillisecondsRefreshTime)
	}
}

func (tc *TermuiConsole) doResizeEvent(e ui.Event, numMillisecondsRefreshTime int) {
	payload := e.Payload.(ui.Resize)
	tc.doResize(payload.Width, payload.Height, numMillisecondsRefreshTime)
}

func (tc *TermuiConsole) doResize(width int, height int, numMillisecondsRefreshTime int) {
	tc.grid.SetRectangle(0, 0, width, height)
	tc.refreshWindow(numMillisecondsRefreshTime)
}

func (tc *TermuiConsole) refreshWindow(numMillisecondsRefreshTime int) {
	tc.mutRefresh.Lock()
	defer tc.mutRefresh.Unlock()

	tc.consoleRender.RefreshData(numMillisecondsRefreshTime)
	ui.Clear()
	ui.Render(tc.grid.TopLeft(), tc.grid.TopRight(), tc.grid.Bottom())
}
