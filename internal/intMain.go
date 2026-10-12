package internal


type Envelope map[string]any


const (
	Action             = "action"
	HelperFileName     = ".clonopher.yaml"
	ClrCreateNewFolder = "create_new"
	ClrStayHere        = "stay_here"
	ClrUncertain       = "uncertain"
)
