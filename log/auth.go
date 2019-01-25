package log

type eventAuth struct {
	Identity     string `json:"identity,omitempty"`
	IdentityType string `json:"identity_type,omitempty"`
}

func (l *eventAuth) attach(le *EventData) {
	le.Auth = l
}

// Auth ...
func Auth(identity, identityType string) option {
	return &eventAuth{
		Identity:     identity,
		IdentityType: identityType,
	}
}
