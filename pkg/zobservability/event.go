package zobservability

type Event interface {
	SetLevel(level Level)
	SetTags(tags map[string]string)
	SetTag(key, value string)
	SetUser(id, email, username string)
	SetFingerprint(fingerprint []string)
	SetError(err error)
	Capture()
}

type EventOption interface {
	ApplyEvent(Event)
}

type eventOptionFunc func(Event)

func (f eventOptionFunc) ApplyEvent(e Event) {
	f(e)
}

func WithEventLevel(level Level) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetLevel(level)
	})
}

func WithEventTags(tags map[string]string) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetTags(tags)
	})
}

func WithEventTag(key, value string) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetTag(key, value)
	})
}

func WithEventUser(id, email, username string) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetUser(id, email, username)
	})
}

func WithEventFingerprint(fingerprint []string) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetFingerprint(fingerprint)
	})
}

func WithEventError(err error) EventOption {
	return eventOptionFunc(func(e Event) {
		e.SetError(err)
	})
}
