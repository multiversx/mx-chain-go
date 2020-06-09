package factory

//Viewer defines the actions which should be handled by a view object
type Viewer interface {
	Start(start chan struct{}) error
}
