package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/zondax/golem/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zondax/golem/pkg/secrets"
)

func SetupConfiguration(c *cobra.Command) {
	var configFileFlag string
	c.PersistentFlags().StringVarP(&configFileFlag, "config", "c", "", "The path to the config file to use.")
	err := viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))
	if err != nil {
		logger.Fatalf("unable to bind config flag: %+v", err)
	}

	viper.SetConfigName("config") // config file name without extension
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // search path

	if defaultConfigHandler != nil {
		defaultConfigHandler()
	}
}

func checkConfig[T Config](opts ...LoadConfigOption) error {
	_, err := LoadConfig[T](opts...)
	if err != nil {
		return fmt.Errorf("invalid config: %s", err.Error())
	}

	return nil
}

// LoadConfig loads the config and resolves secrets using the specified options.
// No secret providers are registered by default; you must specify them via options.
// Example:
//
//	LoadConfig[MyConfigType](WithSecretProviders(providers.GcpProvider{}))
//	LoadConfig[MyConfigType](WithSecretProviders(providers.GcpProvider{}, providers.AwsProvider{}))
func LoadConfig[T Config](opts ...LoadConfigOption) (*T, error) {
	var config T
	var options loadConfigOptions

	config.SetDefaults()

	// adds all default values in viper to struct
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	configFileOverride := viper.GetString("config")
	if configFileOverride != "" {
		viper.SetConfigFile(configFileOverride)
		logger.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatalf("%+v", err)
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	options.RegisterSecretProviders()
	if err := secrets.ResolveSecrets(context.Background()); err != nil {
		logger.Fatalf("error resolving secrets: %+v", err)
	}

	// adds all default+configFile values in viper to struct
	// values in config file overrides default values
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	// To override the value in config.yaml for the key
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Override with environment variables
	viper.AutomaticEnv() // read value ENV variables

	// adds all default+configFile+env values in viper to struct
	// values from env overrides default+configFile values
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}
