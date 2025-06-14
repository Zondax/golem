package clock

import "time"

// Clock abstracts time operations for better testing and flexibility
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// clock implements Clock using the actual system time
type clock struct{}

func New() Clock {
	return &clock{}
}

func (c *clock) Now() time.Time {
	return time.Now()
}
