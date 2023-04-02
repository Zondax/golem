package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

type AppSettings struct {
	Name        string
	Description string
	ConfigPath  string // Where this CLI config.yml?
	EnvPrefix   string // environment variable MYAPP_.....
	GitVersion  string
	GitRevision string
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
				fmt.Printf("%s\n", c.checkConfig().Error())
			} else {
				fmt.Printf("Configuration OK\n")
			}
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", c.GetVersionString())
		},
	}

	c.GetRoot().AddCommand(checkCmd)
	c.GetRoot().AddCommand(versionCmd)

	// TODO: We can make this optional? and more configurable if we see the need
	// Initialize logger
	InitGlobalLogger()
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
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		_ = zap.S().Sync()
		os.Exit(1)
	}
}

func (c *CLI) Close() {
	_ = zap.S().Sync()
}
