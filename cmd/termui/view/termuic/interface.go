package termuic

//TermuiRender defines the actions which should be handled by a presenter
type TermuiRender interface {
	// RefreshData method is used to refresh data that are displayed on a grid
	RefreshData(numMillisecondsRefreshTime int)
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}
