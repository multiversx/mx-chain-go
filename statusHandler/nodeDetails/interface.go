package nodeDetails

// NodeDetails is the interface that defines what a node details handler/provider should do
type NodeDetails interface {
	DetailsMap() (map[string]interface{}, error)
	IsInterfaceNil() bool
}
