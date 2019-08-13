package termuiRenders

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const numLogLinesBigLog = 50

//LogRender will define termui widgets that need to display a termui console
type LogRender struct {
	lLog *widgets.List
	grid *ui.Grid
}

//NewLogRender method is used to created a new LogRender that display termui console
func NewLogRender(grid *ui.Grid) *LogRender {
	self := &LogRender{
		grid: grid,
		lLog: widgets.NewList(),
	}
	self.grid.Set(
		ui.NewRow(1.0, self.lLog),
	)

	return self
}

//RefreshData method is used to prepare data that are displayed on container
func (lr *LogRender) RefreshData(logData []string) {
	lr.lLog.Title = "Log info"
	lr.lLog.Rows = lr.prepareLogLines(logData)
	lr.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	lr.lLog.WrapText = false

	return
}

func (lr *LogRender) prepareLogLines(logData []string) []string {
	logDataLen := len(logData)

	if logDataLen > numLogLinesBigLog {
		return logData[logDataLen-numLogLinesBigLog : logDataLen]
	}
	return logData
}
