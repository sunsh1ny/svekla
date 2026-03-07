package parser

const (
	UnknownCommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

var (
	UnknownCommand = "UNKNOWN"
	SetCommand     = "SET"
	GetCommand     = "GET"
	DelCommand     = "DEL"
)

var namesToId = map[string]int{
	SetCommand: SetCommandID,
	GetCommand: GetCommandID,
	DelCommand: DelCommandID,
}

func commandNameToCommandID(commandName string) int {
	status, ok := namesToId[commandName]
	if !ok {
		return UnknownCommandID
	}

	return status
}

const (
	setCommandArgumentsCount = 2
	getCommandArgumentsCount = 1
	delCommandArgumentsCount = 1
)

var ArgumentsCount = map[int]int{
	SetCommandID: setCommandArgumentsCount,
	GetCommandID: getCommandArgumentsCount,
	DelCommandID: delCommandArgumentsCount,
}

func commandArgumentsCount(commandID int) int {
	return ArgumentsCount[commandID]
}
