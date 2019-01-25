package log

// EventAuth ...
type EventAuth struct {
	Identity     string `json:"identity,omitempty"`
	IdentityType string `json:"identity_type,omitempty"`
}

// Attach ...
func (l *EventAuth) Attach(le *EventData) {
	le.Auth = l
}
