package log

type classification string

func (c classification) attach(le *EventData) {
	le.Classification = string(c)
}

func Classification(classification classification) option {
	return classification
}
