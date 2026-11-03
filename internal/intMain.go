package internal

type Envelope map[string]any

const (
	Action         = "action"
	ConfigFileName = "clonopher.config.yaml"
	HelperFileName = ".clonopher.yaml"
	ActMove        = "ACT_MOVE"
	ActCreateNew   = "ACT_CREATE_NEW"
	ActStay        = "ACT_STAY_HERE"
	ACTUncertain   = "ACT_UNCERTAIN"
	ACTUndefined   = "ACT_UNDEFINED"
)

const (
	// -- run modes --
	CliMode    string = "cli"
	ApiMode    string = "api"
	DryRunMode        = "dry"
	// action policy --
	PlcNotify  = "notify"
	PlcNone    = "none"
	PlcCopy    = "copy"
	PlcMove    = "move"
	PlcAskCopy = "ask-copy"
	PlcAskMove = "ask-move"
)
