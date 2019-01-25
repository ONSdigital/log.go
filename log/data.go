package log

// Data ...
type Data map[string]interface{}

// Attach ...
func (d Data) Attach(le *EventData) {
	le.Data = &d
}
