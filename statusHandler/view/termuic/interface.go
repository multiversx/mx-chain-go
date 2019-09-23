package termuic

//TermuiRender defines the actions which should be handled by a render
type TermuiRender interface {
	// RefreshData method is used to refresh data that are displayed on a grid
	RefreshData()
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}
