package zdbconfig

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestBuildGormConfig() {
	logConfig := LogConfig{
		LogLevel: "info",
	}
	config := BuildGormConfig(logConfig)
	s.NotNil(config.Logger)
}
