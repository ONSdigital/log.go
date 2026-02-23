package log

type classification string

const (
	ProtectiveMonitoring classification = "PROTECTIVE_MONITORING"
)

func (c classification) attach(le *EventData) {
	le.Classification = c
}

func Classification(c classification) option {
	return c
}
