package golem

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func SetupConfiguration(c *cobra.Command, args []string) {
	var configFileFlag string
	c.PersistentFlags().StringVarP(&configFileFlag, "config", "c", "./config.yaml", "The path to the config file to use.")
	err := viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))
	if err != nil {
		zap.S().Fatalf("unable to bind config flag: %+v", err)
	}
	viper.SetConfigFile(configFileFlag)

	viper.SetConfigName("config") // config file name without extension
	viper.AddConfigPath(".")      // search path

	if defaultConfigHandler != nil {
		defaultConfigHandler()
	}

	viper.AutomaticEnv() // read value ENV variables
	err = viper.ReadInConfig()
	if err != nil {
		zap.S().Fatalf("%+v", err)
	}
}

func checkConfig[T Config]() error {
	_, err := LoadConfig[T]()
	if err != nil {
		return fmt.Errorf("invalid config: %s", err.Error())
	}

	return nil
}

func LoadConfig[T Config]() (*T, error) {
	var config T

	config.SetDefaults()

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}
