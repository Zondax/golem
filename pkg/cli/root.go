package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zondax/golem/pkg/logger"
)

type AppSettings struct {
	Name        string
	Description string
	ConfigPath  string // Where this CLI config.yml?
	EnvPrefix   string // environment variable MYAPP_.....
	GitVersion  string
	GitRevision string
	LogLevel    string // Global log level for the app
}

type CLI struct {
	app AppSettings

	checkConfig func() error
	rootCmd     *cobra.Command
}

// New generates a CLI instance
func New[T Config](app AppSettings) *CLI {
	var rootCmd = &cobra.Command{
		Use:   app.Name,
		Short: app.Description,
	}

	c := &CLI{
		app:         app,
		checkConfig: func() error { return checkConfig[T]() },
		rootCmd:     rootCmd,
	}

	c.init()

	return c
}

func (c *CLI) init() {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check configuration",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.checkConfig()
			if err != nil {
				logger.Errorf("%s\n", c.checkConfig().Error())
			} else {
				logger.Infof("Configuration OK\n")
			}
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Infof("%s\n", c.GetVersionString())
		},
	}

	c.GetRoot().AddCommand(checkCmd)
	c.GetRoot().AddCommand(versionCmd)

	// If app log level is defined it is configued, logger.defaultConfig by default
	if len(c.app.LogLevel) > 0 {
		logger.InitLogger(logger.Config{Level: c.app.LogLevel})
	}

	setupCloseHandler(nil)
	// Set Configuration Defaults
	setupDefaultConfiguration(func() {
		viper.AddConfigPath(c.app.ConfigPath)
		viper.SetEnvPrefix(c.app.EnvPrefix)
	})

	SetupConfiguration(c.GetRoot())
}

func (c *CLI) GetRoot() *cobra.Command {
	// ????
	return c.rootCmd
}

func (c *CLI) GetVersionString() string {
	return fmt.Sprintf("version: '%s', revision: '%s' ", c.app.GitVersion, c.app.GitRevision)
}

func (c *CLI) Run() {
	if err := c.rootCmd.Execute(); err != nil {
		logger.Error(err.Error())
		_ = logger.Sync()
		os.Exit(1)
	}
}

func (c *CLI) Close() {
	_ = logger.Sync()
}
