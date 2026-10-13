package internal

type Envelope map[string]any

const (
	Action             = "action"
	HelperFileName     = ".clonopher.yaml"
	ClrCreateNewFolder = "create_new"
	ClrStayHere        = "stay_here"
	ClrUncertain       = "uncertain"
)

const (
	// -- run modes --
	CliMode    string = "cli"
	ApiMode    string = "api"
	DryRunMode        = "dry"
	// action policy --
	PlcNotify  = "notify"
	PlcCopy    = "copy"
	PlcMove    = "move"
	PlcAskCopy = "ask-copy"
	PlcAskMove = "ask-move"
)
