package parser

type Query struct {
	commandID CommandID
	arguments []string
}

func NewQuery(commandID CommandID, arguments []string) Query {
	return Query{
		commandID: commandID,
		arguments: copyArguments(arguments),
	}
}

func (q Query) CommandID() CommandID {
	return q.commandID
}

func (q Query) Arguments() []string {
	return copyArguments(q.arguments)
}

func (q Query) Argument(index int) (string, bool) {
	if index < 0 || index >= len(q.arguments) {
		return "", false
	}

	return q.arguments[index], true
}

func copyArguments(arguments []string) []string {
	if len(arguments) == 0 {
		return nil
	}

	copied := make([]string, len(arguments))
	copy(copied, arguments)

	return copied
}
