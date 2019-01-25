package log

// EventAuth ...
type EventAuth struct {
	Identity     string `json:"identity,omitempty"`
	IdentityType string `json:"identity_type,omitempty"`
}

func (l *EventAuth) attach(le *EventData) {
	le.Auth = l
}
