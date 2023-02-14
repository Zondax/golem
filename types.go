package golem

type CleanUpHandler func()
type DefaultConfigHandler func()

type Config interface {
	Validate() error
}
