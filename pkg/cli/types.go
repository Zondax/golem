package cli

type CleanUpHandler func()
type DefaultConfigHandler func()

type Config interface {
	SetDefaults()
	Validate() error
}
