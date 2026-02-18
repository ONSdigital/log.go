package log

type Classification string

func (c Classification) attach(le *EventData) {
	le.Classification = string(c)
}
