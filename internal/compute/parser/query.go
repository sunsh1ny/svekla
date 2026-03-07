package parser

type Query struct {
	commandID int
	arguments []string
}

func NewQuery(commandID int, arguments []string) Query {
	return Query{
		commandID: commandID,
		arguments: arguments,
	}
}

func (q *Query) CommandID() int {
	return q.commandID
}

func (q *Query) Arguments() []string {
	return q.arguments
}
