package parser

type CommandID int

const (
	UnknownCommandID CommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

const (
	SetCommand = "SET"
	GetCommand = "GET"
	DelCommand = "DEL"
)

type commandSpec struct {
	id            CommandID
	argumentCount int
}

func commandByName(name string) (commandSpec, bool) {
	switch name {
	case SetCommand:
		return commandSpec{id: SetCommandID, argumentCount: 2}, true
	case GetCommand:
		return commandSpec{id: GetCommandID, argumentCount: 1}, true
	case DelCommand:
		return commandSpec{id: DelCommandID, argumentCount: 1}, true
	default:
		return commandSpec{}, false
	}
}
